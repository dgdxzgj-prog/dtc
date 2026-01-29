#!/bin/bash

set -e

BINARY="dtcd"
CHAIN_ID="dtc"
DENOM="udtc"
KEY_NAME="admin"

echo "=========================================="
echo "初始化 Admin 用户脚本"
echo "=========================================="

# 检查 admin 用户是否已存在
if $BINARY keys show $KEY_NAME &>/dev/null; then
    echo "警告: admin 用户已存在，将删除并重新创建..."
    $BINARY keys delete $KEY_NAME -y 2>/dev/null || true
fi

# 1. 创建 admin 用户
echo -e "\n--- 步骤 1: 创建 admin 用户 ---"
echo "正在创建 admin 用户（使用默认密钥环后端）..."
$BINARY keys add $KEY_NAME --output json > /tmp/admin_key.json

# 从 JSON 中提取信息
ADMIN_ADDRESS=$(cat /tmp/admin_key.json | jq -r '.address')
ADMIN_PUBKEY_JSON=$(cat /tmp/admin_key.json | jq -r '.pubkey')
ADMIN_MNEMONIC=$(cat /tmp/admin_key.json | jq -r '.mnemonic')

# 提取公钥的 hex 格式（用于更新链上参数）
# 公钥格式可能是 JSON 对象 {"@type":"/cosmos.crypto.secp256k1.PubKey","key":"base64..."}
ADMIN_PUBKEY_BASE64=$(echo "$ADMIN_PUBKEY_JSON" | jq -r '.key' 2>/dev/null || echo "")
if [ ! -z "$ADMIN_PUBKEY_BASE64" ] && [ "$ADMIN_PUBKEY_BASE64" != "null" ]; then
    # 从 base64 解码并转换为 hex
    ADMIN_PUBKEY_HEX=$(echo "$ADMIN_PUBKEY_BASE64" | base64 -d 2>/dev/null | xxd -p -c 256 | tr -d '\n' || echo "")
else
    # 如果无法提取，使用默认值（与代码中的默认值匹配）
    ADMIN_PUBKEY_HEX="03555db1e9893d6bafff7c3afdb62ddb99cf2f073d25144701966607f63e561a38"
fi

# 导出私钥（需要从助记词恢复，这里我们使用 keys export 命令）
echo "正在导出私钥..."
$BINARY keys export $KEY_NAME --unarmored-hex --unsafe

echo "✓ Admin 用户创建成功"
echo "  地址: $ADMIN_ADDRESS"
echo "  公钥 (JSON): $ADMIN_PUBKEY_JSON"
echo "  公钥 (Hex): $ADMIN_PUBKEY_HEX"

# 2. 获取 alice 地址
echo -e "\n--- 步骤 2: 获取 Alice 地址 ---"
ALICE_ADDRESS=$($BINARY keys show alice -a)
echo "Alice 地址: $ALICE_ADDRESS"

# 检查 alice 余额
ALICE_BALANCE=$($BINARY q bank balances $ALICE_ADDRESS --output json 2>/dev/null | jq -r ".balances[] | select(.denom==\"$DENOM\") | .amount" || echo "0")
echo "Alice 余额: $ALICE_BALANCE $DENOM"

# 3. 从 alice 转账 5000dtc 给 admin
echo -e "\n--- 步骤 3: 从 Alice 转账 5000dtc 给 Admin ---"
AMOUNT="5000000000$DENOM"  # 5000dtc = 5000000000udtc
FEE="200$DENOM"

echo "转账金额: 5000dtc ($AMOUNT)"
echo "手续费: $FEE"

$BINARY tx bank send $ALICE_ADDRESS $ADMIN_ADDRESS $AMOUNT \
    --from alice \
    --chain-id $CHAIN_ID \
    --gas auto \
    --gas-adjustment 1.5 \
    --fees $FEE \
    -y

echo "等待区块确认..."
sleep 5

# 4. 验证转账结果
echo -e "\n--- 步骤 4: 验证转账结果 ---"
ADMIN_BALANCE=$($BINARY q bank balances $ADMIN_ADDRESS --output json 2>/dev/null | jq -r ".balances[] | select(.denom==\"$DENOM\") | .amount" || echo "0")
echo "Admin 余额: $ADMIN_BALANCE $DENOM"

