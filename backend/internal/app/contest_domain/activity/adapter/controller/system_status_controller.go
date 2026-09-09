package activitycontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter"
	espresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/event_stream/system_status"
	"judge-backend/internal/app/contest_domain/activity/application"
	getsubmissionrate "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/get_submission_rate"
	listjudgmenterrors "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/list_judgment_errors"

	"github.com/gin-gonic/gin"
)

// SystemStatusController serves the whole-instance views the admin status page
// reads. The two operations are declared on different spec interfaces
// (submissions and judgments) because that is where their models live, but they
// answer the same question — how is the system doing right now — so they share a
// controller rather than being scattered across the resource controllers.
type SystemStatusController struct {
	app       *application.SystemStatusApplication
	presenter *presenter.SystemStatusPresenter
}

func NewSystemStatusController(
	app *application.SystemStatusApplication,
	presenter *presenter.SystemStatusPresenter,
) *SystemStatusController {
	return &SystemStatusController{
		app:       app,
		presenter: presenter,
	}
}

func (ctl *SystemStatusController) GetSubmissionRate(ctx *gin.Context, request api.GetSubmissionRateRequestObject) (api.GetSubmissionRateResponseObject, error) {
	input := &getsubmissionrate.Input{
		Minutes:          request.Params.Minutes,
		ContestProjectID: request.Params.ContestProjectId,
	}

	output, err := ctl.app.GetSubmissionRate(input)
	if err != nil {
		return ctl.presenter.GetSubmissionRate.Error(err), nil
	}

	return ctl.presenter.GetSubmissionRate.Output(output), nil
}

func (ctl *SystemStatusController) ListJudgmentErrors(ctx *gin.Context, request api.ListJudgmentErrorsRequestObject) (api.ListJudgmentErrorsResponseObject, error) {
	input := &listjudgmenterrors.Input{
		Limit: request.Params.Limit,
	}

	output, err := ctl.app.ListJudgmentErrors(input)
	if err != nil {
		return ctl.presenter.ListJudgmentErrors.Error(err), nil
	}

	return ctl.presenter.ListJudgmentErrors.Output(output), nil
}

// The stream endpoints run their usecase once here, before the stream opens.
// That is what makes a rejected parameter answer 400: a stream that validated
// on its first tick instead would return 200, emit nothing, and close — which
// a client cannot tell apart from a system with nothing to report. The result
// is handed to the presenter rather than discarded so the connection can emit
// immediately without repeating the query.

func (ctl *SystemStatusController) StreamSubmissionRate(ctx *gin.Context, request api.StreamSubmissionRateRequestObject) (api.StreamSubmissionRateResponseObject, error) {
	input := &getsubmissionrate.Input{
		Minutes:          request.Params.Minutes,
		ContestProjectID: request.Params.ContestProjectId,
	}

	first, err := ctl.app.GetSubmissionRate(input)
	if err != nil {
		return ctl.presenter.StreamSubmissionRate.Error(err), nil
	}

	return ctl.presenter.StreamSubmissionRate.Output(ctx.Request.Context(), &espresenter.StreamSubmissionRateInput{
		Input: input,
		First: first,
	}), nil
}

func (ctl *SystemStatusController) StreamJudgmentErrors(ctx *gin.Context, request api.StreamJudgmentErrorsRequestObject) (api.StreamJudgmentErrorsResponseObject, error) {
	input := &listjudgmenterrors.Input{
		Limit: request.Params.Limit,
	}

	first, err := ctl.app.ListJudgmentErrors(input)
	if err != nil {
		return ctl.presenter.StreamJudgmentErrors.Error(err), nil
	}

	return ctl.presenter.StreamJudgmentErrors.Output(ctx.Request.Context(), &espresenter.StreamJudgmentErrorsInput{
		Input: input,
		First: first,
	}), nil
}
