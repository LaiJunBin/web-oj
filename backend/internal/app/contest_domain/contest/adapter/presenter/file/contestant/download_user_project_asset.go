package filepresenter

import (
	"bytes"
	"judge-backend/api"
	downloaduserprojectasset "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/download_user_project_asset"
	"judge-backend/internal/app/shared/adapter/presenter/types"
	"judge-backend/internal/app/shared/types/httperror"
)

type DownloadUserProjectAssetPresenter types.Presenter[downloaduserprojectasset.Output, api.DownloadUserProjectAssetResponseObject]

type downloadUserProjectAssetPresenter struct{}

func NewDownloadUserProjectAssetPresenter() DownloadUserProjectAssetPresenter {
	return &downloadUserProjectAssetPresenter{}
}

func (p *downloadUserProjectAssetPresenter) Output(output *downloaduserprojectasset.Output) api.DownloadUserProjectAssetResponseObject {
	return api.DownloadUserProjectAsset200ApplicationoctetStreamResponse{
		Headers: api.DownloadUserProjectAsset200ResponseHeaders{
			ContentDisposition: "attachment; filename=" + output.File.Name,
		},
		ContentLength: output.File.Size,
		Body:          bytes.NewReader(output.File.Content),
	}
}

func (p *downloadUserProjectAssetPresenter) Error(err error) api.DownloadUserProjectAssetResponseObject {
	return httperror.From(err).Response()
}
