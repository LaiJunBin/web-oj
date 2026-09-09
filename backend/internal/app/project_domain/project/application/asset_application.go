package application

import (
	deleteasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/delete_asset"
	downloadasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/download_asset"
	getasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/get_asset"
	updateasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/update_asset"
	uploadasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/upload_asset"
)

type AssetApplication struct {
	getAsset      getasset.Usecase
	uploadAsset   uploadasset.Usecase
	updateAsset   updateasset.Usecase
	deleteAsset   deleteasset.Usecase
	downloadAsset downloadasset.Usecase
}

func NewAssetApplication(
	getAsset getasset.Usecase,
	uploadAsset uploadasset.Usecase,
	updateAsset updateasset.Usecase,
	deleteAsset deleteasset.Usecase,
	downloadAsset downloadasset.Usecase,
) *AssetApplication {
	return &AssetApplication{
		getAsset:      getAsset,
		uploadAsset:   uploadAsset,
		updateAsset:   updateAsset,
		deleteAsset:   deleteAsset,
		downloadAsset: downloadAsset,
	}
}

func (a AssetApplication) GetAsset(input *getasset.Input) (*getasset.Output, error) {
	return a.getAsset.Execute(input)
}

func (a AssetApplication) UploadAsset(input *uploadasset.Input) (*uploadasset.Output, error) {
	return a.uploadAsset.Execute(input)
}

func (a AssetApplication) UpdateAsset(input *updateasset.Input) (*updateasset.Output, error) {
	return a.updateAsset.Execute(input)
}

func (a AssetApplication) DeleteAsset(input *deleteasset.Input) (*deleteasset.Output, error) {
	return a.deleteAsset.Execute(input)
}

func (a AssetApplication) DownloadAsset(input *downloadasset.Input) (*downloadasset.Output, error) {
	return a.downloadAsset.Execute(input)
}
