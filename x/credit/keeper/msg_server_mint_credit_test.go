package keeper_test

import (
	"context"
	"sync"
	"testing"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	identitytypes "dtc/x/identity/types"

	"dtc/x/credit/keeper"
	module "dtc/x/credit/module"
	"dtc/x/credit/types"
)

// mockIdentityKeeper 用于测试：始终返回已注册（found=true）
type mockIdentityKeeper struct{}

func (mockIdentityKeeper) GetDidDocument(ctx sdk.Context, address string) (identitytypes.DidDocument, bool) {
	return identitytypes.DidDocument{Controller: address}, true
}

// mintCreditBankKeeper 是一个可以跟踪铸币和转账的 mock BankKeeper
type mintCreditBankKeeper struct {
	mu              sync.Mutex
	accountBalances map[string]sdk.Coins
	moduleBalances  map[string]sdk.Coins
	mintCalls       []mintCall
	sendCalls       []sendCall
}

type mintCall struct {
	moduleName string
	amount     sdk.Coins
}

type sendCall struct {
	senderModule string
	recipient    string
	amount       sdk.Coins
}

func newMintCreditBankKeeper() *mintCreditBankKeeper {
	return &mintCreditBankKeeper{
		accountBalances: make(map[string]sdk.Coins),
		moduleBalances:  make(map[string]sdk.Coins),
		mintCalls:       make([]mintCall, 0),
		sendCalls:       make([]sendCall, 0),
	}
}

func (m *mintCreditBankKeeper) SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	m.mu.Lock()
	defer m.mu.Unlock()
	if coins, ok := m.accountBalances[addr.String()]; ok {
		return coins
	}
	return sdk.NewCoins()
}

func (m *mintCreditBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mintCalls = append(m.mintCalls, mintCall{
		moduleName: moduleName,
		amount:     amt,
	})

	// 增加模块账户余额
	currentBalance := m.moduleBalances[moduleName]
	newBalance := currentBalance.Add(amt...)
	m.moduleBalances[moduleName] = newBalance

	return nil
}

func (m *mintCreditBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sendCalls = append(m.sendCalls, sendCall{
		senderModule: senderModule,
		recipient:    recipientAddr.String(),
		amount:       amt,
	})

	// 从模块账户扣除
	if currentBalance, ok := m.moduleBalances[senderModule]; ok {
		newModuleBalance := currentBalance.Sub(amt...)
		m.moduleBalances[senderModule] = newModuleBalance
	}

	// 增加接收账户余额
	currentBalance := m.accountBalances[recipientAddr.String()]
	newBalance := currentBalance.Add(amt...)
	m.accountBalances[recipientAddr.String()] = newBalance

	return nil
}

func (m *mintCreditBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	// 未在测试中使用，但需要实现接口
	return nil
}

func (m *mintCreditBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	// 未在测试中使用，但需要实现接口
	return nil
}

func (m *mintCreditBankKeeper) GetAccountBalance(addr sdk.AccAddress) sdk.Coins {
	m.mu.Lock()
	defer m.mu.Unlock()
	if coins, ok := m.accountBalances[addr.String()]; ok {
		return coins
	}
	return sdk.NewCoins()
}

func (m *mintCreditBankKeeper) GetModuleBalance(moduleName string) sdk.Coins {
	m.mu.Lock()
	defer m.mu.Unlock()
	if coins, ok := m.moduleBalances[moduleName]; ok {
		return coins
	}
	return sdk.NewCoins()
}

func (m *mintCreditBankKeeper) GetMintCalls() []mintCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]mintCall, len(m.mintCalls))
	copy(result, m.mintCalls)
	return result
}

func (m *mintCreditBankKeeper) GetSendCalls() []sendCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]sendCall, len(m.sendCalls))
	copy(result, m.sendCalls)
	return result
}

// mintCreditFixture 用于 MintCredit 测试的 fixture
type mintCreditFixture struct {
	ctx          context.Context
	sdkCtx       sdk.Context
	keeper       keeper.Keeper
	addressCodec address.Codec
	bankKeeper   *mintCreditBankKeeper
}

func initMintCreditFixture(t *testing.T) *mintCreditFixture {
	t.Helper()

	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	storeService := runtime.NewKVStoreService(storeKey)
	testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))
	// 设置 BlockHeight 以便测试能正确获取高度
	sdkCtx := testCtx.Ctx.WithBlockHeight(100)
	// sdk.Context 实现了 context.Context，可以直接使用
	ctx := sdkCtx

	authority := authtypes.NewModuleAddress(types.GovModuleName)

	// 创建可跟踪铸币和转账的 bankKeeper
	bankKeeper := newMintCreditBankKeeper()

	k := keeper.NewKeeper(
		storeService,
		encCfg.Codec,
		addressCodec,
		authority,
		bankKeeper,
		nil,
		mockIdentityKeeper{}, // 测试中视为已注册
	)

	// Initialize params with default gbdp_rate = 100 (1%)
	params := types.DefaultParams()
	params.GbdpRate = 100 // 默认 1%
	if err := k.Params.Set(ctx, params); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	return &mintCreditFixture{
		ctx:          ctx,
		sdkCtx:       sdkCtx,
		keeper:       k,
		addressCodec: addressCodec,
		bankKeeper:   bankKeeper,
	}
}

