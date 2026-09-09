package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	updateasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/update_asset"
)

type UpdateAssetPresenter types.Presenter[updateasset.Output, api.UpdateAssetResponseObject]

type updateAssetPresenter struct{}

func NewUpdateAssetPresenter() UpdateAssetPresenter {
	return &updateAssetPresenter{}
}

func (p *updateAssetPresenter) Output(output *updateasset.Output) api.UpdateAssetResponseObject {
	return api.UpdateAsset200JSONResponse{
		Success: true,
	}
}

func (p *updateAssetPresenter) Error(err error) api.UpdateAssetResponseObject {
	switch {
	case errors.Is(err, updateasset.ErrorInvalidInput):
		fallthrough
	case errors.Is(err, updateasset.ErrorInvalidFileName):
		fallthrough
	case errors.Is(err, updateasset.ErrorInvalidFile):
		return api.UpdateAsset400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, updateasset.ErrorAssetNotFound):
		return api.UpdateAsset404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, updateasset.ErrorInternalServer):
		fallthrough
	default:
		return api.UpdateAsset500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
