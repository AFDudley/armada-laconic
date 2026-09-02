package auction

const (
	EventTypeCreateAuction = "create-auction"
	EventTypeCommitBid     = "commit-bid"
	EventTypeRevealBid     = "reveal-bid"
	EventTypeReleaseFunds  = "release-funds"

	AttributeKeyCommitsDuration = "commits-duration"
	AttributeKeyRevealsDuration = "reveals-duration"
	AttributeKeyCommitFee       = "commit-fee"
	AttributeKeyRevealFee       = "reveal-fee"
	AttributeKeyMinimumBid      = "minimum-bid"
	AttributeKeySigner          = "signer"
	AttributeKeyAuctionId       = "auction-id"
	AttributeKeyCommitHash      = "commit-hash"
	AttributeKeyReveal          = "reveal"

	AttributeValueCategory = ModuleName
)
