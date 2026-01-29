package keeper_test

import (
	"strconv"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"dtc/x/task/keeper"
	"dtc/x/task/types"
)

func TestClaimRecordMsgServerCreate(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)
	creator, err := f.addressCodec.BytesToString([]byte("signerAddr__________________"))
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		expected := &types.MsgCreateClaimRecord{Creator: creator,
			ClaimHash: strconv.Itoa(i),
		}
		_, err := srv.CreateClaimRecord(f.ctx, expected)
		require.NoError(t, err)
		rst, err := f.keeper.ClaimRecord.Get(f.ctx, expected.ClaimHash)
		require.NoError(t, err)
		require.Equal(t, expected.Creator, rst.Creator)
	}
}

func TestClaimRecordMsgServerUpdate(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte("signerAddr__________________"))
	require.NoError(t, err)

	unauthorizedAddr, err := f.addressCodec.BytesToString([]byte("unauthorizedAddr___________"))
	require.NoError(t, err)

	expected := &types.MsgCreateClaimRecord{Creator: creator,
		ClaimHash: strconv.Itoa(0),
	}
	_, err = srv.CreateClaimRecord(f.ctx, expected)
	require.NoError(t, err)

	tests := []struct {
		desc    string
		request *types.MsgUpdateClaimRecord
		err     error
	}{
		{
			desc: "invalid address",
			request: &types.MsgUpdateClaimRecord{Creator: "invalid",
				ClaimHash: strconv.Itoa(0),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			desc: "unauthorized",
			request: &types.MsgUpdateClaimRecord{Creator: unauthorizedAddr,
				ClaimHash: strconv.Itoa(0),
			},
			err: sdkerrors.ErrUnauthorized,
		},
		{
			desc: "key not found",
			request: &types.MsgUpdateClaimRecord{Creator: creator,
				ClaimHash: strconv.Itoa(100000),
			},
			err: sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "completed",
			request: &types.MsgUpdateClaimRecord{Creator: creator,
				ClaimHash: strconv.Itoa(0),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			_, err = srv.UpdateClaimRecord(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				rst, err := f.keeper.ClaimRecord.Get(f.ctx, expected.ClaimHash)
				require.NoError(t, err)
				require.Equal(t, expected.Creator, rst.Creator)
			}
		})
	}
}

func TestClaimRecordMsgServerDelete(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte("signerAddr__________________"))
	require.NoError(t, err)

	unauthorizedAddr, err := f.addressCodec.BytesToString([]byte("unauthorizedAddr___________"))
	require.NoError(t, err)

	_, err = srv.CreateClaimRecord(f.ctx, &types.MsgCreateClaimRecord{Creator: creator,
		ClaimHash: strconv.Itoa(0),
	})
	require.NoError(t, err)

	tests := []struct {
		desc    string
		request *types.MsgDeleteClaimRecord
		err     error
	}{
		{
			desc: "invalid address",
			request: &types.MsgDeleteClaimRecord{Creator: "invalid",
				ClaimHash: strconv.Itoa(0),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			desc: "unauthorized",
			request: &types.MsgDeleteClaimRecord{Creator: unauthorizedAddr,
				ClaimHash: strconv.Itoa(0),
			},
			err: sdkerrors.ErrUnauthorized,
		},
		{
			desc: "key not found",
			request: &types.MsgDeleteClaimRecord{Creator: creator,
				ClaimHash: strconv.Itoa(100000),
			},
			err: sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "completed",
			request: &types.MsgDeleteClaimRecord{Creator: creator,
				ClaimHash: strconv.Itoa(0),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			_, err = srv.DeleteClaimRecord(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				found, err := f.keeper.ClaimRecord.Has(f.ctx, tc.request.ClaimHash)
				require.NoError(t, err)
				require.False(t, found)
			}
		})
	}
}
