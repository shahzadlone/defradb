// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_schema_add_dpi

import (
	"fmt"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddDPISchema_SpecifiedResourceDoesNotExistOnDPI_SchemaRejected(t *testing.T) {
	policyIDOfValidDPI := "c9a09896395d8a173633626253c01d5380a635fe8a0a103b10685e1b2e81f95a"

	test := testUtils.TestCase{

		Description: "Test acp, add dpi schema, but specified resource does not exist on DPI, reject schema",

		Actions: []any{

			testUtils.AddPolicy{

				Identity: actor1Identity,

				Policy: `
                    description: A Valid Defra Policy Interface (DPI)

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
                `,

				ExpectedPolicyID: policyIDOfValidDPI,
			},

			testUtils.SchemaUpdate{
				Schema: fmt.Sprintf(`
					type Users @policy(
						id: "%s",
						resource: "doesntExist"
					) {
						name: String
						age: Int
					}
				`,
					policyIDOfValidDPI,
				),

				ExpectedError: "resource does not exist on the specified policy",
			},

			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Users") {
							name
							fields {
								name
								type {
								name
								kind
								}
							}
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": nil, // NOTE: No "Users" should exist.
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
