package identity

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"dtc/x/identity/types"
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
					RpcMethod: "ListDidDocument",
					Use:       "list-did-document",
					Short:     "List all didDocument",
				},
				{
					RpcMethod:      "GetDidDocument",
					Use:            "get-did-document [id]",
					Short:          "Gets a didDocument",
					Alias:          []string{"show-did-document"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "did"}},
				},
				{
					RpcMethod:      "GetDidByAddress",
					Use:            "get-did-by-address [address]",
					Short:          "Query getDidByAddress",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
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
					RpcMethod:      "CreateDidDocument",
					Use:            "create-did-document [did] [controller] [pubkeys]",
					Short:          "Create a new didDocument",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "did"}, {ProtoField: "controller"}, {ProtoField: "pubkeys"}},
				},
				{
					RpcMethod:      "UpdateDidDocument",
					Use:            "update-did-document [did] [controller] [pubkeys]",
					Short:          "Update didDocument",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "did"}, {ProtoField: "controller"}, {ProtoField: "pubkeys"}},
				},
				{
					RpcMethod:      "DeleteDidDocument",
					Use:            "delete-did-document [did]",
					Short:          "Delete didDocument",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "did"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
