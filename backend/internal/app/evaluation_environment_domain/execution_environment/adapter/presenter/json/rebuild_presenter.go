package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/rebuild"
)

type RebuildPresenter struct{}

func NewRebuildPresenter() types.RebuildPresenterType {
	return &RebuildPresenter{}
}

func (p *RebuildPresenter) Output(output *rebuild.Output) api.RebuildResponseObject {
	return api.Rebuild200JSONResponse{
		Success: true,
	}
}

func (p *RebuildPresenter) Error(err error) api.RebuildResponseObject {
	switch {
	case errors.Is(err, rebuild.ErrorTagIsRequired):
		fallthrough
	case errors.Is(err, rebuild.ErrorImageProjectNotSet):
		fallthrough
	case errors.Is(err, rebuild.ErrorBuildInProgress):
		return &api.Rebuild400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, rebuild.ErrorImageNotFound):
		return &api.Rebuild404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, rebuild.ErrorInternalServer):
		fallthrough
	default:
		return &api.Rebuild500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
