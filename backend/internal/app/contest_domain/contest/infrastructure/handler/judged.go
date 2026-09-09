package handler

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/events/scoreboardupdated"
	"judge-backend/internal/app/contest_domain/events/submissionjudged"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type JudgedHandler struct {
	eventbus sharedtypes.EventBus
	query    query.ProjectQuery
	logger   sharedtypes.Logger
}

func RegisterJudgedHandler(
	eventbus sharedtypes.EventBus,
	eventRepo sharedtypes.EventRepository,
	query query.ProjectQuery,
	logger sharedtypes.Logger,
) {
	handler := &JudgedHandler{
		eventbus: eventbus,
		query:    query,
		logger:   logger,
	}

	// Claimed subscription: read-and-publish, trivially idempotent, so
	// a transient scoreboard-query failure retries instead of the scoreboard
	// never learning about this judgment. The custom key keeps the historical
	// prefix: this event's ID is the judgment ID, which the judge domain's
	// own handlers also claim — an unprefixed key would collide with theirs.
	if err := sharedeventbus.SubscribeClaimedAsync(eventbus, eventRepo, logger, false,
		handler.Judged, sharedeventbus.ClaimedOptions[*submissionjudged.Event]{
			Key: func(e *submissionjudged.Event) string { return "judged-contest-" + e.ID() },
		}); err != nil {
		// Startup-time programming error (the handler is not a valid func):
		// the app must not come up half-wired.
		logger.Panicf("[JudgedHandler] Failed to subscribe: %v", err)
	}
}

func (h *JudgedHandler) Judged(e *submissionjudged.Event) error {
	h.logger.Debugf("Starting to handle judged event", "contestProjectId", e.ContestProjectID())

	scoreboard, err := h.query.GetProjectScoreboardByID(e.ContestProjectID())
	if err != nil {
		h.logger.Errorf("Failed to get project scoreboard", "contestProjectId", e.ContestProjectID(), "error", err)
		return err
	}

	h.eventbus.Publish(scoreboardupdated.New(e.ContestProjectID(), scoreboard))
	h.logger.Debugf("Successfully published scoreboard updated event", "contestProjectId", e.ContestProjectID())
	return nil
}
