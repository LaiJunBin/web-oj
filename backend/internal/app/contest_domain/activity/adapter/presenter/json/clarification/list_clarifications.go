package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	listclarifications "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/list_clarifications"
)

type ListClarificationsPresenter types.Presenter[listclarifications.Output, api.ListClarificationsResponseObject]

type listClarificationsPresenter struct{}

func NewListClarificationsPresenter() ListClarificationsPresenter {
	return &listClarificationsPresenter{}
}

func (p *listClarificationsPresenter) Output(output *listclarifications.Output) api.ListClarificationsResponseObject {
	clarifications := make([]api.AdminProjectSpecClarificationSpecClarificationInfo, 0, len(output.Clarifications))
	for _, clarification := range output.Clarifications {
		var receiverUser *api.AdminProjectSpecClarificationSpecUserInfo
		if clarification.ReceiverUser != nil {
			receiverUser = &api.AdminProjectSpecClarificationSpecUserInfo{
				Id:       clarification.ReceiverUser.ID,
				Name:     clarification.ReceiverUser.Name,
				Username: clarification.ReceiverUser.Username,
			}
		}

		clarifications = append(clarifications, api.AdminProjectSpecClarificationSpecClarificationInfo{
			Id:      clarification.ID,
			Subject: clarification.Subject,
			Content: clarification.Content,
			SenderUser: api.AdminProjectSpecClarificationSpecUserInfo{
				Id:       clarification.SenderUser.ID,
				Name:     clarification.SenderUser.Name,
				Username: clarification.SenderUser.Username,
			},
			ReceiverUser: receiverUser,
			IsAdmin:      clarification.IsAdmin,
			SendToAll:    clarification.SendToAll,
			IsAnswered:   clarification.IsAnswered,
			CreatedAt:    clarification.CreatedAt,
			UpdatedAt:    clarification.UpdatedAt,
		})
	}

	return api.ListClarifications200JSONResponse{
		Success: true,
		Data:    clarifications,
	}
}

func (p *listClarificationsPresenter) Error(err error) api.ListClarificationsResponseObject {
	return httperror.From(err).Response()
}
