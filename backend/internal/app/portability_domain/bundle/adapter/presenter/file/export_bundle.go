package filepresenter

import (
	"bytes"
	"errors"

	"judge-backend/api"
	exportbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/export_bundle"
)

type Presenter[I any, O any] interface {
	Output(data *I) O
	Error(err error) O
}

type ExportBundlePresenter Presenter[exportbundle.Output, api.ExportBundleResponseObject]

type exportBundlePresenter struct{}

func NewExportBundlePresenter() ExportBundlePresenter { return &exportBundlePresenter{} }

func (p *exportBundlePresenter) Output(output *exportbundle.Output) api.ExportBundleResponseObject {
	return api.ExportBundle200ApplicationoctetStreamResponse{
		Headers: api.ExportBundle200ResponseHeaders{
			ContentDisposition: "attachment; filename=" + output.Filename,
		},
		ContentLength: int64(len(output.Content)),
		Body:          bytes.NewReader(output.Content),
	}
}

func (p *exportBundlePresenter) Error(err error) api.ExportBundleResponseObject {
	var notBuilt *exportbundle.NotBuiltError
	switch {
	// The caller's to fix, each with an instruction: build these tags, or
	// pick a mode the API knows, or point at a contest that exists.
	case errors.As(err, &notBuilt), errors.Is(err, exportbundle.ErrorInvalidMode), errors.Is(err, exportbundle.ErrorInvalidContestID):
		return api.ExportBundle400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, exportbundle.ErrorContestNotFound):
		return api.ExportBundle404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	default:
		return api.ExportBundle500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
