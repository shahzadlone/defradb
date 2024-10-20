// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_default

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var minTypeIndexJoinPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"minNode": dataMap{
						"selectNode": dataMap{
							"typeIndexJoin": normalTypeJoinPattern,
						},
					},
				},
			},
		},
	},
}

func TestDefaultExplainRequest_WithMinOnOneToManyJoinedField_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with min on a one-to-many joined field.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						_docID
						TotalPages: _min(
							books: {field: pages}
						)
					}
				}`,

				ExpectedPatterns: minTypeIndexJoinPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "minNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName":      "books",
									"childFieldName": "pages",
									"filter":         nil,
								},
							},
						},
					},
					{
						TargetNodeName:    "typeIndexJoin",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"joinType":    "typeJoinMany",
							"rootName":    immutable.Some("author"),
							"subTypeName": "books",
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of root
						OccurancesToSkip:  0,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter":         nil,
							"spans": []dataMap{
								{
									"start": "/3",
									"end":   "/4",
								},
							},
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of subType (related type)
						OccurancesToSkip:  1,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "2",
							"collectionName": "Book",
							"filter":         nil,
							"spans": []dataMap{
								{
									"start": "/2",
									"end":   "/3",
								},
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequest_WithMinOnOneToManyJoinedFieldWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with min on a one-to-many joined field, with filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						TotalPages: _min(
							articles: {
								field: pages,
								filter: {
									name: {
										_eq: "To my dear readers"
									}
								}
							}
						)
					}
				}`,

				ExpectedPatterns: minTypeIndexJoinPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "minNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName":      "articles",
									"childFieldName": "pages",
									"filter": dataMap{
										"name": dataMap{
											"_eq": "To my dear readers",
										},
									},
								},
							},
						},
					},
					{
						TargetNodeName:    "typeIndexJoin",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"joinType":    "typeJoinMany",
							"rootName":    immutable.Some("author"),
							"subTypeName": "articles",
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of root
						OccurancesToSkip:  0,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter":         nil,
							"spans": []dataMap{
								{
									"start": "/3",
									"end":   "/4",
								},
							},
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of subType (related type)
						OccurancesToSkip:  1,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "1",
							"collectionName": "Article",
							"filter": dataMap{
								"name": dataMap{
									"_eq": "To my dear readers",
								},
							},
							"spans": []dataMap{
								{
									"start": "/1",
									"end":   "/2",
								},
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequest_WithMinOnOneToManyJoinedFieldWithManySources_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with min on a one-to-many joined field with many sources.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						TotalPages: _min(
							books: {field: pages},
							articles: {field: pages}
						)
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"minNode": dataMap{
										"selectNode": dataMap{
											"parallelNode": []dataMap{
												{
													"typeIndexJoin": normalTypeJoinPattern,
												},
												{
													"typeIndexJoin": normalTypeJoinPattern,
												},
											},
										},
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "minNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"childFieldName": "pages",
									"fieldName":      "books",
									"filter":         nil,
								},

								{
									"childFieldName": "pages",
									"fieldName":      "articles",
									"filter":         nil,
								},
							},
						},
					},
					{
						TargetNodeName:    "typeIndexJoin",
						OccurancesToSkip:  0,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"joinType":    "typeJoinMany",
							"rootName":    immutable.Some("author"),
							"subTypeName": "books",
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of 1st root type
						OccurancesToSkip:  0,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter":         nil,
							"spans": []dataMap{
								{
									"start": "/3",
									"end":   "/4",
								},
							},
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of 1st subType (related type)
						OccurancesToSkip:  1,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "2",
							"collectionName": "Book",
							"filter":         nil,
							"spans": []dataMap{
								{
									"start": "/2",
									"end":   "/3",
								},
							},
						},
					},
					{
						TargetNodeName:    "typeIndexJoin",
						OccurancesToSkip:  1,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"joinType":    "typeJoinMany",
							"rootName":    immutable.Some("author"),
							"subTypeName": "articles",
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of 2nd root type (AKA: subType's root)
						OccurancesToSkip:  2,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter":         nil,
							"spans": []dataMap{
								{
									"start": "/3",
									"end":   "/4",
								},
							},
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of 2nd subType (AKA: subType's subtype)
						OccurancesToSkip:  3,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "1",
							"collectionName": "Article",
							"filter":         nil,
							"spans": []dataMap{
								{
									"start": "/1",
									"end":   "/2",
								},
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
