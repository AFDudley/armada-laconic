package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"

	auctionv1 "git.vdb.to/cerc-io/laconicd/api/cerc/auction/v1"
)

var _ autocli.HasAutoCLIConfig = AppModule{}

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: auctionv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Params",
					Use:            "params",
					Short:          "Get the current auction parameters",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
				{
					RpcMethod:      "Auctions",
					Use:            "list",
					Short:          "List auctions",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
				{
					RpcMethod: "GetAuction",
					Use:       "get <auction-id>",
					Short:     "Get auction info by auction id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "AuctionsByOwner",
					Use:       "by-owner <owner-address>",
					Short:     "Get auctions list by owner / creator address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "owner_address"},
					},
				},
				{
					RpcMethod: "GetBid",
					Use:       "get-bid <auction-id> <bidder>",
					Short:     "Get auction bid by auction id and bidder",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "auction_id"},
						{ProtoField: "bidder"},
					},
				},
				{
					RpcMethod: "GetBids",
					Use:       "get-bids <auction-id>",
					Short:     "Get all auction bids",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "auction_id"},
					},
				},
				{
					RpcMethod: "AuctionsByBidder",
					Use:       "by-bidder <bidder>",
					Short:     "Get auctions list by bidder",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "bidder_address"},
					},
				},
				{
					RpcMethod:      "GetAuctionModuleBalance",
					Use:            "balance",
					Short:          "Get auction module account balances",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: auctionv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CreateAuction",
					Use:       "create <commits-duration> <reveals-duration> <commit-fee> <reveal-fee>",
					Short:     "Create an auction",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "commits_duration"},
						{ProtoField: "reveals_duration"},
						{ProtoField: "commit_fee"},
						{ProtoField: "reveal_fee"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"kind":          {Name: "kind", DefaultValue: "vickrey", Usage: "Auction kind (vickrey|provider) (default: vickrey)"},
						"minimum_bid":   {Name: "minimum-bid", Usage: "Minimum bid (required for vickrey auctions)"},
						"max_price":     {Name: "max-price", Usage: "Max price (required for provider auctions)"},
						"num_providers": {Name: "num-providers", Usage: "Number of desired providers (required for provider auctions)"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
			},

			EnhanceCustomCommand: true, // Allow additional manual commands
		},
	}
}