if [ "$ADMIN_BALANCE" -ge "5000000000" ]; then
    echo "✓ 转账成功！"
else
    echo "✗ 转账可能未完成，请检查"
fi

# 5. 导出用户信息
echo -e "\n=========================================="
echo "Admin 用户信息"
echo "=========================================="
echo "地址 (Address): $ADMIN_ADDRESS"
echo "公钥 (Public Key): $ADMIN_PUBKEY"
if [ "$ADMIN_PRIVKEY" != "N/A (使用助记词恢复)" ]; then
    echo "私钥 (Private Key, Hex): $ADMIN_PRIVKEY"
else
    echo "私钥: 无法直接导出（使用助记词恢复）"
    echo "助记词 (Mnemonic): $ADMIN_MNEMONIC"
    echo ""
    echo "注意: 请妥善保管助记词，可以使用以下命令恢复密钥："
    echo "  $BINARY keys add $KEY_NAME --recover"
fi
echo ""
echo "余额: $ADMIN_BALANCE $DENOM"
echo "=========================================="

# 保存信息到文件
INFO_FILE="admin_info.txt"
cat > $INFO_FILE <<EOF
==========================================
Admin 用户信息
生成时间: $(date)
==========================================
地址 (Address): $ADMIN_ADDRESS
公钥 (Public Key, JSON): $ADMIN_PUBKEY_JSON
公钥 (Public Key, Hex): $ADMIN_PUBKEY_HEX
EOF

if [ "$ADMIN_PRIVKEY" != "N/A (使用助记词恢复)" ]; then
    echo "私钥 (Private Key, Hex): $ADMIN_PRIVKEY" >> $INFO_FILE
else
    echo "助记词 (Mnemonic): $ADMIN_MNEMONIC" >> $INFO_FILE
fi

echo "" >> $INFO_FILE
echo "余额: $ADMIN_BALANCE $DENOM" >> $INFO_FILE
echo "==========================================" >> $INFO_FILE

echo -e "\n✓ 用户信息已保存到: $INFO_FILE"

# 清理临时文件
rm -f /tmp/admin_key.json /tmp/admin_privkey_hex.txt
TASKADR=`dtcd q auth module-accounts --output json | jq -r '.accounts[] | select(.value.name=="task") | .value.address'`
$BINARY tx bank send $ALICE_ADDRESS $TASKADR  50000000000$DENOM\
    --from alice \
    --chain-id $CHAIN_ID \
    --gas auto \
    --gas-adjustment 1.5 \
    --fees $FEE \
    -y

echo -e "\n脚本执行完成！"

# 更新链上的 AdminPubkey 参数
# 注意: UpdateParams 命令在 autocli 中被跳过，需要通过其他方式更新
echo -e "\n--- 步骤 5: 更新链上的 AdminPubkey 参数（可选）---"
echo "注意: UpdateParams 需要 authority 权限（默认是 gov 模块）"
echo ""
echo "当前 Admin 公钥 (Hex): $ADMIN_PUBKEY_HEX"
echo ""
echo "要更新链上的 AdminPubkey 参数，可以使用以下方法："
echo ""
echo "方法 1: 使用 JSON 交易（如果 admin 是 authority）"
echo "创建 update_params.json 文件："
cat > update_params.json <<EOF
{
  "body": {
    "messages": [
      {
        "@type": "/dtc.identity.v1.MsgUpdateParams",
        "authority": "$ADMIN_ADDRESS",
        "params": {
          "admin_pubkey": "$ADMIN_PUBKEY_HEX"
        }
      }
    ],
    "memo": "Update identity module AdminPubkey"
  },
  "auth_info": {
    "signer_infos": [],
    "fee": {
      "amount": [{"denom": "$DENOM", "amount": "200"}],
      "gas_limit": "200000"
    }
  }
}
EOF
echo "  文件已创建: update_params.json"
echo ""
echo "方法 2: 通过治理提案（推荐）"
echo "  如果 admin 不是 authority，需要通过治理提案更新参数"
echo ""
echo "方法 3: 在 genesis 文件中设置（仅适用于新链）"
echo "  在 genesis.json 的 identity 模块参数中设置 admin_pubkey"
echo ""
echo "提示: 如果链上未设置 AdminPubkey，将使用代码中的默认值："
echo "  03555db1e9893d6bafff7c3afdb62ddb99cf2f073d25144701966607f63e561a38"