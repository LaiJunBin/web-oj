package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	getuserclarification "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_clarification"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetUserClarificationPresenter types.Presenter[getuserclarification.Output, api.GetUserClarificationResponseObject]

type getUserClarificationPresenter struct{}

func NewGetUserClarificationPresenter() GetUserClarificationPresenter {
	return &getUserClarificationPresenter{}
}

func transformClarification(clarification *dto.ClarificationDetail) *api.UserUserSpecClarificationDetail {
	if clarification == nil {
		return nil
	}

	return &api.UserUserSpecClarificationDetail{
		Id:                 clarification.ID,
		Subject:            clarification.Subject,
		Content:            clarification.Content,
		IsAdminBroadcast:   clarification.IsAdminBroadcast,
		IsAdminResponse:    clarification.IsAdminResponse,
		CreatedAt:          clarification.CreatedAt,
		ReplyClarification: transformClarification(clarification.ReplyClarification),
	}
}

func (p *getUserClarificationPresenter) Output(output *getuserclarification.Output) api.GetUserClarificationResponseObject {
	return api.GetUserClarification200JSONResponse{
		Success: true,
		Data:    *transformClarification(output.Clarification),
	}
}

func (p *getUserClarificationPresenter) Error(err error) api.GetUserClarificationResponseObject {
	return httperror.From(err).Response()
}
