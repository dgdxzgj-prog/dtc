package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	"dtc/x/identity/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	for _, elem := range genState.DidDocumentMap {
		if err := k.DidDocument.Set(ctx, elem.Did, elem); err != nil {
			return err
		}
	}

	return k.Params.Set(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	genesis := types.DefaultGenesis()

	// 如果 Params 不存在，使用默认值（已经在 DefaultGenesis 中设置）
	params, err := k.Params.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	if err == nil {
		genesis.Params = params
	}

	if err := k.DidDocument.Walk(ctx, nil, func(_ string, val types.DidDocument) (stop bool, err error) {
		genesis.DidDocumentMap = append(genesis.DidDocumentMap, val)
		return false, nil
	}); err != nil {
		return nil, err
	}

	return genesis, nil
}
