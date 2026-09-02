package bond

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		Bonds:  []*Bond{},
	}
}

func NewGenesisState(params Params, bonds []*Bond) *GenesisState {
	return &GenesisState{
		Params: params,
		Bonds:  bonds,
	}
}

// Validate performs basic genesis state validation returning an error upon any
func (gs *GenesisState) Validate() error {
	err := gs.Params.Validate()
	if err != nil {
		return err
	}

	return nil
}
