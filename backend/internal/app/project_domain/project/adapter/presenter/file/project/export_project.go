package filepresenter

import (
	"bytes"
	"errors"

	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	exportproject "judge-backend/internal/app/project_domain/project/application/usecase/project/export_project"
)

type ExportProjectPresenter types.Presenter[exportproject.Output, api.ExportProjectResponseObject]

type exportProjectPresenter struct{}

func NewExportProjectPresenter() ExportProjectPresenter {
	return &exportProjectPresenter{}
}

func (p *exportProjectPresenter) Output(output *exportproject.Output) api.ExportProjectResponseObject {
	return api.ExportProject200ApplicationoctetStreamResponse{
		Headers: api.ExportProject200ResponseHeaders{
			ContentDisposition: "attachment; filename=" + output.Filename,
		},
		ContentLength: int64(len(output.Content)),
		Body:          bytes.NewReader(output.Content),
	}
}

func (p *exportProjectPresenter) Error(err error) api.ExportProjectResponseObject {
	switch {
	case errors.Is(err, exportproject.ErrorInvalidProjectID):
		return api.ExportProject400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, exportproject.ErrorProjectNotFound):
		return api.ExportProject404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	default:
		return api.ExportProject500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
