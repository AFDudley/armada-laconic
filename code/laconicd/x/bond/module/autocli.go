package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"

	bondv1 "git.vdb.to/cerc-io/laconicd/api/cerc/bond/v1"
)

var _ autocli.HasAutoCLIConfig = AppModule{}

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: bondv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Params",
					Use:            "params",
					Short:          "Get the current bond parameters",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
				{
					RpcMethod:      "Bonds",
					Use:            "list",
					Short:          "List bonds",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
				{
					RpcMethod: "GetBondById",
					Use:       "get <bond-id>",
					Short:     "Get bond info by bond id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "GetBondsByOwner",
					Use:       "by-owner <owner-address>",
					Short:     "Get bonds list by owner address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "owner"},
					},
				},
				{
					RpcMethod:      "GetBondModuleBalance",
					Use:            "balance",
					Short:          "Get bond module account balances",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: bondv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CreateBond",
					Use:       "create <amount>",
					Short:     "Create bond",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "coins"},
					},
				},
				{
					RpcMethod: "RefillBond",
					Use:       "refill <bond-id> <amount>",
					Short:     "Refill bond",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "id"},
						{ProtoField: "coins"},
					},
				},
				{
					RpcMethod: "WithdrawBond",
					Use:       "withdraw <bond-id> <amount>",
					Short:     "Withdraw amount from bond",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "id"},
						{ProtoField: "coins"},
					},
				},
				{
					RpcMethod: "CancelBond",
					Use:       "cancel <bond-id>",
					Short:     "Cancel bond",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
			},
		},
	}
}
