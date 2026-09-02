package keeper

import (
	"context"

	"cosmossdk.io/core/event"
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/utils"
	"git.vdb.to/cerc-io/laconicd/x/onboarding"
)

type msgServer struct {
	k Keeper
}

var _ onboarding.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper *Keeper) onboarding.MsgServer {
	return &msgServer{k: *keeper}
}

func (ms msgServer) OnboardParticipant(ctx context.Context, msg *onboarding.MsgOnboardParticipant) (*onboarding.MsgOnboardParticipantResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx, logGas := utils.WithCustomGasConfig(ctx, ms.k.gasService)
	defer logGas(ms.k.logger, "OnboardParticipant")

	addrCodec := utils.NewAddressCodec()
	signerAddress, err := addrCodec.StringToBytes(msg.Participant)
	if err != nil {
		return nil, err
	}

	_, err = ms.k.OnboardParticipant(ctx, msg, signerAddress)
	if err != nil {
		return nil, err
	}

	if err := ms.k.eventService.EventManager(ctx).EmitKV(
		ctx,
		onboarding.EventTypeOnboardParticipant,
		event.Attribute{Key: onboarding.AttributeKeySigner, Value: msg.Participant},
		event.Attribute{Key: onboarding.AttributeKeyEthAddress, Value: msg.EthPayload.Address},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", onboarding.EventTypeOnboardParticipant)
	}
	if err := ms.k.eventService.EventManager(ctx).EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: onboarding.AttributeValueCategory},
		event.Attribute{Key: onboarding.AttributeKeySigner, Value: msg.Participant},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &onboarding.MsgOnboardParticipantResponse{}, nil
}
