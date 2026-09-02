package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	onboardingv1 "git.vdb.to/cerc-io/laconicd/api/cerc/onboarding/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: onboardingv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Participants",
					Use:            "list",
					Short:          "List participants",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
				{
					RpcMethod: "GetParticipantByAddress",
					Use:       "get-by-address <address>",
					Short:     "Get participant by address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "GetParticipantByNitroAddress",
					Use:       "get-by-nitro-address <nitro-address>",
					Short:     "Get participant by nitro address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "nitro_address"},
					},
				},
			},
		},
		// TODO: Use JSON file for input
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: onboardingv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "OnboardParticipant",
					Use:       "enroll <eth-payload> <eth-signature> <role> <kyc-id>",
					Short:     "Enroll a testnet validator",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "eth_payload"},
						{ProtoField: "eth_signature"},
						{ProtoField: "role"},
						{ProtoField: "kyc_id"},
					},
				},
			},
		},
	}
}
