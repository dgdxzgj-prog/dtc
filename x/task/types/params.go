package types

import "encoding/hex"

const defaultAdminPubKeyHex = "03555db1e9893d6bafff7c3afdb62ddb99cf2f073d25144701966607f63e561a38"

// NewParams creates a new Params instance.
func NewParams() Params {
	return Params{
		AdminPubkey: defaultAdminPubKeyHex,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams()
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if p.AdminPubkey == "" {
		return nil
	}
	if _, err := hex.DecodeString(p.AdminPubkey); err != nil {
		return err
	}
	return nil
}
