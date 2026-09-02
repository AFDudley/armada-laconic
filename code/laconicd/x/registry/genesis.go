package registry

// DefaultGenesisState sets default evm genesis state with empty accounts and default params and
// chain config values.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		Records:     []Record{},
		Authorities: []AuthorityEntry{},
		Names:       []NameEntry{},
	}
}

func NewGenesisState(params Params, records []Record, authorities []AuthorityEntry, names []NameEntry) GenesisState {
	return GenesisState{
		Params:      params,
		Records:     records,
		Authorities: authorities,
		Names:       names,
	}
}

// Validate performs basic genesis state validation returning an error upon any
func (gs *GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return nil
}
