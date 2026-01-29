package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "credit"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	// It should be synced with the gov module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/gov/types/keys.go#L9
	GovModuleName = "gov"

	// GBDPPoolModuleName 是接收铸币分流 1% 的模块账户名（GBDP 资金池）
	GBDPPoolModuleName = "gbdp_pool"
)

// ParamsKey is the prefix to retrieve all Params
var ParamsKey = collections.NewPrefix("p_credit")

// CreditAccountLiabilityPrefix 按地址存储负债
var CreditAccountLiabilityPrefix = collections.NewPrefix("ca_liability_")

// CreditAccountLastMintHeightPrefix 按地址存储最近铸币高度
var CreditAccountLastMintHeightPrefix = collections.NewPrefix("ca_last_mint_")

// CreditAccountBirthHeightPrefix 按地址存储账户创建高度（用于计算年龄）
var CreditAccountBirthHeightPrefix = collections.NewPrefix("ca_birth_")
