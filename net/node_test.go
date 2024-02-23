// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	badger "github.com/sourcenetwork/badger/v4"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/datastore/memory"
	"github.com/sourcenetwork/defradb/db"
	netutils "github.com/sourcenetwork/defradb/net/utils"
)

// Node.Boostrap is not tested because the underlying, *ipfslite.Peer.Bootstrap is a best-effort function.

func FixtureNewMemoryDBWithBroadcaster(t *testing.T) client.DB {
	var database client.DB
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)
	database, err = db.NewDB(ctx, rootstore, db.WithUpdateEvents())
	require.NoError(t, err)
	return database
}

func TestNewNode_WithEnableRelay_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)
	n, err := NewNode(
		context.Background(),
		db,
		WithEnableRelay(true),
	)
	require.NoError(t, err)
	defer n.Close()
}

func TestNewNode_WithDBClosed_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)

	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)
	db.Close()

	_, err = NewNode(
		context.Background(),
		db,
	)
	require.ErrorContains(t, err, "datastore closed")
}

func TestNewNode_NoPubSub_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)
	n, err := NewNode(
		context.Background(),
		db,
		WithEnablePubSub(false),
	)
	require.NoError(t, err)
	defer n.Close()
	require.Nil(t, n.ps)
}

func TestNewNode_WithEnablePubSub_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)

	n, err := NewNode(
		ctx,
		db,
		WithEnablePubSub(true),
	)

	require.NoError(t, err)
	defer n.Close()
	// overly simple check of validity of pubsub, avoiding the process of creating a PubSub
	require.NotNil(t, n.ps)
}

func TestNodeClose_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	n.Close()
}

func TestNewNode_BootstrapWithNoPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)

	n1, err := NewNode(
		ctx,
		db,
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n1.Bootstrap([]peer.AddrInfo{})
}

func TestNewNode_BootstrapWithOnePeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)

	n1, err := NewNode(
		ctx,
		db,
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n2, err := NewNode(
		ctx,
		db,
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n2.Close()
	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	if err != nil {
		t.Fatal(err)
	}
	n2.Bootstrap(addrs)
}

func TestNewNode_BootstrapWithOneValidPeerAndManyInvalidPeers_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)

	n1, err := NewNode(
		ctx,
		db,
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n2, err := NewNode(
		ctx,
		db,
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n2.Close()
	addrs, err := netutils.ParsePeers([]string{
		n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String(),
		"/ip4/0.0.0.0/tcp/1234/p2p/" + "12D3KooWC8YY6Tx3uAeHsdBmoy7PJPwqXAHE4HkCZ5veankKWci6",
		"/ip4/0.0.0.0/tcp/1235/p2p/" + "12D3KooWC8YY6Tx3uAeHsdBmoy7PJPwqXAHE4HkCZ5veankKWci5",
		"/ip4/0.0.0.0/tcp/1236/p2p/" + "12D3KooWC8YY6Tx3uAeHsdBmoy7PJPwqXAHE4HkCZ5veankKWci4",
	})
	require.NoError(t, err)
	n2.Bootstrap(addrs)
}

func TestListenAddrs_WithListenAddresses_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)
	n, err := NewNode(
		context.Background(),
		db,
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n.Close()

	require.Contains(t, n.ListenAddrs()[0].String(), "/tcp/")
}

func TestPeerConnectionEventEmitter_MultiEvent_NoError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(event.EvtPeerConnectednessChanged))
	require.NoError(t, err)

	// the emitter can take 20 events in the channel. This tests what happens whe go over the 20 events.
	for i := 0; i < 21; i++ {
		err = emitter.Emit(event.EvtPeerConnectednessChanged{})
		require.NoError(t, err)
	}
}

func TestSubscribeToPubSubEvents_SubscriptionError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	n.Peer.host = &mockHost{n.Peer.host}

	n.subscribeToPubSubEvents()
}

func TestPubSubEventEmitter_MultiEvent_NoError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtPubSub))
	require.NoError(t, err)

	// the emitter can take 20 events in the channel. This tests what happens whe go over the 20 events.
	for i := 0; i < 21; i++ {
		err = emitter.Emit(EvtPubSub{})
		require.NoError(t, err)
	}
}

func TestSubscribeToPushLogEvents_SubscriptionError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	n.Peer.host = &mockHost{n.Peer.host}

	n.subscribeToPushLogEvents()
}

func TestPushLogEventEmitter_SingleEvent_NoError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtReceivedPushLog))
	require.NoError(t, err)

	err = emitter.Emit(EvtReceivedPushLog{})
	require.NoError(t, err)
}

func TestPushLogEventEmitter_MultiEvent_NoError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtReceivedPushLog))
	require.NoError(t, err)

	// the emitter can take 20 events in the channel. This tests what happens whe go over the 20 events.
	for i := 0; i < 21; i++ {
		err = emitter.Emit(EvtReceivedPushLog{})
		require.NoError(t, err)
	}
}

func TestWaitForPeerConnectionEvent_WithSamePeer_NoError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(event.EvtPeerConnectednessChanged))
	require.NoError(t, err)

	err = emitter.Emit(event.EvtPeerConnectednessChanged{
		Peer: n.PeerID(),
	})
	require.NoError(t, err)

	err = n.WaitForPeerConnectionEvent(n.PeerID())
	require.NoError(t, err)
}

