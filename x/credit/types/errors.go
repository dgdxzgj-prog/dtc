package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/credit module sentinel errors
var (
	ErrInvalidSigner           = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrIdentityNotRegistered   = errors.Register(ModuleName, 1101, "creator address is not registered (no DID document)")
	ErrMintTooFrequent         = errors.Register(ModuleName, 1102, "mint is too frequent; must wait at least one month since last mint")
)
