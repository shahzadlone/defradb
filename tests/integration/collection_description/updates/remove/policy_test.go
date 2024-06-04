// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package remove

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	acpUtils "github.com/sourcenetwork/defradb/tests/integration/acp"
)

func TestColDescrUpdateRemovePolicy_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddPolicy{

				Identity: acpUtils.Actor1Identity,

				Policy: `
                    description: a test policy which marks a collection in a database as a resource

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                          admin:
                            manages:
                              - reader
                            types:
                              - actor
                `,

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

			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/1/Policy" }
					]
				`,
				ExpectedError: "collection policy cannot be mutated. CollectionID: 1",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
