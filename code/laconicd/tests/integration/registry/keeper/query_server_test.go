package keeper_test

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"

	types "git.vdb.to/cerc-io/laconicd/x/registry"
	"git.vdb.to/cerc-io/laconicd/x/registry/client/cli"
	"git.vdb.to/cerc-io/laconicd/x/registry/helpers"
	registryKeeper "git.vdb.to/cerc-io/laconicd/x/registry/keeper"
)

func (kts *KeeperTestSuite) TestGrpcQueryParams() {
	testCases := []struct {
		msg string
		req *types.QueryParamsRequest
	}{
		{
			"Get Params",
			&types.QueryParamsRequest{},
		},
	}
	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s ", test.msg), func() {
			resp, _ := kts.queryClient.Params(context.Background(), test.req)
			defaultParams := types.DefaultParams()
			kts.Require().Equal(defaultParams.String(), resp.GetParams().String())
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcGetRecordLists() {
	ctx, queryClient := kts.SdkCtx, kts.queryClient
	sr := kts.Require()

	var recordId string
	examples := []string{
		"../../../data/examples/service_provider_example.yml",
		"../../../data/examples/website_registration_example.yml",
		"../../../data/examples/general_record_example.yml",
	}
	testCases := []struct {
		msg           string
		req           *types.QueryRecordsRequest
		createRecords bool
		expErr        bool
		noOfRecords   int
	}{
		{
			"Empty Records",
			&types.QueryRecordsRequest{},
			false,
			false,
			0,
		},
		{
			"List Records",
			&types.QueryRecordsRequest{},
			true,
			false,
			3,
		},
		{
			"Filter with type",
			&types.QueryRecordsRequest{
				Attributes: []*types.QueryRecordsRequest_KeyValueInput{
					{
						Key: "type",
						Value: &types.QueryRecordsRequest_ValueInput{
							Value: &types.QueryRecordsRequest_ValueInput_String_{String_: "WebsiteRegistrationRecord"},
						},
					},
				},
				All: true,
			},
			true,
			false,
			1,
		},
		// Skip the following test as querying with recursive values not supported (PR https://git.vdb.to/cerc-io/laconicd/pulls/112)
		// See function RecordsFromAttributes (QueryValueToJSON call) in the registry keeper implementation (x/registry/keeper/keeper.go)
		// {
		// 	"Filter with tag (extant) (https://git.vdb.to/cerc-io/laconicd/issues/129)",
		// 	&types.QueryRecordsRequest{
		// 		Attributes: []*types.QueryRecordsRequest_KeyValueInput{
		// 			{
		// 				Key: "tags",
		// 				// Value: &types.QueryRecordsRequest_ValueInput{
		// 				// 	Value: &types.QueryRecordsRequest_ValueInput_String_{"tagA"},
		// 				// },
		// 				Value: &types.QueryRecordsRequest_ValueInput{
		// 					Value: &types.QueryRecordsRequest_ValueInput_Array{Array: &types.QueryRecordsRequest_ArrayInput{
		// 						Values: []*types.QueryRecordsRequest_ValueInput{
		// 							{
		// 								Value: &types.QueryRecordsRequest_ValueInput_String_{"tagA"},
		// 							},
		// 						},
		// 					}},
		// 				},
		// 				// Throws: "Recursive query values are not supported"
		// 			},
		// 		},
		// 		All: true,
		// 	},
		// 	true,
		// 	false,
		// 	1,
		// },
		{
			"Filter with tag (non-existent) (https://git.vdb.to/cerc-io/laconicd/issues/129)",
			&types.QueryRecordsRequest{
				Attributes: []*types.QueryRecordsRequest_KeyValueInput{
					{
						Key: "tags",
						Value: &types.QueryRecordsRequest_ValueInput{
							Value: &types.QueryRecordsRequest_ValueInput_String_{String_: "NOEXIST"},
						},
					},
				},
				All: true,
			},
			true,
			false,
			0,
		},
		{
			"Filter test for key collision (https://git.vdb.to/cerc-io/laconicd/issues/122)",
			&types.QueryRecordsRequest{
				Attributes: []*types.QueryRecordsRequest_KeyValueInput{
					{
						Key: "typ",
						Value: &types.QueryRecordsRequest_ValueInput{
							Value: &types.QueryRecordsRequest_ValueInput_String_{String_: "eWebsiteRegistrationRecord"},
						},
					},
				},
				All: true,
			},
			true,
			false,
			0,
		},
		{
			"Filter with attributes ServiceProviderRegistration",
			&types.QueryRecordsRequest{
				Attributes: []*types.QueryRecordsRequest_KeyValueInput{
					{
						Key: "x500state_name",
						Value: &types.QueryRecordsRequest_ValueInput{
							Value: &types.QueryRecordsRequest_ValueInput_String_{String_: "california"},
						},
					},
				},
				All: true,
			},
			true,
			false,
			1,
		},
	}
	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s ", test.msg), func() {
			if test.createRecords {
				for _, example := range examples {
					filePath, err := filepath.Abs(example)
					sr.NoError(err)
					payloadType, err := cli.GetPayloadFromFile(filePath)
					sr.NoError(err)
					payload := payloadType.ToPayload()
					record, err := kts.RegistryKeeper.SetRecord(ctx, types.MsgSetRecord{
						BondId:  kts.bond.GetId(),
						Signer:  kts.accounts[0].String(),
						Payload: payload,
					})
					sr.NoError(err)
					sr.NotNil(record.Id)
				}
			}

			resp, err := queryClient.Records(context.Background(), test.req)

			if test.expErr {
				kts.Error(err)
			} else {
				sr.NoError(err)
				sr.Equal(test.noOfRecords, len(resp.GetRecords()))
				if test.createRecords && test.noOfRecords > 0 {
					recordId = resp.GetRecords()[0].GetId()
					sr.NotZero(resp.GetRecords())
					sr.Equal(resp.GetRecords()[0].GetBondId(), kts.bond.GetId())

					for _, record := range resp.GetRecords() {
						recAttr := helpers.MustUnmarshalJSON[types.AttributeMap](record.Attributes)

						for _, attr := range test.req.GetAttributes() {
							enc, err := registryKeeper.QueryValueToJSON(attr.Value)
							sr.NoError(err)
							av := helpers.MustUnmarshalJSON[any](enc)

							if nil != av && nil != recAttr[attr.Key] &&
								reflect.Slice == reflect.TypeOf(recAttr[attr.Key]).Kind() &&
								reflect.Slice != reflect.TypeOf(av).Kind() {
								found := false
								allValues := recAttr[attr.Key].([]interface{})
								for i := range allValues {
									if av == allValues[i] {
										fmt.Printf("Found %s in %s", allValues[i], recAttr[attr.Key])
										found = true
									}
								}
								sr.Equal(true, found, fmt.Sprintf("Unable to find %s in %s", av, recAttr[attr.Key]))
							} else {
								if attr.Key[:4] == "x500" {
									sr.Equal(av, recAttr["x500"].(map[string]interface{})[attr.Key[4:]])
								} else {
									sr.Equal(av, recAttr[attr.Key])
								}
							}
						}
					}
				}
			}
		})
	}

	// Get the records by record id
	testCases1 := []struct {
		msg          string
		req          *types.QueryGetRecordRequest
		createRecord bool
		expErr       bool
		noOfRecords  int
	}{
		{
			"Invalid Request without record id",
			&types.QueryGetRecordRequest{},
			false,
			true,
			0,
		},
		{
			"With Record ID",
			&types.QueryGetRecordRequest{
				Id: recordId,
			},
			true,
			false,
			1,
		},
	}
	for _, test := range testCases1 {
		kts.Run(fmt.Sprintf("Case %s ", test.msg), func() {
			resp, err := queryClient.GetRecord(context.Background(), test.req)

			if test.expErr {
				kts.Error(err)
			} else {
				sr.NoError(err)
				sr.NotNil(resp.GetRecord())
				if test.createRecord {
					sr.Equal(resp.GetRecord().BondId, kts.bond.GetId())
					sr.Equal(resp.GetRecord().Id, recordId)
				}
			}
		})
	}

	// Get the records by record id
	testCasesByBondId := []struct {
		msg          string
		req          *types.QueryGetRecordsByBondIdRequest
		createRecord bool
		expErr       bool
		noOfRecords  int
	}{
		{
			"Invalid Request without bond id",
			&types.QueryGetRecordsByBondIdRequest{},
			false,
			true,
			0,
		},
		{
			"With Bond ID",
			&types.QueryGetRecordsByBondIdRequest{
				Id: kts.bond.GetId(),
			},
			true,
			false,
			1,
		},
	}
	for _, test := range testCasesByBondId {
		kts.Run(fmt.Sprintf("Case %s ", test.msg), func() {
			resp, err := queryClient.GetRecordsByBondId(context.Background(), test.req)

			if test.expErr {
				sr.Zero(resp.GetRecords())
			} else {
				sr.NoError(err)
				sr.NotNil(resp.GetRecords())
				if test.createRecord {
					sr.NotZero(resp.GetRecords())
					sr.Equal(resp.GetRecords()[0].GetBondId(), kts.bond.GetId())
				}
			}
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcQueryRegistryModuleBalance() {
	queryClient, ctx := kts.queryClient, kts.SdkCtx
	sr := kts.Require()
	examples := []string{
		"../../../data/examples/service_provider_example.yml",
		"../../../data/examples/website_registration_example.yml",
	}
	testCases := []struct {
		msg           string
		req           *types.QueryGetRegistryModuleBalanceRequest
		createRecords bool
		expErr        bool
		noOfRecords   int
	}{
		{
			"Get Module Balance",
			&types.QueryGetRegistryModuleBalanceRequest{},
			true,
			false,
			1,
		},
	}
	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s ", test.msg), func() {
			if test.createRecords {
				for _, example := range examples {
					filePath, err := filepath.Abs(example)
					sr.NoError(err)
					payloadType, err := cli.GetPayloadFromFile(filePath)
					sr.NoError(err)
					payload := payloadType.ToPayload()
					record, err := kts.RegistryKeeper.SetRecord(ctx, types.MsgSetRecord{
						BondId:  kts.bond.GetId(),
						Signer:  kts.accounts[0].String(),
						Payload: payload,
					})
					sr.NoError(err)
					sr.NotNil(record.Id)
				}
			}
			resp, err := queryClient.GetRegistryModuleBalance(context.Background(), test.req)
			if test.expErr {
				kts.Error(err)
			} else {
				sr.NoError(err)
				sr.Equal(test.noOfRecords, len(resp.GetBalances()))
				if test.createRecords {
					balance := resp.GetBalances()[0]
					sr.Equal(balance.AccountName, types.RecordRentModuleAccountName)
				}
			}
		})
	}
}

