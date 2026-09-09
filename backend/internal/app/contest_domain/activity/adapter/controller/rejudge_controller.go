package activitycontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter"
	"judge-backend/internal/app/contest_domain/activity/application"
	approverejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/approve_rejudge"
	cancelrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/cancel_rejudge"
	getrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/get_rejudge"
	getrejudgesubmissioncomparison "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/get_rejudge_submission_comparison"
	listrejudges "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/list_rejudges"
	listrejudgesbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/list_rejudges_by_contest_project"
	rejectrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/reject_rejudge"

	"github.com/gin-gonic/gin"
)

type RejudgeController struct {
	app       *application.RejudgeApplication
	presenter *presenter.RejudgePresenter
}

func NewRejudgeController(
	app *application.RejudgeApplication,
	presenter *presenter.RejudgePresenter,
) *RejudgeController {
	return &RejudgeController{
		app:       app,
		presenter: presenter,
	}
}

func (ctl *RejudgeController) ListRejudges(ctx *gin.Context, request api.ListRejudgesRequestObject) (api.ListRejudgesResponseObject, error) {
	input := &listrejudges.Input{}

	output, err := ctl.app.ListRejudges(input)
	if err != nil {
		return ctl.presenter.ListRejudges.Error(err), nil
	}

	return ctl.presenter.ListRejudges.Output(output), nil
}

func (ctl *RejudgeController) ListRejudgesByContestProject(ctx *gin.Context, request api.ListRejudgesByContestProjectRequestObject) (api.ListRejudgesByContestProjectResponseObject, error) {
	input := &listrejudgesbycontestproject.Input{
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.ListRejudgesByContestProject(input)
	if err != nil {
		return ctl.presenter.ListRejudgesByContestProject.Error(err), nil
	}

	return ctl.presenter.ListRejudgesByContestProject.Output(output), nil
}

func (ctl *RejudgeController) GetRejudge(ctx *gin.Context, request api.GetRejudgeRequestObject) (api.GetRejudgeResponseObject, error) {
	input := &getrejudge.Input{
		RejudgeID: request.Id,
	}

	output, err := ctl.app.GetRejudge(input)
	if err != nil {
		return ctl.presenter.GetRejudge.Error(err), nil
	}

	return ctl.presenter.GetRejudge.Output(output), nil
}

func (ctl *RejudgeController) GetRejudgeSubmissionComparison(ctx *gin.Context, request api.GetRejudgeSubmissionComparisonRequestObject) (api.GetRejudgeSubmissionComparisonResponseObject, error) {
	input := &getrejudgesubmissioncomparison.Input{
		RejudgeID:    request.Id,
		SubmissionID: request.SubmissionId,
	}

	output, err := ctl.app.GetRejudgeSubmissionComparison(input)
	if err != nil {
		return ctl.presenter.GetRejudgeSubmissionComparison.Error(err), nil
	}

	return ctl.presenter.GetRejudgeSubmissionComparison.Output(output), nil
}

func (ctl *RejudgeController) CancelRejudge(ctx *gin.Context, request api.CancelRejudgeRequestObject) (api.CancelRejudgeResponseObject, error) {
	input := &cancelrejudge.Input{
		RejudgeID: request.Id,
	}

	output, err := ctl.app.CancelRejudge(input)
	if err != nil {
		return ctl.presenter.CancelRejudge.Error(err), nil
	}

	return ctl.presenter.CancelRejudge.Output(output), nil
}

func (ctl *RejudgeController) ApproveRejudge(ctx *gin.Context, request api.ApproveRejudgeRequestObject) (api.ApproveRejudgeResponseObject, error) {
	input := &approverejudge.Input{
		RejudgeID: request.Id,
	}

	output, err := ctl.app.ApproveRejudge(input)
	if err != nil {
		return ctl.presenter.ApproveRejudge.Error(err), nil
	}

	return ctl.presenter.ApproveRejudge.Output(output), nil
}

func (ctl *RejudgeController) RejectRejudge(ctx *gin.Context, request api.RejectRejudgeRequestObject) (api.RejectRejudgeResponseObject, error) {
	input := &rejectrejudge.Input{
		RejudgeID: request.Id,
	}

	output, err := ctl.app.RejectRejudge(input)
	if err != nil {
		return ctl.presenter.RejectRejudge.Error(err), nil
	}

	return ctl.presenter.RejectRejudge.Output(output), nil
}
