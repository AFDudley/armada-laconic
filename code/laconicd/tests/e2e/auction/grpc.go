package auction

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil"

	auctiontypes "git.vdb.to/cerc-io/laconicd/x/auction"
)

const (
	randomAuctionId     = "randomAuctionId"
	randomBidderAddress = "randomBidderAddress"
	randomOwnerAddress  = "randomOwnerAddress"
)

func (ets *E2ETestSuite) TestQueryParamsGrpc() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := fmt.Sprintf("%s/cerc/auction/v1/params", val.APIAddress)

	ets.Run("valid request to get auction params", func() {
		resp, err := testutil.GetRequest(reqURL)
		ets.Require().NoError(err)

		var params auctiontypes.QueryParamsResponse
		err = val.ClientCtx.Codec.UnmarshalJSON(resp, &params)

		sr.NoError(err)
		sr.Equal(*params.GetParams(), auctiontypes.DefaultParams())
	})
}

func (ets *E2ETestSuite) TestGetAllAuctionsGrpc() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := fmt.Sprintf("%s/cerc/auction/v1/auctions", val.APIAddress)

	testCases := []struct {
		msg             string
		url             string
		errorMsg        string
		isErrorExpected bool
	}{
		{
			"invalid request to get all auctions",
			reqURL + randomAuctionId,
			"",
			true,
		},
		{
			"valid request to get all auctions",
			reqURL,
			"",
			false,
		},
	}
	for _, tc := range testCases {
		ets.Run(tc.msg, func() {
			resp, err := testutil.GetRequest(tc.url)
			if tc.isErrorExpected {
				sr.Contains(string(resp), tc.errorMsg)
			} else {
				sr.NoError(err)
				var auctions auctiontypes.QueryAuctionsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &auctions)
				sr.NoError(err)
				sr.NotZero(len(auctions.Auctions.Auctions))
			}
		})
	}
}

func (ets *E2ETestSuite) TestGetAuctionGrpc() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := fmt.Sprintf("%s/cerc/auction/v1/auctions/", val.APIAddress)

	testCases := []struct {
		msg             string
		url             string
		errorMsg        string
		isErrorExpected bool
		preRun          func() string
	}{
		{
			"invalid request to get an auction",
			reqURL + randomAuctionId,
			"",
			true,
			func() string { return "" },
		},
		{
			"valid request to get an auction",
			reqURL,
			"",
			false,
			func() string { return ets.defaultAuctionId },
		},
	}
	for _, tc := range testCases {
		ets.Run(tc.msg, func() {
			auctionId := tc.preRun()
			resp, err := testutil.GetRequest(tc.url + auctionId)
			if tc.isErrorExpected {
				sr.Contains(string(resp), tc.errorMsg)
			} else {
				sr.NoError(err)
				var auction auctiontypes.QueryGetAuctionResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &auction)
				sr.NoError(err)
				sr.Equal(auctionId, auction.Auction.Id)
			}
		})
	}
}

func (ets *E2ETestSuite) TestGetBidsGrpc() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := fmt.Sprintf("%s/cerc/auction/v1/bids/", val.APIAddress)
	testCases := []struct {
		msg             string
		url             string
		errorMsg        string
		isErrorExpected bool
		preRun          func() string
	}{
		{
			"invalid request to get all bids",
			reqURL,
			"",
			true,
			func() string { return "" },
		},
		{
			"valid request to get all bids",
			reqURL,
			"",
			false,
			func() string { return ets.createAuctionAndBid(false, true) },
		},
	}

	for _, tc := range testCases {
		ets.Run(tc.msg, func() {
			auctionId := tc.preRun()
			tc.url += auctionId
			resp, err := testutil.GetRequest(tc.url)
			if tc.isErrorExpected {
				sr.Contains(string(resp), tc.errorMsg)
			} else {
				sr.NoError(err)
				var bids auctiontypes.QueryGetBidsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &bids)
				sr.NoError(err)
				sr.Equal(auctionId, bids.Bids[0].AuctionId)
			}
		})
	}
}

func (ets *E2ETestSuite) TestGetBidGrpc() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := fmt.Sprintf("%s/cerc/auction/v1/bids/", val.APIAddress)
	testCases := []struct {
		msg             string
		url             string
		errorMsg        string
		isErrorExpected bool
		preRun          func() string
	}{
		{
			"invalid request to get bid",
			reqURL,
			"",
			true,
			func() string { return randomAuctionId },
		},
		{
			"valid request to get bid",
			reqURL,
			"",
			false,
			func() string { return ets.createAuctionAndBid(false, true) },
		},
	}
	for _, tc := range testCases {
		ets.Run(tc.msg, func() {
			auctionId := tc.preRun()
			tc.url += auctionId + "/" + bidderAddress
			resp, err := testutil.GetRequest(tc.url)

			if tc.isErrorExpected {
				sr.Contains(string(resp), tc.errorMsg)
			} else {
				sr.NoError(err)
				var bid auctiontypes.QueryGetBidResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &bid)
				sr.NoError(err)
			}
		})
	}
}

func (ets *E2ETestSuite) TestGetAuctionsByOwnerGrpc() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := fmt.Sprintf("%s/cerc/auction/v1/by-owner/", val.APIAddress)
	testCases := []struct {
		msg             string
		url             string
		errorMsg        string
		isErrorExpected bool
	}{
		{
			"invalid request to get auctions by owner",
			reqURL,
			"",
			true,
		},
		{
			"valid request to get auctions by owner",
			fmt.Sprintf("%s/%s", reqURL, ownerAddress),
			"",
			false,
		},
	}
	for _, tc := range testCases {
		ets.Run(tc.msg, func() {
			resp, err := testutil.GetRequest(tc.url)
			if tc.isErrorExpected {
				sr.Contains(string(resp), tc.errorMsg)
			} else {
				sr.NoError(err)
				var auctions auctiontypes.QueryAuctionsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp, &auctions)
				sr.NoError(err)
			}
		})
	}
}

func (ets *E2ETestSuite) TestQueryBalanceGrpc() {
	val := ets.network.Validators[0]
	sr := ets.Require()
	reqURL := fmt.Sprintf("%s/cerc/auction/v1/balance", val.APIAddress)
	msg := "valid request to get the auction module balance"

	ets.createAuctionAndBid(false, true)

	ets.Run(msg, func() {
		resp, err := testutil.GetRequest(reqURL)
		sr.NoError(err)

		var response auctiontypes.QueryGetAuctionModuleBalanceResponse
		err = val.ClientCtx.Codec.UnmarshalJSON(resp, &response)

		sr.NoError(err)
		sr.NotZero(len(response.GetBalance()))
	})
}
