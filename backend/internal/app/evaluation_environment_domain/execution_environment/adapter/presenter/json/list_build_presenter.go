package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	listbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/list_build"
)

type ListBuildPresenter struct{}

func NewListBuildPresenter() types.ListBuildPresenterType {
	return &ListBuildPresenter{}
}
func (p *ListBuildPresenter) Output(output *listbuild.Output) api.ListBuildResponseObject {
	buildInfos := make([]api.AdminBuildSpecBuildInfo, 0, len(output.Builds))
	for _, b := range output.Builds {
		buildInfos = append(buildInfos, api.AdminBuildSpecBuildInfo{
			Batch:  b.Batch,
			Status: api.AdminBuildSpecBuildStatus(b.Status),
		})
	}

	return &api.ListBuild200JSONResponse{
		Success: true,
		Data:    buildInfos,
	}
}

func (p *ListBuildPresenter) Error(err error) api.ListBuildResponseObject {
	switch {
	case errors.Is(err, listbuild.ErrorTagIsRequired):
		return &api.ListBuild400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, listbuild.ErrorImageNotFound):
		return &api.ListBuild404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, listbuild.ErrorInternalServer):
		fallthrough
	default:
		return &api.ListBuild500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
