package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	sendclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/send_clarification"
)

type SendClarificationPresenter types.Presenter[sendclarification.Output, api.SendClarificationResponseObject]

type sendClarificationPresenter struct{}

func NewSendClarificationPresenter() SendClarificationPresenter {
	return &sendClarificationPresenter{}
}

func (p *sendClarificationPresenter) Output(output *sendclarification.Output) api.SendClarificationResponseObject {
	return api.SendClarification200JSONResponse{
		Success: true,
		Data:    output.ID,
	}
}

func (p *sendClarificationPresenter) Error(err error) api.SendClarificationResponseObject {
	return httperror.From(err).Response()
}
