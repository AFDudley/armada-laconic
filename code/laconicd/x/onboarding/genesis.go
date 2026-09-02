package onboarding

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:       DefaultParams(),
		Participants: []Participant{},
	}
}

func NewGenesisState(params Params, participants []Participant) *GenesisState {
	return &GenesisState{
		Params:       params,
		Participants: participants,
	}
}

// Validate performs basic genesis state validation returning an error upon any
func (gs *GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return nil
}
