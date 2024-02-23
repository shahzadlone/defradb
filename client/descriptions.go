// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
)

// CollectionDescription with no known root will take this ID as their temporary RootID.
//
// Orphan CollectionDescriptions are typically created when setting migrations from schema versions
// that do not yet exist.  The OrphanRootID will be replaced with the actual RootID once a full chain
// of schema versions leading back to a schema version used by a collection with a non-orphan RootID
// has been established.
const OrphanRootID uint32 = math.MaxUint32

// CollectionDescription describes a Collection and all its associated metadata.
type CollectionDescription struct {
	// Name contains the name of the collection.
	//
	// It is conceptually local to the node hosting the DefraDB instance, but currently there
	// is no means to update the local value so that it differs from the (global) schema name.
	Name immutable.Option[string]

	// ID is the local identifier of this collection.
	//
	// It is immutable.
	ID uint32

	// RootID is the local root identifier of this collection, linking together a chain of
	// collection instances on different schema versions.
	//
	// Collections sharing the same RootID will be compatable with each other, with the documents
	// within them shared and yielded as if they were in the same set, using Lens transforms to
	// migrate between schema versions when provided.
	RootID uint32

	// The ID of the schema version that this collection is at.
	SchemaVersionID string

	// Sources is the set of sources from which this collection draws data.
	//
	// Currently supported source types are:
	// - [QuerySource]
	// - [CollectionSource]
	Sources []any

	// Indexes contains the secondary indexes that this Collection has.
	Indexes []IndexDescription

	// Policy contains the policy information on this collection.
	//
	// It is possible for a collection to not have a policy.
	//
	// Note: The policy information must be validated using the acp module
	// right after parsing is done, to avoid storing an invalid policyID
	// or policy resource that may not even exist on acp module.
	Policy immutable.Option[PolicyDescription]
}

// IDString returns the collection ID as a string.
func (col CollectionDescription) IDString() string {
	return fmt.Sprint(col.ID)
}

// GetFieldByID searches for a field with the given ID. If such a field is found it
// will return it and true, if it is not found it will return false.
func (col CollectionDescription) GetFieldByID(id FieldID, schema *SchemaDescription) (FieldDescription, bool) {
	for _, field := range schema.Fields {
		if field.ID == id {
			return field, true
		}
	}
	return FieldDescription{}, false
}

// GetFieldByName returns the field for the given field name. If such a field is found it
// will return it and true, if it is not found it will return false.
func (col CollectionDescription) GetFieldByName(fieldName string, schema *SchemaDescription) (FieldDescription, bool) {
	for _, field := range schema.Fields {
		if field.Name == fieldName {
			return field, true
		}
	}
	return FieldDescription{}, false
}

// GetFieldByRelation returns the field that supports the relation of the given name.
func (col CollectionDescription) GetFieldByRelation(
	relationName string,
	otherCollectionName string,
	otherFieldName string,
	schema *SchemaDescription,
) (FieldDescription, bool) {
	for _, field := range schema.Fields {
		if field.RelationName == relationName &&
			!(col.Name.Value() == otherCollectionName && otherFieldName == field.Name) &&
			field.Kind != FieldKind_DocID {
			return field, true
		}
	}
	return FieldDescription{}, false
}

// QuerySources returns all the Sources of type [QuerySource]
func (col CollectionDescription) QuerySources() []*QuerySource {
	return sourcesOfType[*QuerySource](col)
}

// CollectionSources returns all the Sources of type [CollectionSource]
func (col CollectionDescription) CollectionSources() []*CollectionSource {
	return sourcesOfType[*CollectionSource](col)
}

func sourcesOfType[ResultType any](col CollectionDescription) []ResultType {
	result := []ResultType{}
	for _, source := range col.Sources {
		if typedSource, isOfType := source.(ResultType); isOfType {
			result = append(result, typedSource)
		}
	}
	return result
}

