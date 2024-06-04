// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_CreateWithoutIdentityAndReadWithoutIdentity_CanRead(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, create without identity, and read without identity, can read",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: Actor1Identity,

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

			testUtils.CreateDoc{
				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.Request{
				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: []map[string]any{
					{
						"_docID": "bae-1e608f7d-b01e-5dd5-ad4a-9c6cc3005a36",
						"name":   "Shahzad",
						"age":    int64(28),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithoutIdentityAndReadWithIdentity_CanRead(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, create without identity, and read with identity, can read",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: Actor1Identity,

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

			testUtils.CreateDoc{
				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.Request{
				Identity: Actor1Identity,

				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: []map[string]any{
					{
						"_docID": "bae-1e608f7d-b01e-5dd5-ad4a-9c6cc3005a36",
						"name":   "Shahzad",
						"age":    int64(28),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithIdentityAndReadWithIdentity_CanRead(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and read with identity, can read",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: Actor1Identity,

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

			testUtils.CreateDoc{
				CollectionID: 0,

				Identity: Actor1Identity,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.Request{
				Identity: Actor1Identity,

				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: []map[string]any{
					{
						"_docID": "bae-1e608f7d-b01e-5dd5-ad4a-9c6cc3005a36",
						"name":   "Shahzad",
						"age":    int64(28),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithIdentityAndReadWithoutIdentity_CanNotRead(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and read without identity, can not read",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: Actor1Identity,

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

			testUtils.CreateDoc{
				CollectionID: 0,

				Identity: Actor1Identity,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.Request{
				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithIdentityAndReadWithWrongIdentity_CanNotRead(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and read without identity, can not read",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: Actor1Identity,

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

			testUtils.CreateDoc{
				CollectionID: 0,

				Identity: Actor1Identity,

				Doc: `
 					{
 						"name": "Shahzad",
 						"age": 28
 					}
 				`,
			},

			testUtils.Request{
				Identity: Actor2Identity,

				Request: `
 					query {
 						Users {
 							_docID
 							name
 							age
 						}
 					}
 				`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
