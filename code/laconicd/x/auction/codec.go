package auction

import (
	registry "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry registry.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateAuction{},
		&MsgCommitBid{},
		&MsgRevealBid{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &Msg_serviceDesc)
}
