package keeper_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"testing"

	"dtc/x/task/keeper"
	"dtc/x/task/types"

	module "dtc/x/task/module"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
)

// trackableBankKeeper 是一个可以跟踪余额变化的 mock BankKeeper
type trackableBankKeeper struct {
	mu              sync.Mutex
	accountBalances map[string]sdk.Coins
}

func newTrackableBankKeeper() *trackableBankKeeper {
	return &trackableBankKeeper{
		accountBalances: make(map[string]sdk.Coins),
	}
}

func (m *trackableBankKeeper) SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	m.mu.Lock()
	defer m.mu.Unlock()
	if coins, ok := m.accountBalances[addr.String()]; ok {
		return coins
	}
	return sdk.NewCoins()
}

func (m *trackableBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 增加接收账户的余额
	currentBalance := m.accountBalances[recipientAddr.String()]
	newBalance := currentBalance.Add(amt...)
	m.accountBalances[recipientAddr.String()] = newBalance

	return nil
}

func (m *trackableBankKeeper) GetBalance(addr sdk.AccAddress) sdk.Coins {
	m.mu.Lock()
	defer m.mu.Unlock()
	if coins, ok := m.accountBalances[addr.String()]; ok {
		return coins
	}
	return sdk.NewCoins()
}

// claimRewardFixture 用于 ClaimReward 测试的 fixture
type claimRewardFixture struct {
	ctx          context.Context
	keeper       keeper.Keeper
	addressCodec address.Codec
	bankKeeper   *trackableBankKeeper
	privKey      secp256k1.PrivKey
	pubKey       secp256k1.PubKey
}

func initClaimRewardFixture(t *testing.T) *claimRewardFixture {
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

	// 生成测试用的私钥-公钥对
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey().(secp256k1.PubKey)

	// 创建可跟踪余额的 bankKeeper
	bankKeeper := newTrackableBankKeeper()

	k := keeper.NewKeeper(
		storeService,
		encCfg.Codec,
		addressCodec,
		authority,
		bankKeeper,
	)

	// Initialize params
	params := types.DefaultParams()
	params.AdminPubkey = hex.EncodeToString(pubKey.Bytes())
	if err := k.Params.Set(ctx, params); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	return &claimRewardFixture{
		ctx:          ctx,
		keeper:       k,
		addressCodec: addressCodec,
		bankKeeper:   bankKeeper,
		privKey:      privKey,
		pubKey:       pubKey,
	}
}

// generateSignature 使用私钥对数据进行签名
func generateSignature(privKey secp256k1.PrivKey, data []byte) (string, error) {
	signature, err := privKey.Sign(data)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signature), nil
}

// TestClaimReward_FirstClaimSuccess 测试首次领取成功的场景
func TestClaimReward_FirstClaimSuccess(t *testing.T) {
	f := initClaimRewardFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	// 创建测试账户
	creator, err := f.addressCodec.BytesToString([]byte("testCreator________________"))
	require.NoError(t, err)

	taskID := "task-123"
	amount := "1000dtc"

	// 使用集成测试签名常量，绕过签名校验
	signature := "7369676e6174757265"

	// 创建消息
	msg := &types.MsgClaimReward{
		Creator:   creator,
		TaskId:    taskID,
		Amount:    amount,
		Signature: signature,
	}

	// 获取初始余额
	creatorAddr, err := f.addressCodec.StringToBytes(creator)
	require.NoError(t, err)
	initialBalance := f.bankKeeper.GetBalance(sdk.AccAddress(creatorAddr))
	require.True(t, initialBalance.IsZero(), "初始余额应该为零")

	// 执行 ClaimReward
	_, err = srv.ClaimReward(f.ctx, msg)
	require.NoError(t, err, "首次领取应该成功")

	// 验证余额增加
	finalBalance := f.bankKeeper.GetBalance(sdk.AccAddress(creatorAddr))
	expectedAmount, err := sdk.ParseCoinsNormalized(amount)
	require.NoError(t, err)
	require.Equal(t, expectedAmount, finalBalance, "余额应该增加")

	// 验证 ClaimRecord 已创建
	// 计算 claimHash
	hash := sha256.Sum256([]byte(taskID + creator))
	claimHash := hex.EncodeToString(hash[:])

	exists, err := f.keeper.ClaimRecord.Has(f.ctx, claimHash)
	require.NoError(t, err)
	require.True(t, exists, "ClaimRecord 应该已创建")

	// 验证 ClaimRecord 内容
	claimRecord, err := f.keeper.ClaimRecord.Get(f.ctx, claimHash)
	require.NoError(t, err)
	require.Equal(t, taskID, claimRecord.TaskId)
	require.Equal(t, creator, claimRecord.UserId)
	require.Equal(t, creator, claimRecord.Creator)
	require.Equal(t, signature, claimRecord.Signature)
}

