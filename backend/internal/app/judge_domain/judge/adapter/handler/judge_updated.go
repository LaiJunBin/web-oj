package handler

import (
	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	sharedevent "judge-backend/internal/app/shared/events"
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
	h := &JudgeUpdatedHandler{eventbus: eventbus, logger: logger}
	eventbus.SubscribeAsync(h.Handle, false)
	logger.Infof("[JudgeUpdatedHandler] Registered")
}

func (h *JudgeUpdatedHandler) Handle(e *judgeupdated.JudgeUpdatedEvent) {
	h.logger.Debugf("[JudgeUpdatedHandler] Translating judge updated event %s", e.ID())
	d := e.Data()
	gradingItems := make([]sharedevent.JudgeUpdatedGradingItemData, 0, len(d.GradingItems))
	for _, gi := range d.GradingItems {
		testcases := make([]sharedevent.JudgeUpdatedTestcaseData, 0, len(gi.Testcases))
		for _, tc := range gi.Testcases {
			testcases = append(testcases, sharedevent.JudgeUpdatedTestcaseData{
				ID:        tc.ID,
				Status:    tc.Status,
				ExecTime:  tc.ExecTime,
				ExitCode:  tc.ExitCode,
				Stdout:    tc.Stdout,
				Stderr:    tc.Stderr,
				HasResult: tc.HasResult,
			})
		}
		gradingItems = append(gradingItems, sharedevent.JudgeUpdatedGradingItemData{
			ID:        gi.ID,
			Score:     gi.Score,
			Status:    gi.Status,
			Testcases: testcases,
		})
	}
	h.eventbus.Publish(sharedevent.NewJudgeUpdated(sharedevent.JudgeUpdatedData{
		JudgmentID:        d.JudgmentID,
		RejudgeID:         d.RejudgeID,
		ContestProjectID:  d.ContestProjectID,
		SubmissionID:      d.SubmissionID,
		IsRejudge:         d.IsRejudge,
		Status:            d.Status,
		StatusIsError:     d.StatusIsError,
		Reason:            d.Reason,
		StaffNote:         d.StaffNote,
		StatusIsCancelled: d.StatusIsCancelled,
		StatusIsRejudged:  d.StatusIsRejudged,
		Score:             d.Score,
		GradingItems:      gradingItems,
	}))
}
