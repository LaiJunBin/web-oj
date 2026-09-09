package bundlecontroller

import (
	"judge-backend/api"
	filepresenter "judge-backend/internal/app/portability_domain/bundle/adapter/presenter/file"
	jsonpresenter "judge-backend/internal/app/portability_domain/bundle/adapter/presenter/json"
	"judge-backend/internal/app/portability_domain/bundle/application"
	exportbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/export_bundle"
	importbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/import_bundle"
	previewbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/preview_bundle"
	fileheader "judge-backend/internal/app/shared/types/file_header"
	formrequest "judge-backend/internal/app/shared/types/form_request"

	"github.com/gin-gonic/gin"
)

type BundleController struct {
	app     *application.BundleApplication
	preview jsonpresenter.PreviewBundlePresenter
	imports jsonpresenter.ImportBundlePresenter
	exports filepresenter.ExportBundlePresenter
}

func NewBundleController(
	app *application.BundleApplication,
	preview jsonpresenter.PreviewBundlePresenter,
	imports jsonpresenter.ImportBundlePresenter,
	exports filepresenter.ExportBundlePresenter,
) *BundleController {
	return &BundleController{app: app, preview: preview, imports: imports, exports: exports}
}

type PreviewBundlePayload struct {
	Bundle *fileheader.FileHeader `form:"bundle" validate:"required"`
}

type ImportBundlePayload struct {
	Bundle *fileheader.FileHeader `form:"bundle" validate:"required"`
	Force  bool                   `form:"force"`
}

func (ctl *BundleController) PreviewBundle(ctx *gin.Context, request api.PreviewBundleRequestObject) (api.PreviewBundleResponseObject, error) {
	payload := &PreviewBundlePayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		return ctl.preview.Error(&previewbundle.InvalidBundleError{
			Details: []string{"沒有收到 bundle 檔案"},
		}), nil
	}

	output, err := ctl.app.PreviewBundle(&previewbundle.Input{Bundle: payload.Bundle.Buffer()})
	if err != nil {
		return ctl.preview.Error(err), nil
	}

	return ctl.preview.Output(output), nil
}

func (ctl *BundleController) ImportBundle(ctx *gin.Context, request api.ImportBundleRequestObject) (api.ImportBundleResponseObject, error) {
	payload := &ImportBundlePayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		return ctl.imports.Error(&previewbundle.InvalidBundleError{
			Details: []string{"沒有收到 bundle 檔案"},
		}), nil
	}

	output, err := ctl.app.ImportBundle(&importbundle.Input{
		Bundle: payload.Bundle.Buffer(),
		Force:  payload.Force,
	})
	if err != nil {
		return ctl.imports.Error(err), nil
	}

	return ctl.imports.Output(output), nil
}

func (ctl *BundleController) ExportBundle(ctx *gin.Context, request api.ExportBundleRequestObject) (api.ExportBundleResponseObject, error) {
	output, err := ctl.app.ExportBundle(&exportbundle.Input{
		ContestID: request.Id,
		Mode:      string(request.Params.Mode),
	})
	if err != nil {
		return ctl.exports.Error(err), nil
	}

	return ctl.exports.Output(output), nil
}
