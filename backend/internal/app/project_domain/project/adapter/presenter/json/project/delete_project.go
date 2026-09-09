package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	deleteproject "judge-backend/internal/app/project_domain/project/application/usecase/project/delete_project"
)

type DeleteProjectPresenter types.Presenter[deleteproject.Output, api.DeleteProjectResponseObject]

type deleteProjectPresenter struct{}

func NewDeleteProjectPresenter() DeleteProjectPresenter {
	return &deleteProjectPresenter{}
}

func (p *deleteProjectPresenter) Output(output *deleteproject.Output) api.DeleteProjectResponseObject {
	return api.DeleteProject200JSONResponse{
		Success: true,
	}
}

func (p *deleteProjectPresenter) Error(err error) api.DeleteProjectResponseObject {
	switch {
	case errors.Is(err, deleteproject.ErrorInvalidID):
		return api.DeleteProject400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, deleteproject.ErrorProjectNotFound):
		return api.DeleteProject404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, deleteproject.ErrorInternalServer):
		fallthrough
	default:
		return api.DeleteProject500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
