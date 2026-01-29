package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"cosmossdk.io/math"

	"dtc/x/credit/types"
)

// EndBlocker 在每个区块结束时执行自动清偿逻辑
func (k Keeper) EndBlocker(ctx sdk.Context) error {
	// 将 sdk.Context 转换为 context.Context
	ctxContext := ctx.Context()
	currentHeight := uint64(ctx.BlockHeight())

	// 遍历所有信用账户
	err := k.IterateCreditAccount(ctxContext, func(addr string, liability uint64, birthHeight uint64) error {
		// 跳过没有负债的账户
		if liability == 0 {
			return nil
		}

		// 计算账户年龄（当前高度 - 出生高度）
		var age uint64
		if birthHeight > 0 {
			if currentHeight < birthHeight {
				// 防御性检查：如果当前高度小于出生高度，跳过
				return nil
			}
			age = currentHeight - birthHeight
		} else {
			// 如果没有 BirthHeight，年龄为 0，跳过
			return nil
		}

		// 如果年龄不超过 100 个区块，跳过
		if age <= 100 {
			return nil
		}

		// 解析地址
		addrBytes, err := k.addressCodec.StringToBytes(addr)
		if err != nil {
			// 地址格式错误，跳过该账户
			return nil
		}
		accAddr := sdk.AccAddress(addrBytes)

		// 查询账户可用余额
		spendableCoins := k.bankKeeper.SpendableCoins(ctxContext, accAddr)
		if spendableCoins.IsZero() {
			// 没有可用余额，跳过
			return nil
		}

		// 计算需要偿还的金额：可用余额的 5%
		// 使用 MulInt 计算：先乘以 5，再除以 100
		repayCoins := spendableCoins.MulInt(math.NewInt(5)).QuoInt(math.NewInt(100))
		if repayCoins.IsZero() {
			return nil
		}

		// 确保偿还金额不超过负债（以 udtc 为单位）
		liabilityCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(liability)))
		if repayCoins.AmountOf(sdk.DefaultBondDenom).GT(liabilityCoins.AmountOf(sdk.DefaultBondDenom)) {
			repayCoins = liabilityCoins
		}

		// 从账户划转到 credit 模块
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctxContext, accAddr, types.ModuleName, repayCoins); err != nil {
			// 如果划转失败（例如余额不足），跳过该账户
			return nil
		}

		// 销毁代币
		if err := k.bankKeeper.BurnCoins(ctxContext, types.ModuleName, repayCoins); err != nil {
			// 如果销毁失败，返回错误（因为代币已经在模块账户中）
			return errorsmod.Wrap(sdkerrors.ErrLogic, "burn coins: "+err.Error())
		}

		// 从负债中等额扣除
		repayAmount := repayCoins.AmountOf(sdk.DefaultBondDenom).Uint64()
		if repayAmount > liability {
			repayAmount = liability
		}
		newLiability := liability - repayAmount

		if newLiability == 0 {
			// 如果负债清零，删除记录
			if err := k.CreditAccountLiability.Remove(ctxContext, addr); err != nil {
				return errorsmod.Wrap(sdkerrors.ErrLogic, "remove credit account liability: "+err.Error())
			}
		} else {
			// 更新负债
			if err := k.CreditAccountLiability.Set(ctxContext, addr, newLiability); err != nil {
				return errorsmod.Wrap(sdkerrors.ErrLogic, "set credit account liability: "+err.Error())
			}
		}

		return nil
	})

	return err
}
