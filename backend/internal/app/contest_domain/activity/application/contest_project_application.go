package application

import (
	exportscoredetail "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/export_score_detail"
	getcontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_contest_project"
	getjudgmentanalytics "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_judgment_analytics"
	getuserscores "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_user_scores"
	rejudgecontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/rejudge_contest_project"
)

type ContestProjectApplication struct {
	getContestProject     getcontestproject.Usecase
	getUserScores         getuserscores.Usecase
	rejudgeContestProject rejudgecontestproject.Usecase
	exportScoreDetail     exportscoredetail.Usecase
	getJudgmentAnalytics  getjudgmentanalytics.Usecase
}

func NewContestProjectApplication(
	getContestProject getcontestproject.Usecase,
	getUserScores getuserscores.Usecase,
	rejudgeContestProject rejudgecontestproject.Usecase,
	exportScoreDetail exportscoredetail.Usecase,
	getJudgmentAnalytics getjudgmentanalytics.Usecase,
) *ContestProjectApplication {
	return &ContestProjectApplication{
		getContestProject:     getContestProject,
		getUserScores:         getUserScores,
		rejudgeContestProject: rejudgeContestProject,
		exportScoreDetail:     exportScoreDetail,
		getJudgmentAnalytics:  getJudgmentAnalytics,
	}
}

func (a *ContestProjectApplication) GetContestProject(input *getcontestproject.Input) (*getcontestproject.Output, error) {
	return a.getContestProject.Execute(input)
}

func (a *ContestProjectApplication) GetUserScores(input *getuserscores.Input) (*getuserscores.Output, error) {
	return a.getUserScores.Execute(input)
}

func (a *ContestProjectApplication) RejudgeContestProject(input *rejudgecontestproject.Input) (*rejudgecontestproject.Output, error) {
	return a.rejudgeContestProject.Execute(input)
}

func (a *ContestProjectApplication) ExportScoreDetail(input *exportscoredetail.Input) (*exportscoredetail.Output, error) {
	return a.exportScoreDetail.Execute(input)
}

func (a *ContestProjectApplication) GetJudgmentAnalytics(input *getjudgmentanalytics.Input) (*getjudgmentanalytics.Output, error) {
	return a.getJudgmentAnalytics.Execute(input)
}
