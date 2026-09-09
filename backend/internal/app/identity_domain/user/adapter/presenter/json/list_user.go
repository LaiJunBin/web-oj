package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/user/adapter/presenter/types"
	listuser "judge-backend/internal/app/identity_domain/user/application/usecase/list_user"
)

type ListUserPresenter struct{}

func NewListUserPresenter() types.ListUserPresenter {
	return &ListUserPresenter{}
}

func (p *ListUserPresenter) Output(output *listuser.Output) api.ListUserResponseObject {
	users := []api.AdminUserSpecUser{}
	for _, user := range output.Users {
		users = append(users, api.AdminUserSpecUser{
			Id:        user.ID,
			Username:  user.Username,
			Name:      user.Name,
			Roles:     user.Roles,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}
	return &api.ListUser200JSONResponse{
		Success: true,
		Data:    users,
	}
}

func (p *ListUserPresenter) Error(err error) api.ListUserResponseObject {
	switch {
	case errors.Is(err, listuser.ErrorInternalServer):
		fallthrough
	default:
		return &api.ListUser500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
