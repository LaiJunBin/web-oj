package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	getasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/get_asset"
)

type GetAssetPresenter types.Presenter[getasset.Output, api.GetAssetResponseObject]

type getAssetPresenter struct{}

func NewGetAssetPresenter() GetAssetPresenter {
	return &getAssetPresenter{}
}

func (p *getAssetPresenter) Output(output *getasset.Output) api.GetAssetResponseObject {
	return api.GetAsset200JSONResponse{
		Success: true,
		Data: api.SharedAsset{
			Id:           output.Asset.ID,
			Name:         output.Asset.Name,
			Size:         output.Asset.Size,
			IsEnabled:    output.Asset.Enabled,
			LastModified: output.Asset.LastModified,
		},
	}
}

func (p *getAssetPresenter) Error(err error) api.GetAssetResponseObject {
	switch {
	case errors.Is(err, getasset.ErrorInvalidID):
		return api.GetAsset400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, getasset.ErrorAssetNotFound):
		return api.GetAsset404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, getasset.ErrorInternalServer):
		fallthrough
	default:
		return api.GetAsset500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
