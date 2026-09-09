package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	uploadasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/upload_asset"
)

type UploadAssetPresenter types.Presenter[uploadasset.Output, api.UploadAssetResponseObject]

type uploadAssetPresenter struct{}

func NewUploadAssetPresenter() UploadAssetPresenter {
	return &uploadAssetPresenter{}
}

func (p *uploadAssetPresenter) Output(output *uploadasset.Output) api.UploadAssetResponseObject {
	return api.UploadAsset200JSONResponse{
		Success: true,
		Data:    output.ID,
	}
}

func (p *uploadAssetPresenter) Error(err error) api.UploadAssetResponseObject {
	switch {
	case errors.Is(err, uploadasset.ErrorInvalidInput):
		fallthrough
	case errors.Is(err, uploadasset.ErrorInvalidProjectID):
		fallthrough
	case errors.Is(err, uploadasset.ErrorInvalidFileName):
		fallthrough
	case errors.Is(err, uploadasset.ErrorInvalidFile):
		return api.UploadAsset400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, uploadasset.ErrorProjectNotFound):
		return api.UploadAsset404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, uploadasset.ErrorInternalServer):
		fallthrough
	default:
		return api.UploadAsset500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
