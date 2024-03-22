// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package update

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithInvalidID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with invalid document id",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 42.1
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        -1,
				Doc: `{
					"points": 59
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"points": float64(59),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
