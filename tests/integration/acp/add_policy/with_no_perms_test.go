// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_add_policy

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Note: Eventhough this file shows we can load a policy, that has no permissions. It is important
// to know that DPI always has a set of permissions it requires. Therefore when a schema is loaded,
// and it has policyID and resource defined on the collection, then before we accept that schema
// the validation occurs.
// Inotherwords, we do not allow a non-DPI compliant policy to be specified on a collection schema, if
// it is the schema would be rejected. However we register the policy with acp even if
// the policy is not DPI compliant.

func TestACP_AddPolicy_NoPermissionsOnlyOwner_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, no permissions only owner relation",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:

                        relations:
                          owner:
                            types:
                              - actor

                `,

				ExpectedPolicyID: "39b436f0c28e7ce5ed7e1c592bb578590d62ccfacef0df565ac97520c880c017",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_NoPermissionsMultiRelations_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, no permissions with multi relations",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor

                `,

				ExpectedPolicyID: "4966d643ba8b0b555624882d5ecc722a38228d99f5f7c930bfbb8843126ba349",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_NoPermissionsLabelOnlyOwner_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, no permissions label only owner relation",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        relations:
                          owner:
                            types:
                              - actor

                `,

				ExpectedPolicyID: "39b436f0c28e7ce5ed7e1c592bb578590d62ccfacef0df565ac97520c880c017",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_NoPermissionsLabelMultiRelations_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, no permissions label with multi relations",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor

                `,

				ExpectedPolicyID: "4966d643ba8b0b555624882d5ecc722a38228d99f5f7c930bfbb8843126ba349",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