func (kts *KeeperTestSuite) TestGrpcQueryWhoIs() {
	queryClient, ctx := kts.queryClient, kts.SdkCtx
	sr := kts.Require()
	authorityName := "TestGrpcQueryWhoIs"

	testCases := []struct {
		msg         string
		req         *types.QueryWhoisRequest
		createName  bool
		expErr      bool
		noOfRecords int
	}{
		{
			"Invalid Request without name",
			&types.QueryWhoisRequest{},
			false,
			true,
			1,
		},
		{
			"Success",
			&types.QueryWhoisRequest{},
			true,
			false,
			1,
		},
	}
	for _, test := range testCases {
		kts.Run(fmt.Sprintf("Case %s ", test.msg), func() {
			if test.createName {
				err := kts.RegistryKeeper.ReserveAuthority(ctx, types.MsgReserveAuthority{
					Name:   authorityName,
					Signer: kts.accounts[0].String(),
					Owner:  kts.accounts[0].String(),
				})
				sr.NoError(err)
				test.req = &types.QueryWhoisRequest{Name: authorityName}
			}
			resp, err := queryClient.Whois(context.Background(), test.req)
			if test.expErr {
				kts.Error(err)
				sr.Nil(resp)
			} else {
				sr.NoError(err)
				if test.createName {
					nameAuth := resp.NameAuthority
					sr.NotNil(nameAuth)
					sr.Equal(nameAuth.OwnerAddress, kts.accounts[0].String())
					sr.Equal(types.AuthorityActive, nameAuth.Status)
				}
			}
		})
	}
}
