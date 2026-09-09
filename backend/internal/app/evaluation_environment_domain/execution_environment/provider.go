package execution_environment

import (
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/controller"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application"
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
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/factory"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/handler"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/provider"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/repository"
	eeservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/build_service"
	imageservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/image_service"
	buildquery "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/query/build_query"
	imagequery "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/query/image_query"
	coreprovider "judge-backend/internal/app/shared/types/provider"
)

func NewProvider() *coreprovider.Provider {
	providers := coreprovider.New(coreprovider.Config{
		Infrastructures: []interface{}{
			buildquery.NewService,
			imagequery.NewService,
			repository.NewImageRepository,
			repository.NewBuildRepository,
			buildservice.NewService,
			imageservice.NewService,

			// Bundle ports: the daemon adapters for save/load/digest, and the
			// tag-usage lookup a conflict reports its blast radius from.
			eeservice.NewImageArchiver,
			eeservice.NewImageLoader,
			eeservice.NewTagUsage,
		},
		Controllers: []interface{}{
			controller.NewImageController,
		},
		Presenters: []interface{}{
			presenter.NewImagePresenter,
		},
		Usecases: []interface{}{
			listimage.NewUsecase,
			getimage.NewUsecase,
			createimage.NewUsecase,
			updateimage.NewUsecase,
			deleteimage.NewUsecase,
			exportimage.NewUsecase,
			importimage.NewUsecase,
			listbuild.NewUsecase,
			getbuild.NewUsecase,
			cancelbuild.NewUsecase,
			rebuild.NewUsecase,
			streambuildlog.NewUsecase,
		},
		Applications: []interface{}{
			application.NewImageApplication,
		},
		Factories: []interface{}{
			factory.NewImageFactory,
			factory.NewBuildFactory,
		},
		Handlers: []interface{}{
			handler.RegisterBuildEventHandler,
		},
		Providers: []interface{}{
			provider.NewBuildQueueProvider,
			provider.NewDockerClient,
			provider.NewStoragePathProvider,
		},
	})

	return providers
}
