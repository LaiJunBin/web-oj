package presenter

import (
	espresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/event_stream/system_status"
	jsonpresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/json/system_status"
)

type SystemStatusPresenter struct {
	GetSubmissionRate    jsonpresenter.GetSubmissionRatePresenter
	ListJudgmentErrors   jsonpresenter.ListJudgmentErrorsPresenter
	StreamSubmissionRate espresenter.StreamSubmissionRatePresenter
	StreamJudgmentErrors espresenter.StreamJudgmentErrorsPresenter
}

func NewSystemStatusPresenter(
	streamSubmissionRate espresenter.StreamSubmissionRatePresenter,
	streamJudgmentErrors espresenter.StreamJudgmentErrorsPresenter,
) *SystemStatusPresenter {
	return &SystemStatusPresenter{
		GetSubmissionRate:    jsonpresenter.NewGetSubmissionRatePresenter(),
		ListJudgmentErrors:   jsonpresenter.NewListJudgmentErrorsPresenter(),
		StreamSubmissionRate: streamSubmissionRate,
		StreamJudgmentErrors: streamJudgmentErrors,
	}
}
