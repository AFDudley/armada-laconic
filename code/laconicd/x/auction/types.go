package auction

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Auction status values.
const (
	// Auction is in commit phase.
	AuctionStatusCommitPhase = "commit"

	// Auction is in reveal phase.
	AuctionStatusRevealPhase = "reveal"

	// Auction has ended (no reveals allowed).
	AuctionStatusExpired = "expired"

	// Auction has completed (winner selected).
	AuctionStatusCompleted = "completed"
)

// Bid status values.
const (
	BidStatusCommitted = "commit"
	BidStatusRevealed  = "reveal"
)

// Auction kinds
const (
	AuctionKindVickrey  = "vickrey"
	AuctionKindProvider = "provider"
)

// AuctionId simplifies generation of auction ids.
type AuctionId struct {
	Address  sdk.Address
	AccNum   uint64
	Sequence uint64
}

// Generate creates the auction id.
func (auctionId AuctionId) Generate(ac address.Codec) string {
	hasher := sha256.New()
	addrStr, err := ac.BytesToString(auctionId.Address.Bytes())
	if err != nil {
		panic(err)
	}
	str := fmt.Sprintf("%s:%d:%d", addrStr, auctionId.AccNum, auctionId.Sequence)
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (auction Auction) GetCreateTime() string {
	return string(sdk.FormatTimeBytes(auction.CreateTime))
}

func (auction Auction) GetCommitsEndTime() string {
	return string(sdk.FormatTimeBytes(auction.CommitsEndTime))
}

func (auction Auction) GetRevealsEndTime() string {
	return string(sdk.FormatTimeBytes(auction.RevealsEndTime))
}

func (bid Bid) GetCommitTime() string {
	return string(sdk.FormatTimeBytes(bid.CommitTime))
}

func (bid Bid) GetRevealTime() string {
	return string(sdk.FormatTimeBytes(bid.RevealTime))
}
