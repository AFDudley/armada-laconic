package system

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/tidwall/gjson"

	"git.vdb.to/cerc-io/laconicd/x/registry"
)

const badPath = "/asdasd"

func (s *registrySuite) TestGRPCQueryParams() {
	reqURL := s.endpointURL("registry/v1/params")

	testCases := []struct {
		name   string
		url    string
		errmsg string
	}{
		{
			"invalid request",
			reqURL + badPath,
			"Not Implemented",
		},
		{
			"valid request",
			reqURL,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)

			if tc.errmsg != "" {
				message := gjson.Get(string(resp), "message").String()
				require.Contains(message, tc.errmsg, tc.url)
			} else {
				var response registry.QueryParamsResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &response), string(resp))

				params := registry.DefaultParams()
				s.updateParams(&params)
				require.Equal(params.String(), response.GetParams().String())
			}
		})
	}
}

func (s *registrySuite) TestGRPCQueryWhoIs() {
	reqURL := s.endpointURL("registry/v1/whois/%s")
	authorityName := "QueryWhoIS"

	testCases := []struct {
		name   string
		url    string
		errmsg string
		prerun func(string)
	}{
		{
			"invalid url",
			reqURL + badPath,
			"Not Implemented",
			func(string) {},
		},
		{
			"valid request",
			reqURL,
			"",
			func(name string) { s.reserveName(name) },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			tc.prerun(authorityName)
			url := fmt.Sprintf(tc.url, authorityName)

			resp, err := testutil.GetRequest(url)
			require.NoError(err)

			if tc.errmsg != "" {
				message := gjson.Get(string(resp), "message").String()
				require.Contains(message, tc.errmsg, tc.url)
			} else {
				var response registry.QueryWhoisResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &response), string(resp))
				require.Equal(registry.AuthorityActive, response.GetNameAuthority().Status)
			}
		})
	}
}

func (s *registrySuite) TestGRPCQueryLookup() {
	reqURL := s.endpointURL("registry/v1/lookup")
	authorityName := "QueryLookUp"

	s.createNameRecord(authorityName)

	testCases := []struct {
		name   string
		url    string
		errmsg string
	}{
		{
			"invalid url",
			reqURL + badPath,
			"Not Implemented",
		},
		{
			"nonexistent LRN",
			fmt.Sprintf(reqURL+"?lrn=lrn://%s/", "nonexistent"),
			"not found",
		},
		{
			"valid request",
			fmt.Sprintf(reqURL+"?lrn=lrn://%s/", authorityName),
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)

			if tc.errmsg != "" {
				message := gjson.Get(string(resp), "message").String()
				require.Contains(message, tc.errmsg, tc.url)
			} else {
				var response registry.QueryLookupLrnResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &response), string(resp))
				require.NotEmpty(response.Name.Latest.Id)
			}
		})
	}
}

func (s *registrySuite) TestGRPCQueryListRecords() {
	reqURL := s.endpointURL("registry/v1/records")
	bondId := s.bondId

	testCases := []struct {
		name   string
		url    string
		errmsg string
		prerun func(string)
	}{
		{
			"invalid url",
			reqURL + badPath,
			"not found",
			func(string) {},
		},
		{
			"valid request",
			reqURL,
			"",
			func(id string) { s.createRecord(id) },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			tc.prerun(bondId)

			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)

			if tc.errmsg != "" {
				message := gjson.Get(string(resp), "message").String()
				require.Contains(message, tc.errmsg, tc.url)
			} else {
				var response registry.QueryRecordsResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &response), string(resp))
				require.NotEmpty(response.GetRecords())
				require.Equal(bondId, response.GetRecords()[0].GetBondId())
			}
		})
	}
}

func (s *registrySuite) TestGRPCQueryGetRecordById() {
	reqURL := s.endpointURL("registry/v1/records/%s")
	bondId := s.bondId

	testCases := []struct {
		name   string
		url    string
		errmsg string
		prerun func(string) string
	}{
		{
			"invalid url",
			reqURL + badPath,
			"not found",
			func(string) string { return "" },
		},
		{
			"valid request",
			reqURL,
			"",
			func(id string) string {
				require := s.Require()

				// create a record and get the id
				s.createRecord(id)
				out := s.cli().CustomQuery("q", "registry", "list")
				var response registry.QueryRecordsResponse
				require.NoError(s.codec.UnmarshalJSON([]byte(out), &response), out)
				require.NotEmpty(response.GetRecords())

				return response.GetRecords()[0].Id
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			recordId := tc.prerun(bondId)
			url := fmt.Sprintf(tc.url, recordId)

			resp, err := testutil.GetRequest(url)
			require.NoError(err)

			if tc.errmsg != "" {
				message := gjson.Get(string(resp), "message").String()
				require.Contains(message, tc.errmsg, tc.url)
			} else {
				var response registry.QueryGetRecordResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &response), string(resp))
				record := response.GetRecord()
				require.NotEmpty(record.GetId())
				require.Equal(recordId, record.GetId())
			}
		})
	}
}

func (s *registrySuite) TestGRPCQueryGetRecordByBondId() {
	reqURL := s.endpointURL("registry/v1/records-by-bond-id/%s")
	bondId := s.bondId

	testCases := []struct {
		name   string
		url    string
		errmsg string
		prerun func(string)
	}{
		{
			"invalid url",
			reqURL + badPath,
			"Not Implemented",
			func(string) {},
		},
		{
			"valid request",
			reqURL,
			"",
			func(id string) { s.createRecord(id) },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			tc.prerun(bondId)
			url := fmt.Sprintf(tc.url, bondId)

			resp, err := testutil.GetRequest(url)
			require.NoError(err)

			if tc.errmsg != "" {
				message := gjson.Get(string(resp), "message").String()
				require.Contains(message, tc.errmsg, tc.url)
			} else {
				var response registry.QueryGetRecordsByBondIdResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &response), string(resp))
				records := response.GetRecords()
				require.NotEmpty(records)
				require.Equal(bondId, records[0].GetBondId())
			}
		})
	}
}

func (s *registrySuite) TestGRPCQueryGetRegistryModuleBalance() {
	reqURL := s.endpointURL("registry/v1/balance")
	bondId := s.bondId

	testCases := []struct {
		name   string
		url    string
		errmsg string
		prerun func(string)
	}{
		{
			"invalid url",
			reqURL + badPath,
			"Not Implemented",
			func(string) {},
		},
		{
			"valid request",
			reqURL,
			"",
			func(id string) { s.createRecord(id) },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			tc.prerun(bondId)

			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)

			if tc.errmsg != "" {
				message := gjson.Get(string(resp), "message").String()
				require.Contains(message, tc.errmsg, tc.url)
			} else {
				var response registry.QueryGetRegistryModuleBalanceResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &response), string(resp))
				require.NotEmpty(response.GetBalances())
			}
		})
	}
}

func (s *registrySuite) TestGRPCQueryNamesList() {
	reqURL := s.endpointURL("registry/v1/names")

	testCases := []struct {
		name   string
		url    string
		errmsg string
		prerun func(string)
	}{
		{
			"invalid url",
			reqURL + badPath,
			"Not Implemented",
			func(string) {},
		},
		{
			"valid request",
			reqURL,
			"",
			func(name string) { s.createNameRecord(name) },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			tc.prerun("ListNameRecords")

			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)

			if tc.errmsg != "" {
				message := gjson.Get(string(resp), "message").String()
				require.Contains(message, tc.errmsg, tc.url)
			} else {
				var response registry.QueryNameRecordsResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &response), string(resp))
				require.NotEmpty(response.GetNames())
			}
		})
	}
}
