package presenter

import (
	commonpresenter "judge-backend/internal/app/shared/adapter/presenter/common"
	jsonpresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/json/rejudge"
)

type RejudgePresenter struct {
	ListRejudges                   jsonpresenter.ListRejudgesPresenter
	ListRejudgesByContestProject   jsonpresenter.ListRejudgesByContestProjectPresenter
	GetRejudge                     jsonpresenter.GetRejudgePresenter
	GetRejudgeSubmissionComparison jsonpresenter.GetRejudgeSubmissionComparisonPresenter
	ApproveRejudge                 jsonpresenter.ApproveRejudgePresenter
	RejectRejudge                  jsonpresenter.RejectRejudgePresenter
	CancelRejudge                  *commonpresenter.JSONPresenter
}

func NewRejudgePresenter() *RejudgePresenter {
	return &RejudgePresenter{
		ListRejudges:                   jsonpresenter.NewListRejudgesPresenter(),
		ListRejudgesByContestProject:   jsonpresenter.NewListRejudgesByContestProjectPresenter(),
		GetRejudge:                     jsonpresenter.NewGetRejudgePresenter(),
		GetRejudgeSubmissionComparison: jsonpresenter.NewGetRejudgeSubmissionComparisonPresenter(),
		ApproveRejudge:                 jsonpresenter.NewApproveRejudgePresenter(),
		RejectRejudge:                  jsonpresenter.NewRejectRejudgePresenter(),
		CancelRejudge:                  commonpresenter.NewJSONPresenter(),
	}
}
