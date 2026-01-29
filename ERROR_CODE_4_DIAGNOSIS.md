# 错误码 4 诊断文档

## 错误信息
```
code=4, rawLog=failed to execute message; message index: 0: invalid admin signature: unauthorized
```

## 错误码说明

**错误码 4** 对应 Cosmos SDK 的 `sdkerrors.ErrUnauthorized`，表示未授权错误。

在 `CreateDidDocument` 函数中，这个错误发生在第90行：
```go
return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "invalid admin signature")
```

## 签名验证流程

链端的签名验证逻辑（`x/identity/keeper/msg_server_did_document.go`）：

1. **构造待签名数据**（第55行）：
   ```go
   data := msg.Did + controller + msg.FaceHash
   ```
   - `controller` 如果为空，则使用 `msg.Creator`
   - 数据格式：字符串直接拼接，**无分隔符**

2. **SHA256 哈希**（第56行）：
   ```go
   hash := sha256.Sum256([]byte(data))
   hashBytes := hash[:]
   ```

3. **签名验证**（第88行）：
   ```go
   valid := ecdsa.Verify(pubKey, hashBytes, r, s)
   ```
   - 使用 ECDSA 验证
   - 签名格式：**64字节的 R||S**（前32字节是R，后32字节是S）
   - **不是 DER 编码格式**

## 可能的问题原因

### 1. 公钥不匹配 ⚠️ **最可能的原因**

**问题**：中台使用的私钥对应的公钥与链上配置的 AdminPubkey 不一致。

**检查方法**：
```bash
# 查询链上的 AdminPubkey 参数
dtcd q identity params --output json | jq -r '.params.admin_pubkey'

# 检查 admin 用户的公钥
dtcd keys show admin --output json | jq -r '.pubkey'
```

**解决方案**：
- **方案A**：更新链上的 AdminPubkey 参数，使其与中台使用的私钥对应的公钥一致
  ```bash
  dtcd tx identity update-params \
    --admin-pubkey <中台使用的公钥hex> \
    --from <authority账户> \
    --chain-id dtc \
    -y
  ```

- **方案B**：确保中台使用与链上 AdminPubkey 对应的私钥进行签名

### 2. 签名数据构造不正确

**问题**：中台构造的签名数据与链端不一致。

**链端构造方式**：
```
数据 = Did + Controller + FaceHash
```
- **无分隔符**，直接字符串拼接
- 如果 `Controller` 为空，使用 `Creator`

**中台需要确保**：
- 使用相同的顺序：`Did + Controller + FaceHash`
- 如果 `Controller` 为空，使用 `Creator`
- 无任何分隔符或额外字符

### 3. 签名格式不正确

**问题**：签名格式不符合链端要求。

**链端要求**：
- 签名必须是 **64字节**
- 格式：`R || S`（前32字节是R，后32字节是S）
- **不是 DER 编码格式**

**中台需要确保**：
- 签名是64字节的原始 R||S 格式
- 不是 DER 编码
- 不是其他格式（如以太坊的 v||r||s）

### 4. 链上参数未设置

**问题**：链上的 AdminPubkey 参数未设置，使用了默认值。

**默认值**（代码第21行）：
```go
const defaultAdminPubKeyHex = "03555db1e9893d6bafff7c3afdb62ddb99cf2f073d25144701966607f63e561a38"
```

**检查方法**：
```bash
dtcd q identity params --output json | jq -r '.params.admin_pubkey'
```

如果返回空或 null，说明使用了默认值。

## 诊断步骤

1. **运行诊断脚本**：
   ```bash
   ./check_admin_signature.sh
   ```

2. **检查链上参数**：
   ```bash
   dtcd q identity params
   ```

3. **验证公钥匹配**：
   - 获取链上的 AdminPubkey
   - 获取中台使用的私钥对应的公钥
   - 比较两者是否一致

4. **验证签名数据**：
   - 确认中台构造的数据格式：`Did + Controller + FaceHash`
   - 确认签名格式：64字节 R||S

## 解决方案

### 快速修复（推荐）

如果中台已经使用了正确的私钥，但链上参数未设置或设置错误：

```bash
# 1. 获取中台使用的公钥（hex格式，33字节压缩公钥）
ADMIN_PUBKEY_HEX="<中台使用的公钥hex>"

# 2. 更新链上参数（需要 authority 权限）
dtcd tx identity update-params \
  --admin-pubkey $ADMIN_PUBKEY_HEX \
  --from <authority账户> \
  --chain-id dtc \
  --gas auto \
  --gas-adjustment 1.5 \
  --fees 200udtc \
  -y
```

### 中台签名实现示例（参考）

```javascript
// 伪代码示例
const data = did + controller + faceHash;  // 字符串拼接，无分隔符
const hash = sha256(data);  // SHA256 哈希
const signature = ecdsaSign(hash, privateKey);  // ECDSA 签名
// signature 应该是 64 字节的 R||S 格式
```

## 当前配置信息

从 `admin_info.txt` 可以看到：
- Admin 地址：`dtc1cpyfdz7xg69smlfg32vd9a69dfl3u7awlwfxp8`
- Admin 公钥（base64）：`A1VdsemJPWuv/3w6/bYt25nPLwc9JRRHAZZmB/Y+Vho4`
- Admin 公钥（hex）：`03555db1e9893d6bafff7c3afdb62ddb99cf2f073d25144701966607f63e561a38`

**注意**：代码中的默认公钥与 admin 用户的公钥是匹配的。如果链上参数未设置，应该使用这个默认值。

## 总结

错误码 4 表示签名验证失败。最可能的原因是：
1. **公钥不匹配**：中台使用的私钥对应的公钥与链上配置不一致
2. **签名数据格式错误**：数据构造方式与链端不一致
3. **签名格式错误**：不是64字节的 R||S 格式

建议首先检查公钥是否匹配，这是最常见的问题。
