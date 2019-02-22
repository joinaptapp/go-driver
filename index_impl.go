//
// DISCLAIMER
//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package driver

import (
	"context"
	"path"
	"strings"
)

// indexStringToType converts a string representation of an index to IndexType
func indexStringToType(indexTypeString string) (IndexType, error) {
	switch indexTypeString {
	case string(FullTextIndex):
		return FullTextIndex, nil
	case string(HashIndex):
		return HashIndex, nil
	case string(SkipListIndex):
		return SkipListIndex, nil
	case string(PrimaryIndex):
		return PrimaryIndex, nil
	case string(PersistentIndex):
		return PersistentIndex, nil
	case string(GeoIndex), "geo1", "geo2":
		return GeoIndex, nil

	default:
		return "", WithStack(InvalidArgumentError{Message: "unknown index type"})
	}
}

// newIndex creates a new Index implementation.
func newIndex(data indexData, col *collection) (Index, error) {
	if data.id == "" {
		return nil, WithStack(InvalidArgumentError{Message: "id is empty"})
	}
	parts := strings.Split(data.id, "/")
	if len(parts) != 2 {
		return nil, WithStack(InvalidArgumentError{Message: "id must be `collection/name`"})
	}
	if col == nil {
		return nil, WithStack(InvalidArgumentError{Message: "col is nil"})
	}
	indexType, err := indexStringToType(data.typestr)
	if err != nil {
		return nil, WithStack(err)
	}
	return &index{
		data:      data,
		indexType: indexType,
		col:       col,
		db:        col.db,
		conn:      col.conn,
	}, nil
}

type indexData struct {
	id          string   `json:"id,omitempty"`
	typestr     string   `json:"type"`
	fields      []string `json:"fields,omitempty"`
	unique      *bool    `json:"unique,omitempty"`
	deduplicate *bool    `json:"deduplicate,omitempty"`
	sparse      *bool    `json:"sparse,omitempty"`
	geoJSON     *bool    `json:"geoJson,omitempty"`
	minLength   int      `json:"minLength,omitempty"`
}

type index struct {
	data      indexData
	indexType IndexType
	db        *database
	col       *collection
	conn      Connection
}

// relPath creates the relative path to this index (`_db/<db-name>/_api/index`)
func (i *index) relPath() string {
	return path.Join(i.db.relPath(), "_api", "index")
}

// Name returns the name of the index.
func (i *index) Name() string {
	parts := strings.Split(i.data.id, "/")
	return parts[1]
}

// Type returns the type of the index
func (i *index) Type() IndexType {
	return i.indexType
}

// Remove removes the entire index.
// If the index does not exist, a NotFoundError is returned.
func (i *index) Remove(ctx context.Context) error {
	req, err := i.conn.NewRequest("DELETE", path.Join(i.relPath(), i.data.id))
	if err != nil {
		return WithStack(err)
	}
	resp, err := i.conn.Do(ctx, req)
	if err != nil {
		return WithStack(err)
	}
	if err := resp.CheckStatus(200); err != nil {
		return WithStack(err)
	}
	return nil
}

// Fields returns the fields covered by this index
func (i *index) Fields() []string {
	return i.data.fields
}

func boolOrFalse(ptr *bool) bool {
	if ptr != nil {
		return *ptr
	}

	return false
}

// IsUnique returns the Unique attribute if the index supports this attribute, false otherwise.
func (i *index) IsUnique() bool {
	return boolOrFalse(i.data.unique)
}

// IsSparse returns the Sparse attribute if the index supports this attribute, false otherwise.
func (i *index) IsSparse() bool {
	return boolOrFalse(i.data.unique)
}

// IsDeduplicate returns the Deduplicate attribute if the index supports this attribute, false otherwise.
func (i *index) IsDeduplicate() bool {
	return boolOrFalse(i.data.deduplicate)
}

// IsGeoJSON returns the GeoJSON attribute if the index is a GeoIndex, false otherwise.
func (i *index) IsGeoJSON() bool {
	return boolOrFalse(i.data.geoJSON)
}

// MinLength returns the MinLength attribute if the index is a full-text index, 0 otherwise.
func (i *index) MinLength() int {
	return i.data.minLength
}
