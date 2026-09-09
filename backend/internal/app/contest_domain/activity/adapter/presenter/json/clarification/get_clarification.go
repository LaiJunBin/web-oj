package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	getclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/get_clarification"
)

type GetClarificationPresenter types.Presenter[getclarification.Output, api.GetClarificationResponseObject]

type getClarificationPresenter struct{}

func NewGetClarificationPresenter() GetClarificationPresenter {
	return &getClarificationPresenter{}
}

func transformClarification(clarification *dto.ClarificationDetail) *api.AdminProjectSpecClarificationSpecClarification {
	if clarification == nil {
		return nil
	}

	var receiverUser *api.AdminProjectSpecClarificationSpecUserInfo
	if clarification.ReceiverUser != nil {
		receiverUser = &api.AdminProjectSpecClarificationSpecUserInfo{
			Id:       clarification.ReceiverUser.ID,
			Name:     clarification.ReceiverUser.Name,
			Username: clarification.ReceiverUser.Username,
		}
	}

	return &api.AdminProjectSpecClarificationSpecClarification{
		Id:      clarification.ID,
		Subject: clarification.Subject,
		Content: clarification.Content,
		SenderUser: api.AdminProjectSpecClarificationSpecUserInfo{
			Id:       clarification.SenderUser.ID,
			Name:     clarification.SenderUser.Name,
			Username: clarification.SenderUser.Username,
		},
		ReceiverUser:       receiverUser,
		IsAdmin:            clarification.IsAdmin,
		SendToAll:          clarification.SendToAll,
		IsAnswered:         clarification.IsAnswered,
		CreatedAt:          clarification.CreatedAt,
		UpdatedAt:          clarification.UpdatedAt,
		ReplyClarification: transformClarification(clarification.ReplyClarification),
	}
}

func (p *getClarificationPresenter) Output(output *getclarification.Output) api.GetClarificationResponseObject {
	clarification := transformClarification(output.Clarification)

	return api.GetClarification200JSONResponse{
		Success: true,
		Data:    *clarification,
	}
}

func (p *getClarificationPresenter) Error(err error) api.GetClarificationResponseObject {
	return httperror.From(err).Response()
}
