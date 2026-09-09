package bundle

import (
	bundlecontroller "judge-backend/internal/app/portability_domain/bundle/adapter/controller"
	filepresenter "judge-backend/internal/app/portability_domain/bundle/adapter/presenter/file"
	jsonpresenter "judge-backend/internal/app/portability_domain/bundle/adapter/presenter/json"
	"judge-backend/internal/app/portability_domain/bundle/application"
	exportbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/export_bundle"
	importbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/import_bundle"
	previewbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/preview_bundle"
	"judge-backend/internal/app/portability_domain/bundle/infrastructure/service"
	coreprovider "judge-backend/internal/app/shared/types/provider"
)

func NewProvider() *coreprovider.Provider {
	return coreprovider.New(coreprovider.Config{
		Controllers: []interface{}{
			bundlecontroller.NewBundleController,
		},
		Presenters: []interface{}{
			jsonpresenter.NewPreviewBundlePresenter,
			jsonpresenter.NewImportBundlePresenter,
			filepresenter.NewExportBundlePresenter,
		},
		Usecases: []interface{}{
			exportbundle.NewUsecase,
			previewbundle.NewUsecase,
			importbundle.NewUsecase,
		},
		Applications: []interface{}{
			application.NewBundleApplication,
		},
		Services: []interface{}{
			service.NewClock,
		},
	})
}
