package keeper

import (
	"context"

	"dtc/x/identity/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) GetDidByAddress(ctx context.Context, req *types.QueryGetDidByAddressRequest) (*types.QueryGetDidByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// 遍历 DidDocument，查找 Controller 等于输入 address 的文档
	var foundDidDocument *types.DidDocument
	err := q.k.DidDocument.Walk(ctx, nil, func(key string, value types.DidDocument) (stop bool, err error) {
		if value.Controller == req.Address {
			foundDidDocument = &value
			return true, nil // 找到后停止遍历
		}
		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 如果没找到，返回 isRegistered: false
	if foundDidDocument == nil {
		return &types.QueryGetDidByAddressResponse{
			IsRegistered: false,
			Did:          "",
			FaceHash:     "",
		}, nil
	}

	// 找到后返回 did 和 faceHash
	return &types.QueryGetDidByAddressResponse{
		IsRegistered: true,
		Did:          foundDidDocument.Did,
		FaceHash:     foundDidDocument.FaceHash,
	}, nil
}
