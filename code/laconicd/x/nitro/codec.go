package nitro

import (
	registry "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	"git.vdb.to/cerc-io/laconicd/x/nitro/types/v1"
)

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry registry.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&v1.MsgOpenChannel{},
		&v1.MsgCloseChannel{},
		&v1.MsgCreatePaymentChannel{},
		&v1.MsgClosePaymentChannel{},
		&v1.MsgPay{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &v1.Msg_serviceDesc)
}
