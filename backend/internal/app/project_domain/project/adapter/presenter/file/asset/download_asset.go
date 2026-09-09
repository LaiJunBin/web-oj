package filepresenter

import (
	"bytes"
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	downloadasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/download_asset"
)

type DownloadAssetPresenter types.Presenter[downloadasset.Output, api.DownloadAssetResponseObject]

type downloadAssetPresenter struct{}

func NewDownloadAssetPresenter() DownloadAssetPresenter {
	return &downloadAssetPresenter{}
}

func (p *downloadAssetPresenter) Output(output *downloadasset.Output) api.DownloadAssetResponseObject {
	return api.DownloadAsset200ApplicationoctetStreamResponse{
		Headers: api.DownloadAsset200ResponseHeaders{
			ContentDisposition: "attachment; filename=" + output.File.Name,
		},
		ContentLength: output.File.Size,
		Body:          bytes.NewReader(output.File.Content),
	}
}

func (p *downloadAssetPresenter) Error(err error) api.DownloadAssetResponseObject {
	switch {
	case errors.Is(err, downloadasset.ErrorInvalidFileName):
		return api.DownloadAsset400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, downloadasset.ErrorAssetNotFound):
		return api.DownloadAsset404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, downloadasset.ErrorInternalServer):
		fallthrough
	default:
		return api.DownloadAsset500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
