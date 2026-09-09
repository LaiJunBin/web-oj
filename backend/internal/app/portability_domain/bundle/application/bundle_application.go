package application

import (
	exportbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/export_bundle"
	importbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/import_bundle"
	previewbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/preview_bundle"
)

type BundleApplication struct {
	exportBundle  exportbundle.Usecase
	previewBundle previewbundle.Usecase
	importBundle  importbundle.Usecase
}

func NewBundleApplication(
	exportBundle exportbundle.Usecase,
	previewBundle previewbundle.Usecase,
	importBundle importbundle.Usecase,
) *BundleApplication {
	return &BundleApplication{
		exportBundle:  exportBundle,
		previewBundle: previewBundle,
		importBundle:  importBundle,
	}
}

func (a *BundleApplication) ExportBundle(input *exportbundle.Input) (*exportbundle.Output, error) {
	return a.exportBundle.Execute(input)
}

func (a *BundleApplication) PreviewBundle(input *previewbundle.Input) (*previewbundle.Output, error) {
	return a.previewBundle.Execute(input)
}

func (a *BundleApplication) ImportBundle(input *importbundle.Input) (*importbundle.Output, error) {
	return a.importBundle.Execute(input)
}
