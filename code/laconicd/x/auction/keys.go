package auction

import "cosmossdk.io/collections"

const (
	ModuleName = "auction"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// AuctionBurnModuleAccountName is the name of the auction burn module account.
	AuctionBurnModuleAccountName = "auction_burn"
)

// Store prefixes
var (
	// ParamsKey is the prefix for params key
	ParamsPrefix = collections.NewPrefix(0)

	AuctionsPrefix          = collections.NewPrefix(1)
	AuctionOwnerIndexPrefix = collections.NewPrefix(2)

	BidsPrefix                 = collections.NewPrefix(3)
	BidderAuctionIdIndexPrefix = collections.NewPrefix(4)
)
