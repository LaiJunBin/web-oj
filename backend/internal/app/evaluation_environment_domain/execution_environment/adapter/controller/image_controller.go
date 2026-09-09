package controller

import (
	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application"
	cancelbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/cancel_build"
	createimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/create_image"
	deleteimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/delete_image"
	exportimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/export_image"
	getbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/get_build"
	getimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/get_image"
	importimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/import_image"
	listbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/list_build"
	listimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/list_image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/rebuild"
	streambuildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/stream_build_log"
	updateimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/update_image"
	fileheader "judge-backend/internal/app/shared/types/file_header"
	formrequest "judge-backend/internal/app/shared/types/form_request"

	"github.com/gin-gonic/gin"
)

type ImageController struct {
	presenter *presenter.ImagePresenter
	app       *application.ImageApplication
}

func NewImageController(
	presenter *presenter.ImagePresenter,
	app *application.ImageApplication,

) *ImageController {
	return &ImageController{
		presenter: presenter,
		app:       app,
	}
}

func (ctl *ImageController) ListImage(ctx *gin.Context, request api.ListImageRequestObject) (api.ListImageResponseObject, error) {
	input := &listimage.Input{}
	output, err := ctl.app.ListImage(input)
	if err != nil {
		return ctl.presenter.ListImage.Error(err), nil
	}

	return ctl.presenter.ListImage.Output(output), nil
}

func (ctl *ImageController) GetImage(ctx *gin.Context, request api.GetImageRequestObject) (api.GetImageResponseObject, error) {
	input := &getimage.Input{
		Tag: request.Tag,
	}

	output, err := ctl.app.GetImage(input)
	if err != nil {
		return ctl.presenter.GetImage.Error(err), nil
	}

	return ctl.presenter.GetImage.Output(output), nil
}

type CreateImagePayload struct {
	Tag         string                 `form:"tag"`
	MountPath   string                 `form:"mountPath"`
	Description string                 `form:"description"`
	Project     *fileheader.FileHeader `form:"project" binding:"required"`
	NoCache     bool                   `form:"noCache"`
}

func (ctl *ImageController) CreateImage(ctx *gin.Context, request api.CreateImageRequestObject) (api.CreateImageResponseObject, error) {
	payload := &CreateImagePayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		return ctl.presenter.CreateImage.Error(createimage.ErrorInvalidInput), nil
	}

	input := &createimage.Input{
		Tag:            payload.Tag,
		MountPath:      payload.MountPath,
		Description:    payload.Description,
		ProjectZipFile: payload.Project.Buffer(),
		NoCache:        payload.NoCache,
	}

	output, err := ctl.app.CreateImage(input)
	if err != nil {
		return ctl.presenter.CreateImage.Error(err), nil
	}

	return ctl.presenter.CreateImage.Output(output), nil
}

type UpdateImagePayload struct {
	MountPath   *string                `form:"mount_path"`
	Description *string                `form:"description"`
	Project     *fileheader.FileHeader `form:"project"`
	NoCache     bool                   `form:"noCache"`
}

func (ctl *ImageController) UpdateImage(ctx *gin.Context, request api.UpdateImageRequestObject) (api.UpdateImageResponseObject, error) {
	payload := &UpdateImagePayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		return ctl.presenter.UpdateImage.Error(updateimage.ErrorInvalidInput), nil
	}

	input := &updateimage.Input{
		Tag:         request.Tag,
		MountPath:   payload.MountPath,
		Description: payload.Description,
		NoCache:     payload.NoCache,
	}

	if payload.Project != nil {
		input.ProjectZipFile = payload.Project.Buffer()
	}

	output, err := ctl.app.UpdateImage(input)
	if err != nil {
		return ctl.presenter.UpdateImage.Error(err), nil
	}

	return ctl.presenter.UpdateImage.Output(output), nil
}

