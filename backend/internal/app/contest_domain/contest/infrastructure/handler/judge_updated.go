package handler

import (
	"judge-backend/internal/app/contest_domain/events/submissionjudgingupdated"
	"judge-backend/internal/app/contest_domain/events/submissionupdated"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type JudgeUpdatedHandler struct {
	eventbus sharedtypes.EventBus
	logger   sharedtypes.Logger
}

func RegisterJudgeUpdatedHandler(
	eventbus sharedtypes.EventBus,
	logger sharedtypes.Logger,
) {
	handler := &JudgeUpdatedHandler{
		eventbus: eventbus,
		logger:   logger,
	}

	eventbus.SubscribeAsync(handler.JudgingUpdated, false)
}

// JudgingUpdated deliberately does NOT claim/dedupe the event: a judgment
// publishes many judging updates over its life and every one of them must
// reach the submission-updated stream (per-testcase progress, score changes).
// Claiming by event ID here used to swallow every update after the first per
// judgment, which starved the live submission feed mid-run.
func (h *JudgeUpdatedHandler) JudgingUpdated(e *submissionjudgingupdated.Event) {
	h.logger.Debugf("Starting to handle judging updated event", "judgmentId", e.ID())

	h.eventbus.Publish(submissionupdated.New(e.ContestProjectID(), e.Judgment()))
	h.logger.Debugf("Successfully published submission updated event", "judgmentId", e.ID())
}
