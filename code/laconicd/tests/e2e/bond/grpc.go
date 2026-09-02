package bond

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil"

	bondtypes "git.vdb.to/cerc-io/laconicd/x/bond"
)

func (ets *E2ETestSuite) TestGRPCGetParams() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := fmt.Sprintf("%s/cerc/bond/v1/params", val.APIAddress)

	resp, err := testutil.GetRequest(reqURL)
	ets.Require().NoError(err)

	var params bondtypes.QueryParamsResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(resp, &params)

	sr.NoError(err)
	sr.Equal(params.GetParams().MaxBondAmount, bondtypes.DefaultParams().MaxBondAmount)
}

func (ets *E2ETestSuite) TestGRPCGetBonds() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := fmt.Sprintf("%s/cerc/bond/v1/bonds", val.APIAddress)

	testCases := []struct {
		name     string
		url      string
		expErr   bool
		errorMsg string
		preRun   func() string
	}{
		{
			"invalid request with headers",
			reqURL + "asdasdas",
			true,
			"",
			func() string { return "" },
		},
		{
			"valid request",
			reqURL,
			false,
			"",
			func() string { return ets.createBond() },
		},
	}
	for _, tc := range testCases {
		ets.Run(tc.name, func() {
			tc.preRun()

			resp, _ := testutil.GetRequest(tc.url)
			if tc.expErr {
				sr.Contains(string(resp), tc.errorMsg)
			} else {
				var response bondtypes.QueryBondsResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &response)
				sr.NoError(err)
				sr.NotZero(len(response.GetBonds()))
			}
		})
	}
}

func (ets *E2ETestSuite) TestGRPCGetBondsByOwner() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := val.APIAddress + "/cerc/bond/v1/by-owner/%s"

	testCases := []struct {
		name   string
		url    string
		expErr bool
		preRun func() string
	}{
		{
			"empty list",
			fmt.Sprintf(reqURL, "asdasd"),
			true,
			func() string { return "" },
		},
		{
			"valid request",
			fmt.Sprintf(reqURL, ets.accountAddress),
			false,
			func() string { return ets.createBond() },
		},
	}

	for _, tc := range testCases {
		ets.Run(tc.name, func() {
			tc.preRun()

			resp, err := testutil.GetRequest(tc.url)
			ets.Require().NoError(err)

			var bonds bondtypes.QueryGetBondsByOwnerResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &bonds)
			sr.NoError(err)
			if tc.expErr {
				sr.Empty(bonds.GetBonds())
			} else {
				bondsList := bonds.GetBonds()
				sr.NotZero(len(bondsList))
				sr.Equal(ets.accountAddress, bondsList[0].GetOwner())
			}
		})
	}
}

func (ets *E2ETestSuite) TestGRPCGetBondById() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := val.APIAddress + "/cerc/bond/v1/bonds/%s"

	testCases := []struct {
		name   string
		url    string
		expErr bool
		preRun func() string
	}{
		{
			"invalid request",
			fmt.Sprintf(reqURL, "asdadad"),
			true,
			func() string { return "" },
		},
		{
			"valid request",
			reqURL,
			false,
			func() string { return ets.createBond() },
		},
	}
	for _, tc := range testCases {
		ets.Run(tc.name, func() {
			var bondId string
			if !tc.expErr {
				bondId = tc.preRun()
				tc.url = fmt.Sprintf(reqURL, bondId)
			}

			resp, err := testutil.GetRequest(tc.url)
			ets.Require().NoError(err)

			var bonds bondtypes.QueryGetBondByIdResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &bonds)

			if tc.expErr {
				sr.Empty(bonds.GetBond().GetId())
			} else {
				sr.NoError(err)
				sr.NotZero(bonds.GetBond().GetId())
				sr.Equal(bonds.GetBond().GetId(), bondId)
			}
		})
	}
}

func (ets *E2ETestSuite) TestGRPCGetBondModuleBalance() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := fmt.Sprintf("%s/cerc/bond/v1/balance", val.APIAddress)

	// creating the bond
	ets.createBond()

	ets.Run("valid request", func() {
		resp, err := testutil.GetRequest(reqURL)
		sr.NoError(err)

		var response bondtypes.QueryGetBondModuleBalanceResponse
		err = val.ClientCtx.Codec.UnmarshalJSON(resp, &response)

		sr.NoError(err)
		sr.False(response.GetBalance().IsZero())
	})
}
