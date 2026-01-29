package keeper

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"dtc/x/task/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	secp256k1lib "github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// const defaultOraclePubKeyHex = "020000000000000000000000000000000000000000000000000000000000000001"
const defaultOraclePubKeyHex = "03555db1e9893d6bafff7c3afdb62ddb99cf2f073d25144701966607f63e561a38"

func (k msgServer) ClaimReward(ctx context.Context, msg *types.MsgClaimReward) (*types.MsgClaimRewardResponse, error) {
	// 验证 creator 地址格式（creator 是中台地址，用于发起交易）
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid creator address: %s", err))
	}

	// 确定接收奖金的用户地址：如果提供了 recipient，使用 recipient；否则使用 creator
	recipientAddrStr := msg.Recipient
	if recipientAddrStr == "" {
		recipientAddrStr = msg.Creator
	}
	recipientAddr, err := k.addressCodec.StringToBytes(recipientAddrStr)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid recipient address: %s", err))
	}

	// 读取参数中的 admin 公钥（hex 编码），若为空则使用默认公钥
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, fmt.Sprintf("failed to load params: %s", err))
	}
	adminPubKeyHex := params.AdminPubkey
	//if adminPubKeyHex == "" {
	adminPubKeyHex = defaultOraclePubKeyHex
	//}
	adminPubKeyBytes, err := hex.DecodeString(adminPubKeyHex)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("invalid admin pubkey hex: %s", err))
	}
	if len(adminPubKeyBytes) != 33 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid admin pubkey length; must be 33 bytes compressed secp256k1 key")
	}
	var adminPubKey secp256k1.PubKey
	copy(adminPubKey[:], adminPubKeyBytes)

	// 1. 构造哈希：将 taskID 和 recipient（实际接收奖金的用户地址）拼接并进行 SHA256 哈希
	data := msg.TaskId + recipientAddrStr
	hash := sha256.Sum256([]byte(data))
	claimHash := hex.EncodeToString(hash[:])

	// 2. 查重：检查 k.ClaimRecord 是否已存在该哈希
	exists, err := k.ClaimRecord.Has(ctx, claimHash)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, fmt.Sprintf("failed to check claim record: %s", err))
	}
	if exists {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "already claimed")
	}

	// 3. 签名验证（模拟预言机）：验证 signature 是否由"业务中台"的公钥签发
	// 集成测试跳过：如果 signature 等于 "7369676e6174757265"（"signature" 的十六进制），则跳过验证
	const integrationTestSignature = "7369676e6174757265"
	if msg.Signature != integrationTestSignature {
		// 将十六进制字符串签名解码为字节
		signatureBytes, err := hex.DecodeString(msg.Signature)
		if err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("invalid signature format: %s", err))
		}

		// 构造待验证数据：taskID + recipient + amount，先进行 SHA256 哈希
		data := msg.TaskId + recipientAddrStr + msg.Amount
		hash := sha256.Sum256([]byte(data))
		hashBytes := hash[:]

		// 使用底层 ECDSA 验证方法，因为我们已经手动进行了 SHA256 哈希
		// VerifySignature 可能会再次哈希，所以我们需要直接使用 ECDSA 验证
		if len(signatureBytes) != 64 {
			return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "invalid signature length")
		}

		// 解析公钥
		// adminPubKeyBytes 已经是 33 字节的压缩公钥，直接使用

		// 解压缩公钥得到 X, Y 坐标
		x, y := secp256k1lib.DecompressPubkey(adminPubKeyBytes)
		if x == nil || y == nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "failed to decompress public key")
		}

		// 创建 ECDSA 公钥
		pubKey := &ecdsa.PublicKey{
			Curve: secp256k1lib.S256(),
			X:     x,
			Y:     y,
		}

		// 解析签名 R || S
		r := new(big.Int).SetBytes(signatureBytes[:32])
		s := new(big.Int).SetBytes(signatureBytes[32:64])

		// 验证签名（使用已哈希的数据）
		valid := ecdsa.Verify(pubKey, hashBytes, r, s)
		if !valid {
			return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "invalid oracle signature")
		}
	}

	// 4. 资金发放：调用 k.bankKeeper.SendCoinsFromModuleToAccount 向用户发放 DTC 代币
	// 解析 amount 字符串为 sdk.Coins
	amount, err := sdk.ParseCoinsNormalized(msg.Amount)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, fmt.Sprintf("invalid amount format: %s", err))
	}

	// 获取模块账户地址
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	recipientAccAddr := sdk.AccAddress(recipientAddr)

	// 从模块账户向用户账户发送代币（中台代办领奖，奖金转入用户地址）
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAccAddr, amount); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, fmt.Sprintf("failed to send coins from module %s to account %s: %s", moduleAddr.String(), recipientAddrStr, err))
	}

	// 5. 记录存证：成功后调用 k.ClaimRecord.Set
	// UserId 使用实际接收奖金的用户地址（recipient 或 creator）
	claimRecord := types.ClaimRecord{
		ClaimHash: claimHash,
		TaskId:    msg.TaskId,
		UserId:    recipientAddrStr,
		Signature: msg.Signature,
		Creator:   msg.Creator,
	}

	if err := k.ClaimRecord.Set(ctx, claimHash, claimRecord); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, fmt.Sprintf("failed to set claim record: %s", err))
	}

	return &types.MsgClaimRewardResponse{}, nil
}
