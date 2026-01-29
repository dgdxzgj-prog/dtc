package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"dtc/x/credit/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema collections.Schema
	Params collections.Item[types.Params]

	// CreditAccount 按地址：Liability（负债）、LastMintHeight（最近铸币高度）、BirthHeight（账户创建高度）
	CreditAccountLiability      collections.Map[string, uint64]
	CreditAccountLastMintHeight collections.Map[string, uint64]
	CreditAccountBirthHeight    collections.Map[string, uint64]

	bankKeeper    types.BankKeeper
	authKeeper   types.AuthKeeper
	identityKeeper types.IdentityKeeper
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,

	bankKeeper types.BankKeeper,
	authKeeper types.AuthKeeper,
	identityKeeper types.IdentityKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,

		bankKeeper:     bankKeeper,
		authKeeper:     authKeeper,
		identityKeeper: identityKeeper,
		Params:                    collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		CreditAccountLiability:    collections.NewMap(sb, types.CreditAccountLiabilityPrefix, "ca_liability", collections.StringKey, collections.Uint64Value),
		CreditAccountLastMintHeight: collections.NewMap(sb, types.CreditAccountLastMintHeightPrefix, "ca_last_mint", collections.StringKey, collections.Uint64Value),
		CreditAccountBirthHeight:  collections.NewMap(sb, types.CreditAccountBirthHeightPrefix, "ca_birth", collections.StringKey, collections.Uint64Value),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}

// IterateCreditAccount 遍历所有信用账户，对每个账户调用回调函数
// 回调函数参数：地址、负债、出生高度
func (k Keeper) IterateCreditAccount(ctx context.Context, cb func(addr string, liability uint64, birthHeight uint64) error) error {
	iter, err := k.CreditAccountLiability.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}
		addr := kv.Key
		liability := kv.Value

		var birthHeight uint64
		birthHeight, err = k.CreditAccountBirthHeight.Get(ctx, addr)
		if err != nil && !errors.Is(err, collections.ErrNotFound) {
			return err
		}
		// 如果未找到 BirthHeight，使用 0（表示账户年龄为 0）

		if err := cb(addr, liability, birthHeight); err != nil {
			return err
		}
	}

	return nil
}
