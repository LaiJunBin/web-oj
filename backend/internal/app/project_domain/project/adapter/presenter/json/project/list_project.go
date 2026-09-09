package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	listproject "judge-backend/internal/app/project_domain/project/application/usecase/project/list_project"
)

type ListProjectPresenter types.Presenter[listproject.Output, api.ListProjectResponseObject]

type listProjectPresenter struct{}

func NewListProjectPresenter() ListProjectPresenter {
	return &listProjectPresenter{}
}

func (p *listProjectPresenter) Output(output *listproject.Output) api.ListProjectResponseObject {
	projects := make([]api.AdminProjectSpecProject, 0, len(output.Projects))
	for _, project := range output.Projects {
		projects = append(projects, api.AdminProjectSpecProject{
			Id:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			Image:       project.Image,
			TestImage:   project.TestImage,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		})
	}

	return api.ListProject200JSONResponse{
		Success: true,
		Data:    projects,
	}
}

func (p *listProjectPresenter) Error(err error) api.ListProjectResponseObject {
	switch {
	case errors.Is(err, listproject.ErrorInternalServer):
		fallthrough
	default:
		return api.ListProject500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
