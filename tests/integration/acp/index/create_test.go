// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	acpUtils "github.com/sourcenetwork/defradb/tests/integration/acp"
)

func TestACP_IndexCreateWithSeparateRequest_OnCollectionWithPolicy_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, with creating new index using separate request on permissioned collection, no error",
		Actions: []any{

			testUtils.AddPolicy{
				Identity:         acpUtils.Actor1Identity,
				Policy:           userPolicy,
				ExpectedPolicyID: "9b5bf263a047605cce43360bf0546911b3d5da78b2b12a894318ad2a084a4a21",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "9b5bf263a047605cce43360bf0546911b3d5da78b2b12a894318ad2a084a4a21",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			testUtils.CreateIndex{
				CollectionID: 0,
				IndexName:    "some_index",
				FieldName:    "name",
			},

			testUtils.Request{
				Request: `
					query  {
						Users {
							name
							age
						}
					}`,

				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_IndexCreateWithDirective_OnCollectionWithPolicy_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, with creating new index using directive on permissioned collection, no error",
		Actions: []any{

			testUtils.AddPolicy{
				Identity:         acpUtils.Actor1Identity,
				Policy:           userPolicy,
				ExpectedPolicyID: "9b5bf263a047605cce43360bf0546911b3d5da78b2b12a894318ad2a084a4a21",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "9b5bf263a047605cce43360bf0546911b3d5da78b2b12a894318ad2a084a4a21",
						resource: "users"
					) {
						name: String @index
						age: Int
					}
				`,
			},

			testUtils.Request{
				Request: `
					query  {
						Users {
							name
							age
						}
					}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
