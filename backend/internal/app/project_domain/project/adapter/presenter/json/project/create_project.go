package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	createproject "judge-backend/internal/app/project_domain/project/application/usecase/project/create_project"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/domain/entity/project/dynamicfield"
)

type CreateProjectPresenter types.Presenter[createproject.Output, api.CreateProjectResponseObject]

type createProjectPresenter struct{}

func NewCreateProjectPresenter() CreateProjectPresenter {
	return &createProjectPresenter{}
}

func (p *createProjectPresenter) Output(output *createproject.Output) api.CreateProjectResponseObject {
	return api.CreateProject200JSONResponse{
		Success: true,
		Data:    output.ID,
	}
}

func (p *createProjectPresenter) Error(err error) api.CreateProjectResponseObject {
	badRequestErrors := []error{
		createproject.ErrorNameIsRequired,
		createproject.ErrorImageTagIsRequired,
		createproject.ErrorTestImageTagIsRequired,
		createproject.ErrorImageNotFound,
		createproject.ErrorTestImageNotFound,
		createproject.ErrorInvalidResultParser,
		project.ErrorInvalidEnv,
		dynamicfield.ErrorInvalidInputType,
	}

	for _, e := range badRequestErrors {
		if errors.Is(err, e) {
			return api.CreateProject400JSONResponse{
				Success: false,
				Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			}
		}
	}

	switch {
	case errors.Is(err, createproject.ErrorInternalServer):
		fallthrough
	default:
		return api.CreateProject500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
