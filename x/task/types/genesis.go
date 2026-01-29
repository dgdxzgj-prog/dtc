package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:         DefaultParams(),
		ClaimRecordMap: []ClaimRecord{}}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	claimRecordIndexMap := make(map[string]struct{})

	for _, elem := range gs.ClaimRecordMap {
		index := fmt.Sprint(elem.ClaimHash)
		if _, ok := claimRecordIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for claimRecord")
		}
		claimRecordIndexMap[index] = struct{}{}
	}

	return gs.Params.Validate()
}
