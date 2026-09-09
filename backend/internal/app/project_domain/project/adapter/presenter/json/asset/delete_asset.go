package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	deleteasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/delete_asset"
)

type DeleteAssetPresenter types.Presenter[deleteasset.Output, api.DeleteAssetResponseObject]

type deleteAssetPresenter struct{}

func NewDeleteAssetPresenter() DeleteAssetPresenter {
	return &deleteAssetPresenter{}
}

func (p *deleteAssetPresenter) Output(output *deleteasset.Output) api.DeleteAssetResponseObject {
	return api.DeleteAsset200JSONResponse{
		Success: true,
	}
}

func (p *deleteAssetPresenter) Error(err error) api.DeleteAssetResponseObject {
	switch {
	case errors.Is(err, deleteasset.ErrorInvalidInput):
		return api.DeleteAsset400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, deleteasset.ErrorAssetNotFound):
		return api.DeleteAsset404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, deleteasset.ErrorInternalServer):
		fallthrough
	default:
		return api.DeleteAsset500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
