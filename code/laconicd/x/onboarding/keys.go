package onboarding

import "cosmossdk.io/collections"

const (
	ModuleName = "onboarding"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

var (
	ParamsPrefix = collections.NewPrefix(0)

	ParticipantsPrefix = collections.NewPrefix(1)
)
