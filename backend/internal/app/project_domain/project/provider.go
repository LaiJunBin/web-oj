package project

import (
	projectcontroller "judge-backend/internal/app/project_domain/project/adapter/controller"
	"judge-backend/internal/app/project_domain/project/adapter/presenter"
	"judge-backend/internal/app/project_domain/project/application"
	deleteasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/delete_asset"
	downloadasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/download_asset"
	getasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/get_asset"
	updateasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/update_asset"
	uploadasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/upload_asset"
	creategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/create_grading_item"
	deletegradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/delete_grading_item"
	getgradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/get_grading_item"
	listgradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/list_grading_item"
	movegradingitemorder "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/move_grading_item_order"
	updategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/update_grading_item"
	downloadgradingproject "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/download_grading_project"
	listgradingprojectfiles "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/list_grading_project_files"
	putgradingproject "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/put_grading_project"
	readgradingprojectfile "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/read_grading_project_file"
	createlifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/create_lifecycle"
	deletelifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/delete_lifecycle"
	getlifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/get_lifecycle"
	updatelifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/update_lifecycle"
	createproject "judge-backend/internal/app/project_domain/project/application/usecase/project/create_project"
	deleteproject "judge-backend/internal/app/project_domain/project/application/usecase/project/delete_project"
	exportproject "judge-backend/internal/app/project_domain/project/application/usecase/project/export_project"
	getproject "judge-backend/internal/app/project_domain/project/application/usecase/project/get_project"
	importproject "judge-backend/internal/app/project_domain/project/application/usecase/project/import_project"
	listproject "judge-backend/internal/app/project_domain/project/application/usecase/project/list_project"
	updateproject "judge-backend/internal/app/project_domain/project/application/usecase/project/update_project"
	createtestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/create_testcase"
	deletetestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/delete_testcase"
	gettestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/get_testcase"
	listprojecttestcases "judge-backend/internal/app/project_domain/project/application/usecase/testcase/list_project_testcases"
	updatetestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/update_testcase"
	"judge-backend/internal/app/project_domain/project/infrastructure/provider"
	"judge-backend/internal/app/project_domain/project/infrastructure/query"
	"judge-backend/internal/app/project_domain/project/infrastructure/repository"
	"judge-backend/internal/app/project_domain/project/infrastructure/service"
	coreprovider "judge-backend/internal/app/shared/types/provider"
)

func NewProvider() *coreprovider.Provider {
	return coreprovider.New(coreprovider.Config{
		Infrastructures: []interface{}{
			query.NewProjectQuery,
			query.NewAssetQuery,
			query.NewLifecycleQuery,
			query.NewGradingItemQuery,
			query.NewTestcaseQuery,
			query.NewProjectExportQuery,

			repository.NewProjectRepository,
			repository.NewImageRepository,
			repository.NewAssetRepository,
			repository.NewLifecycleRepository,
			repository.NewGradingItemRepository,
			repository.NewTestcaseRepository,
		},
		Controllers: []interface{}{
			projectcontroller.NewProjectController,
			projectcontroller.NewAssetController,
			projectcontroller.NewLifecycleController,
			projectcontroller.NewGradingItemController,
			projectcontroller.NewTestcaseController,
			projectcontroller.NewGradingProjectController,
		},
		Presenters: []interface{}{
			presenter.NewProjectPresenter,
			presenter.NewAssetPresenter,
			presenter.NewLifecyclePresenter,
			presenter.NewGradingItemPresenter,
			presenter.NewTestcasePresenter,
			presenter.NewGradingProjectPresenter,
		},
		Usecases: []interface{}{
			listproject.NewUsecase,
			getproject.NewUsecase,
			createproject.NewUsecase,
			updateproject.NewUsecase,
			deleteproject.NewUsecase,
			exportproject.NewUsecase,
			importproject.NewUsecase,

			getasset.NewUsecase,
			uploadasset.NewUsecase,
			updateasset.NewUsecase,
			deleteasset.NewUsecase,
			downloadasset.NewUsecase,

			getlifecycle.NewUsecase,
			createlifecycle.NewUsecase,
			updatelifecycle.NewUsecase,
			deletelifecycle.NewUsecase,

			listgradingitem.NewUsecase,
			getgradingitem.NewUsecase,
			creategradingitem.NewUsecase,
			updategradingitem.NewUsecase,
			deletegradingitem.NewUsecase,
			movegradingitemorder.NewUsecase,

			listprojecttestcases.NewUsecase,
			gettestcase.NewUsecase,
			createtestcase.NewUsecase,
			updatetestcase.NewUsecase,
			deletetestcase.NewUsecase,

			putgradingproject.NewUsecase,
			listgradingprojectfiles.NewUsecase,
			readgradingprojectfile.NewUsecase,
			downloadgradingproject.NewUsecase,
		},
		Applications: []interface{}{
			application.NewProjectApplication,
			application.NewAssetApplication,
			application.NewLifecycleApplication,
			application.NewGradingItemApplication,
			application.NewTestcaseApplication,
			application.NewGradingProjectApplication,
		},
		Services: []interface{}{
			service.NewAssetService,

			service.NewGradingProjectService,
			service.NewBase64Service,
			service.NewResultParserValidatorService,

			// Bundle ports: the import usecase reaches the image catalogue and
			// the two file uploaders through these rather than importing them.
			service.NewImageCatalogue,
			service.NewProjectNamer,
			service.NewAssetUploader,
			service.NewGradingFilesUploader,
			service.NewExportClock,
		},
		Providers: []interface{}{
			provider.NewStoragePathProvider,
		},
	})
}
