package keeper

import (
	"context"

	"git.vdb.to/cerc-io/laconicd/x/onboarding"
)

// InitGenesis initializes the module state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *onboarding.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}

	for _, participant := range data.Participants {
		if err := k.Participants.Set(ctx, participant.CosmosAddress, participant); err != nil {
			return err
		}
	}

	return nil
}

// ExportGenesis exports the module state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) (*onboarding.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	var participants []onboarding.Participant
	if err := k.Participants.Walk(ctx, nil, func(cosmosAddress string, participant onboarding.Participant) (bool, error) {
		participants = append(participants, participant)
		return false, nil
	}); err != nil {
		return nil, err
	}

	return &onboarding.GenesisState{
		Params:       params,
		Participants: participants,
	}, nil
}
