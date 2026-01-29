package keeper_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"dtc/x/task/keeper"
	"dtc/x/task/types"
)

func createNClaimRecord(keeper keeper.Keeper, ctx context.Context, n int) []types.ClaimRecord {
	items := make([]types.ClaimRecord, n)
	for i := range items {
		items[i].ClaimHash = strconv.Itoa(i)
		items[i].TaskId = strconv.Itoa(i)
		items[i].UserId = strconv.Itoa(i)
		items[i].Signature = strconv.Itoa(i)
		_ = keeper.ClaimRecord.Set(ctx, items[i].ClaimHash, items[i])
	}
	return items
}

func TestClaimRecordQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNClaimRecord(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetClaimRecordRequest
		response *types.QueryGetClaimRecordResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetClaimRecordRequest{
				ClaimHash: msgs[0].ClaimHash,
			},
			response: &types.QueryGetClaimRecordResponse{ClaimRecord: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetClaimRecordRequest{
				ClaimHash: msgs[1].ClaimHash,
			},
			response: &types.QueryGetClaimRecordResponse{ClaimRecord: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetClaimRecordRequest{
				ClaimHash: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := qs.GetClaimRecord(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestClaimRecordQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNClaimRecord(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllClaimRecordRequest {
		return &types.QueryAllClaimRecordRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListClaimRecord(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.ClaimRecord), step)
			require.Subset(t, msgs, resp.ClaimRecord)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListClaimRecord(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.ClaimRecord), step)
			require.Subset(t, msgs, resp.ClaimRecord)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListClaimRecord(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.ClaimRecord)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListClaimRecord(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
