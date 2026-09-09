package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	getbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/get_build"
)

type GetBuildPresenter struct{}

func NewGetBuildPresenter() types.GetBuildPresenterType {
	return &GetBuildPresenter{}
}
func (p *GetBuildPresenter) Output(output *getbuild.Output) api.GetBuildResponseObject {
	logs := make([]api.AdminBuildSpecLog, 0, len(output.Logs))
	for _, log := range output.Logs {
		logs = append(logs, api.AdminBuildSpecLog{
			Level:   log.Level,
			Message: log.Message,
			Time:    log.Time,
		})
	}

	return &api.GetBuild200JSONResponse{
		Success: true,
		Data: api.AdminBuildSpecBuild{
			Batch:  output.Batch,
			Status: api.AdminBuildSpecBuildStatus(output.Status),
			Logs:   logs,
		},
	}
}

func (p *GetBuildPresenter) Error(err error) api.GetBuildResponseObject {
	switch {
	case errors.Is(err, getbuild.ErrorTagIsRequired):
		fallthrough
	case errors.Is(err, getbuild.ErrorBatchMustBeGreaterThanZero):
		return &api.GetBuild400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, getbuild.ErrorBuildNotFound):
		return &api.GetBuild404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	default:
		return &api.GetBuild500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
