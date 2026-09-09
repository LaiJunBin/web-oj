package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	cancelbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/cancel_build"
)

type CancelBuildPresenter struct{}

func NewCancelBuildPresenter() types.CancelBuildPresenterType {
	return &CancelBuildPresenter{}
}
func (p *CancelBuildPresenter) Output(output *cancelbuild.Output) api.CancelBuildResponseObject {
	return &api.CancelBuild200JSONResponse{
		Success: true,
	}
}

func (p *CancelBuildPresenter) Error(err error) api.CancelBuildResponseObject {
	switch {
	case errors.Is(err, cancelbuild.ErrorInvalidInput):
		fallthrough
	case errors.Is(err, cancelbuild.ErrorBuildNotInProgress):
		return &api.CancelBuild400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, cancelbuild.ErrorBuildNotFound):
		return &api.CancelBuild404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, cancelbuild.ErrorInternalServer):
		fallthrough
	default:
		return &api.CancelBuild500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
