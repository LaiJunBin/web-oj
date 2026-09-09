package presenter

import (
	filepresenter "judge-backend/internal/app/project_domain/project/adapter/presenter/file/gradingproject"
	jsonpresenter "judge-backend/internal/app/project_domain/project/adapter/presenter/json/gradingproject"
)

type GradingProjectPresenter struct {
	PutGradingProject       jsonpresenter.PutGradingProjectPresenter
	ListGradingProjectFiles jsonpresenter.ListGradingProjectFilesPresenter
	ReadGradingProjectFile  jsonpresenter.ReadGradingProjectFilePresenter
	DownloadGradingProject  filepresenter.DownloadGradingProjectPresenter
}

func NewGradingProjectPresenter() *GradingProjectPresenter {
	return &GradingProjectPresenter{
		PutGradingProject:       jsonpresenter.NewPutGradingProjectPresenter(),
		ListGradingProjectFiles: jsonpresenter.NewListGradingProjectFilesPresenter(),
		ReadGradingProjectFile:  jsonpresenter.NewReadGradingProjectFilePresenter(),
		DownloadGradingProject:  filepresenter.NewDownloadGradingProjectPresenter(),
	}
}
