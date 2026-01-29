package task

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"dtc/testutil/sample"
	tasksimulation "dtc/x/task/simulation"
	"dtc/x/task/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	taskGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		ClaimRecordMap: []types.ClaimRecord{{Creator: sample.AccAddress(),
			ClaimHash: "0",
		}, {Creator: sample.AccAddress(),
			ClaimHash: "1",
		}}}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&taskGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgCreateClaimRecord          = "op_weight_msg_task"
		defaultWeightMsgCreateClaimRecord int = 100
	)

	var weightMsgCreateClaimRecord int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateClaimRecord, &weightMsgCreateClaimRecord, nil,
		func(_ *rand.Rand) {
			weightMsgCreateClaimRecord = defaultWeightMsgCreateClaimRecord
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateClaimRecord,
		tasksimulation.SimulateMsgCreateClaimRecord(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgUpdateClaimRecord          = "op_weight_msg_task"
		defaultWeightMsgUpdateClaimRecord int = 100
	)

	var weightMsgUpdateClaimRecord int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateClaimRecord, &weightMsgUpdateClaimRecord, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateClaimRecord = defaultWeightMsgUpdateClaimRecord
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateClaimRecord,
		tasksimulation.SimulateMsgUpdateClaimRecord(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgDeleteClaimRecord          = "op_weight_msg_task"
		defaultWeightMsgDeleteClaimRecord int = 100
	)

	var weightMsgDeleteClaimRecord int
	simState.AppParams.GetOrGenerate(opWeightMsgDeleteClaimRecord, &weightMsgDeleteClaimRecord, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteClaimRecord = defaultWeightMsgDeleteClaimRecord
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteClaimRecord,
		tasksimulation.SimulateMsgDeleteClaimRecord(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgClaimReward          = "op_weight_msg_task"
		defaultWeightMsgClaimReward int = 100
	)

	var weightMsgClaimReward int
	simState.AppParams.GetOrGenerate(opWeightMsgClaimReward, &weightMsgClaimReward, nil,
		func(_ *rand.Rand) {
			weightMsgClaimReward = defaultWeightMsgClaimReward
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgClaimReward,
		tasksimulation.SimulateMsgClaimReward(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
