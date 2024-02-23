// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package db provides the implementation of the [client.DB] interface, collection operations,
and related components.
*/
package db

import (
	"context"
	"sync"
	"sync/atomic"

	blockstore "github.com/ipfs/boxo/blockstore"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/lens"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/request/graphql"
)

var (
	log = logging.MustNewLogger("db")
)

// make sure we match our client interface
var (
	_ client.Collection = (*collection)(nil)
)

const (
	defaultMaxTxnRetries  = 5
	updateEventBufferSize = 100
)

// DB is the main interface for interacting with the
// DefraDB storage system.
type db struct {
	glock sync.RWMutex

	rootstore  datastore.RootStore
	multistore datastore.MultiStore

	events events.Events

	parser       core.Parser
	lensRegistry client.LensRegistry

	// The maximum number of retries per transaction.
	maxTxnRetries immutable.Option[int]

	// The maximum number of cached migrations instances to preserve per schema version.
	lensPoolSize immutable.Option[int]

	// The options used to init the database
	options any

	// The ID of the last transaction created.
	previousTxnID atomic.Uint64

	// ACP module
	acp immutable.Option[acp.ACPModule]
}

// Functional option type.
type Option func(*db)

// WithACPModule enables access control. If path is empty, then acp module runs in-memory.
func WithACPModule(path string) Option {
	return func(db *db) {
		db.acp = acp.SomeNewLocalACPModule(context.Background(), path)
	}
}

// WithUpdateEvents enables the update events channel.
func WithUpdateEvents() Option {
	return func(db *db) {
		db.events = events.Events{
			Updates: immutable.Some(events.New[events.Update](0, updateEventBufferSize)),
		}
	}
}

// WithMaxRetries sets the maximum number of retries per transaction.
func WithMaxRetries(num int) Option {
	return func(db *db) {
		db.maxTxnRetries = immutable.Some(num)
	}
}

// WithLensPoolSize sets the maximum number of cached migrations instances to preserve per schema version.
//
// Will default to `5` if not set.
func WithLensPoolSize(num int) Option {
	return func(db *db) {
		db.lensPoolSize = immutable.Some(num)
	}
}

// NewDB creates a new instance of the DB using the given options.
func NewDB(
	ctx context.Context,
	rootstore datastore.RootStore,
	options ...Option,
) (client.DB, error) {
	return newDB(ctx, rootstore, options...)
}

func newDB(
	ctx context.Context,
	rootstore datastore.RootStore,
	options ...Option,
) (*implicitTxnDB, error) {
	log.Debug(ctx, "Loading: internal datastores")
	multistore := datastore.MultiStoreFrom(rootstore)

	parser, err := graphql.NewParser()
	if err != nil {
		return nil, err
	}

	db := &db{
		rootstore:  rootstore,
		multistore: multistore,
		acp:        acp.NoACPModule,
		parser:     parser,
		options:    options,
	}

	// apply options
	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(db)
	}

	// lensPoolSize may be set by `options`, and because they are funcs on db
	// we have to mutate `db` here to set the registry.
	db.lensRegistry = lens.NewRegistry(db.lensPoolSize, db)

	err = db.initialize(ctx)
	if err != nil {
		return nil, err
	}

	return &implicitTxnDB{db}, nil
}

// NewTxn creates a new transaction.
func (db *db) NewTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	return datastore.NewTxnFrom(ctx, db.rootstore, txnId, readonly)
}

// NewConcurrentTxn creates a new transaction that supports concurrent API calls.
func (db *db) NewConcurrentTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	return datastore.NewConcurrentTxnFrom(ctx, db.rootstore, txnId, readonly)
}

// WithTxn returns a new [client.Store] that respects the given transaction.
func (db *db) WithTxn(txn datastore.Txn) client.Store {
	return &explicitTxnDB{
		db:           db,
		txn:          txn,
		lensRegistry: db.lensRegistry.WithTxn(txn),
	}
}

// Root returns the root datastore.
func (db *db) Root() datastore.RootStore {
	return db.rootstore
}

// Blockstore returns the internal DAG store which contains IPLD blocks.
func (db *db) Blockstore() blockstore.Blockstore {
	return db.multistore.DAGstore()
}

// Peerstore returns the internal DAG store which contains IPLD blocks.
func (db *db) Peerstore() datastore.DSBatching {
	return db.multistore.Peerstore()
}

func (db *db) LensRegistry() client.LensRegistry {
	return db.lensRegistry
}

func (db *db) ACPModule() immutable.Option[acp.ACPModule] {
	return db.acp
}

// Initialize is called when a database is first run and creates all the db global meta data
// like Collection ID counters.
func (db *db) initialize(ctx context.Context) error {
	db.glock.Lock()
	defer db.glock.Unlock()

	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	log.Debug(ctx, "Checking if DB has already been initialized...")
	exists, err := txn.Systemstore().Has(ctx, ds.NewKey("init"))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	// if we're loading an existing database, just load the schema
	// and migrations and finish initialization
	if exists {
		log.Debug(ctx, "DB has already been initialized, continuing")
		err = db.loadSchema(ctx, txn)
		if err != nil {
			return err
		}

		//err = db.Acp.Reload(ctx)
		//if err != nil {
		//	return err
		//}

		err = db.lensRegistry.ReloadLenses(ctx)
		if err != nil {
			return err
		}

		// The query language types are only updated on successful commit
		// so we must not forget to do so on success regardless of whether
		// we have written to the datastores.
		return txn.Commit(ctx)
	}

	log.Debug(ctx, "Opened a new DB, needs full initialization")

	// init meta data
	// collection sequence
	_, err = db.getSequence(ctx, txn, core.COLLECTION)
	if err != nil {
		return err
	}
	// err = db.Acp.Initialize(ctx, db)
	// if err != nil {
	// return err
	// }

	err = txn.Systemstore().Put(ctx, ds.NewKey("init"), []byte{1})
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// Events returns the events Channel.
func (db *db) Events() events.Events {
	return db.events
}

// MaxRetries returns the maximum number of retries per transaction.
// Defaults to `defaultMaxTxnRetries` if not explicitely set
func (db *db) MaxTxnRetries() int {
	if db.maxTxnRetries.HasValue() {
		return db.maxTxnRetries.Value()
	}
	return defaultMaxTxnRetries
}

// PrintDump prints the entire database to console.
func (db *db) PrintDump(ctx context.Context) error {
	return printStore(ctx, db.multistore.Rootstore())
}

// Close is called when we are shutting down the database.
// This is the place for any last minute cleanup or releasing of resources (i.e.: Badger instance).
func (db *db) Close() {
	log.Info(context.Background(), "Closing DefraDB process...")
	if db.events.Updates.HasValue() {
		db.events.Updates.Value().Close()
	}

	// TODO: This might be optional so use `immutable.Option` Perhaps.
	//if err := db.acp.Close(); err != nil {
	//	log.ErrorE(context.Background(), "Failure closing acp module", err)

	//}

	if err := db.rootstore.Close(); err != nil {
		log.ErrorE(context.Background(), "Failure closing rootstore", err)
	}

	log.Info(context.Background(), "Successfully closed running process")
}

func printStore(ctx context.Context, store datastore.DSReaderWriter) error {
	q := dsq.Query{
		Prefix:   "",
		KeysOnly: false,
		Orders:   []dsq.Order{dsq.OrderByKey{}},
	}

	results, err := store.Query(ctx, q)
	if err != nil {
		return err
	}

	for r := range results.Next() {
		log.Info(ctx, "", logging.NewKV(r.Key, r.Value))
	}

	return results.Close()
}