// TestMintCredit_Success 测试铸币成功的场景
func TestMintCredit_Success(t *testing.T) {
	f := initMintCreditFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	// 创建测试账户
	creatorBytes := []byte("testCreator________________")
	creator, err := f.addressCodec.BytesToString(creatorBytes)
	require.NoError(t, err)

	creatorAddr, err := f.addressCodec.StringToBytes(creator)
	require.NoError(t, err)
	creatorAccAddr := sdk.AccAddress(creatorAddr)

	// 获取 GBDP 池地址
	gbdpPoolAddr := authtypes.NewModuleAddress(types.GBDPPoolModuleName)

	// 验证初始状态
	initialCreatorBalance := f.bankKeeper.GetAccountBalance(creatorAccAddr)
	require.True(t, initialCreatorBalance.IsZero(), "用户初始余额应该为零")

	initialGBDPBalance := f.bankKeeper.GetAccountBalance(gbdpPoolAddr)
	require.True(t, initialGBDPBalance.IsZero(), "GBDP 池初始余额应该为零")

	// 创建铸币消息
	msg := &types.MsgMintCredit{
		Creator: creator,
	}

	// 执行 MintCredit
	_, err = srv.MintCredit(f.ctx, msg)
	require.NoError(t, err, "MintCredit 应该成功")

	// 验证铸币调用
	mintCalls := f.bankKeeper.GetMintCalls()
	require.Len(t, mintCalls, 1, "应该有一次铸币调用")
	require.Equal(t, types.ModuleName, mintCalls[0].moduleName, "应该铸造到 credit 模块")
	expectedTotalCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))
	require.Equal(t, expectedTotalCoins, mintCalls[0].amount, "应该铸造总量 1000000 udtc")

	// 验证转账调用
	sendCalls := f.bankKeeper.GetSendCalls()
	require.Len(t, sendCalls, 2, "应该有两次转账调用：一次给用户，一次给 GBDP 池")

	// 找到给用户的转账
	var creatorSendCall *sendCall
	var gbdpSendCall *sendCall
	for i := range sendCalls {
		if sendCalls[i].recipient == creatorAccAddr.String() {
			creatorSendCall = &sendCalls[i]
		}
		if sendCalls[i].recipient == gbdpPoolAddr.String() {
			gbdpSendCall = &sendCalls[i]
		}
	}

	require.NotNil(t, creatorSendCall, "应该有一次转账给用户")
	require.NotNil(t, gbdpSendCall, "应该有一次转账给 GBDP 池")

	// 验证用户收到的金额：总量的 99% (1000000 * 99 / 100 = 990000)
	expectedCreatorAmount := int64(990000) // 99% of 1000000
	creatorCoins := creatorSendCall.amount
	require.Equal(t, sdk.NewInt64Coin(sdk.DefaultBondDenom, expectedCreatorAmount), creatorCoins[0], "用户应该收到总量的 99%")

	// 验证 GBDP 池收到的金额：总量的 1% (1000000 * 1 / 100 = 10000)
	expectedGBDPAmount := int64(10000) // 1% of 1000000
	gbdpCoins := gbdpSendCall.amount
	require.Equal(t, sdk.NewInt64Coin(sdk.DefaultBondDenom, expectedGBDPAmount), gbdpCoins[0], "GBDP 池应该收到总量的 1%")

	// 验证用户余额增加了总量的 99%
	finalCreatorBalance := f.bankKeeper.GetAccountBalance(creatorAccAddr)
	expectedCreatorBalance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, expectedCreatorAmount))
	require.Equal(t, expectedCreatorBalance, finalCreatorBalance, "用户余额应该增加总量的 99%")

	// 验证 GBDP 池余额增加了总量的 1%
	finalGBDPBalance := f.bankKeeper.GetAccountBalance(gbdpPoolAddr)
	expectedGBDPBalance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, expectedGBDPAmount))
	require.Equal(t, expectedGBDPBalance, finalGBDPBalance, "GBDP 池余额应该增加总量的 1%")

	// 验证 CreditAccount.Liability 等于总量的 100%
	liability, err := f.keeper.CreditAccountLiability.Get(f.ctx, creator)
	require.NoError(t, err, "应该能获取用户的负债记录")
	expectedLiability := uint64(1000000) // 总量的 100%
	require.Equal(t, expectedLiability, liability, "用户的负债应该等于总量的 100%")

	// 验证 BirthHeight 已设置（新账户）
	birthHeight, err := f.keeper.CreditAccountBirthHeight.Get(f.ctx, creator)
	require.NoError(t, err, "应该能获取用户的出生高度")
	require.Greater(t, birthHeight, uint64(0), "出生高度应该大于 0")

	// 验证 LastMintHeight 已设置
	lastMintHeight, err := f.keeper.CreditAccountLastMintHeight.Get(f.ctx, creator)
	require.NoError(t, err, "应该能获取用户的最近铸币高度")
	require.Greater(t, lastMintHeight, uint64(0), "最近铸币高度应该大于 0")
}
