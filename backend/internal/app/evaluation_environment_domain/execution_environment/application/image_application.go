package application

import (
	cancelbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/cancel_build"
	createimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/create_image"
	deleteimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/delete_image"
	exportimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/export_image"
	getbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/get_build"
	getimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/get_image"
	importimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/import_image"
	listbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/list_build"
	listimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/list_image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/rebuild"
	streambuildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/stream_build_log"
	updateimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/update_image"
)

type ImageApplication struct {
	listImage      listimage.Usecase
	getImage       getimage.Usecase
	createImage    createimage.Usecase
	updateImage    updateimage.Usecase
	deleteImage    deleteimage.Usecase
	exportImage    exportimage.Usecase
	importImage    importimage.Usecase
	listBuild      listbuild.Usecase
	getBuild       getbuild.Usecase
	cancelBuild    cancelbuild.Usecase
	rebuild        rebuild.Usecase
	streamBuildLog streambuildlog.Usecase
}

func NewImageApplication(
	listImage listimage.Usecase,
	getImage getimage.Usecase,
	createImage createimage.Usecase,
	updateImage updateimage.Usecase,
	deleteImage deleteimage.Usecase,
	exportImage exportimage.Usecase,
	importImage importimage.Usecase,
	listBuild listbuild.Usecase,
	getBuild getbuild.Usecase,
	cancelBuild cancelbuild.Usecase,
	rebuild rebuild.Usecase,
	streamBuildLog streambuildlog.Usecase,
) *ImageApplication {
	return &ImageApplication{
		listImage:      listImage,
		getImage:       getImage,
		createImage:    createImage,
		updateImage:    updateImage,
		deleteImage:    deleteImage,
		exportImage:    exportImage,
		importImage:    importImage,
		listBuild:      listBuild,
		getBuild:       getBuild,
		cancelBuild:    cancelBuild,
		rebuild:        rebuild,
		streamBuildLog: streamBuildLog,
	}
}

func (a *ImageApplication) ListImage(input *listimage.Input) (*listimage.Output, error) {
	return a.listImage.Execute(input)
}

func (a *ImageApplication) GetImage(input *getimage.Input) (*getimage.Output, error) {
	return a.getImage.Execute(input)
}

func (a *ImageApplication) CreateImage(input *createimage.Input) (*createimage.Output, error) {
	return a.createImage.Execute(input)
}

func (a *ImageApplication) UpdateImage(input *updateimage.Input) (*updateimage.Output, error) {
	return a.updateImage.Execute(input)
}

func (a *ImageApplication) DeleteImage(input *deleteimage.Input) (*deleteimage.Output, error) {
	return a.deleteImage.Execute(input)
}

func (a *ImageApplication) ListBuild(input *listbuild.Input) (*listbuild.Output, error) {
	return a.listBuild.Execute(input)
}

func (a *ImageApplication) GetBuild(input *getbuild.Input) (*getbuild.Output, error) {
	return a.getBuild.Execute(input)
}

func (a *ImageApplication) CancelBuild(input *cancelbuild.Input) (*cancelbuild.Output, error) {
	return a.cancelBuild.Execute(input)
}

func (a *ImageApplication) Rebuild(input *rebuild.Input) (*rebuild.Output, error) {
	return a.rebuild.Execute(input)
}

func (a *ImageApplication) StreamBuildLog(input *streambuildlog.Input) (*streambuildlog.Output, error) {
	return a.streamBuildLog.Execute(input)
}

func (a *ImageApplication) ExportImage(input *exportimage.Input) (*exportimage.Output, error) {
	return a.exportImage.Execute(input)
}

func (a *ImageApplication) ImportImage(input *importimage.Input) (*importimage.Output, error) {
	return a.importImage.Execute(input)
}
