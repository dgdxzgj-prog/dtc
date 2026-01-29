package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"dtc/x/identity/keeper"
	module "dtc/x/identity/module"
	"dtc/x/identity/types"
)

type fixture struct {
	ctx          context.Context
	keeper       keeper.Keeper
	addressCodec address.Codec
}

func initFixture(t *testing.T) *fixture {
	t.Helper()

	// 初始化 SDK 配置以支持 dtc 前缀
	// 如果配置还没有设置，则设置它（app/config.go 的 init() 可能已经设置了）
	config := sdk.GetConfig()
	currentPrefix := config.GetBech32AccountAddrPrefix()
	if currentPrefix != "dtc" {
		// 尝试设置配置，如果已经 seal 会 panic，但我们捕获它
		func() {
			defer func() {
				_ = recover() // 忽略 panic，说明配置已经 seal
			}()
			config.SetBech32PrefixForAccount("dtc", "dtcpub")
			config.SetBech32PrefixForValidator("dtcvaloper", "dtcvaloperpub")
			config.SetBech32PrefixForConsensusNode("dtcvalcons", "dtcvalconspub")
			config.Seal()
		}()
	}

	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	authority := authtypes.NewModuleAddress(types.GovModuleName)

	k := keeper.NewKeeper(
		storeService,
		encCfg.Codec,
		addressCodec,
		authority,
	)

	// Initialize params
	if err := k.Params.Set(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	return &fixture{
		ctx:          ctx,
		keeper:       k,
		addressCodec: addressCodec,
	}
}
