package keeper

import (
	"context"
	"errors"
	"fmt"

	"dtc/x/task/types"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateClaimRecord(ctx context.Context, msg *types.MsgCreateClaimRecord) (*types.MsgCreateClaimRecordResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid address: %s", err))
	}

	// Check if the value already exists
	ok, err := k.ClaimRecord.Has(ctx, msg.ClaimHash)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, err.Error())
	} else if ok {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "index already set")
	}

	var claimRecord = types.ClaimRecord{
		Creator:   msg.Creator,
		ClaimHash: msg.ClaimHash,
		TaskId:    msg.TaskId,
		UserId:    msg.UserId,
		Signature: msg.Signature,
	}

	if err := k.ClaimRecord.Set(ctx, claimRecord.ClaimHash, claimRecord); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, err.Error())
	}

	return &types.MsgCreateClaimRecordResponse{}, nil
}

func (k msgServer) UpdateClaimRecord(ctx context.Context, msg *types.MsgUpdateClaimRecord) (*types.MsgUpdateClaimRecordResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid signer address: %s", err))
	}

	// Check if the value exists
	val, err := k.ClaimRecord.Get(ctx, msg.ClaimHash)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
		}

		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, err.Error())
	}

	// Checks if the msg creator is the same as the current owner
	if msg.Creator != val.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	var claimRecord = types.ClaimRecord{
		Creator:   msg.Creator,
		ClaimHash: msg.ClaimHash,
		TaskId:    msg.TaskId,
		UserId:    msg.UserId,
		Signature: msg.Signature,
	}

	if err := k.ClaimRecord.Set(ctx, claimRecord.ClaimHash, claimRecord); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to update claimRecord")
	}

	return &types.MsgUpdateClaimRecordResponse{}, nil
}

func (k msgServer) DeleteClaimRecord(ctx context.Context, msg *types.MsgDeleteClaimRecord) (*types.MsgDeleteClaimRecordResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid signer address: %s", err))
	}

	// Check if the value exists
	val, err := k.ClaimRecord.Get(ctx, msg.ClaimHash)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
		}

		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, err.Error())
	}

	// Checks if the msg creator is the same as the current owner
	if msg.Creator != val.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	if err := k.ClaimRecord.Remove(ctx, msg.ClaimHash); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to remove claimRecord")
	}

	return &types.MsgDeleteClaimRecordResponse{}, nil
}
