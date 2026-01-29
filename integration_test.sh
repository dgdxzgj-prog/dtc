#!/bin/bash

BINARY="dtcd"
CHAIN_ID="dtc"
DENOM="udtc"

# 获取 Alice 地址
ALICE=$( $BINARY keys show alice -a )

echo "--- 准备阶段 ---"
# 优化余额获取逻辑，确保只拿数字
INITIAL_BAL=$($BINARY q bank balances $ALICE --output json | jq -r '.balances[] | select(.denom=="'$DENOM'") | .amount')
echo "Alice 初始余额: $INITIAL_BAL $DENOM"

TASK_ID="task_$(date +%s)"
AMOUNT="500$DENOM"
SIGNATURE="7369676e6174757265"

echo -e "\n--- 步骤 2: 用户 Alice 领取奖励 ---"
# 修正 flag: --gas-adjustment
$BINARY tx task claim-reward "$TASK_ID" "$AMOUNT" "$SIGNATURE" \
  --from alice \
  --chain-id $CHAIN_ID \
  --gas auto \
  --gas-adjustment 1.5 \
  --fees 200$DENOM \
  -y

echo "等待 6 秒让区块确认..."
sleep 6

echo -e "\n--- 步骤 3: 验证结果 ---"
FINAL_BAL=$($BINARY q bank balances $ALICE --output json | jq -r '.balances[] | select(.denom=="'$DENOM'") | .amount')
echo "Alice 新余额: $FINAL_BAL $DENOM"

# 计算差值
DIFF=$((FINAL_BAL - INITIAL_BAL))
echo "余额增加量: $DIFF $DENOM"

echo -e "\n--- 步骤 4: 防刷测试 ---"
# 捕获错误输出
# 修改这行
RESULT=$($BINARY tx task claim-reward "$TASK_ID" "$AMOUNT" "$SIGNATURE" --from alice --chain-id $CHAIN_ID -y --broadcast-mode sync)
echo "重复领取返回结果: $(echo $RESULT | jq -r '.raw_log')"