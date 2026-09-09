package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	readgradingprojectfile "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/read_grading_project_file"
)

type ReadGradingProjectFilePresenter types.Presenter[readgradingprojectfile.Output, api.ReadGradingProjectFileResponseObject]

type readGradingProjectFilePresenter struct{}

func NewReadGradingProjectFilePresenter() ReadGradingProjectFilePresenter {
	return &readGradingProjectFilePresenter{}
}

func (p *readGradingProjectFilePresenter) Output(output *readgradingprojectfile.Output) api.ReadGradingProjectFileResponseObject {
	return api.ReadGradingProjectFile200JSONResponse{
		Success: true,
		Data:    output.Base64Content,
	}
}

func (p *readGradingProjectFilePresenter) Error(err error) api.ReadGradingProjectFileResponseObject {
	switch {
	case errors.Is(err, readgradingprojectfile.ErrorInvalidProjectID):
		fallthrough
	case errors.Is(err, readgradingprojectfile.ErrorBase64EncodedPathRequired):
		fallthrough
	case errors.Is(err, readgradingprojectfile.ErrorInvalidBase64EncodedPath):
		return api.ReadGradingProjectFile400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, readgradingprojectfile.ErrorProjectNotFound):
		return api.ReadGradingProjectFile404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, readgradingprojectfile.ErrorInternalServer):
		fallthrough
	default:
		return api.ReadGradingProjectFile500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
