package registry

import (
	registry "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(registry registry.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetName{},
		&MsgReserveAuthority{},
		&MsgDeleteName{},
		&MsgSetAuthorityBond{},

		&MsgSetRecord{},
		&MsgRenewRecord{},
		&MsgAssociateBond{},
		&MsgDissociateBond{},
		&MsgDissociateRecords{},
		&MsgReassociateRecords{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &Msg_serviceDesc)
}
