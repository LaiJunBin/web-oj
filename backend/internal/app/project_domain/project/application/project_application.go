package application

import (
	createproject "judge-backend/internal/app/project_domain/project/application/usecase/project/create_project"
	deleteproject "judge-backend/internal/app/project_domain/project/application/usecase/project/delete_project"
	exportproject "judge-backend/internal/app/project_domain/project/application/usecase/project/export_project"
	getproject "judge-backend/internal/app/project_domain/project/application/usecase/project/get_project"
	importproject "judge-backend/internal/app/project_domain/project/application/usecase/project/import_project"
	listproject "judge-backend/internal/app/project_domain/project/application/usecase/project/list_project"
	updateproject "judge-backend/internal/app/project_domain/project/application/usecase/project/update_project"
)

type ProjectApplication struct {
	listProject   listproject.Usecase
	getProject    getproject.Usecase
	createProject createproject.Usecase
	updateProject updateproject.Usecase
	deleteProject deleteproject.Usecase
	exportProject exportproject.Usecase
	importProject importproject.Usecase
}

func NewProjectApplication(
	listProject listproject.Usecase,
	getProject getproject.Usecase,
	createProject createproject.Usecase,
	updateProject updateproject.Usecase,
	deleteProject deleteproject.Usecase,
	exportProject exportproject.Usecase,
	importProject importproject.Usecase,
) *ProjectApplication {
	return &ProjectApplication{
		listProject:   listProject,
		getProject:    getProject,
		createProject: createProject,
		updateProject: updateProject,
		deleteProject: deleteProject,
		exportProject: exportProject,
		importProject: importProject,
	}
}

func (a ProjectApplication) ListProject(input *listproject.Input) (*listproject.Output, error) {
	return a.listProject.Execute(input)
}

func (a ProjectApplication) GetProject(input *getproject.Input) (*getproject.Output, error) {
	return a.getProject.Execute(input)
}

func (a ProjectApplication) CreateProject(input *createproject.Input) (*createproject.Output, error) {
	return a.createProject.Execute(input)
}

func (a ProjectApplication) UpdateProject(input *updateproject.Input) (*updateproject.Output, error) {
	return a.updateProject.Execute(input)
}

func (a ProjectApplication) DeleteProject(input *deleteproject.Input) (*deleteproject.Output, error) {
	return a.deleteProject.Execute(input)
}

func (a ProjectApplication) ExportProject(input *exportproject.Input) (*exportproject.Output, error) {
	return a.exportProject.Execute(input)
}

func (a ProjectApplication) ImportProject(input *importproject.Input) (*importproject.Output, error) {
	return a.importProject.Execute(input)
}
