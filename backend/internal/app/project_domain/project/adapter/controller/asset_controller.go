package projectcontroller

import (
	"judge-backend/api"
	fileheader "judge-backend/internal/app/shared/types/file_header"
	formrequest "judge-backend/internal/app/shared/types/form_request"
	"judge-backend/internal/app/project_domain/project/adapter/presenter"
	"judge-backend/internal/app/project_domain/project/application"
	deleteasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/delete_asset"
	downloadasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/download_asset"
	getasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/get_asset"
	updateasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/update_asset"
	uploadasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/upload_asset"

	"github.com/gin-gonic/gin"
)

type AssetController struct {
	app       *application.AssetApplication
	presenter *presenter.AssetPresenter
}

func NewAssetController(
	app *application.AssetApplication,
	presenter *presenter.AssetPresenter,
) *AssetController {
	return &AssetController{
		app:       app,
		presenter: presenter,
	}
}

type UploadAssetPayload struct {
	Name    string                 `form:"name" validate:"required"`
	Project *fileheader.FileHeader `form:"file" validate:"required"`
}

func (ctl *AssetController) GetAsset(ctx *gin.Context, request api.GetAssetRequestObject) (api.GetAssetResponseObject, error) {
	input := &getasset.Input{
		ID: request.Id,
	}

	output, err := ctl.app.GetAsset(input)
	if err != nil {
		return ctl.presenter.GetAsset.Error(err), nil
	}

	return ctl.presenter.GetAsset.Output(output), nil
}

func (ctl *AssetController) UploadAsset(ctx *gin.Context, request api.UploadAssetRequestObject) (api.UploadAssetResponseObject, error) {
	payload := &UploadAssetPayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		return ctl.presenter.UploadAsset.Error(uploadasset.ErrorInvalidInput), nil
	}

	input := &uploadasset.Input{
		ProjectID: request.Id,
		FileName:  payload.Name,
		File:      payload.Project.Buffer(),
	}

	output, err := ctl.app.UploadAsset(input)
	if err != nil {
		return ctl.presenter.UploadAsset.Error(err), nil
	}

	return ctl.presenter.UploadAsset.Output(output), nil
}

type UpdateAssetPayload struct {
	Name      *string                `form:"name"`
	File      *fileheader.FileHeader `form:"file"`
	IsEnabled *bool                  `form:"isEnabled"`
}

func (ctl *AssetController) UpdateAsset(ctx *gin.Context, request api.UpdateAssetRequestObject) (api.UpdateAssetResponseObject, error) {
	payload := &UpdateAssetPayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		return ctl.presenter.UpdateAsset.Error(updateasset.ErrorInvalidInput), nil
	}

	input := &updateasset.Input{
		ID:        request.Id,
		FileName:  payload.Name,
		IsEnabled: payload.IsEnabled,
	}

	if payload.File != nil {
		input.File = payload.File.Buffer()
	}

	output, err := ctl.app.UpdateAsset(input)
	if err != nil {
		return ctl.presenter.UpdateAsset.Error(err), nil
	}

	return ctl.presenter.UpdateAsset.Output(output), nil
}

func (ctl *AssetController) DeleteAsset(ctx *gin.Context, request api.DeleteAssetRequestObject) (api.DeleteAssetResponseObject, error) {
	input := &deleteasset.Input{
		ID: request.Id,
	}

	output, err := ctl.app.DeleteAsset(input)
	if err != nil {
		return ctl.presenter.DeleteAsset.Error(err), nil
	}

	return ctl.presenter.DeleteAsset.Output(output), nil
}

func (ctl *AssetController) DownloadAsset(ctx *gin.Context, request api.DownloadAssetRequestObject) (api.DownloadAssetResponseObject, error) {
	input := &downloadasset.Input{
		FileName: request.Id,
	}

	output, err := ctl.app.DownloadAsset(input)
	if err != nil {
		return ctl.presenter.DownloadAsset.Error(err), nil
	}

	return ctl.presenter.DownloadAsset.Output(output), nil
}
