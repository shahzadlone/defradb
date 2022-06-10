// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithFloatLEFilterBlockWithEqualValue(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with basic le float filter with equal value",
		Query: `query {
					users(filter: {HeightM: {_le: 1.82}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"HeightM": 2.1
				}`,
				`{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "Bob",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithFloatLEFilterBlockWithGreaterValue(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with basic le float filter with greater value",
		Query: `query {
					users(filter: {HeightM: {_le: 1.820000000001}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"HeightM": 2.1
				}`,
				`{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "Bob",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithFloatLEFilterBlockWithGreaterIntValue(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with basic le float filter with greater int value",
		Query: `query {
					users(filter: {HeightM: {_le: 2}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"HeightM": 2.1
				}`,
				`{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "Bob",
			},
		},
	}

	executeTestCase(t, test)
}