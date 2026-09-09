package presenter

import (
	commonpresenter "judge-backend/internal/app/shared/adapter/presenter/common"
	espresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/event_stream/submission"
	jsonpresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/json/submission"
	"judge-backend/internal/app/contest_domain/activity/application/query"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type SubmissionPresenter struct {
	GetSubmissionsByContestProject    jsonpresenter.GetSubmissionsByContestProjectPresenter
	StreamSubmissionsByContestProject espresenter.StreamSubmissionsByContestProjectPresenter
	GetSubmission                     jsonpresenter.GetSubmissionPresenter
	GetSubmissionsByUser              jsonpresenter.GetSubmissionsByUserPresenter
	RejudgeSubmission                 jsonpresenter.RejudgeSubmissionPresenter
	IgnoreSubmission                  *commonpresenter.JSONPresenter
	UnignoreSubmission                *commonpresenter.JSONPresenter
}

func NewSubmissionPresenter(eventBus sharedtypes.EventBus, submissionQuery query.SubmissionQuery) *SubmissionPresenter {
	return &SubmissionPresenter{
		GetSubmissionsByContestProject:    jsonpresenter.NewGetSubmissionsByContestProjectPresenter(),
		StreamSubmissionsByContestProject: espresenter.NewStreamSubmissionsByContestProjectPresenter(eventBus, submissionQuery),
		GetSubmission:                     jsonpresenter.NewGetSubmissionPresenter(),
		GetSubmissionsByUser:              jsonpresenter.NewGetSubmissionsByUserPresenter(),
		RejudgeSubmission:                 jsonpresenter.NewRejudgeSubmissionPresenter(),
		IgnoreSubmission:                  commonpresenter.NewJSONPresenter(),
		UnignoreSubmission:                commonpresenter.NewJSONPresenter(),
	}
}
