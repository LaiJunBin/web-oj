package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	listgradingprojectfiles "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/list_grading_project_files"
)

type ListGradingProjectFilesPresenter types.Presenter[listgradingprojectfiles.Output, api.ListGradingProjectFilesResponseObject]

type listGradingProjectFilesPresenter struct{}

func NewListGradingProjectFilesPresenter() ListGradingProjectFilesPresenter {
	return &listGradingProjectFilesPresenter{}
}

func (p *listGradingProjectFilesPresenter) Output(output *listgradingprojectfiles.Output) api.ListGradingProjectFilesResponseObject {
	return api.ListGradingProjectFiles200JSONResponse{
		Success: true,
		Data:    output.Files,
	}
}

func (p *listGradingProjectFilesPresenter) Error(err error) api.ListGradingProjectFilesResponseObject {
	switch {
	case errors.Is(err, listgradingprojectfiles.ErrorInvalidProjectID):
		return api.ListGradingProjectFiles400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, listgradingprojectfiles.ErrorProjectNotFound):
		return api.ListGradingProjectFiles404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, listgradingprojectfiles.ErrorInternalServer):
		fallthrough
	default:
		return api.ListGradingProjectFiles500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
