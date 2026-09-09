package filepresenter

import (
	"bytes"
	"errors"

	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	exportimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/export_image"
)

type exportImagePresenter struct{}

func NewExportImagePresenter() types.ExportImagePresenterType { return &exportImagePresenter{} }

func (p *exportImagePresenter) Output(output *exportimage.Output) api.ExportImageResponseObject {
	return api.ExportImage200ApplicationoctetStreamResponse{
		Headers: api.ExportImage200ResponseHeaders{
			ContentDisposition: "attachment; filename=" + output.Filename,
		},
		ContentLength: int64(len(output.Content)),
		Body:          bytes.NewReader(output.Content),
	}
}

func (p *exportImagePresenter) Error(err error) api.ExportImageResponseObject {
	switch {
	case errors.Is(err, exportimage.ErrorImageNotFound):
		return api.ExportImage404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	// A mode the API does not offer, or an archive of something never built:
	// both are the caller's to fix, so neither is a 500.
	case errors.Is(err, exportimage.ErrorInvalidMode), errors.Is(err, exportimage.ErrorNotBuilt):
		return api.ExportImage400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	default:
		return api.ExportImage500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
