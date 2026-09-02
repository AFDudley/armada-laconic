package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"git.vdb.to/cerc-io/laconicd/x/auction"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/stretchr/testify/suite"
)

const (
	sampleCommitTime     = "90s"
	sampleRevealTime     = "5s"
	placeholderAuctionId = "placeholder_auction_id"
)

type auctionSuite struct {
	testSuite

	ownerAccount     string
	bidderAccount    string
	defaultAuctionId string
}

func TestAuction(t *testing.T) {
	suite.Run(t, new(auctionSuite))
}

func (s *auctionSuite) SetupTest() {
	s.resetChain()

	s.ownerAccount = s.account(0)
	s.bidderAccount = s.account(1)

	Sut.StartChain(s.T())
}

func (s *auctionSuite) TearDownTest() {
	s.cleanupBidFiles()
}

func (s *auctionSuite) TestQueryList() {
	sr := s.Require()

	testCases := []struct {
		name          string
		createAuction bool
	}{
		{
			"when no auctions exist",
			false,
		},
		{
			"after creating an auction",
			true,
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			if test.createAuction {
				s.createAuction()
			}

			out := s.cli().CustomQuery("q", "auction", "list")
			var response auction.QueryAuctionsResponse
			sr.NoError(s.codec.UnmarshalJSON([]byte(out), &response), out)
			if test.createAuction {
				sr.NotEmpty(response.Auctions.Auctions)
			}
		})
	}
}

func (s *auctionSuite) TestTxCommitBid() {
	auctionId := s.createAuction()
	testCases := []struct {
		name string
		args []string
		err  bool
	}{
		{
			"with missing args",
			[]string{auctionId},
			true,
		},
		{
			"with zero bid",
			[]string{auctionId, ("0" + s.bondDenom)},
			true,
		},
		{
			"with invalid auction",
			[]string{"fake", ("200" + s.bondDenom)},
			true,
		},
		{
			"with valid args",
			[]string{auctionId, ("200" + s.bondDenom)},
			false,
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			cmd := []string{
				"tx", "auction", "commit-bid",
			}
			cmd = append(cmd, test.args...)
			s.runTx(cmd, test.err)
		})
	}
}

func (s *auctionSuite) createAuction() string {
	require := s.Require()

	cmd := []string{
		"tx", "auction", "create",
		sampleCommitTime,
		sampleRevealTime,
		"10" + s.bondDenom,
		"10" + s.bondDenom,
		"--kind", auction.AuctionKindVickrey,
		"--minimum-bid", "100" + s.bondDenom,
		"--max-price", "0" + s.bondDenom,
		"--num-providers", "0",
	}
	s.runTx(cmd, false)

	out := s.cli().CustomQuery("q", "auction", "list")
	var response auction.QueryAuctionsResponse
	require.NoError(s.codec.UnmarshalJSON([]byte(out), &response), out)
	return response.Auctions.Auctions[0].Id
}

func (s *auctionSuite) createBid(auctionId string) {
	cmd := []string{
		"tx", "auction", "commit-bid",
		auctionId,
		"200" + s.bondDenom,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.bidderAccount),
	}
	s.runTx(cmd, false)
}

func (s *auctionSuite) cleanupBidFiles() {
	matches, err := filepath.Glob(fmt.Sprintf("%s-*.json", s.bidderAccount))
	if err != nil {
		s.T().Errorf("Error matching bidder files: %v\n", err)
	}

	for _, match := range matches {
		err := os.Remove(match)
		if err != nil {
			s.T().Errorf("Error removing bidder file: %v\n", err)
		}
	}
}
