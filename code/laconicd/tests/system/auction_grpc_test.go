package system

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/tidwall/gjson"

	"git.vdb.to/cerc-io/laconicd/x/auction"
)

const (
	randomAuctionId     = "randomAuctionId"
	randomBidderAddress = "randomBidderAddress"
	randomOwnerAddress  = "randomOwnerAddress"
)

func (s *auctionSuite) TestGRPCQueryParams() {
	reqURL := s.endpointURL("auction/v1/params")

	s.Run("valid request to get auction params", func() {
		require := s.Require()

		resp, err := testutil.GetRequest(reqURL)
		require.NoError(err)

		var params auction.QueryParamsResponse
		require.NoError(s.codec.UnmarshalJSON(resp, &params), string(resp))
		require.Equal(auction.DefaultParams(), *params.GetParams())
	})
}

func (s *auctionSuite) TestGRPCGetAllAuctions() {
	reqURL := s.endpointURL("auction/v1/auctions")

	s.createAuction()

	testCases := []struct {
		name   string
		url    string
		errmsg string
	}{
		{
			"invalid request",
			reqURL + "-asdasd",
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
				var auctions auction.QueryAuctionsResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &auctions), string(resp))
				require.NotEmpty(auctions.Auctions.Auctions)
			}
		})
	}
}

func (s *auctionSuite) TestGRPCGetAuction() {
	reqURL := s.endpointURL("auction/v1/auctions/%s")

	auctionId := s.createAuction()
	s.createBid(auctionId)

	testCases := []struct {
		name   string
		url    string
		exists bool
	}{
		{
			"nonexistent auction",
			fmt.Sprintf(reqURL, randomAuctionId),
			false,
		},
		{
			"valid request",
			fmt.Sprintf(reqURL, auctionId),
			true,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)

			if tc.exists {
				var auction auction.QueryGetAuctionResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &auction), string(resp))
				require.Equal(auctionId, auction.Auction.Id)
			} else {
				message := gjson.Get(string(resp), "message").String()
				require.Contains(message, "not found", tc.url)
			}
		})
	}
}

func (s *auctionSuite) TestGRPCGetBids() {
	reqURL := s.endpointURL("auction/v1/bids/%s")

	auctionId := s.createAuction()
	s.createBid(auctionId)

	testCases := []struct {
		name   string
		url    string
		exists bool
	}{
		{
			"nonexistent auction",
			fmt.Sprintf(reqURL, randomAuctionId),
			false,
		},
		{
			"valid request",
			fmt.Sprintf(reqURL, auctionId),
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)

			var bids auction.QueryGetBidsResponse
			require.NoError(s.codec.UnmarshalJSON(resp, &bids), string(resp))
			if tc.exists {
				require.Equal(auctionId, bids.Bids[0].AuctionId)
			} else {
				require.Empty(bids.Bids)
			}
		})
	}
}

func (s *auctionSuite) TestGRPCGetBid() {
	reqURL := s.endpointURL("auction/v1/bids/%s/%s")
	bidderAddress := s.cli().GetKeyAddr(s.bidderAccount)

	auctionId := s.createAuction()
	s.createBid(auctionId)

	testCases := []struct {
		name   string
		url    string
		exists bool
	}{
		{
			"nonexistent auction",
			fmt.Sprintf(reqURL, randomAuctionId, bidderAddress),
			false,
		},
		{
			"nonexistent bidder",
			fmt.Sprintf(reqURL, auctionId, randomBidderAddress),
			false,
		},
		{
			"valid request",
			fmt.Sprintf(reqURL, auctionId, bidderAddress),
			true,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)

			if tc.exists {
				var bid auction.QueryGetBidResponse
				require.NoError(s.codec.UnmarshalJSON(resp, &bid), string(resp))
				require.Equal(auctionId, bid.Bid.AuctionId)
			} else {
				message := gjson.Get(string(resp), "message").String()
				require.Contains(message, "not found", tc.url)
			}
		})
	}
}

func (s *auctionSuite) TestGRPCGetAuctionsByOwner() {
	reqURL := s.endpointURL("auction/v1/by-owner/%s")
	ownerAddress := s.cli().GetKeyAddr(s.ownerAccount)

	auctionId := s.createAuction()
	s.createBid(auctionId)

	testCases := []struct {
		name   string
		url    string
		exists bool
	}{
		{
			"nonexistent owner",
			fmt.Sprintf(reqURL, randomOwnerAddress),
			false,
		},
		{
			"valid request",
			fmt.Sprintf(reqURL, ownerAddress),
			true,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			require := s.Require()

			resp, err := testutil.GetRequest(tc.url)
			require.NoError(err)

			var auctions auction.QueryAuctionsResponse
			require.NoError(s.codec.UnmarshalJSON(resp, &auctions), string(resp))
			if tc.exists {
				require.Equal(auctionId, auctions.Auctions.Auctions[0].Id)
			} else {
				require.Empty(auctions.Auctions.Auctions)
			}
		})
	}
}

func (s *auctionSuite) TestGRPCQueryBalance() {
	reqURL := s.endpointURL("auction/v1/balance")

	s.Run("valid request", func() {
		require := s.Require()

		auctionId := s.createAuction()
		s.createBid(auctionId)

		resp, err := testutil.GetRequest(reqURL)
		require.NoError(err)

		var response auction.QueryGetAuctionModuleBalanceResponse
		require.NoError(s.codec.UnmarshalJSON(resp, &response), string(resp))
		require.NotEmpty(response.GetBalance())
	})
}
