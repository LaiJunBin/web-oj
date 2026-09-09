package handler

import (
	"judge-backend/internal/app/judge_domain/events/judgmentcreated"
	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	"judge-backend/internal/app/judge_domain/judge/domain/query"
	"judge-backend/internal/app/judge_domain/judge/domain/queue"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type JudgmentCreatedHandler struct {
	priorityStrategyQuery    query.PriorityStrategyQuery
	maxJudgeDurationQuery    query.MaxJudgeDurationQuery
	priorityStrategyProvider prioritystrategies.Provider
	queueProvider            queue.JudgeQueueProvider
	logger                   sharedtypes.Logger
}

func RegisterJudgmentCreatedHandler(
	eventBus sharedtypes.EventBus,
	eventRepo sharedtypes.EventRepository,
	priorityStrategyQuery query.PriorityStrategyQuery,
	maxJudgeDurationQuery query.MaxJudgeDurationQuery,
	priorityStrategyProvider prioritystrategies.Provider,
	queueProvider queue.JudgeQueueProvider,
	logger sharedtypes.Logger,
) {
	handler := &JudgmentCreatedHandler{
		priorityStrategyQuery:    priorityStrategyQuery,
		maxJudgeDurationQuery:    maxJudgeDurationQuery,
		priorityStrategyProvider: priorityStrategyProvider,
		queueProvider:            queueProvider,
		logger:                   logger,
	}
	// Claimed subscription. transactional=true is load-bearing and must
	// stay: enqueues for the same judge type are serialized. The only
	// retryable failure (GetQueue) happens BEFORE Push, so a retry can never
	// double-enqueue. Cost of transactional + blocking retries: the handler
	// lock is taken on the PUBLISHER's goroutine inside Publish(), so while a
	// (rare) retry chain sleeps, the next judgment_created publisher — an
	// event-handler goroutine, never an HTTP request — blocks in Publish().
	if err := sharedeventbus.SubscribeClaimedAsync(eventBus, eventRepo, logger, true,
		handler.DistributeJudgmentToQueue, sharedeventbus.ClaimedOptions[*judgmentcreated.JudgmentCreatedEvent]{}); err != nil {
		// Startup-time programming error (the handler is not a valid func):
		// the app must not come up half-wired.
		logger.Panicf("[JudgmentCreatedHandler] Failed to subscribe: %v", err)
	}
	logger.Infof("[JudgmentCreatedHandler] Registered judgment created handler")
}

func (h *JudgmentCreatedHandler) DistributeJudgmentToQueue(event *judgmentcreated.JudgmentCreatedEvent) error {
	h.logger.Debugf("[JudgmentCreatedHandler] Processing judgment created event for judgment %s", event.Judgment().ID())

	strategyName, err := h.priorityStrategyQuery.GetStrategyNameByContestProjectID(event.Judgment().ContestProjectID())
	if err != nil {
		h.logger.Errorf("[JudgmentCreatedHandler] Failed to get strategy name for contest project %d: %v", event.Judgment().ContestProjectID(), err)
	}

	strategy, err := h.priorityStrategyProvider.GetStrategy(strategyName)
	if err != nil {
		h.logger.Errorf("[JudgmentCreatedHandler] Failed to get priority strategy %s: %v, using default", strategyName, err)
		strategy = prioritystrategies.NewDefault()
	}

	// Resolved here — at every (re)enqueue, restarts included — so the worker
	// never touches the DB under its own mutex; lookup failure degrades to
	// the global default instead of blocking judging.
	maxDuration, err := h.maxJudgeDurationQuery.GetMaxJudgeDurationMinutesByContestProjectID(event.Judgment().ContestProjectID())
	if err != nil {
		h.logger.Errorf("[JudgmentCreatedHandler] Failed to get max judge duration for contest project %d: %v, using global default", event.Judgment().ContestProjectID(), err)
	} else {
		event.Judgment().SetMaxJudgeDurationMinutes(maxDuration)
	}

	q, err := h.queueProvider.GetQueue(event.Judgment().JudgeType().String())
	if err != nil {
		// Retryable: this happens BEFORE Push, so a retry cannot
		// double-enqueue. Before, the judgment stayed QUEUED in the DB but
		// in no queue until the next restart.
		h.logger.Errorf("[JudgmentCreatedHandler] Failed to get queue for judge type %s: %v", event.Judgment().JudgeType().String(), err)
		return err
	}

	q.Push(event.Judgment(), strategy)
	h.logger.Infof("[JudgmentCreatedHandler] Successfully pushed judgment %s to queue", event.Judgment().ID())
	return nil
}
