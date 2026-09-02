package registry

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"

	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
	"git.vdb.to/cerc-io/laconicd/x/registry/client/cli"
)

const badPath = "/asdasd"

func (ets *E2ETestSuite) TestGRPCQueryParams() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := val.APIAddress + "/cerc/registry/v1/params"

	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
	}{
		{
			"invalid request",
			reqURL + badPath,
			true,
			"",
		},
		{
			"valid request",
			reqURL,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		ets.Run(tc.name, func() {
			resp, err := testutil.GetRequest(tc.url)
			ets.NoError(err)
			require := ets.Require()
			if tc.expectErr {
				require.Contains(string(resp), tc.errorMsg)
			} else {
				var response registrytypes.QueryParamsResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &response)
				sr.NoError(err)

				params := registrytypes.DefaultParams()
				ets.updateParams(&params)
				sr.Equal(params.String(), response.GetParams().String())
			}
		})
	}
}

func (ets *E2ETestSuite) TestGRPCQueryWhoIs() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqUrl := val.APIAddress + "/cerc/registry/v1/whois/%s"
	authorityName := "QueryWhoIS"
	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
		preRun    func(authorityName string)
	}{
		{
			"invalid url",
			reqUrl + badPath,
			true,
			"",
			func(authorityName string) {
			},
		},
		{
			"valid request",
			reqUrl,
			false,
			"",
			func(authorityName string) { ets.reserveName(authorityName) },
		},
	}

	for _, tc := range testCases {
		ets.Run(tc.name, func() {
			tc.preRun(authorityName)
			tc.url = fmt.Sprintf(tc.url, authorityName)

			resp, err := testutil.GetRequest(tc.url)
			ets.NoError(err)
			require := ets.Require()
			if tc.expectErr {
				require.Contains(string(resp), tc.errorMsg)
			} else {
				var response registrytypes.QueryWhoisResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &response)
				sr.NoError(err)
				sr.Equal(registrytypes.AuthorityActive, response.GetNameAuthority().Status)
			}
		})
	}
}

func (ets *E2ETestSuite) TestGRPCQueryLookup() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := val.APIAddress + "/cerc/registry/v1/lookup"
	authorityName := "QueryLookUp"

	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
		preRun    func(authorityName string)
	}{
		{
			"invalid url",
			reqURL + badPath,
			true,
			"",
			func(authorityName string) {
			},
		},
		{
			"valid request",
			fmt.Sprintf(reqURL+"?lrn=lrn://%s/", authorityName),
			false,
			"",
			func(authorityName string) {
				// create name record
				ets.createNameRecord(authorityName)
			},
		},
	}

	for _, tc := range testCases {
		ets.Run(tc.name, func() {
			tc.preRun(authorityName)
			resp, err := testutil.GetRequest(tc.url)
			ets.NoError(err)
			if tc.expectErr {
				sr.Contains(string(resp), tc.errorMsg)
			} else {
				var response registrytypes.QueryLookupLrnResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &response)
				sr.NoError(err)
				sr.NotZero(len(response.Name.Latest.Id))
			}
		})
	}
}

func (ets *E2ETestSuite) TestGRPCQueryListRecords() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqUrl := val.APIAddress + "/cerc/registry/v1/records"

	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
		preRun    func(bondId string)
	}{
		{
			"invalid url",
			reqUrl + badPath,
			true,
			"",
			func(bondId string) {
			},
		},
		{
			"valid request",
			reqUrl,
			false,
			"",
			func(bondId string) { ets.createRecord(bondId) },
		},
	}

	for _, tc := range testCases {
		ets.Run(tc.name, func() {
			tc.preRun(ets.bondId)
			resp, err := testutil.GetRequest(tc.url)
			ets.NoError(err)
			require := ets.Require()
			if tc.expectErr {
				require.Contains(string(resp), tc.errorMsg)
			} else {
				var response registrytypes.QueryRecordsResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &response)
				sr.NoError(err)
				sr.NotZero(len(response.GetRecords()))
				sr.Equal(ets.bondId, response.GetRecords()[0].GetBondId())
			}
		})
	}
}