// QuerySource represents a collection data source from a query.
//
// The query will be executed when data from this source is requested, and the query results
// yielded to the consumer.
type QuerySource struct {
	// Query contains the base query of this data source.
	Query request.Select

	// Transform is a optional Lens configuration.  If specified, data drawn from the [Query] will have the
	// transform applied before being returned.
	//
	// The transform is not limited to just transforming the input documents, it may also yield new ones, or filter out
	// those passed in from the underlying query.
	Transform immutable.Option[model.Lens]
}

// CollectionSource represents a collection data source from another collection instance.
//
// Data against all collection instances in a CollectionSource chain will be returned as-if
// from the same dataset when queried.  Lens transforms may be applied between instances.
//
// Typically these are used to link together multiple schema versions into the same dataset.
type CollectionSource struct {
	// SourceCollectionID is the local identifier of the source [CollectionDescription] from which to
	// share data.
	//
	// This is a bi-directional relationship, and documents in the host collection instance will also
	// be available to the source collection instance.
	SourceCollectionID uint32

	// Transform is a optional Lens configuration.  If specified, data drawn from the source will have the
	// transform applied before being returned by any operation on the host collection instance.
	//
	// If the transform supports an inverse operation, that inverse will be applied when the source collection
	// draws data from this host.
	Transform immutable.Option[model.Lens]
}

// SchemaDescription describes a Schema and its associated metadata.
type SchemaDescription struct {
	// Root is the version agnostic identifier for this schema.
	//
	// It remains constant throughout the lifetime of this schema.
	Root string

	// VersionID is the version-specific identifier for this schema.
	//
	// It is generated on mutation of this schema and can be used to uniquely
	// identify a schema at a specific version.
	VersionID string

	// Name is the name of this Schema.
	//
	// It is currently used to define the Collection Name, and as such these two properties
	// will currently share the same name.
	//
	// It is immutable.
	Name string

	// Fields contains the fields within this Schema.
	//
	// Currently new fields may be added after initial declaration, but they cannot be removed.
	Fields []FieldDescription
}

// GetField returns the field of the given name.
func (sd SchemaDescription) GetField(name string) (FieldDescription, bool) {
	for _, field := range sd.Fields {
		if field.Name == name {
			return field, true
		}
	}
	return FieldDescription{}, false
}

// FieldKind describes the type of a field.
type FieldKind uint8

func (f FieldKind) String() string {
	switch f {
	case FieldKind_DocID:
		return "ID"
	case FieldKind_NILLABLE_BOOL:
		return "Boolean"
	case FieldKind_NILLABLE_BOOL_ARRAY:
		return "[Boolean]"
	case FieldKind_BOOL_ARRAY:
		return "[Boolean!]"
	case FieldKind_NILLABLE_INT:
		return "Int"
	case FieldKind_NILLABLE_INT_ARRAY:
		return "[Int]"
	case FieldKind_INT_ARRAY:
		return "[Int!]"
	case FieldKind_NILLABLE_DATETIME:
		return "DateTime"
	case FieldKind_NILLABLE_FLOAT:
		return "Float"
	case FieldKind_NILLABLE_FLOAT_ARRAY:
		return "[Float]"
	case FieldKind_FLOAT_ARRAY:
		return "[Float!]"
	case FieldKind_NILLABLE_STRING:
		return "String"
	case FieldKind_NILLABLE_STRING_ARRAY:
		return "[String]"
	case FieldKind_STRING_ARRAY:
		return "[String!]"
	case FieldKind_NILLABLE_BLOB:
		return "Blob"
	case FieldKind_NILLABLE_JSON:
		return "JSON"
	default:
		return fmt.Sprint(uint8(f))
	}
}

