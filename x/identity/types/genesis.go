package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:         DefaultParams(),
		DidDocumentMap: []DidDocument{}}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	didDocumentIndexMap := make(map[string]struct{})

	for _, elem := range gs.DidDocumentMap {
		index := fmt.Sprint(elem.Did)
		if _, ok := didDocumentIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for didDocument")
		}
		didDocumentIndexMap[index] = struct{}{}
	}

	return gs.Params.Validate()
}
