package jsonpresenter

import (
	"errors"

	"judge-backend/api"
	importbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/import_bundle"
	previewbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/preview_bundle"
)

type ImportBundlePresenter Presenter[importbundle.Output, api.ImportBundleResponseObject]

type importBundlePresenter struct{}

func NewImportBundlePresenter() ImportBundlePresenter { return &importBundlePresenter{} }

func (p *importBundlePresenter) Output(output *importbundle.Output) api.ImportBundleResponseObject {
	report := api.AdminBundleSpecBundleImportReport{
		CreatedImages:      output.CreatedImages,
		ReusedImages:       output.ReusedImages,
		QueuedBuilds:       output.QueuedBuilds,
		CreatedProjectIds:  output.CreatedProjectIDs,
		ContestId:          output.ContestID,
		UnresolvedProjects: output.UnresolvedProjects,
		Warnings:           output.Warnings,
	}

	return api.ImportBundle200JSONResponse{
		Success: true,
		Data:    report,
	}
}

// Error keeps the three refusals apart, because the operator's next move
// differs for each: a malformed bundle has to be fixed (400), a conflict has
// to be decided (409, with force as the override), and anything else is this
// deployment's problem (500).
func (p *importBundlePresenter) Error(err error) api.ImportBundleResponseObject {
	// A partial import is not a success: some parts landed and then a later one
	// failed. Report it as a failure with the whole story (what failed and what
	// already exists) folded into the message, so automation cannot read a
	// half-done import as done. Deliberately not a 409 — there is nothing to
	// "force" past; the operator finishes or removes the created parts by hand.
	var partial *importbundle.PartialImportError
	if errors.As(err, &partial) {
		return api.ImportBundle500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage("匯入未完全成功"),
			Details: partial.Details(),
		}
	}

	var conflict *importbundle.ConflictError
	if errors.As(err, &conflict) {
		return api.ImportBundle409JSONResponse{
			Success: false,
			Message: err.Error(),
			Details: conflict.Details,
		}
	}

	var invalid *previewbundle.InvalidBundleError
	if errors.As(err, &invalid) {
		return api.ImportBundle400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			Details: invalid.Details,
		}
	}

	return api.ImportBundle500JSONResponse{
		Success: false,
		Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
	}
}
