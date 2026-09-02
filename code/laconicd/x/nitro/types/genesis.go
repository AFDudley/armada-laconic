package types

// TODO
type GenesisState struct{}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return "" }
func (*GenesisState) ProtoMessage()    {}

func (GenesisState) Validate() error { return nil }

func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}