func TestWaitForPeerConnectionEvent_WithDifferentPeer_TimeoutError(t *testing.T) {
	evtWaitTimeout = 100 * time.Microsecond
	defer func() {
		evtWaitTimeout = 10 * time.Second
	}()
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(event.EvtPeerConnectednessChanged))
	require.NoError(t, err)

	err = emitter.Emit(event.EvtPeerConnectednessChanged{})
	require.NoError(t, err)

	err = n.WaitForPeerConnectionEvent(n.PeerID())
	require.ErrorIs(t, err, ErrPeerConnectionWaitTimout)
}

func TestWaitForPeerConnectionEvent_WithDifferentPeerAndContextClosed_NoError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(event.EvtPeerConnectednessChanged))
	require.NoError(t, err)

	err = emitter.Emit(event.EvtPeerConnectednessChanged{})
	require.NoError(t, err)

	n.cancel()

	err = n.WaitForPeerConnectionEvent(n.PeerID())
	require.NoError(t, err)
}

func TestWaitForPubSubEvent_WithSamePeer_NoError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtPubSub))
	require.NoError(t, err)

	err = emitter.Emit(EvtPubSub{
		Peer: n.PeerID(),
	})
	require.NoError(t, err)

	err = n.WaitForPubSubEvent(n.PeerID())
	require.NoError(t, err)
}

func TestWaitForPubSubEvent_WithDifferentPeer_TimeoutError(t *testing.T) {
	evtWaitTimeout = 100 * time.Microsecond
	defer func() {
		evtWaitTimeout = 10 * time.Second
	}()
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtPubSub))
	require.NoError(t, err)

	err = emitter.Emit(EvtPubSub{})
	require.NoError(t, err)

	err = n.WaitForPubSubEvent(n.PeerID())
	require.ErrorIs(t, err, ErrPubSubWaitTimeout)
}

func TestWaitForPubSubEvent_WithDifferentPeerAndContextClosed_NoError(t *testing.T) {
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtPubSub))
	require.NoError(t, err)

	err = emitter.Emit(EvtPubSub{})
	require.NoError(t, err)

	n.cancel()

	err = n.WaitForPubSubEvent(n.PeerID())
	require.NoError(t, err)
}

func TestWaitForPushLogByPeerEvent_WithSamePeer_NoError(t *testing.T) {
	ctx := context.Background()
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		ctx,
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtReceivedPushLog))
	require.NoError(t, err)

	err = emitter.Emit(EvtReceivedPushLog{
		ByPeer: n.PeerID(),
	})
	require.NoError(t, err)

	err = n.WaitForPushLogByPeerEvent(n.PeerID())
	require.NoError(t, err)
}

func TestWaitForPushLogByPeerEvent_WithDifferentPeer_TimeoutError(t *testing.T) {
	evtWaitTimeout = 100 * time.Microsecond
	defer func() {
		evtWaitTimeout = 10 * time.Second
	}()
	ctx := context.Background()
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		ctx,
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtReceivedPushLog))
	require.NoError(t, err)

	err = emitter.Emit(EvtReceivedPushLog{})
	require.NoError(t, err)

	err = n.WaitForPushLogByPeerEvent(n.PeerID())
	require.ErrorIs(t, err, ErrPushLogWaitTimeout)
}

func TestWaitForPushLogByPeerEvent_WithDifferentPeerAndContextClosed_NoError(t *testing.T) {
	ctx := context.Background()
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		ctx,
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtReceivedPushLog))
	require.NoError(t, err)

	err = emitter.Emit(EvtReceivedPushLog{})
	require.NoError(t, err)

	n.cancel()

	err = n.WaitForPushLogByPeerEvent(n.PeerID())
	require.NoError(t, err)
}

func TestWaitForPushLogFromPeerEvent_WithSamePeer_NoError(t *testing.T) {
	ctx := context.Background()
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		ctx,
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtReceivedPushLog))
	require.NoError(t, err)

	err = emitter.Emit(EvtReceivedPushLog{
		FromPeer: n.PeerID(),
	})
	require.NoError(t, err)

	err = n.WaitForPushLogFromPeerEvent(n.PeerID())
	require.NoError(t, err)
}

func TestWaitForPushLogFromPeerEvent_WithDifferentPeer_TimeoutError(t *testing.T) {
	evtWaitTimeout = 100 * time.Microsecond
	defer func() {
		evtWaitTimeout = 10 * time.Second
	}()
	ctx := context.Background()
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		ctx,
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtReceivedPushLog))
	require.NoError(t, err)

	err = emitter.Emit(EvtReceivedPushLog{})
	require.NoError(t, err)

	err = n.WaitForPushLogFromPeerEvent(n.PeerID())
	require.ErrorIs(t, err, ErrPushLogWaitTimeout)
}

func TestWaitForPushLogFromPeerEvent_WithDifferentPeerAndContextClosed_NoError(t *testing.T) {
	ctx := context.Background()
	db := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		ctx,
		db,
	)
	require.NoError(t, err)
	defer n.Close()

	emitter, err := n.host.EventBus().Emitter(new(EvtReceivedPushLog))
	require.NoError(t, err)

	err = emitter.Emit(EvtReceivedPushLog{})
	require.NoError(t, err)

	n.cancel()

	err = n.WaitForPushLogFromPeerEvent(n.PeerID())
	require.NoError(t, err)
}
