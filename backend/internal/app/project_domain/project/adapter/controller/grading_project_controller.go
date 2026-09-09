package projectcontroller

import (
	"judge-backend/api"
	fileheader "judge-backend/internal/app/shared/types/file_header"
	formrequest "judge-backend/internal/app/shared/types/form_request"
	"judge-backend/internal/app/project_domain/project/adapter/presenter"
	"judge-backend/internal/app/project_domain/project/application"
	downloadgradingproject "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/download_grading_project"
	listgradingprojectfiles "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/list_grading_project_files"
	putgradingproject "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/put_grading_project"
	readgradingprojectfile "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/read_grading_project_file"

	"github.com/gin-gonic/gin"
)

type GradingProjectController struct {
	app       *application.GradingProjectApplication
	presenter *presenter.GradingProjectPresenter
}

func NewGradingProjectController(
	app *application.GradingProjectApplication,
	presenter *presenter.GradingProjectPresenter,
) *GradingProjectController {
	return &GradingProjectController{
		app:       app,
		presenter: presenter,
	}
}

type PutGradingProjectPayload struct {
	Project *fileheader.FileHeader `form:"project" binding:"required"`
}

func (ctl *GradingProjectController) PutGradingProject(ctx *gin.Context, request api.PutGradingProjectRequestObject) (api.PutGradingProjectResponseObject, error) {
	payload := &PutGradingProjectPayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		return ctl.presenter.PutGradingProject.Error(putgradingproject.ErrorInvalidInput), nil
	}

	input := &putgradingproject.Input{
		ProjectID:  request.Id,
		ZipProject: payload.Project.Buffer(),
	}

	output, err := ctl.app.PutGradingProject(input)
	if err != nil {
		return ctl.presenter.PutGradingProject.Error(err), nil
	}

	return ctl.presenter.PutGradingProject.Output(output), nil
}

func (ctl *GradingProjectController) ListGradingProjectFiles(ctx *gin.Context, request api.ListGradingProjectFilesRequestObject) (api.ListGradingProjectFilesResponseObject, error) {
	input := &listgradingprojectfiles.Input{
		ProjectID: request.Id,
	}

	output, err := ctl.app.ListGradingProjectFiles(input)
	if err != nil {
		return ctl.presenter.ListGradingProjectFiles.Error(err), nil
	}

	return ctl.presenter.ListGradingProjectFiles.Output(output), nil
}

func (ctl *GradingProjectController) ReadGradingProjectFile(ctx *gin.Context, request api.ReadGradingProjectFileRequestObject) (api.ReadGradingProjectFileResponseObject, error) {
	input := &readgradingprojectfile.Input{
		ProjectID:         request.Id,
		Base64EncodedPath: request.File,
	}

	output, err := ctl.app.ReadGradingProjectFile(input)
	if err != nil {
		return ctl.presenter.ReadGradingProjectFile.Error(err), nil
	}

	return ctl.presenter.ReadGradingProjectFile.Output(output), nil
}

func (ctl *GradingProjectController) DownloadGradingProject(ctx *gin.Context, request api.DownloadGradingProjectRequestObject) (api.DownloadGradingProjectResponseObject, error) {
	input := &downloadgradingproject.Input{
		ProjectID: request.Id,
	}

	output, err := ctl.app.DownloadGradingProject(input)
	if err != nil {
		return ctl.presenter.DownloadGradingProject.Error(err), nil
	}

	return ctl.presenter.DownloadGradingProject.Output(output), nil
}
