package system

import (
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/cosmos/cosmos-sdk/testutil"

	"git.vdb.to/cerc-io/laconicd/x/bond"
)

func (s *bondSuite) TestGRPCGetParams() {
	reqURL := s.endpointURL("bond/v1/params")

	var params bond.QueryParamsResponse
	resp, err := testutil.GetRequest(reqURL)
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.codec.UnmarshalJSON(resp, &params))

	require.Equal(s.T(), bond.DefaultParams().MaxBondAmount, params.GetParams().MaxBondAmount)
}

func (s *bondSuite) TestGRPCGetBond() {
	reqURL := s.endpointURL("bond/v1/bonds")

	testcases := []struct {
		name   string
		url    string
		errmsg string
		prerun func()
	}{
		{
			"invalid request with headers",
			reqURL + "asdasdas",
			"Not Implemented",
			func() {},
		},
		{
			"valid request",
			reqURL,
			"",
			func() { s.createBond() },
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			require := s.Require()
			tc.prerun()

			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)
			if tc.errmsg != "" {
				message := gjson.Get(string(resp), "message").String()
				require.Contains(message, tc.errmsg, string(resp))
			} else {
				var bonds bond.QueryBondsResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &bonds), string(resp))
				require.NotEmpty(bonds.GetBonds())
			}
		})
	}
}

func (s *bondSuite) TestGRPCGetBondsByOwner() {
	reqURL := s.endpointURL("bond/v1/by-owner/%s")
	accountAddress := s.cli().GetKeyAddr(s.account(0))

	testcases := []struct {
		name   string
		url    string
		err    bool
		prerun func()
	}{
		{
			"empty list",
			fmt.Sprintf(reqURL, "asdasd"),
			true,
			func() {},
		},
		{
			"valid request",
			fmt.Sprintf(reqURL, accountAddress),
			false,
			func() { s.createBond() },
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			require := s.Require()
			tc.prerun()

			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)

			var bonds bond.QueryGetBondsByOwnerResponse
			require.NoError(s.codec.UnmarshalJSON(resp, &bonds))
			if tc.err {
				require.Empty(bonds.GetBonds())
			} else {
				bondsList := bonds.GetBonds()
				require.NotEmpty(bondsList)
				require.Equal(accountAddress, bondsList[0].GetOwner())
			}
		})
	}
}

func (s *bondSuite) TestGRPCGetBondById() {
	reqURL := s.endpointURL("bond/v1/bonds/%s")

	testcases := []struct {
		name   string
		url    string
		err    bool
		prerun func() string
	}{
		{
			"invalid request",
			reqURL,
			true,
			func() string { return "asdadad" },
		},
		{
			"valid request",
			reqURL,
			false,
			func() string { return s.createBond() },
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			require := s.Require()

			bondId := tc.prerun()
			tc.url = fmt.Sprintf(reqURL, bondId)

			var bond bond.QueryGetBondByIdResponse
			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)
			if tc.err {
				require.Error(s.codec.UnmarshalJSON(resp, &bond))
			} else {
				require.NoError(s.codec.UnmarshalJSON(resp, &bond))
				require.Equal(bondId, bond.GetBond().GetId())
			}
		})
	}
}

func (s *bondSuite) TestGRPCGetBondModuleBalance() {
	reqURL := s.endpointURL("bond/v1/balance")

	s.createBond()

	require := s.Require()

	resp, err := testutil.GetRequest(reqURL)
	require.NoError(err)

	var response bond.QueryGetBondModuleBalanceResponse
	require.NoError(s.codec.UnmarshalJSON(resp, &response))
	require.False(response.GetBalance().IsZero())
}
