package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	replyclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/reply_clarification"
)

type ReplyClarificationPresenter types.Presenter[replyclarification.Output, api.ReplyClarificationToUserResponseObject]

type replyClarificationPresenter struct{}

func NewReplyClarificationPresenter() ReplyClarificationPresenter {
	return &replyClarificationPresenter{}
}

func (p *replyClarificationPresenter) Output(output *replyclarification.Output) api.ReplyClarificationToUserResponseObject {
	return api.ReplyClarificationToUser200JSONResponse{
		Success: true,
		Data:    output.ID,
	}
}

func (p *replyClarificationPresenter) Error(err error) api.ReplyClarificationToUserResponseObject {
	return httperror.From(err).Response()
}
