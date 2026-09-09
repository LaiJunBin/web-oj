package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	updateproject "judge-backend/internal/app/project_domain/project/application/usecase/project/update_project"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/domain/entity/project/dynamicfield"
)

type UpdateProjectPresenter types.Presenter[updateproject.Output, api.UpdateProjectResponseObject]

type updateProjectPresenter struct{}

func NewUpdateProjectPresenter() UpdateProjectPresenter {
	return &updateProjectPresenter{}
}

func (p *updateProjectPresenter) Output(output *updateproject.Output) api.UpdateProjectResponseObject {
	return api.UpdateProject200JSONResponse{
		Success: true,
	}
}

func (p *updateProjectPresenter) Error(err error) api.UpdateProjectResponseObject {
	badRequestErrors := []error{
		updateproject.ErrorNameIsRequired,
		updateproject.ErrorImageTagIsRequired,
		updateproject.ErrorTestImageTagIsRequired,
		updateproject.ErrorImageNotFound,
		updateproject.ErrorTestImageNotFound,
		updateproject.ErrorInvalidResultParser,
		project.ErrorInvalidEnv,
		dynamicfield.ErrorInvalidInputType,
	}

	for _, e := range badRequestErrors {
		if errors.Is(err, e) {
			return api.UpdateProject400JSONResponse{
				Success: false,
				Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			}
		}
	}

	switch {
	case errors.Is(err, updateproject.ErrorProjectNotFound):
		return api.UpdateProject404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, updateproject.ErrorInternalServer):
		fallthrough
	default:
		return api.UpdateProject500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
