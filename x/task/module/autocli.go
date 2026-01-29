package task

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"dtc/x/task/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "ListClaimRecord",
					Use:       "list-claim-record",
					Short:     "List all claimRecord",
				},
				{
					RpcMethod:      "GetClaimRecord",
					Use:            "get-claim-record [id]",
					Short:          "Gets a claimRecord",
					Alias:          []string{"show-claim-record"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "claim_hash"}},
				},
				// this line is used by ignite scaffolding # autocli/query
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod:      "CreateClaimRecord",
					Use:            "create-claim-record [claim_hash] [task-id] [user-id] [signature]",
					Short:          "Create a new claimRecord",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "claim_hash"}, {ProtoField: "task_id"}, {ProtoField: "user_id"}, {ProtoField: "signature"}},
				},
				{
					RpcMethod:      "UpdateClaimRecord",
					Use:            "update-claim-record [claim_hash] [task-id] [user-id] [signature]",
					Short:          "Update claimRecord",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "claim_hash"}, {ProtoField: "task_id"}, {ProtoField: "user_id"}, {ProtoField: "signature"}},
				},
				{
					RpcMethod:      "DeleteClaimRecord",
					Use:            "delete-claim-record [claim_hash]",
					Short:          "Delete claimRecord",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "claim_hash"}},
				},
				{
					RpcMethod:      "ClaimReward",
					Use:            "claim-reward [task-id] [amount] [signature]",
					Short:          "Send a claimReward tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "task_id"}, {ProtoField: "amount"}, {ProtoField: "signature"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
