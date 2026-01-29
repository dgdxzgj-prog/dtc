package types

// CreditAccount 表示某地址的信用账户：负债与最近铸币高度。
type CreditAccount struct {
	Liability      uint64 // 该地址的负债（与铸币总量同单位，如 1 DTC = 1000000）
	LastMintHeight uint64 // 最近一次铸币时的区块高度
}
