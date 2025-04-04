package metrics

import (
	"context"
	"time"

	"github.com/NilFoundation/nil/nil/internal/telemetry"
	"github.com/NilFoundation/nil/nil/internal/telemetry/telattr"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	attrTaskType     = "task.type"
	attrTaskExecutor = "task.executor.id"
)

type taskStorageMetricsHandler struct {
	attributes metric.MeasurementOption

	currentActiveTasks  telemetry.UpDownCounter
	currentPendingTasks telemetry.UpDownCounter

	totalTasksCreated     telemetry.Counter
	totalTasksSucceeded   telemetry.Counter
	totalTasksRescheduled telemetry.Counter
	totalTasksFailed      telemetry.Counter

	taskExecutionTimeMs telemetry.Histogram
}

func (h *taskStorageMetricsHandler) init(attributes metric.MeasurementOption, meter telemetry.Meter) error {
	h.attributes = attributes
	var err error
	const tasksNamespace = namespace + "tasks."

	if h.currentActiveTasks, err = meter.Int64UpDownCounter(tasksNamespace + "current_active"); err != nil {
		return err
	}

	if h.currentPendingTasks, err = meter.Int64UpDownCounter(tasksNamespace + "current_pending"); err != nil {
		return err
	}

	if h.totalTasksCreated, err = meter.Int64Counter(tasksNamespace + "total_created"); err != nil {
		return err
	}

	if h.totalTasksSucceeded, err = meter.Int64Counter(tasksNamespace + "total_succeeded"); err != nil {
		return err
	}

	if h.totalTasksRescheduled, err = meter.Int64Counter(tasksNamespace + "total_rescheduled"); err != nil {
		return err
	}

	if h.totalTasksFailed, err = meter.Int64Counter(tasksNamespace + "total_failed"); err != nil {
		return err
	}

	if h.taskExecutionTimeMs, err = meter.Int64Histogram(tasksNamespace + "execution_time_ms"); err != nil {
		return err
	}

	return nil
}

func (h *taskStorageMetricsHandler) RecordTaskAdded(ctx context.Context, taskEntry *types.TaskEntry) {
	taskAttributes := h.getAttrTypeOnly(taskEntry)
	h.totalTasksCreated.Add(ctx, 1, h.attributes, taskAttributes)
	h.currentPendingTasks.Add(ctx, 1, h.attributes, taskAttributes)
}

func (h *taskStorageMetricsHandler) RecordTaskStarted(ctx context.Context, taskEntry *types.TaskEntry) {
	pendingAttributes := h.getAttrTypeOnly(taskEntry)
	h.currentPendingTasks.Add(ctx, -1, h.attributes, pendingAttributes)

	activeAttributes := h.getAttrTypeAndOwner(taskEntry)
	h.currentActiveTasks.Add(ctx, 1, h.attributes, activeAttributes)
}

func (h *taskStorageMetricsHandler) RecordTaskTerminated(
	ctx context.Context,
	taskEntry *types.TaskEntry,
	taskResult *types.TaskResult,
) {
	taskAttributes := h.getAttrTypeAndOwner(taskEntry)

	h.currentActiveTasks.Add(ctx, -1, h.attributes, taskAttributes)

	if taskResult.IsSuccess() {
		executionTimeMs := time.Since(*taskEntry.Started).Milliseconds()
		h.taskExecutionTimeMs.Record(ctx, executionTimeMs, h.attributes, taskAttributes)
		h.totalTasksSucceeded.Add(ctx, 1, h.attributes, taskAttributes)
	} else {
		h.totalTasksFailed.Add(ctx, 1, h.attributes, taskAttributes)
	}
}

func (h *taskStorageMetricsHandler) RecordTaskRescheduled(
	ctx context.Context,
	taskType types.TaskType,
	previousExecutor types.TaskExecutorId,
) {
	taskAttributes := telattr.With(
		attribute.Stringer(attrTaskType, taskType),
		attribute.Int64(attrTaskExecutor, int64(previousExecutor)),
	)

	h.totalTasksRescheduled.Add(ctx, 1, h.attributes, taskAttributes)
	h.currentActiveTasks.Add(ctx, -1, h.attributes, taskAttributes)

	pendingAttributes := telattr.With(
		attribute.Stringer(attrTaskType, taskType),
	)
	h.currentPendingTasks.Add(ctx, 1, h.attributes, pendingAttributes)
}

func (h *taskStorageMetricsHandler) getAttrTypeOnly(taskEntry *types.TaskEntry) metric.MeasurementOption {
	return telattr.With(
		attribute.Stringer(attrTaskType, taskEntry.Task.TaskType),
	)
}

func (h *taskStorageMetricsHandler) getAttrTypeAndOwner(taskEntry *types.TaskEntry) metric.MeasurementOption {
	return telattr.With(
		attribute.Stringer(attrTaskType, taskEntry.Task.TaskType),
		attribute.Int64(attrTaskExecutor, int64(taskEntry.Owner)),
	)
}
