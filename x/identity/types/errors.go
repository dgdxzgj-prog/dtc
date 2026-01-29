package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/identity module sentinel errors
var (
	ErrInvalidSigner     = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrDuplicateFaceHash = errors.Register(ModuleName, 1101, "face hash already registered")
)
