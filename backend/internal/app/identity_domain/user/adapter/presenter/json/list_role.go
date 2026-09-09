package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/user/adapter/presenter/types"
	listrole "judge-backend/internal/app/identity_domain/user/application/usecase/list_role"
)

type ListRolePresenter struct{}

func NewListRolePresenter() types.ListRolePresenter {
	return &ListRolePresenter{}
}

func (p *ListRolePresenter) Output(output *listrole.Output) api.ListRoleResponseObject {
	roles := []api.AdminRoleSpecRole{}
	for _, user := range output.Roles {
		roles = append(roles, api.AdminRoleSpecRole{
			Id:   user.ID,
			Name: user.Name,
		})
	}
	return &api.ListRole200JSONResponse{
		Success: true,
		Data:    roles,
	}
}

func (p *ListRolePresenter) Error(err error) api.ListRoleResponseObject {
	switch {
	case errors.Is(err, listrole.ErrorInternalServer):
		fallthrough
	default:
		return &api.ListRole500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
