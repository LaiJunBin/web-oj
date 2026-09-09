package presenter

import (
	espresenter "judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/event_stream"
	filepresenter "judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/file"
	jsonpresenter "judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/json"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
)

type ImagePresenter struct {
	ListImage   types.ListImagePresenterType
	GetImage    types.GetImagePresenterType
	CreateImage types.CreateImagePresenterType
	UpdateImage types.UpdateImagePresenterType
	DeleteImage types.DeleteImagePresenterType
	ListBuild   types.ListBuildPresenterType
	GetBuild    types.GetBuildPresenterType
	CancelBuild types.CancelBuildPresenterType
	Rebuild     types.RebuildPresenterType

	StreamBuildLog types.StreamBuildLogPresenterType

	ExportImage types.ExportImagePresenterType
	ImportImage types.ImportImagePresenterType
}

func NewImagePresenter() *ImagePresenter {
	return &ImagePresenter{
		ListImage:   jsonpresenter.NewListImagePresenter(),
		GetImage:    jsonpresenter.NewGetImagePresenter(),
		CreateImage: jsonpresenter.NewCreateImagePresenter(),
		UpdateImage: jsonpresenter.NewUpdateImagePresenter(),
		DeleteImage: jsonpresenter.NewDeleteImagePresenter(),
		ListBuild:   jsonpresenter.NewListBuildPresenter(),
		GetBuild:    jsonpresenter.NewGetBuildPresenter(),
		CancelBuild: jsonpresenter.NewCancelBuildPresenter(),
		Rebuild:     jsonpresenter.NewRebuildPresenter(),

		StreamBuildLog: espresenter.NewStreamBuildLogPresenter(),

		ExportImage: filepresenter.NewExportImagePresenter(),
		ImportImage: jsonpresenter.NewImportImagePresenter(),
	}
}
