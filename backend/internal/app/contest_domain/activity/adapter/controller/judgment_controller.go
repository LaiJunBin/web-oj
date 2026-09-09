package activitycontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter"
	"judge-backend/internal/app/contest_domain/activity/application"
	canceljudge "judge-backend/internal/app/contest_domain/activity/application/usecase/judgment/cancel_judge"
	setstaffnote "judge-backend/internal/app/contest_domain/activity/application/usecase/judgment/set_staff_note"

	"github.com/gin-gonic/gin"
)

type JudgmentController struct {
	app       *application.JudgmentApplication
	presenter *presenter.JudgmentPresenter
}

func NewJudgmentController(
	app *application.JudgmentApplication,
	presenter *presenter.JudgmentPresenter,
) *JudgmentController {
	return &JudgmentController{
		app:       app,
		presenter: presenter,
	}
}

func (ctl *JudgmentController) CancelJudge(ctx *gin.Context, request api.CancelJudgeRequestObject) (api.CancelJudgeResponseObject, error) {
	input := &canceljudge.Input{
		JudgmentID: request.Id,
		Reason:     request.Body.Reason,
	}

	output, err := ctl.app.CancelJudge(input)
	if err != nil {
		return ctl.presenter.CancelJudge.Error(err), nil
	}

	return ctl.presenter.CancelJudge.Output(output), nil
}

func (ctl *JudgmentController) SetJudgmentStaffNote(ctx *gin.Context, request api.SetJudgmentStaffNoteRequestObject) (api.SetJudgmentStaffNoteResponseObject, error) {
	input := &setstaffnote.Input{
		JudgmentID: request.Id,
		StaffNote:  request.Body.StaffNote,
	}

	output, err := ctl.app.SetStaffNote(input)
	if err != nil {
		return ctl.presenter.SetStaffNote.Error(err), nil
	}

	return ctl.presenter.SetStaffNote.Output(output), nil
}
