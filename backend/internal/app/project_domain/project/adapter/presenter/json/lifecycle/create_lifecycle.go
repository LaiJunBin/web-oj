package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	createlifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/create_lifecycle"
)

type CreateLifecyclePresenter types.Presenter[createlifecycle.Output, api.CreateLifecycleResponseObject]

type createLifecyclePresenter struct{}

func NewCreateLifecyclePresenter() CreateLifecyclePresenter {
	return &createLifecyclePresenter{}
}

func (p *createLifecyclePresenter) Output(output *createlifecycle.Output) api.CreateLifecycleResponseObject {
	return api.CreateLifecycle200JSONResponse{
		Success: true,
		Data:    output.ID,
	}
}

func (p *createLifecyclePresenter) Error(err error) api.CreateLifecycleResponseObject {
	badRequestErrors := []error{
		createlifecycle.ErrorInvalidProjectID,
		createlifecycle.ErrorEmptyType,
		createlifecycle.ErrorEmptyName,
		createlifecycle.ErrorEmptyCommand,
		createlifecycle.ErrorEmptyExecutionEnvironment,
		createlifecycle.ErrorInvalidType,
		createlifecycle.ErrorInvalidExecutionEnvironment,
	}

	for _, e := range badRequestErrors {
		if errors.Is(err, e) {
			return api.CreateLifecycle400JSONResponse{
				Success: false,
				Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			}
		}
	}

	switch {
	case errors.Is(err, createlifecycle.ErrorProjectNotFound):
		return api.CreateLifecycle404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, createlifecycle.ErrorInternalServer):
		fallthrough
	default:
		return api.CreateLifecycle500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
