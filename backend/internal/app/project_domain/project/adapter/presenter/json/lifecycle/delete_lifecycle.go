package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	deletelifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/delete_lifecycle"
)

type DeleteLifecyclePresenter types.Presenter[deletelifecycle.Output, api.DeleteLifecycleResponseObject]

type deleteLifecyclePresenter struct{}

func NewDeleteLifecyclePresenter() DeleteLifecyclePresenter {
	return &deleteLifecyclePresenter{}
}

func (p *deleteLifecyclePresenter) Output(output *deletelifecycle.Output) api.DeleteLifecycleResponseObject {
	return api.DeleteLifecycle200JSONResponse{
		Success: true,
	}
}

func (p *deleteLifecyclePresenter) Error(err error) api.DeleteLifecycleResponseObject {
	switch {
	case errors.Is(err, deletelifecycle.ErrorInvalidID):
		return api.DeleteLifecycle400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, deletelifecycle.ErrorLifecycleNotFound):
		return api.DeleteLifecycle404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, deletelifecycle.ErrorInternalServer):
		fallthrough
	default:
		return api.DeleteLifecycle500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
