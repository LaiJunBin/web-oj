package application

import (
	downloadgradingproject "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/download_grading_project"
	listgradingprojectfiles "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/list_grading_project_files"
	putgradingproject "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/put_grading_project"
	readgradingprojectfile "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/read_grading_project_file"
)

type GradingProjectApplication struct {
	putGradingProject       putgradingproject.Usecase
	listGradingProjectFiles listgradingprojectfiles.Usecase
	readGradingProjectFile  readgradingprojectfile.Usecase
	downloadGradingProject  downloadgradingproject.Usecase
}

func NewGradingProjectApplication(
	putGradingProject putgradingproject.Usecase,
	listGradingProjectFiles listgradingprojectfiles.Usecase,
	readGradingProjectFile readgradingprojectfile.Usecase,
	downloadGradingProject downloadgradingproject.Usecase,
) *GradingProjectApplication {
	return &GradingProjectApplication{
		putGradingProject:       putGradingProject,
		listGradingProjectFiles: listGradingProjectFiles,
		readGradingProjectFile:  readGradingProjectFile,
		downloadGradingProject:  downloadGradingProject,
	}
}

func (a GradingProjectApplication) PutGradingProject(input *putgradingproject.Input) (*putgradingproject.Output, error) {
	return a.putGradingProject.Execute(input)
}

func (a GradingProjectApplication) ListGradingProjectFiles(input *listgradingprojectfiles.Input) (*listgradingprojectfiles.Output, error) {
	return a.listGradingProjectFiles.Execute(input)
}

func (a GradingProjectApplication) ReadGradingProjectFile(input *readgradingprojectfile.Input) (*readgradingprojectfile.Output, error) {
	return a.readGradingProjectFile.Execute(input)
}

func (a GradingProjectApplication) DownloadGradingProject(input *downloadgradingproject.Input) (*downloadgradingproject.Output, error) {
	return a.downloadGradingProject.Execute(input)
}
