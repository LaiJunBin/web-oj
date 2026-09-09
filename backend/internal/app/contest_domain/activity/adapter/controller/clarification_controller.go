package activitycontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter"
	"judge-backend/internal/app/contest_domain/activity/application"
	getclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/get_clarification"
	listclarifications "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/list_clarifications"
	replyclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/reply_clarification"
	sendclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/send_clarification"
	updateclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/update_clarification"

	"github.com/gin-gonic/gin"
)

type ClarificationController struct {
	app       *application.ClarificationApplication
	presenter *presenter.ClarificationPresenter
}

func NewClarificationController(
	app *application.ClarificationApplication,
	presenter *presenter.ClarificationPresenter,
) *ClarificationController {
	return &ClarificationController{
		app:       app,
		presenter: presenter,
	}
}

func (ctl *ClarificationController) ListClarifications(ctx *gin.Context, request api.ListClarificationsRequestObject) (api.ListClarificationsResponseObject, error) {
	input := &listclarifications.Input{
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.ListClarifications(input)
	if err != nil {
		return ctl.presenter.ListClarifications.Error(err), nil
	}

	return ctl.presenter.ListClarifications.Output(output), nil
}

func (ctl *ClarificationController) GetClarification(ctx *gin.Context, request api.GetClarificationRequestObject) (api.GetClarificationResponseObject, error) {
	input := &getclarification.Input{
		ClarificationID: request.Id,
	}

	output, err := ctl.app.GetClarification(input)
	if err != nil {
		return ctl.presenter.GetClarification.Error(err), nil
	}

	return ctl.presenter.GetClarification.Output(output), nil
}

func (ctl *ClarificationController) userID(ctx *gin.Context) int32 {
	return ctx.MustGet(api.UserIDScopes).(int32)
}

func (ctl *ClarificationController) SendClarification(ctx *gin.Context, request api.SendClarificationRequestObject) (api.SendClarificationResponseObject, error) {
	input := &sendclarification.Input{
		ContestProjectID: request.Id,
		Subject:          request.Body.Subject,
		Content:          request.Body.Content,
		SenderUserID:     ctl.userID(ctx),
		ReceiverUserID:   request.Body.ReceiverUserId,
		SendToAll:        request.Body.SendToAll,
	}

	output, err := ctl.app.SendClarification(input)
	if err != nil {
		return ctl.presenter.SendClarification.Error(err), nil
	}

	return ctl.presenter.SendClarification.Output(output), nil
}

func (ctl *ClarificationController) ReplyClarificationToUser(ctx *gin.Context, request api.ReplyClarificationToUserRequestObject) (api.ReplyClarificationToUserResponseObject, error) {
	input := &replyclarification.Input{
		ID:           request.Id,
		Subject:      request.Body.Subject,
		Content:      request.Body.Content,
		SenderUserID: ctl.userID(ctx),
		SendToAll:    request.Body.SendToAll,
	}

	output, err := ctl.app.ReplyClarification(input)
	if err != nil {
		return ctl.presenter.ReplyClarification.Error(err), nil
	}

	return ctl.presenter.ReplyClarification.Output(output), nil
}

func (ctl *ClarificationController) MarkClarificationAsAnswered(ctx *gin.Context, request api.MarkClarificationAsAnsweredRequestObject) (api.MarkClarificationAsAnsweredResponseObject, error) {
	isAnswered := true
	input := &updateclarification.Input{
		ID:         request.Id,
		IsAnswered: &isAnswered,
	}

	output, err := ctl.app.UpdateClarification(input)
	if err != nil {
		return ctl.presenter.MarkClarificationAsAnswered.Error(err), nil
	}

	return ctl.presenter.MarkClarificationAsAnswered.Output(output), nil
}

func (ctl *ClarificationController) MarkClarificationAsUnanswered(ctx *gin.Context, request api.MarkClarificationAsUnansweredRequestObject) (api.MarkClarificationAsUnansweredResponseObject, error) {
	isAnswered := false
	input := &updateclarification.Input{
		ID:         request.Id,
		IsAnswered: &isAnswered,
	}

	output, err := ctl.app.UpdateClarification(input)
	if err != nil {
		return ctl.presenter.MarkClarificationAsUnanswered.Error(err), nil
	}

	return ctl.presenter.MarkClarificationAsUnanswered.Output(output), nil
}
