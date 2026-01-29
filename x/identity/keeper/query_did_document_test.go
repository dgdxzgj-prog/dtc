package keeper_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"dtc/x/identity/keeper"
	"dtc/x/identity/types"
)

func createNDidDocument(keeper keeper.Keeper, ctx context.Context, n int) []types.DidDocument {
	items := make([]types.DidDocument, n)
	for i := range items {
		items[i].Did = strconv.Itoa(i)
		items[i].Controller = strconv.Itoa(i)
		items[i].Pubkeys = strconv.Itoa(i)
		_ = keeper.DidDocument.Set(ctx, items[i].Did, items[i])
	}
	return items
}

func TestDidDocumentQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNDidDocument(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetDidDocumentRequest
		response *types.QueryGetDidDocumentResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetDidDocumentRequest{
				Did: msgs[0].Did,
			},
			response: &types.QueryGetDidDocumentResponse{DidDocument: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetDidDocumentRequest{
				Did: msgs[1].Did,
			},
			response: &types.QueryGetDidDocumentResponse{DidDocument: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetDidDocumentRequest{
				Did: strconv.Itoa(100000),
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
			response, err := qs.GetDidDocument(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestDidDocumentQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNDidDocument(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllDidDocumentRequest {
		return &types.QueryAllDidDocumentRequest{
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
			resp, err := qs.ListDidDocument(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.DidDocument), step)
			require.Subset(t, msgs, resp.DidDocument)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListDidDocument(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.DidDocument), step)
			require.Subset(t, msgs, resp.DidDocument)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListDidDocument(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.DidDocument)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListDidDocument(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
