package keeper

import (
	"context"

	"dtc/x/credit/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) SubmitDeathCertificate(ctx context.Context, msg *types.MsgSubmitDeathCertificate) (*types.MsgSubmitDeathCertificateResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgSubmitDeathCertificateResponse{}, nil
}