// Note: These values are serialized and persisted in the database, avoid modifying existing values.
const (
	FieldKind_None              FieldKind = 0
	FieldKind_DocID             FieldKind = 1
	FieldKind_NILLABLE_BOOL     FieldKind = 2
	FieldKind_BOOL_ARRAY        FieldKind = 3
	FieldKind_NILLABLE_INT      FieldKind = 4
	FieldKind_INT_ARRAY         FieldKind = 5
	FieldKind_NILLABLE_FLOAT    FieldKind = 6
	FieldKind_FLOAT_ARRAY       FieldKind = 7
	_                           FieldKind = 8 // safe to repurpose (was never used)
	_                           FieldKind = 9 // safe to repurpose (previously old field)
	FieldKind_NILLABLE_DATETIME FieldKind = 10
	FieldKind_NILLABLE_STRING   FieldKind = 11
	FieldKind_STRING_ARRAY      FieldKind = 12
	FieldKind_NILLABLE_BLOB     FieldKind = 13
	FieldKind_NILLABLE_JSON     FieldKind = 14
	_                           FieldKind = 15 // safe to repurpose (was never used)

	// Embedded object, but accessed via foreign keys
	FieldKind_FOREIGN_OBJECT FieldKind = 16

	// Array of embedded objects, accessed via foreign keys
	FieldKind_FOREIGN_OBJECT_ARRAY FieldKind = 17

	FieldKind_NILLABLE_BOOL_ARRAY   FieldKind = 18
	FieldKind_NILLABLE_INT_ARRAY    FieldKind = 19
	FieldKind_NILLABLE_FLOAT_ARRAY  FieldKind = 20
	FieldKind_NILLABLE_STRING_ARRAY FieldKind = 21
)

// FieldKindStringToEnumMapping maps string representations of [FieldKind] values to
// their enum values.
//
// It is currently used to by [db.PatchSchema] to allow string representations of
// [FieldKind] to be provided instead of their raw int values.  This usage may expand
// in the future.  They currently roughly correspond to the GQL field types, but this
// equality is not guaranteed.
var FieldKindStringToEnumMapping = map[string]FieldKind{
	"ID":         FieldKind_DocID,
	"Boolean":    FieldKind_NILLABLE_BOOL,
	"[Boolean]":  FieldKind_NILLABLE_BOOL_ARRAY,
	"[Boolean!]": FieldKind_BOOL_ARRAY,
	"Int":        FieldKind_NILLABLE_INT,
	"[Int]":      FieldKind_NILLABLE_INT_ARRAY,
	"[Int!]":     FieldKind_INT_ARRAY,
	"DateTime":   FieldKind_NILLABLE_DATETIME,
	"Float":      FieldKind_NILLABLE_FLOAT,
	"[Float]":    FieldKind_NILLABLE_FLOAT_ARRAY,
	"[Float!]":   FieldKind_FLOAT_ARRAY,
	"String":     FieldKind_NILLABLE_STRING,
	"[String]":   FieldKind_NILLABLE_STRING_ARRAY,
	"[String!]":  FieldKind_STRING_ARRAY,
	"Blob":       FieldKind_NILLABLE_BLOB,
	"JSON":       FieldKind_NILLABLE_JSON,
}

// RelationType describes the type of relation between two types.
type RelationType uint8

// FieldID is a unique identifier for a field in a schema.
type FieldID uint32

func (f FieldID) String() string {
	return fmt.Sprint(uint32(f))
}

// FieldDescription describes a field on a Schema and its associated metadata.
type FieldDescription struct {
	// Name contains the name of this field.
	//
	// It is currently immutable.
	Name string

	// ID contains the internal ID of this field.
	//
	// Whilst this ID will typically match the field's index within the Schema's Fields
	// slice, there is no guarantee that they will be the same.
	//
	// It is immutable.
	ID FieldID

	// The data type that this field holds.
	//
	// Must contain a valid value. It is currently immutable.
	Kind FieldKind

	// Schema contains the schema name of the type this field contains if this field is
	// a relation field.  Otherwise this will be empty.
	Schema string

	// RelationName the name of the relationship that this field represents if this field is
	// a relation field.  Otherwise this will be empty.
	RelationName string

	// The CRDT Type of this field. If no type has been provided it will default to [LWW_REGISTER].
	//
	// It is currently immutable.
	Typ CType

	IsPrimaryRelation bool
}

