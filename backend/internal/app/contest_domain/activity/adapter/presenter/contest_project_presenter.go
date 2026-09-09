package presenter

import (
	filepresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/file/contest_project"
	jsonpresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/json/contest_project"
)

type ContestProjectPresenter struct {
	GetContestProject     jsonpresenter.GetContestProjectPresenter
	GetUserScores         jsonpresenter.GetUserScoresPresenter
	RejudgeContestProject jsonpresenter.RejudgeContestProjectPresenter
	ExportScoreDetail     filepresenter.ExportScoreDetailPresenter
	GetJudgmentAnalytics  jsonpresenter.GetJudgmentAnalyticsPresenter
}

func NewContestProjectPresenter() *ContestProjectPresenter {
	return &ContestProjectPresenter{
		GetContestProject:     jsonpresenter.NewGetContestProjectPresenter(),
		GetUserScores:         jsonpresenter.NewGetUserScoresPresenter(),
		RejudgeContestProject: jsonpresenter.NewRejudgeContestProjectPresenter(),
		ExportScoreDetail:     filepresenter.NewExportScoreDetailPresenter(),
		GetJudgmentAnalytics:  jsonpresenter.NewGetJudgmentAnalyticsPresenter(),
	}
}
