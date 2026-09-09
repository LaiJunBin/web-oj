package presenter

import (
	filepresenter "judge-backend/internal/app/project_domain/project/adapter/presenter/file/project"
	jsonpresenter "judge-backend/internal/app/project_domain/project/adapter/presenter/json/project"
)

type ProjectPresenter struct {
	ListProject   jsonpresenter.ListProjectPresenter
	GetProject    jsonpresenter.GetProjectPresenter
	CreateProject jsonpresenter.CreateProjectPresenter
	UpdateProject jsonpresenter.UpdateProjectPresenter
	DeleteProject jsonpresenter.DeleteProjectPresenter
	ExportProject filepresenter.ExportProjectPresenter
	ImportProject jsonpresenter.ImportProjectPresenter
}

func NewProjectPresenter() *ProjectPresenter {
	return &ProjectPresenter{
		ListProject:   jsonpresenter.NewListProjectPresenter(),
		GetProject:    jsonpresenter.NewGetProjectPresenter(),
		CreateProject: jsonpresenter.NewCreateProjectPresenter(),
		UpdateProject: jsonpresenter.NewUpdateProjectPresenter(),
		DeleteProject: jsonpresenter.NewDeleteProjectPresenter(),
		ExportProject: filepresenter.NewExportProjectPresenter(),
		ImportProject: jsonpresenter.NewImportProjectPresenter(),
	}
}