// IsObject returns true if this field is an object type.
func (f FieldDescription) IsObject() bool {
	return (f.Kind == FieldKind_FOREIGN_OBJECT) ||
		(f.Kind == FieldKind_FOREIGN_OBJECT_ARRAY)
}

// IsObjectArray returns true if this field is an object array type.
func (f FieldDescription) IsObjectArray() bool {
	return (f.Kind == FieldKind_FOREIGN_OBJECT_ARRAY)
}

// IsRelation returns true if this field is a relation.
func (f FieldDescription) IsRelation() bool {
	return f.RelationName != ""
}

// IsArray returns true if this field is an array type which includes inline arrays as well
// as relation arrays.
func (f FieldDescription) IsArray() bool {
	return f.Kind == FieldKind_BOOL_ARRAY ||
		f.Kind == FieldKind_INT_ARRAY ||
		f.Kind == FieldKind_FLOAT_ARRAY ||
		f.Kind == FieldKind_STRING_ARRAY ||
		f.Kind == FieldKind_FOREIGN_OBJECT_ARRAY ||
		f.Kind == FieldKind_NILLABLE_BOOL_ARRAY ||
		f.Kind == FieldKind_NILLABLE_INT_ARRAY ||
		f.Kind == FieldKind_NILLABLE_FLOAT_ARRAY ||
		f.Kind == FieldKind_NILLABLE_STRING_ARRAY
}

// IsSet returns true if the target relation type is set.
func (m RelationType) IsSet(target RelationType) bool {
	return m&target > 0
}

// collectionDescription is a private type used to facilitate the unmarshalling
// of json to a [CollectionDescription].
type collectionDescription struct {
	// These properties are unmarshalled using the default json unmarshaller
	Name            immutable.Option[string]
	ID              uint32
	RootID          uint32
	SchemaVersionID string
	Indexes         []IndexDescription

	// Properties below this line are unmarshalled using custom logic in [UnmarshalJSON]
	Sources []map[string]json.RawMessage

	// TODO-ACP: Do we need this here?
	//Policy immutable.Option[PolicyDescription]
}

func (c *CollectionDescription) UnmarshalJSON(bytes []byte) error {
	var descMap collectionDescription
	err := json.Unmarshal(bytes, &descMap)
	if err != nil {
		return err
	}

	c.Name = descMap.Name
	c.ID = descMap.ID
	c.RootID = descMap.RootID
	c.SchemaVersionID = descMap.SchemaVersionID
	c.Indexes = descMap.Indexes
	c.Sources = make([]any, len(descMap.Sources))
	// TODO-ACP: Do we need this?
	// c.Policy = descMap.Policy

	for i, source := range descMap.Sources {
		sourceJson, err := json.Marshal(source)
		if err != nil {
			return err
		}

		var sourceValue any
		// We detect which concrete type each `Source` object is by detecting
		// non-nillable fields, if the key is present it must be of that type.
		// They must be non-nillable as nil values may have their keys omitted from
		// the json. This also relies on the fields being unique.  We may wish to change
		// this later to custom-serialize with a `_type` property.
		if _, ok := source["Query"]; ok {
			// This must be a QuerySource, as only the `QuerySource` type has a `Query` field
			var querySource QuerySource
			err := json.Unmarshal(sourceJson, &querySource)
			if err != nil {
				return err
			}
			sourceValue = &querySource
		} else if _, ok := source["SourceCollectionID"]; ok {
			// This must be a CollectionSource, as only the `CollectionSource` type has a `SourceCollectionID` field
			var collectionSource CollectionSource
			err := json.Unmarshal(sourceJson, &collectionSource)
			if err != nil {
				return err
			}
			sourceValue = &collectionSource
		} else {
			return ErrFailedToUnmarshalCollection
		}

		c.Sources[i] = sourceValue
	}

	return nil
}