func (ctl *ImageController) DeleteImage(ctx *gin.Context, request api.DeleteImageRequestObject) (api.DeleteImageResponseObject, error) {
	input := &deleteimage.Input{
		Tag: request.Tag,
	}

	output, err := ctl.app.DeleteImage(input)
	if err != nil {
		return ctl.presenter.DeleteImage.Error(err), nil
	}

	return ctl.presenter.DeleteImage.Output(output), nil
}

func (ctl *ImageController) ListBuild(ctx *gin.Context, request api.ListBuildRequestObject) (api.ListBuildResponseObject, error) {
	input := &listbuild.Input{
		Tag: request.Tag,
	}

	output, err := ctl.app.ListBuild(input)
	if err != nil {
		return ctl.presenter.ListBuild.Error(err), nil
	}

	return ctl.presenter.ListBuild.Output(output), nil
}

func (ctl *ImageController) GetBuild(ctx *gin.Context, request api.GetBuildRequestObject) (api.GetBuildResponseObject, error) {
	input := &getbuild.Input{
		Tag:   request.Tag,
		Batch: int(request.Batch),
	}

	output, err := ctl.app.GetBuild(input)
	if err != nil {
		return ctl.presenter.GetBuild.Error(err), nil
	}

	return ctl.presenter.GetBuild.Output(output), nil
}

func (ctl *ImageController) CancelBuild(ctx *gin.Context, request api.CancelBuildRequestObject) (api.CancelBuildResponseObject, error) {
	input := &cancelbuild.Input{
		Tag: request.Tag,
	}

	output, err := ctl.app.CancelBuild(input)
	if err != nil {
		return ctl.presenter.CancelBuild.Error(err), nil
	}

	return ctl.presenter.CancelBuild.Output(output), nil
}

func (ctl *ImageController) Rebuild(ctx *gin.Context, request api.RebuildRequestObject) (api.RebuildResponseObject, error) {
	input := &rebuild.Input{
		Tag: request.Tag,
	}

	output, err := ctl.app.Rebuild(input)
	if err != nil {
		return ctl.presenter.Rebuild.Error(err), nil
	}

	return ctl.presenter.Rebuild.Output(output), nil
}

func (ctl *ImageController) StreamLogs(ctx *gin.Context, request api.StreamLogsRequestObject) (api.StreamLogsResponseObject, error) {
	input := &streambuildlog.Input{
		Tag:   request.Tag,
		Batch: int(request.Batch),
	}

	output, err := ctl.app.StreamBuildLog(input)
	if err != nil {
		return ctl.presenter.StreamBuildLog.Error(err), nil
	}

	return ctl.presenter.StreamBuildLog.Output(ctx.Request.Context(), output), nil
}

type ImportImagePayload struct {
	Bundle *fileheader.FileHeader `form:"bundle" validate:"required"`
	Force  bool                   `form:"force"`
}

func (ctl *ImageController) ExportImage(ctx *gin.Context, request api.ExportImageRequestObject) (api.ExportImageResponseObject, error) {
	output, err := ctl.app.ExportImage(&exportimage.Input{
		Tag:  request.Tag,
		Mode: string(request.Params.Mode),
	})
	if err != nil {
		return ctl.presenter.ExportImage.Error(err), nil
	}

	return ctl.presenter.ExportImage.Output(output), nil
}

func (ctl *ImageController) ImportImage(ctx *gin.Context, request api.ImportImageRequestObject) (api.ImportImageResponseObject, error) {
	payload := &ImportImagePayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		return ctl.presenter.ImportImage.Error(&importimage.InvalidBundleError{
			Details: []string{"沒有收到 bundle 檔案"},
		}), nil
	}

	output, err := ctl.app.ImportImage(&importimage.Input{
		Bundle: payload.Bundle.Buffer(),
		Force:  payload.Force,
	})
	if err != nil {
		return ctl.presenter.ImportImage.Error(err), nil
	}

	return ctl.presenter.ImportImage.Output(output), nil
}
