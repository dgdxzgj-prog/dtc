package keeper

import (
	"context"
	"errors"

	"dtc/x/identity/types"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) ListDidDocument(ctx context.Context, req *types.QueryAllDidDocumentRequest) (*types.QueryAllDidDocumentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	didDocuments, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.DidDocument,
		req.Pagination,
		func(_ string, value types.DidDocument) (types.DidDocument, error) {
			return value, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllDidDocumentResponse{DidDocument: didDocuments, Pagination: pageRes}, nil
}

func (q queryServer) GetDidDocument(ctx context.Context, req *types.QueryGetDidDocumentRequest) (*types.QueryGetDidDocumentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, err := q.k.DidDocument.Get(ctx, req.Did)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetDidDocumentResponse{DidDocument: val}, nil
}
