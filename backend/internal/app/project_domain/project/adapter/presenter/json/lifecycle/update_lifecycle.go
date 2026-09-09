package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	updatelifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/update_lifecycle"
)

type UpdateLifecyclePresenter types.Presenter[updatelifecycle.Output, api.UpdateLifecycleResponseObject]

type updateLifecyclePresenter struct{}

func NewUpdateLifecyclePresenter() UpdateLifecyclePresenter {
	return &updateLifecyclePresenter{}
}

func (p *updateLifecyclePresenter) Output(output *updatelifecycle.Output) api.UpdateLifecycleResponseObject {
	return api.UpdateLifecycle200JSONResponse{
		Success: true,
	}
}

func (p *updateLifecyclePresenter) Error(err error) api.UpdateLifecycleResponseObject {
	switch {
	case errors.Is(err, updatelifecycle.ErrorInvalidID):
		fallthrough
	case errors.Is(err, updatelifecycle.ErrorInvalidName):
		fallthrough
	case errors.Is(err, updatelifecycle.ErrorEmptyInput):
		fallthrough
	case errors.Is(err, updatelifecycle.ErrorInvalidType):
		fallthrough
	case errors.Is(err, updatelifecycle.ErrorInvalidExecutionEnvironment):
		return api.UpdateLifecycle400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, updatelifecycle.ErrorLifecycleNotFound):
		return api.UpdateLifecycle404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, updatelifecycle.ErrorInternalServer):
		fallthrough
	default:
		return api.UpdateLifecycle500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
