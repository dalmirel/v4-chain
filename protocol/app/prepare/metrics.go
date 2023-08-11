package prepare

import (
	gometrics "github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/dydxprotocol/v4/lib/metrics"
)

const (
	ModuleName = "prepare_proposal"
)

// successMetricParams defines params needed for reporting a successful `PrepareProposal`.
type successMetricParams struct {
	txs                 PrepareProposalTxs
	pricesTx            PricesTxResponse
	fundingTx           FundingTxResponse
	operationsTx        OperationsTxResponse
	numTxsToReturn      int
	numTxsInOriginalReq int
}

// recordErrorMetricsWithLabel records an error metric in `PrepareProposalHandler` with a label.
func recordErrorMetricsWithLabel(label string) {
	telemetry.IncrCounterWithLabels(
		[]string{ModuleName, metrics.Handler, metrics.Error, metrics.Count},
		1,
		[]gometrics.Label{metrics.GetLabelForStringValue(metrics.Detail, label)},
	)
}

// recordSuccessMetrics records a success metric details for `PrepareProposalHandler`.
func recordSuccessMetrics(params successMetricParams) {
	// Record success.
	telemetry.IncrCounter(
		1,
		ModuleName,
		metrics.Handler,
		metrics.Success,
		metrics.Count,
	)

	// Prices tx.
	telemetry.SetGauge(
		float32(params.pricesTx.NumMarkets),
		ModuleName,
		metrics.NumMarketPricesToUpdate,
	)

	// Funding tx.
	// TODO(DEC-1254): add more metrics for Funding tx.
	telemetry.SetGauge(
		float32(params.fundingTx.NumVotes),
		ModuleName,
		metrics.NumPremiumVotes,
	)

	// Other txs.
	telemetry.SetGauge(
		float32(len(params.txs.OtherTxs)),
		ModuleName,
		metrics.NumOtherTxs,
	)

	// Total # of txs in the new proposal.
	telemetry.SetGauge(
		float32(params.numTxsToReturn),
		ModuleName,
		metrics.TotalNumTxs,
	)

	// Total # of bytes in txs in the new proposal.
	telemetry.SetGauge(
		float32(params.txs.UsedBytes),
		ModuleName,
		metrics.TotalNumBytes,
	)

	// Total # of txs in the original req.
	telemetry.SetGauge(
		float32(params.numTxsInOriginalReq),
		ModuleName,
		metrics.OriginalNumTxs,
	)
}
