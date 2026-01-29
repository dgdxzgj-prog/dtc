package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"dtc/x/identity/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema          collections.Schema
	Params          collections.Item[types.Params]
	DidDocument     collections.Map[string, types.DidDocument]
	FaceHashToIndex collections.Map[string, string] // faceHash -> did
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,

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

		Params:          collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		DidDocument:     collections.NewMap(sb, types.DidDocumentKey, "didDocument", collections.StringKey, codec.CollValue[types.DidDocument](cdc)),
		FaceHashToIndex: collections.NewMap(sb, types.FaceHashToIndexKey, "faceHashToIndex", collections.StringKey, collections.StringValue),
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

// GetDidDocument returns the DidDocument whose Controller equals the given address.
// It is used by the credit module's IdentityKeeper interface.
func (k Keeper) GetDidDocument(ctx sdk.Context, address string) (val types.DidDocument, found bool) {
	var foundDoc *types.DidDocument
	err := k.DidDocument.Walk(ctx.Context(), nil, func(key string, value types.DidDocument) (stop bool, err error) {
		if value.Controller == address {
			foundDoc = &value
			return true, nil
		}
		return false, nil
	})
	if err != nil || foundDoc == nil {
		return types.DidDocument{}, false
	}
	return *foundDoc, true
}
