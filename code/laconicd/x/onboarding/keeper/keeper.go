package keeper

import (
	"context"
	"encoding/json"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"

	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"git.vdb.to/cerc-io/laconicd/utils"
	"git.vdb.to/cerc-io/laconicd/x/onboarding"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec
	eventService event.Service
	gasService   gas.Service
	logger       log.Logger

	// authority is the address capable of executing a MsgUpdateParams and other authority-gated message.
	// typically, this should be the x/gov module account.
	authority types.AccAddress

	// state management
	Schema       collections.Schema
	Params       collections.Item[onboarding.Params]
	Participants collections.Map[string, onboarding.Participant]
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	addressCodec address.Codec,
	storeService storetypes.KVStoreService,
	eventService event.Service,
	gasService gas.Service,
	authority types.AccAddress,
	logger log.Logger,
) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:          cdc,
		addressCodec: addressCodec,
		eventService: eventService,
		gasService:   gasService,
		logger:       logger.With(log.ModuleKey, "x/"+onboarding.ModuleName),
		authority:    authority,
		Params:       collections.NewItem(sb, onboarding.ParamsPrefix, "params", codec.CollValue[onboarding.Params](cdc)),
		Participants: collections.NewMap(
			sb, onboarding.ParticipantsPrefix, "participants", collections.StringKey, codec.CollValue[onboarding.Participant](cdc),
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	return &k
}

func (k Keeper) OnboardParticipant(
	ctx context.Context,
	msg *onboarding.MsgOnboardParticipant,
	signerAddress sdk.AccAddress,
) (*onboarding.MsgOnboardParticipantResponse, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	if !params.OnboardingEnabled {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Onboarding is disabled")
	}

	message, err := json.Marshal(msg.EthPayload)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid format for payload")
	}

	// Decode eth pubkey from signature. The derived address should be the nitro address of the participant
	nitroPubKey, err := utils.DecodeEthereumPubKey(message, msg.EthSignature)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Failed to decode Ethereum public key")
	}
	nitroAddress, err := utils.EthAddressFromPubKey(nitroPubKey)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Failed to derive Ethereum address from public key")
	}

	if nitroAddress.String() != msg.EthPayload.Address {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Recovered Ethereum address does not match the address set in payload")
	}

	cosmosAddr, err := k.addressCodec.BytesToString(signerAddress)
	if err != nil {
		return nil, err
	}
	participant := &onboarding.Participant{
		CosmosAddress: cosmosAddr,
		PublicKey:     nitroPubKey,
		Role:          msg.Role,
		KycId:         msg.KycId,
	}

	if err := k.StoreParticipant(ctx, participant); err != nil {
		return nil, err
	}

	return &onboarding.MsgOnboardParticipantResponse{}, nil
}

func (k Keeper) StoreParticipant(ctx context.Context, participant *onboarding.Participant) error {
	key := participant.CosmosAddress
	return k.Participants.Set(ctx, key, *participant)
}

// ListParticipants - get all participants.
func (k Keeper) ListParticipants(ctx context.Context) ([]*onboarding.Participant, error) {
	var participants []*onboarding.Participant

	iter, err := k.Participants.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}

	for ; iter.Valid(); iter.Next() {
		participant, err := iter.Value()
		if err != nil {
			return nil, err
		}

		participants = append(participants, &participant)
	}

	return participants, nil
}

// GetParticipantByAddress - get participant by cosmos (laconic) address.
// Returns nil if participant is not found.
func (k Keeper) GetParticipantByAddress(ctx context.Context, address string) (*onboarding.Participant, error) {
	participant, err := k.Participants.Get(ctx, address)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &participant, nil
}

// GetParticipantByNitroAddress - get participant by nitro address.
// Returns nil if participant is not found.
func (k Keeper) GetParticipantByNitroAddress(ctx context.Context, nitroAddress string) (*onboarding.Participant, error) {
	var participant *onboarding.Participant

	err := k.Participants.Walk(ctx, nil, func(key string, value onboarding.Participant) (bool, error) {
		ethAddr, err := utils.EthAddressFromPubKey(value.PublicKey)
		if err != nil {
			return false, err
		}
		if ethAddr.String() == nitroAddress {
			participant = &value
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return participant, nil
}
