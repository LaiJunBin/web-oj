package filepresenter

import (
	"bytes"
	"errors"

	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	exportcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/export_contest"
)

type ExportContestPresenter types.Presenter[exportcontest.Output, api.ExportContestResponseObject]

type exportContestPresenter struct{}

func NewExportContestPresenter() ExportContestPresenter { return &exportContestPresenter{} }

func (p *exportContestPresenter) Output(output *exportcontest.Output) api.ExportContestResponseObject {
	return api.ExportContest200ApplicationoctetStreamResponse{
		Headers: api.ExportContest200ResponseHeaders{
			ContentDisposition: "attachment; filename=" + output.Filename,
		},
		ContentLength: int64(len(output.Content)),
		Body:          bytes.NewReader(output.Content),
	}
}

func (p *exportContestPresenter) Error(err error) api.ExportContestResponseObject {
	switch {
	case errors.Is(err, exportcontest.ErrorInvalidContestID):
		return api.ExportContest400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, exportcontest.ErrorContestNotFound):
		return api.ExportContest404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	default:
		return api.ExportContest500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
