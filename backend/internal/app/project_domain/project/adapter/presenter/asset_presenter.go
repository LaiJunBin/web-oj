package presenter

import (
	filepresenter "judge-backend/internal/app/project_domain/project/adapter/presenter/file/asset"
	jsonpresenter "judge-backend/internal/app/project_domain/project/adapter/presenter/json/asset"
)

type AssetPresenter struct {
	GetAsset      jsonpresenter.GetAssetPresenter
	UploadAsset   jsonpresenter.UploadAssetPresenter
	UpdateAsset   jsonpresenter.UpdateAssetPresenter
	DeleteAsset   jsonpresenter.DeleteAssetPresenter
	DownloadAsset filepresenter.DownloadAssetPresenter
}

func NewAssetPresenter() *AssetPresenter {
	return &AssetPresenter{
		GetAsset:      jsonpresenter.NewGetAssetPresenter(),
		UploadAsset:   jsonpresenter.NewUploadAssetPresenter(),
		UpdateAsset:   jsonpresenter.NewUpdateAssetPresenter(),
		DeleteAsset:   jsonpresenter.NewDeleteAssetPresenter(),
		DownloadAsset: filepresenter.NewDownloadAssetPresenter(),
	}
}
