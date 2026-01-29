package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	"dtc/x/dtc/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
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

	return genesis, nil
}
