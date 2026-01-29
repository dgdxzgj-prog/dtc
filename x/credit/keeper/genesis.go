package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	"dtc/x/credit/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	return k.Params.Set(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	genesis := types.DefaultGenesis()

	params, err := k.Params.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			// Params 未设置时（例如从未执行 InitGenesis）使用默认值
			genesis.Params = types.DefaultParams()
			return genesis, nil
		}
		return nil, err
	}
	genesis.Params = params
	return genesis, nil
}