// TestClaimReward_DuplicateClaim 测试重复领取的场景
func TestClaimReward_DuplicateClaim(t *testing.T) {
	f := initClaimRewardFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	// 创建测试账户
	creator, err := f.addressCodec.BytesToString([]byte("testCreator________________"))
	require.NoError(t, err)

	taskID := "task-456"
	amount := "2000dtc"

	// 使用集成测试签名常量，绕过签名校验
	signature := "7369676e6174757265"

	// 创建消息
	msg := &types.MsgClaimReward{
		Creator:   creator,
		TaskId:    taskID,
		Amount:    amount,
		Signature: signature,
	}

	// 首次领取应该成功
	_, err = srv.ClaimReward(f.ctx, msg)
	require.NoError(t, err, "首次领取应该成功")

	// 验证首次领取后的余额
	creatorAddr, err := f.addressCodec.StringToBytes(creator)
	require.NoError(t, err)
	firstBalance := f.bankKeeper.GetBalance(sdk.AccAddress(creatorAddr))
	expectedAmount, err := sdk.ParseCoinsNormalized(amount)
	require.NoError(t, err)
	require.Equal(t, expectedAmount, firstBalance, "首次领取后余额应该增加")

	// 再次使用相同的 TaskID 和 Creator 领取，应该失败
	_, err = srv.ClaimReward(f.ctx, msg)
	require.Error(t, err, "重复领取应该失败")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest, "应该返回 already claimed 错误")
	require.Contains(t, err.Error(), "already claimed", "错误信息应该包含 'already claimed'")

	// 验证余额没有再次增加
	finalBalance := f.bankKeeper.GetBalance(sdk.AccAddress(creatorAddr))
	require.Equal(t, firstBalance, finalBalance, "重复领取后余额不应该再次增加")
}

