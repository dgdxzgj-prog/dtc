package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"dtc/x/identity/keeper"
	"dtc/x/identity/types"
)

func TestDidDocumentMsgServerCreate(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)
	creator, err := f.addressCodec.BytesToString([]byte("signerAddr__________________"))
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		expected := &types.MsgCreateDidDocument{Creator: creator,
			Did:       strconv.Itoa(i),
			Signature: []byte("7369676e6174757265"), // 集成测试签名，跳过验证
		}
		_, err := srv.CreateDidDocument(f.ctx, expected)
		require.NoError(t, err)
		rst, err := f.keeper.DidDocument.Get(f.ctx, expected.Did)
		require.NoError(t, err)
		require.Equal(t, expected.Did, rst.Did)
	}
}

func TestDidDocumentMsgServerUpdate(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte("signerAddr__________________"))
	require.NoError(t, err)
	testSig := []byte("7369676e6174757265")

	unauthorizedAddr, err := f.addressCodec.BytesToString([]byte("unauthorizedAddr___________"))
	require.NoError(t, err)

	expected := &types.MsgCreateDidDocument{Creator: creator,
		Did:       strconv.Itoa(0),
		Signature: testSig, // 集成测试签名，跳过验证
	}
	_, err = srv.CreateDidDocument(f.ctx, expected)
	require.NoError(t, err)

	tests := []struct {
		desc    string
		request *types.MsgUpdateDidDocument
		err     error
	}{
		{
			desc: "invalid address",
			request: &types.MsgUpdateDidDocument{Creator: "invalid",
				Did: strconv.Itoa(0),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			desc: "unauthorized",
			request: &types.MsgUpdateDidDocument{Creator: unauthorizedAddr,
				Did: strconv.Itoa(0),
			},
			err: sdkerrors.ErrUnauthorized,
		},
		{
			desc: "key not found",
			request: &types.MsgUpdateDidDocument{Creator: creator,
				Did: strconv.Itoa(100000),
			},
			err: sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "completed",
			request: &types.MsgUpdateDidDocument{Creator: creator,
				Did: strconv.Itoa(0),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			_, err = srv.UpdateDidDocument(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				rst, err := f.keeper.DidDocument.Get(f.ctx, expected.Did)
				require.NoError(t, err)
				require.Equal(t, expected.Did, rst.Did)
			}
		})
	}
}

func TestDidDocumentMsgServerDelete(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte("signerAddr__________________"))
	require.NoError(t, err)
	testSig := []byte("7369676e6174757265")

	unauthorizedAddr, err := f.addressCodec.BytesToString([]byte("unauthorizedAddr___________"))
	require.NoError(t, err)

	_, err = srv.CreateDidDocument(f.ctx, &types.MsgCreateDidDocument{Creator: creator,
		Did:       strconv.Itoa(0),
		Signature: testSig,
	})
	require.NoError(t, err)

	tests := []struct {
		desc    string
		request *types.MsgDeleteDidDocument
		err     error
	}{
		{
			desc: "invalid address",
			request: &types.MsgDeleteDidDocument{Creator: "invalid",
				Did: strconv.Itoa(0),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			desc: "unauthorized",
			request: &types.MsgDeleteDidDocument{Creator: unauthorizedAddr,
				Did: strconv.Itoa(0),
			},
			err: sdkerrors.ErrUnauthorized,
		},
		{
			desc: "key not found",
			request: &types.MsgDeleteDidDocument{Creator: creator,
				Did: strconv.Itoa(100000),
			},
			err: sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "completed",
			request: &types.MsgDeleteDidDocument{Creator: creator,
				Did: strconv.Itoa(0),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			_, err = srv.DeleteDidDocument(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				found, err := f.keeper.DidDocument.Has(f.ctx, tc.request.Did)
				require.NoError(t, err)
				require.False(t, found)
			}
		})
	}
}

// TestCreateDidDocument_WithSignature 测试使用真实签名创建 DID 文档
func TestCreateDidDocument_WithSignature(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	// 测试参数
	controller := "dtc1m4g397xg68ja9ufal99tk84a9wfrjt6vpaueex"
	did := "did:dtc:4577a4061029e1e2c053a27cafbcc83a"
	faceHash := "4577a4061029e1e2c053a27cafbcc83a"

	// TODO: 替换为实际的 admin 私钥（hex 编码，64 个十六进制字符，即 32 字节）
	// 占位符：请将下面的私钥替换为实际的 admin 私钥
	adminPrivKeyHex := "da22b1840dbce304ed6b3e46da143e1f15d9e3012dd31446b0277af6c409cd57"

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

	// 构造待签名数据：Did + Controller + FaceHash
	data := did + controller + faceHash

	// 使用私钥对原始数据进行签名
	// 注意：Sign 方法会自动对输入进行 SHA256 哈希
	// 验证时我们会手动进行 SHA256 哈希，然后传给 VerifySignature（它不会再哈希）
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
	msg := &types.MsgCreateDidDocument{
		Creator:    controller,
		Did:        did,
		Controller: controller,
		FaceHash:   faceHash,
		Pubkeys:    "",
		Signature:  signature,
	}

	// 执行 CreateDidDocument
	_, err = srv.CreateDidDocument(f.ctx, msg)
	require.NoError(t, err, "CreateDidDocument should succeed")

	// 验证 DID 文档已创建
	didDoc, err := f.keeper.DidDocument.Get(f.ctx, did)
	require.NoError(t, err, "should be able to get created DID document")
	require.Equal(t, did, didDoc.Did, "DID should match")
	require.Equal(t, controller, didDoc.Controller, "Controller should match")
	require.Equal(t, faceHash, didDoc.FaceHash, "FaceHash should match")

	// 验证 FaceHashToIndex 索引已创建
	indexedDid, err := f.keeper.FaceHashToIndex.Get(f.ctx, faceHash)
	require.NoError(t, err, "should be able to get indexed DID")
	require.Equal(t, did, indexedDid, "indexed DID should match")
}
