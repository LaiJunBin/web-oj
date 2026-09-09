package activitycontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter"
	"judge-backend/internal/app/contest_domain/activity/application"
	exportscoredetail "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/export_score_detail"
	getcontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_contest_project"
	getjudgmentanalytics "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_judgment_analytics"
	getuserscores "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_user_scores"
	rejudgecontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/rejudge_contest_project"

	"github.com/gin-gonic/gin"
)

type ContestProjectController struct {
	app       *application.ContestProjectApplication
	presenter *presenter.ContestProjectPresenter
}

func NewContestProjectController(
	app *application.ContestProjectApplication,
	presenter *presenter.ContestProjectPresenter,
) *ContestProjectController {
	return &ContestProjectController{
		app:       app,
		presenter: presenter,
	}
}

func (ctl *ContestProjectController) RejudgeContestProject(ctx *gin.Context, request api.RejudgeContestProjectRequestObject) (api.RejudgeContestProjectResponseObject, error) {
	input := &rejudgecontestproject.Input{
		ContestProjectID: request.Id,
		GradingItemIds:   request.Body.GradingItemIds,
		Reason:           request.Body.Reason,
	}

	output, err := ctl.app.RejudgeContestProject(input)
	if err != nil {
		return ctl.presenter.RejudgeContestProject.Error(err), nil
	}

	return ctl.presenter.RejudgeContestProject.Output(output), nil
}

// This is a placeholder, may be removed in the future
func (ctl *ContestProjectController) GetContestProject(ctx *gin.Context, request api.GetContestProjectRequestObject) (api.GetContestProjectResponseObject, error) {
	input := &getcontestproject.Input{
		ID: request.Id,
	}

	output, err := ctl.app.GetContestProject(input)
	if err != nil {
		return ctl.presenter.GetContestProject.Error(err), nil
	}

	return ctl.presenter.GetContestProject.Output(output), nil
}

func (ctl *ContestProjectController) GetUserScores(ctx *gin.Context, request api.GetUserScoresRequestObject) (api.GetUserScoresResponseObject, error) {
	input := &getuserscores.Input{
		ID: request.Id,
	}

	output, err := ctl.app.GetUserScores(input)
	if err != nil {
		return ctl.presenter.GetUserScores.Error(err), nil
	}

	return ctl.presenter.GetUserScores.Output(output), nil
}

func (ctl *ContestProjectController) ExportContestProjectScoreDetail(ctx *gin.Context, request api.ExportContestProjectScoreDetailRequestObject) (api.ExportContestProjectScoreDetailResponseObject, error) {
	input := &exportscoredetail.Input{
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.ExportScoreDetail(input)
	if err != nil {
		return ctl.presenter.ExportScoreDetail.Error(err), nil
	}

	return ctl.presenter.ExportScoreDetail.Output(output), nil
}

func (ctl *ContestProjectController) GetContestProjectJudgmentAnalytics(ctx *gin.Context, request api.GetContestProjectJudgmentAnalyticsRequestObject) (api.GetContestProjectJudgmentAnalyticsResponseObject, error) {
	input := &getjudgmentanalytics.Input{
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.GetJudgmentAnalytics(input)
	if err != nil {
		return ctl.presenter.GetJudgmentAnalytics.Error(err), nil
	}

	return ctl.presenter.GetJudgmentAnalytics.Output(output), nil
}
