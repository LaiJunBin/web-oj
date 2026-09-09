package handler

import (
	"judge-backend/internal/app/judge_domain/events/judgecancelled"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/worker"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type JudgeCancelledHandler struct {
	workerPool *worker.JudgeWorkerPool
	logger     sharedtypes.Logger
}

func RegisterJudgeCancelledHandler(
	eventbus sharedtypes.EventBus,
	eventRepo sharedtypes.EventRepository,
	workerPool *worker.JudgeWorkerPool,
	logger sharedtypes.Logger,
) {
	handler := &JudgeCancelledHandler{
		workerPool: workerPool,
		logger:     logger,
	}

	// Claimed subscription: the handler body has no failure path (pure
	// in-memory worker scan), so the retry side is dormant — the claim keeps
	// its historical dedupe role. CancelJudgeIfCurrent is atomic either way.
	if err := sharedeventbus.SubscribeClaimedAsync(eventbus, eventRepo, logger, false,
		handler.JudgeCancelled, sharedeventbus.ClaimedOptions[*judgecancelled.JudgeCancelledEvent]{}); err != nil {
		// Startup-time programming error (the handler is not a valid func):
		// the app must not come up half-wired.
		logger.Panicf("[JudgeCancelledHandler] Failed to subscribe: %v", err)
	}
	logger.Infof("[JudgeCancelledHandler] Registered judge cancelled handler")
}

func (h *JudgeCancelledHandler) JudgeCancelled(event *judgecancelled.JudgeCancelledEvent) error {
	h.logger.Debugf("[JudgeCancelledHandler] Processing judge cancelled event for judgment %s", event.JudgmentID())

	for _, worker := range h.workerPool.ListJudgeWorkers() {
		// CancelJudgeIfCurrent re-checks CurrentJudgeID itself, atomically
		// with the cancel: the earlier CurrentJudgeID() here is only a
		// cheap pre-filter to pick which worker to try, not the actual
		// guard. Without the atomic re-check, this worker could have
		// finished event.JudgmentID() and already claimed a DIFFERENT
		// judgment by the time the cancel call below actually runs — a
		// plain CancelJudge() would then cancel that new judgment instead.
		if worker.CurrentJudgeID() != event.JudgmentID() {
			continue
		}
		if !worker.CancelJudgeIfCurrent(event.JudgmentID()) {
			h.logger.Debugf("[JudgeCancelledHandler] Judgment %s was no longer current on worker %s by the time cancellation ran, skipping", event.JudgmentID(), worker.ID())
			continue
		}
		h.logger.Debugf("[JudgeCancelledHandler] Found worker handling judgment %s, cancelling", event.JudgmentID())
		worker.AppendLog("--- Judge cancelled ---")
		h.logger.Infof("[JudgeCancelledHandler] Successfully cancelled judgment %s", event.JudgmentID())
		return nil
	}
	h.logger.Debugf("[JudgeCancelledHandler] No worker found handling judgment %s", event.JudgmentID())
	return nil
}
