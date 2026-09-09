package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	listuserprojectclarifications "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_project_clarifications"
	"judge-backend/internal/app/shared/types/httperror"
)

type ListUserProjectClarificationsPresenter types.Presenter[listuserprojectclarifications.Output, api.ListUserProjectClarificationsResponseObject]

type listUserProjectClarificationsPresenter struct{}

func NewListUserProjectClarificationsPresenter() ListUserProjectClarificationsPresenter {
	return &listUserProjectClarificationsPresenter{}
}

func (p *listUserProjectClarificationsPresenter) Output(output *listuserprojectclarifications.Output) api.ListUserProjectClarificationsResponseObject {
	clarifications := make([]api.UserUserSpecClarification, 0, len(output.Clarifications))
	for _, clarification := range output.Clarifications {
		clarifications = append(clarifications, api.UserUserSpecClarification{
			Id:               clarification.ID,
			Subject:          clarification.Subject,
			Content:          clarification.Content,
			IsAdminBroadcast: clarification.IsAdminBroadcast,
			IsAdminResponse:  clarification.IsAdminResponse,
			CreatedAt:        clarification.CreatedAt,
		})
	}

	return api.ListUserProjectClarifications200JSONResponse{
		Success: true,
		Data:    clarifications,
	}
}

func (p *listUserProjectClarificationsPresenter) Error(err error) api.ListUserProjectClarificationsResponseObject {
	return httperror.From(err).Response()
}
