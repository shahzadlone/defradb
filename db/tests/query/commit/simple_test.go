// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package commit

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

func TestQueryOneCommit(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by CID",
		Query: `query {
					commit(cid: "bafkreiercmxn6e3qryxvuped5pplg733c5fj6gjypj5wykk63ouvcfb25m") {
						cid
						height
						delta
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"cid":    "bafkreiercmxn6e3qryxvuped5pplg733c5fj6gjypj5wykk63ouvcfb25m",
				"height": int64(1),
				// cbor encoded delta
				"delta": []uint8{0xa2, 0x63, 0x41, 0x67, 0x65, 0x15, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x64, 0x4a, 0x6f, 0x68, 0x6e},
			},
		},
	}

	executeTestCase(t, test)
}