func (ets *E2ETestSuite) TestGRPCQueryGetRecordById() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := val.APIAddress + "/cerc/registry/v1/records/%s"

	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
		preRun    func(bondId string) string
	}{
		{
			"invalid url",
			reqURL + badPath,
			true,
			"",
			func(bondId string) string {
				return ""
			},
		},
		{
			"valid request",
			reqURL,
			false,
			"",
			func(bondId string) string {
				// creating the record
				ets.createRecord(bondId)

				// list the records
				clientCtx := val.ClientCtx
				cmd := cli.GetCmdList()
				args := []string{
					fmt.Sprintf("--%s=json", flags.FlagOutput),
				}
				out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
				sr.NoError(err)
				var records []registrytypes.ReadableRecord
				err = json.Unmarshal(out.Bytes(), &records)
				sr.NoError(err)
				return records[0].Id
			},
		},
	}

	for _, tc := range testCases {
		ets.Run(tc.name, func() {
			recordId := tc.preRun(ets.bondId)
			tc.url = fmt.Sprintf(reqURL, recordId)

			resp, err := testutil.GetRequest(tc.url)
			ets.NoError(err)
			require := ets.Require()
			if tc.expectErr {
				require.Contains(string(resp), tc.errorMsg)
			} else {
				var response registrytypes.QueryGetRecordResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &response)
				sr.NoError(err)
				record := response.GetRecord()
				sr.NotZero(len(record.GetId()))
				sr.Equal(record.GetId(), recordId)
			}
		})
	}
}

func (ets *E2ETestSuite) TestGRPCQueryGetRecordByBondId() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := val.APIAddress + "/cerc/registry/v1/records-by-bond-id/%s"

	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
		preRun    func(bondId string)
	}{
		{
			"invalid url",
			reqURL + badPath,
			true,
			"",
			func(bondId string) {
			},
		},
		{
			"valid request",
			reqURL,
			false,
			"",
			func(bondId string) {
				// creating the record
				ets.createRecord(bondId)
			},
		},
	}

	for _, tc := range testCases {
		ets.Run(tc.name, func() {
			tc.preRun(ets.bondId)
			tc.url = fmt.Sprintf(reqURL, ets.bondId)

			resp, err := testutil.GetRequest(tc.url)
			ets.NoError(err)
			require := ets.Require()
			if tc.expectErr {
				require.Contains(string(resp), tc.errorMsg)
			} else {
				var response registrytypes.QueryGetRecordsByBondIdResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &response)
				sr.NoError(err)
				records := response.GetRecords()
				sr.NotZero(len(records))
				sr.Equal(records[0].GetBondId(), ets.bondId)
			}
		})
	}
}

func (ets *E2ETestSuite) TestGRPCQueryGetRegistryModuleBalance() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := val.APIAddress + "/cerc/registry/v1/balance"

	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
		preRun    func(bondId string)
	}{
		{
			"invalid url",
			reqURL + badPath,
			true,
			"",
			func(bondId string) {
			},
		},
		{
			"Success",
			reqURL,
			false,
			"",
			func(bondId string) {
				// creating the record
				ets.createRecord(bondId)
			},
		},
	}

	for _, tc := range testCases {
		ets.Run(tc.name, func() {
			tc.preRun(ets.bondId)
			resp, err := testutil.GetRequest(tc.url)
			ets.NoError(err)
			require := ets.Require()
			if tc.expectErr {
				require.Contains(string(resp), tc.errorMsg)
			} else {
				var response registrytypes.QueryGetRegistryModuleBalanceResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &response)
				sr.NoError(err)
				sr.NotZero(len(response.GetBalances()))
			}
		})
	}
}

func (ets *E2ETestSuite) TestGRPCQueryNamesList() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := val.APIAddress + "/cerc/registry/v1/names"

	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
		preRun    func(authorityName string)
	}{
		{
			"invalid url",
			reqURL + badPath,
			true,
			"",
			func(authorityName string) {
			},
		},
		{
			"valid request",
			reqURL,
			false,
			"",
			func(authorityName string) {
				// create name record
				ets.createNameRecord(authorityName)
			},
		},
	}

	for _, tc := range testCases {
		ets.Run(tc.name, func() {
			tc.preRun("ListNameRecords")
			resp, err := testutil.GetRequest(tc.url)
			ets.NoError(err)
			require := ets.Require()
			if tc.expectErr {
				require.Contains(string(resp), tc.errorMsg)
			} else {
				var response registrytypes.QueryNameRecordsResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &response)
				sr.NoError(err)
				sr.NotZero(len(response.GetNames()))
			}
		})
	}
}
