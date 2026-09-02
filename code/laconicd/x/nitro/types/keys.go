package types

import "cosmossdk.io/collections"

const (
	ModuleName = "nitro"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

// Store prefixes
var (
	ParamsPrefix = collections.NewPrefix(0)

	ChannelProposalsPrefix = collections.NewPrefix(1)
	PaymentChannelsPrefix  = collections.NewPrefix(2)
)
