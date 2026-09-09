package handler

import (
	"judge-backend/internal/app/contest_domain/events/submissionjudgingupdated"
	"judge-backend/internal/app/contest_domain/events/types"
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

func (h *JudgeUpdatedHandler) Handle(e *sharedevent.JudgeUpdated) {
	h.logger.Debugf("[JudgeUpdatedHandler] Translating shared event %s", e.ID())
	h.eventbus.Publish(submissionjudgingupdated.New(e.ID(), e.Data().ContestProjectID, toJudgmentData(e.Data())))
}

func toJudgmentData(d sharedevent.JudgeUpdatedData) types.JudgmentData {
	gradingItems := make([]types.GradingItemData, 0, len(d.GradingItems))
	for _, gi := range d.GradingItems {
		testcases := make([]types.TestcaseData, 0, len(gi.Testcases))
		for _, tc := range gi.Testcases {
			td := types.TestcaseData{
				ID:     tc.ID,
				Status: tc.Status,
			}
			if tc.HasResult {
				td.ExecTime = tc.ExecTime
				td.ExitCode = tc.ExitCode
				td.Stdout = tc.Stdout
				td.Stderr = tc.Stderr
			}
			testcases = append(testcases, td)
		}
		gradingItems = append(gradingItems, types.GradingItemData{
			ID:        gi.ID,
			Score:     gi.Score,
			Status:    gi.Status,
			Testcases: testcases,
		})
	}
	return types.JudgmentData{
		SubmissionID: d.SubmissionID,
		IsRejudge:    d.IsRejudge,
		Status:       d.Status,
		Score:        d.Score,
		Reason:       d.Reason,
		StaffNote:    d.StaffNote,
		GradingItems: gradingItems,
	}
}
