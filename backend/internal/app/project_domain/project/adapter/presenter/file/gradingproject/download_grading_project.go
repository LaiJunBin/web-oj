package filepresenter

import (
	"bytes"
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	downloadgradingproject "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/download_grading_project"
)

type DownloadGradingProjectPresenter types.Presenter[downloadgradingproject.Output, api.DownloadGradingProjectResponseObject]

type downloadGradingProjectPresenter struct{}

func NewDownloadGradingProjectPresenter() DownloadGradingProjectPresenter {
	return &downloadGradingProjectPresenter{}
}

func (p *downloadGradingProjectPresenter) Output(output *downloadgradingproject.Output) api.DownloadGradingProjectResponseObject {
	return api.DownloadGradingProject200ApplicationoctetStreamResponse{
		Headers: api.DownloadGradingProject200ResponseHeaders{
			ContentDisposition: "attachment; filename=" + output.File.Name,
		},
		ContentLength: output.File.Size,
		Body:          bytes.NewReader(output.File.Content),
	}
}

func (p *downloadGradingProjectPresenter) Error(err error) api.DownloadGradingProjectResponseObject {
	switch {
	case errors.Is(err, downloadgradingproject.ErrorInvalidProjectID):
		return api.DownloadGradingProject400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, downloadgradingproject.ErrorProjectNotFound):
		return api.DownloadGradingProject404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, downloadgradingproject.ErrorInternalServer):
		fallthrough
	default:
		return api.DownloadGradingProject500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
