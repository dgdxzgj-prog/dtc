package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"dtc/x/credit/types"
)

// 铸币总量：1 DTC = 1000000（与 udtc 最小单位一致）
const totalMintAmount = 100000000

// 分流计算基数：gbdp_rate 100 表示 1%
const rateBase = 10000

// BlocksPerMonth 按 5 秒一区块计算，30 天约 518400 个区块
const BlocksPerMonth = 518400

func (k msgServer) MintCredit(ctx context.Context, msg *types.MsgMintCredit) (*types.MsgMintCreditResponse, error) {
	creatorAddr, err := k.addressCodec.StringToBytes(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// 1. 身份准入检查：Creator 必须已注册 DID
	_, found := k.identityKeeper.GetDidDocument(sdkCtx, msg.Creator)
	if !found {
		return nil, errorsmod.Wrap(types.ErrIdentityNotRegistered, msg.Creator)
	}

	// 2. 铸币间隔检查：距上次铸币不足 BlocksPerMonth 则拒绝
	lastMintHeight, err := k.CreditAccountLastMintHeight.Get(ctx, msg.Creator)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "get last mint height: "+err.Error())
	}
	if !errors.Is(err, collections.ErrNotFound) && lastMintHeight > 0 {
		blocksSince := sdkCtx.BlockHeight() - int64(lastMintHeight)
		if blocksSince < int64(BlocksPerMonth) {
			return nil, errorsmod.Wrap(types.ErrMintTooFrequent, "must wait at least one month since last mint")
		}
	}

	// 3. 获取参数
	params, err := k.Params.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to get params")
	}
	gbdpRate := params.GbdpRate
	if gbdpRate > rateBase {
		gbdpRate = rateBase
	}

	// 4. 铸造总量到 credit 模块
	totalCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, totalMintAmount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, totalCoins); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "mint coins: "+err.Error())
	}

	// 5. 按 gbdp_rate 分流：gbdp_rate/10000 给 GBDP 池，其余给 Creator
	gbdpAmount := (totalMintAmount * int64(gbdpRate)) / int64(rateBase)
	creatorAmount := totalMintAmount - gbdpAmount

	creatorCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, creatorAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, creatorCoins); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "send coins to creator: "+err.Error())
	}

	if gbdpAmount > 0 {
		gbdpPoolAddr := authtypes.NewModuleAddress(types.GBDPPoolModuleName)
		gbdpCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, gbdpAmount))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, gbdpPoolAddr, gbdpCoins); err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "send coins to GBDP pool: "+err.Error())
		}
	}

	// 6. 获取或创建该地址的 CreditAccount：Liability += totalAmount，LastMintHeight = 当前高度
	height := uint64(sdkCtx.BlockHeight())
	addrStr := msg.Creator

	var liability uint64
	liability, err = k.CreditAccountLiability.Get(ctx, addrStr)
	isNewAccount := errors.Is(err, collections.ErrNotFound)
	if err != nil && !isNewAccount {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "get credit account liability: "+err.Error())
	}
	if isNewAccount {
		liability = 0
		// 新账户：设置 BirthHeight 为当前高度
		if err := k.CreditAccountBirthHeight.Set(ctx, addrStr, height); err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "set credit account birth height: "+err.Error())
		}
	}
	liability += uint64(totalMintAmount)
	if err := k.CreditAccountLiability.Set(ctx, addrStr, liability); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "set credit account liability: "+err.Error())
	}
	if err := k.CreditAccountLastMintHeight.Set(ctx, addrStr, height); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "set credit account last mint height: "+err.Error())
	}

	return &types.MsgMintCreditResponse{}, nil
}
