package auction

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{}
}

// Validate a set of params.
func (p Params) Validate() error {
	return nil
}
