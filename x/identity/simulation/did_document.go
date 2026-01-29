package simulation

import (
	"math/rand"
	"strconv"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"dtc/x/identity/keeper"
	"dtc/x/identity/types"
)

func SimulateMsgCreateDidDocument(
	ak types.AuthKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
	txGen client.TxConfig,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		i := r.Int()
		msg := &types.MsgCreateDidDocument{
			Creator:   simAccount.Address.String(),
			Did:       strconv.Itoa(i),
			Signature: []byte("7369676e6174757265"), // integration test signature to bypass verification
		}

		found, err := k.DidDocument.Has(ctx, msg.Did)
		if err == nil && found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "DidDocument already exist"), nil, nil
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txGen,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      simAccount,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgUpdateDidDocument(
	ak types.AuthKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
	txGen client.TxConfig,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var (
			simAccount  = simtypes.Account{}
			didDocument = types.DidDocument{}
			msg         = &types.MsgUpdateDidDocument{}
			found       = false
		)

		var allDidDocument []types.DidDocument
		err := k.DidDocument.Walk(ctx, nil, func(key string, value types.DidDocument) (stop bool, err error) {
			allDidDocument = append(allDidDocument, value)
			return false, nil
		})
		if err != nil {
			panic(err)
		}

		for _, obj := range allDidDocument {
			acc, err := ak.AddressCodec().StringToBytes(obj.Controller)
			if err != nil {
				return simtypes.OperationMsg{}, nil, err
			}

			simAccount, found = simtypes.FindAccount(accs, sdk.AccAddress(acc))
			if found {
				didDocument = obj
				break
			}
		}
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "didDocument controller not found"), nil, nil
		}
		msg.Creator = simAccount.Address.String()
		msg.Did = didDocument.Did

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txGen,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      simAccount,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgDeleteDidDocument(
	ak types.AuthKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
	txGen client.TxConfig,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var (
			simAccount  = simtypes.Account{}
			didDocument = types.DidDocument{}
			msg         = &types.MsgUpdateDidDocument{}
			found       = false
		)

		var allDidDocument []types.DidDocument
		err := k.DidDocument.Walk(ctx, nil, func(key string, value types.DidDocument) (stop bool, err error) {
			allDidDocument = append(allDidDocument, value)
			return false, nil
		})
		if err != nil {
			panic(err)
		}

		for _, obj := range allDidDocument {
			acc, err := ak.AddressCodec().StringToBytes(obj.Controller)
			if err != nil {
				return simtypes.OperationMsg{}, nil, err
			}

			simAccount, found = simtypes.FindAccount(accs, sdk.AccAddress(acc))
			if found {
				didDocument = obj
				break
			}
		}
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "didDocument controller not found"), nil, nil
		}
		msg.Creator = simAccount.Address.String()
		msg.Did = didDocument.Did

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txGen,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      simAccount,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
