package projectcontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter"
	"judge-backend/internal/app/project_domain/project/application"
	createproject "judge-backend/internal/app/project_domain/project/application/usecase/project/create_project"
	deleteproject "judge-backend/internal/app/project_domain/project/application/usecase/project/delete_project"
	exportproject "judge-backend/internal/app/project_domain/project/application/usecase/project/export_project"
	getproject "judge-backend/internal/app/project_domain/project/application/usecase/project/get_project"
	importproject "judge-backend/internal/app/project_domain/project/application/usecase/project/import_project"
	listproject "judge-backend/internal/app/project_domain/project/application/usecase/project/list_project"
	updateproject "judge-backend/internal/app/project_domain/project/application/usecase/project/update_project"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	fileheader "judge-backend/internal/app/shared/types/file_header"
	formrequest "judge-backend/internal/app/shared/types/form_request"

	"github.com/gin-gonic/gin"
)

type ProjectController struct {
	app       *application.ProjectApplication
	presenter *presenter.ProjectPresenter
}

func NewProjectController(
	app *application.ProjectApplication,
	presenter *presenter.ProjectPresenter,
) *ProjectController {
	return &ProjectController{
		app:       app,
		presenter: presenter,
	}
}

func (ctl *ProjectController) ListProject(ctx *gin.Context, request api.ListProjectRequestObject) (api.ListProjectResponseObject, error) {
	input := &listproject.Input{}

	output, err := ctl.app.ListProject(input)
	if err != nil {
		return ctl.presenter.ListProject.Error(err), nil
	}

	return ctl.presenter.ListProject.Output(output), nil
}

func (ctl *ProjectController) GetProject(ctx *gin.Context, request api.GetProjectRequestObject) (api.GetProjectResponseObject, error) {
	input := &getproject.Input{
		ID: request.Id,
	}

	output, err := ctl.app.GetProject(input)
	if err != nil {
		return ctl.presenter.GetProject.Error(err), nil
	}

	return ctl.presenter.GetProject.Output(output), nil
}

func (ctl *ProjectController) CreateProject(ctx *gin.Context, request api.CreateProjectRequestObject) (api.CreateProjectResponseObject, error) {
	input := &createproject.Input{
		Name:         request.Body.Name,
		Description:  request.Body.Description,
		ImageTag:     request.Body.ProjectImageTag,
		TestImageTag: request.Body.TestImageTag,
		Env:          request.Body.ProjectEnv,
		TestEnv:      request.Body.TestEnv,
	}

	if request.Body.ResultParser != nil {
		input.ResultParser = *request.Body.ResultParser
	}

	ds := make([]*project.DynamicFieldConfig, 0, len(request.Body.DynamicFields))
	for _, df := range request.Body.DynamicFields {
		dfc := &project.DynamicFieldConfig{
			Name:       df.Name,
			Type:       df.Type,
			Label:      df.Label,
			DefaultVal: df.DefaultValue,
			IsRequired: df.IsRequired,
		}
		ds = append(ds, dfc)
	}

	input.DynamicFieldConfigs = ds

	output, err := ctl.app.CreateProject(input)
	if err != nil {
		return ctl.presenter.CreateProject.Error(err), nil
	}

	return ctl.presenter.CreateProject.Output(output), nil
}

func (ctl *ProjectController) DeleteProject(ctx *gin.Context, request api.DeleteProjectRequestObject) (api.DeleteProjectResponseObject, error) {
	input := &deleteproject.Input{
		ID: request.Id,
	}

	output, err := ctl.app.DeleteProject(input)
	if err != nil {
		return ctl.presenter.DeleteProject.Error(err), nil
	}

	return ctl.presenter.DeleteProject.Output(output), nil
}

func (ctl *ProjectController) UpdateProject(ctx *gin.Context, request api.UpdateProjectRequestObject) (api.UpdateProjectResponseObject, error) {
	input := &updateproject.Input{
		ID:           request.Id,
		Name:         request.Body.Name,
		Description:  request.Body.Description,
		ImageTag:     request.Body.ProjectImageTag,
		TestImageTag: request.Body.TestImageTag,
		Env:          request.Body.ProjectEnv,
		TestEnv:      request.Body.TestEnv,
		ResultParser: request.Body.ResultParser,
	}

	if request.Body.DynamicFields != nil {
		ds := make([]*project.DynamicFieldConfig, 0, len(*request.Body.DynamicFields))
		for _, df := range *request.Body.DynamicFields {
			dfc := &project.DynamicFieldConfig{
				Name:       df.Name,
				Type:       df.Type,
				Label:      df.Label,
				DefaultVal: df.DefaultValue,
				IsRequired: df.IsRequired,
			}
			ds = append(ds, dfc)
		}
		input.DynamicFieldConfigs = &ds
	}

	output, err := ctl.app.UpdateProject(input)
	if err != nil {
		return ctl.presenter.UpdateProject.Error(err), nil
	}

	return ctl.presenter.UpdateProject.Output(output), nil
}

type ImportProjectPayload struct {
	Bundle *fileheader.FileHeader `form:"bundle" validate:"required"`
}

func (ctl *ProjectController) ExportProject(ctx *gin.Context, request api.ExportProjectRequestObject) (api.ExportProjectResponseObject, error) {
	output, err := ctl.app.ExportProject(&exportproject.Input{ProjectID: request.Id})
	if err != nil {
		return ctl.presenter.ExportProject.Error(err), nil
	}

	return ctl.presenter.ExportProject.Output(output), nil
}

func (ctl *ProjectController) ImportProject(ctx *gin.Context, request api.ImportProjectRequestObject) (api.ImportProjectResponseObject, error) {
	payload := &ImportProjectPayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		return ctl.presenter.ImportProject.Error(&importproject.InvalidBundleError{
			Details: []string{"沒有收到 bundle 檔案"},
		}), nil
	}

	output, err := ctl.app.ImportProject(&importproject.Input{Bundle: payload.Bundle.Buffer()})
	if err != nil {
		return ctl.presenter.ImportProject.Error(err), nil
	}

	return ctl.presenter.ImportProject.Output(output), nil
}