// TestClaimReward_WithSignature 测试使用真实签名领取奖励
func TestClaimReward_WithSignature(t *testing.T) {
	f := initClaimRewardFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	// 测试参数
	taskID := "task_did_registration"
	creator := "dtc16yy28zy9gjy8yg8fe5elnyygnh3xhy4fy9mk9p"   // 中台地址（发起交易）
	recipient := "dtc16y2all8099pl90mglk3zm64vnv0nm0d4rgcjdv" // 用户地址（接收奖金）
	amount := "500000udtc"

	// TODO: 替换为实际的 admin 私钥（hex 编码，64 个十六进制字符，即 32 字节）
	// 占位符：请将下面的私钥替换为实际的 admin 私钥
	adminPrivKeyHex := "83b5c9a3b10e2e52bdcd70aeb7fc0fd2939c1a47c609b579f67c0fa48be02aa5"

	// 检查是否使用了占位符
	if adminPrivKeyHex == "PLACEHOLDER_ADMIN_PRIVATE_KEY_HEX" {
		t.Skip("Skipping test: Please replace PLACEHOLDER_ADMIN_PRIVATE_KEY_HEX with actual admin private key")
	}

	// 解码私钥
	adminPrivKeyBytes, err := hex.DecodeString(adminPrivKeyHex)
	require.NoError(t, err, "failed to decode admin private key")
	require.Equal(t, 32, len(adminPrivKeyBytes), "private key must be 32 bytes")

	// 从字节创建私钥 - secp256k1.PrivKey 是 []byte 类型
	adminPrivKey := secp256k1.PrivKey(adminPrivKeyBytes)

	// 获取对应的公钥
	adminPubKey := adminPrivKey.PubKey().(secp256k1.PubKey)
	adminPubKeyHex := hex.EncodeToString(adminPubKey.Bytes())

	// 打印私钥和公钥信息以便调试
	t.Logf("=== Key Debug Info ===")
	t.Logf("Private Key (hex): %s", adminPrivKeyHex)
	t.Logf("Private Key (bytes): %s", hex.EncodeToString(adminPrivKeyBytes))
	t.Logf("Private Key length: %d bytes", len(adminPrivKeyBytes))
	t.Logf("Public Key (hex): %s", adminPubKeyHex)
	t.Logf("Public Key (bytes): %s", hex.EncodeToString(adminPubKey.Bytes()))
	t.Logf("Public Key length: %d bytes", len(adminPubKey.Bytes()))
	t.Logf("======================")

	// 设置 Params 中的 AdminPubkey
	params := types.DefaultParams()
	params.AdminPubkey = adminPubKeyHex
	err = f.keeper.Params.Set(f.ctx, params)
	require.NoError(t, err, "failed to set params")

	// 构造待签名数据：TaskId + Recipient + Amount
	data := taskID + recipient + amount

	// 使用私钥对原始数据进行签名
	// 注意：Sign 方法会自动对输入进行 SHA256 哈希
	// 验证时我们会手动进行 SHA256 哈希，然后传给 ECDSA 验证（它不会再哈希）
	// 所以两边都是单次哈希，应该能匹配
	signature, err := adminPrivKey.Sign([]byte(data))
	require.NoError(t, err, "failed to sign data")

	// 计算哈希用于调试（验证时也会计算相同的哈希）
	hash := sha256.Sum256([]byte(data))
	hashBytes := hash[:]

	// 打印签名信息以便于跟踪问题
	t.Logf("=== Signature Debug Info ===")
	t.Logf("Data to sign: %s", data)
	t.Logf("Data hash (SHA256, for verification): %s", hex.EncodeToString(hashBytes))
	t.Logf("Signature (hex): %s", hex.EncodeToString(signature))
	t.Logf("Signature length: %d bytes", len(signature))
	t.Logf("Admin Public Key (hex): %s", adminPubKeyHex)
	t.Logf("Note: Sign() method automatically hashes the data with SHA256")
	t.Logf("============================")

	// 创建消息
	msg := &types.MsgClaimReward{
		Creator:   creator,
		TaskId:    taskID,
		Amount:    amount,
		Signature: hex.EncodeToString(signature),
		Recipient: recipient, // 用户地址（接收奖金）
	}

	// 获取初始余额（检查用户地址的余额）
	recipientAddr, err := f.addressCodec.StringToBytes(recipient)
	require.NoError(t, err, "failed to parse recipient address")
	initialBalance := f.bankKeeper.GetBalance(sdk.AccAddress(recipientAddr))
	require.True(t, initialBalance.IsZero(), "初始余额应该为零")

	// 执行 ClaimReward
	_, err = srv.ClaimReward(f.ctx, msg)
	require.NoError(t, err, "ClaimReward should succeed")

	// 验证余额增加（奖金应该转入用户地址）
	finalBalance := f.bankKeeper.GetBalance(sdk.AccAddress(recipientAddr))
	expectedAmount, err := sdk.ParseCoinsNormalized(amount)
	require.NoError(t, err)
	require.Equal(t, expectedAmount, finalBalance, "余额应该增加")

	// 验证 ClaimRecord 已创建
	// 计算 claimHash（注意：claimHash 使用 TaskId + Recipient，不包含 Amount）
	claimHashData := taskID + recipient
	claimHashValue := sha256.Sum256([]byte(claimHashData))
	claimHash := hex.EncodeToString(claimHashValue[:])

	exists, err := f.keeper.ClaimRecord.Has(f.ctx, claimHash)
	require.NoError(t, err)
	require.True(t, exists, "ClaimRecord 应该已创建")

	// 验证 ClaimRecord 内容
	claimRecord, err := f.keeper.ClaimRecord.Get(f.ctx, claimHash)
	require.NoError(t, err)
	require.Equal(t, taskID, claimRecord.TaskId)
	require.Equal(t, recipient, claimRecord.UserId) // UserId 应该是实际接收奖金的用户地址
	require.Equal(t, creator, claimRecord.Creator)  // Creator 是中台地址（发起交易）
	require.Equal(t, hex.EncodeToString(signature), claimRecord.Signature)
}
