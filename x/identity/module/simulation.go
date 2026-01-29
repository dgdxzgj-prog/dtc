package identity

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"dtc/testutil/sample"
	identitysimulation "dtc/x/identity/simulation"
	"dtc/x/identity/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	identityGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		DidDocumentMap: []types.DidDocument{{
			Did:        "0",
			Controller: sample.AccAddress(),
		}, {
			Did:        "1",
			Controller: sample.AccAddress(),
		}}}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&identityGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgCreateDidDocument          = "op_weight_msg_identity"
		defaultWeightMsgCreateDidDocument int = 100
	)

	var weightMsgCreateDidDocument int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateDidDocument, &weightMsgCreateDidDocument, nil,
		func(_ *rand.Rand) {
			weightMsgCreateDidDocument = defaultWeightMsgCreateDidDocument
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateDidDocument,
		identitysimulation.SimulateMsgCreateDidDocument(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgUpdateDidDocument          = "op_weight_msg_identity"
		defaultWeightMsgUpdateDidDocument int = 100
	)

	var weightMsgUpdateDidDocument int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateDidDocument, &weightMsgUpdateDidDocument, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateDidDocument = defaultWeightMsgUpdateDidDocument
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateDidDocument,
		identitysimulation.SimulateMsgUpdateDidDocument(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgDeleteDidDocument          = "op_weight_msg_identity"
		defaultWeightMsgDeleteDidDocument int = 100
	)

	var weightMsgDeleteDidDocument int
	simState.AppParams.GetOrGenerate(opWeightMsgDeleteDidDocument, &weightMsgDeleteDidDocument, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteDidDocument = defaultWeightMsgDeleteDidDocument
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteDidDocument,
		identitysimulation.SimulateMsgDeleteDidDocument(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
