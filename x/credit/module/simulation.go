package credit

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	creditsimulation "dtc/x/credit/simulation"
	"dtc/x/credit/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	creditGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&creditGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgMintCredit          = "op_weight_msg_credit"
		defaultWeightMsgMintCredit int = 100
	)

	var weightMsgMintCredit int
	simState.AppParams.GetOrGenerate(opWeightMsgMintCredit, &weightMsgMintCredit, nil,
		func(_ *rand.Rand) {
			weightMsgMintCredit = defaultWeightMsgMintCredit
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgMintCredit,
		creditsimulation.SimulateMsgMintCredit(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgSubmitDeathCertificate          = "op_weight_msg_credit"
		defaultWeightMsgSubmitDeathCertificate int = 100
	)

	var weightMsgSubmitDeathCertificate int
	simState.AppParams.GetOrGenerate(opWeightMsgSubmitDeathCertificate, &weightMsgSubmitDeathCertificate, nil,
		func(_ *rand.Rand) {
			weightMsgSubmitDeathCertificate = defaultWeightMsgSubmitDeathCertificate
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSubmitDeathCertificate,
		creditsimulation.SimulateMsgSubmitDeathCertificate(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
