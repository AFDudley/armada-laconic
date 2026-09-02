package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/event"
	"git.vdb.to/cerc-io/laconicd/x/nitro/types/v1"
)

type msgServer struct {
	k Keeper
}

func NewMsgServer(k Keeper) *msgServer {
	return &msgServer{k}
}

func (m msgServer) OpenChannel(ctx context.Context, req *v1.MsgOpenChannel) (*v1.MsgOpenChannelResponse, error) {
	if err := m.k.OpenChannel(ctx, req); err != nil {
		return nil, err
	}
	eventManager := m.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		"open_channel",
		event.Attribute{Key: "signer", Value: req.Signer},
		event.Attribute{Key: "funds", Value: req.Funds.String()},
		event.Attribute{Key: "nitro_address", Value: req.NitroAddress},
		event.Attribute{Key: "counterparty", Value: req.Counterparty},
		event.Attribute{Key: "channel_id", Value: req.ChannelId},
	); err != nil {
		return nil, err
	}

	return &v1.MsgOpenChannelResponse{}, nil
}

func (m msgServer) CloseChannel(ctx context.Context, req *v1.MsgCloseChannel) (*v1.MsgCloseChannelResponse, error) {
	if err := m.k.CloseChannel(ctx, req); err != nil {
		return nil, err
	}
	eventManager := m.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		"close_channel",
		event.Attribute{Key: "signer", Value: req.Signer},
		event.Attribute{Key: "channel_id", Value: req.ChannelId},
		event.Attribute{Key: "is_challenge", Value: fmt.Sprintf("%t", req.IsChallenge)},
	); err != nil {
		return nil, err
	}

	return &v1.MsgCloseChannelResponse{}, nil
}

func (m msgServer) CreatePaymentChannel(ctx context.Context, req *v1.MsgCreatePaymentChannel) (*v1.MsgCreatePaymentChannelResponse, error) {
	if err := m.k.CreatePaymentChannel(ctx, req); err != nil {
		return nil, err
	}
	eventManager := m.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		"create_payment_channel",
		event.Attribute{Key: "signer", Value: req.Signer},
		event.Attribute{Key: "intermediaries", Value: fmt.Sprintf("%v", req.Intermediaries)},
		event.Attribute{Key: "counterparty", Value: req.Counterparty},
		event.Attribute{Key: "challenge_duration", Value: fmt.Sprintf("%d", req.ChallengeDuration)},
		event.Attribute{Key: "funds", Value: req.Funds.String()},
		event.Attribute{Key: "channel_id", Value: req.ChannelId},
	); err != nil {
		return nil, err
	}

	return &v1.MsgCreatePaymentChannelResponse{}, nil
}

func (m msgServer) ClosePaymentChannel(ctx context.Context, req *v1.MsgClosePaymentChannel) (*v1.MsgClosePaymentChannelResponse, error) {
	if err := m.k.ClosePaymentChannel(ctx, req); err != nil {
		return nil, err
	}
	eventManager := m.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		"close_payment_channel",
		event.Attribute{Key: "signer", Value: req.Signer},
		event.Attribute{Key: "channel_id", Value: req.ChannelId},
	); err != nil {
		return nil, err
	}

	return &v1.MsgClosePaymentChannelResponse{}, nil
}

func (m msgServer) Pay(ctx context.Context, req *v1.MsgPay) (*v1.MsgPayResponse, error) {
	if err := m.k.Pay(ctx, req); err != nil {
		return nil, err
	}
	eventManager := m.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		"pay",
		event.Attribute{Key: "signer", Value: req.Signer},
		event.Attribute{Key: "channel_id", Value: req.ChannelId},
		event.Attribute{Key: "amount", Value: req.Amount.String()},
	); err != nil {
		return nil, err
	}

	return &v1.MsgPayResponse{}, nil
}
