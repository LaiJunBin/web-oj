package activitycontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter"
	"judge-backend/internal/app/contest_domain/activity/application"
	getsubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submission"
	getsubmissionsbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submissions_by_contest_project"
	getsubmissionsbyuserid "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submissions_by_user_id"
	ignoresubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/ignore_submission"
	rejudgesubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/rejudge_submission"
	streamsubmissionsbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/stream_submissions_by_contest_project"
	unignoresubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/unignore_submission"

	"github.com/gin-gonic/gin"
)

type SubmissionController struct {
	app       *application.SubmissionApplication
	presenter *presenter.SubmissionPresenter
}

func NewSubmissionController(
	app *application.SubmissionApplication,
	presenter *presenter.SubmissionPresenter,
) *SubmissionController {
	return &SubmissionController{
		app:       app,
		presenter: presenter,
	}
}

func (ctl *SubmissionController) GetSubmissionsByContestProject(ctx *gin.Context, request api.GetSubmissionsByContestProjectRequestObject) (api.GetSubmissionsByContestProjectResponseObject, error) {
	input := &getsubmissionsbycontestproject.Input{
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.GetSubmissionsByContestProject(input)
	if err != nil {
		return ctl.presenter.GetSubmissionsByContestProject.Error(err), nil
	}

	return ctl.presenter.GetSubmissionsByContestProject.Output(output), nil
}

func (ctl *SubmissionController) GetSubmission(ctx *gin.Context, request api.GetSubmissionRequestObject) (api.GetSubmissionResponseObject, error) {
	input := &getsubmission.Input{
		SubmissionID: request.Id,
	}

	output, err := ctl.app.GetSubmission(input)
	if err != nil {
		return ctl.presenter.GetSubmission.Error(err), nil
	}

	return ctl.presenter.GetSubmission.Output(output), nil
}

func (ctl *SubmissionController) GetSubmissionsByUser(ctx *gin.Context, request api.GetSubmissionsByUserRequestObject) (api.GetSubmissionsByUserResponseObject, error) {
	input := &getsubmissionsbyuserid.Input{
		UserID: request.Id,
	}

	output, err := ctl.app.GetSubmissionsByUserID(input)
	if err != nil {
		return ctl.presenter.GetSubmissionsByUser.Error(err), nil
	}

	return ctl.presenter.GetSubmissionsByUser.Output(output), nil
}

func (ctl *SubmissionController) RejudgeSubmission(ctx *gin.Context, request api.RejudgeSubmissionRequestObject) (api.RejudgeSubmissionResponseObject, error) {
	input := &rejudgesubmission.Input{
		SubmissionID: request.Id,
		Reason:       request.Body.Reason,
	}

	output, err := ctl.app.RejudgeSubmission(input)
	if err != nil {
		return ctl.presenter.RejudgeSubmission.Error(err), nil
	}

	return ctl.presenter.RejudgeSubmission.Output(output), nil
}

func (ctl *SubmissionController) IgnoreSubmission(ctx *gin.Context, request api.IgnoreSubmissionRequestObject) (api.IgnoreSubmissionResponseObject, error) {
	input := &ignoresubmission.Input{
		SubmissionID: request.Id,
	}

	output, err := ctl.app.IgnoreSubmission(input)
	if err != nil {
		return ctl.presenter.IgnoreSubmission.Error(err), nil
	}

	return ctl.presenter.IgnoreSubmission.Output(output), nil
}

func (ctl *SubmissionController) StreamSubmissionsByContestProject(ctx *gin.Context, request api.StreamSubmissionsByContestProjectRequestObject) (api.StreamSubmissionsByContestProjectResponseObject, error) {
	input := &streamsubmissionsbycontestproject.Input{
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.StreamSubmissionsByContestProject(input)
	if err != nil {
		return ctl.presenter.StreamSubmissionsByContestProject.Error(err), nil
	}

	return ctl.presenter.StreamSubmissionsByContestProject.Output(ctx.Request.Context(), output), nil
}

func (ctl *SubmissionController) UnignoreSubmission(ctx *gin.Context, request api.UnignoreSubmissionRequestObject) (api.UnignoreSubmissionResponseObject, error) {
	input := &unignoresubmission.Input{
		SubmissionID: request.Id,
	}

	output, err := ctl.app.UnignoreSubmission(input)
	if err != nil {
		return ctl.presenter.UnignoreSubmission.Error(err), nil
	}

	return ctl.presenter.UnignoreSubmission.Output(output), nil
}
