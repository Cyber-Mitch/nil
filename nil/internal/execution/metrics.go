package execution

import (
	"context"

	"github.com/NilFoundation/nil/nil/internal/telemetry"
	"github.com/NilFoundation/nil/nil/internal/telemetry/telattr"
	"github.com/NilFoundation/nil/nil/internal/types"
	"go.opentelemetry.io/otel/metric"
)

type MetricsHandler struct {
	option metric.MeasurementOption

	measurer *telemetry.Measurer

	// Counters
	internalTxnCounter telemetry.Counter
	externalTxnCounter telemetry.Counter
	deployTxnCounter   telemetry.Counter
	execTxnCounter     telemetry.Counter

	// Gauges
	coinsUsed telemetry.Gauge
	gasPrice  telemetry.Gauge
	blockId   telemetry.Gauge
}

type BlockGeneratorCounters struct {
	InternalTransactions int64
	ExternalTransactions int64
	DeployTransactions   int64
	ExecTransactions     int64
	CoinsUsed            types.Value
}

func NewBlockGeneratorCounters() *BlockGeneratorCounters {
	return &BlockGeneratorCounters{
		CoinsUsed: types.NewZeroValue(),
	}
}

func NewMetricsHandler(name string, shardId types.ShardId) (*MetricsHandler, error) {
	meter := telemetry.NewMeter(name)
	measurer, err := telemetry.NewMeasurer(
		meter, "block_generation", telattr.ShardId(shardId),
	)
	if err != nil {
		return nil, err
	}

	handler := &MetricsHandler{
		measurer: measurer,
		option:   telattr.With(telattr.ShardId(shardId)),
	}

	if err := handler.initMetrics(meter); err != nil {
		return nil, err
	}

	return handler, nil
}

func (mh *MetricsHandler) initMetrics(meter metric.Meter) error {
	var err error

	// Initialize counters

	if mh.internalTxnCounter, err = meter.Int64Counter("internal_transactions_processed"); err != nil {
		return err
	}

	if mh.externalTxnCounter, err = meter.Int64Counter("external_transactions_processed"); err != nil {
		return err
	}

	if mh.deployTxnCounter, err = meter.Int64Counter("deploy_transactions_processed"); err != nil {
		return err
	}

	if mh.execTxnCounter, err = meter.Int64Counter("execution_transactions_processed"); err != nil {
		return err
	}

	// Initialize gauges
	if mh.gasPrice, err = meter.Int64Gauge("gas_price"); err != nil {
		return err
	}

	if mh.blockId, err = meter.Int64Gauge("block_id"); err != nil {
		return err
	}

	if mh.coinsUsed, err = meter.Int64Gauge("coins_used"); err != nil {
		return err
	}

	return nil
}

func (mh *MetricsHandler) RecordGasPrice(ctx context.Context, gasPrice types.Value) {
	mh.gasPrice.Record(ctx, int64(gasPrice.Uint64()), mh.option)
}

func (mh *MetricsHandler) StartProcessingMeasurement(ctx context.Context, blockId types.BlockNumber) {
	mh.measurer.Restart()
	mh.blockId.Record(ctx, int64(blockId), mh.option)
}

func (mh *MetricsHandler) EndProcessingMeasurement(ctx context.Context, counters *BlockGeneratorCounters) {
	mh.measurer.Measure(ctx)
	mh.internalTxnCounter.Add(ctx, counters.InternalTransactions, mh.option)
	mh.externalTxnCounter.Add(ctx, counters.ExternalTransactions, mh.option)
	mh.deployTxnCounter.Add(ctx, counters.DeployTransactions, mh.option)
	mh.execTxnCounter.Add(ctx, counters.ExecTransactions, mh.option)
	mh.coinsUsed.Record(ctx, int64(counters.CoinsUsed.Uint64()), mh.option)
}
