package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	onboardingtypes "git.vdb.to/cerc-io/laconicd/x/onboarding"
)

var _ onboardingtypes.QueryServer = queryServer{}

type queryServer struct {
	k *Keeper
}

// NewQueryServerImpl returns an implementation of the module QueryServer.
func NewQueryServerImpl(k *Keeper) onboardingtypes.QueryServer {
	return queryServer{k}
}

// Participants implements Participants.QueryServer.
func (qs queryServer) Participants(
	ctx context.Context,
	_ *onboardingtypes.QueryParticipantsRequest,
) (*onboardingtypes.QueryParticipantsResponse, error) {
	resp, err := qs.k.ListParticipants(ctx)
	if err != nil {
		return nil, err
	}

	return &onboardingtypes.QueryParticipantsResponse{Participants: resp}, nil
}

// GetParticipantByAddress implements the GetParticipantByAddress query.
func (qs queryServer) GetParticipantByAddress(
	ctx context.Context,
	req *onboardingtypes.QueryGetParticipantByAddressRequest,
) (*onboardingtypes.QueryGetParticipantByAddressResponse, error) {
	if req.Address == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cosmos (laconic) address is required")
	}

	participant, err := qs.k.GetParticipantByAddress(ctx, req.Address)
	if err != nil {
		return nil, err
	}
	if participant == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "participant with given address not found")
	}
	return &onboardingtypes.QueryGetParticipantByAddressResponse{Participant: participant}, nil
}

// GetParticipantByNitroAddress implements the GetParticipantByNitroAddress query.
func (qs queryServer) GetParticipantByNitroAddress(
	ctx context.Context,
	req *onboardingtypes.QueryGetParticipantByNitroAddressRequest,
) (*onboardingtypes.QueryGetParticipantByNitroAddressResponse, error) {
	if req.NitroAddress == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "nitro address is required")
	}

	participant, err := qs.k.GetParticipantByNitroAddress(ctx, req.NitroAddress)
	if err != nil {
		return nil, err
	}
	if participant == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "participant with given Nitro address not found")
	}
	return &onboardingtypes.QueryGetParticipantByNitroAddressResponse{Participant: participant}, nil
}
