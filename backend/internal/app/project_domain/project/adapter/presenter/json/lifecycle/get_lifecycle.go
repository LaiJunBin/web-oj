package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	getlifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/get_lifecycle"
)

type GetLifecyclePresenter types.Presenter[getlifecycle.Output, api.GetLifecycleResponseObject]

type getLifecyclePresenter struct{}

func NewGetLifecyclePresenter() GetLifecyclePresenter {
	return &getLifecyclePresenter{}
}

func (p *getLifecyclePresenter) Output(output *getlifecycle.Output) api.GetLifecycleResponseObject {
	return api.GetLifecycle200JSONResponse{
		Success: true,
		Data: api.AdminProjectSpecLifecycleSpecLifecycleDetail{
			Id:          output.Lifecycle.ID,
			Type:        api.AdminProjectSpecLifecycleSpecLifecycleType(output.Lifecycle.Type),
			Name:        output.Lifecycle.Name,
			Description: output.Lifecycle.Description,
			Command:     output.Lifecycle.Command,
			RunsOn:      api.SharedRunsOn(output.Lifecycle.ExecutionEnvironment),
			CreatedAt:   output.Lifecycle.CreatedAt,
			UpdatedAt:   output.Lifecycle.UpdatedAt,
		},
	}
}

func (p *getLifecyclePresenter) Error(err error) api.GetLifecycleResponseObject {
	switch {
	case errors.Is(err, getlifecycle.ErrorInvalidID):
		return api.GetLifecycle400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, getlifecycle.ErrorLifecycleNotFound):
		return api.GetLifecycle404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, getlifecycle.ErrorInternalServer):
		fallthrough
	default:
		return api.GetLifecycle500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
