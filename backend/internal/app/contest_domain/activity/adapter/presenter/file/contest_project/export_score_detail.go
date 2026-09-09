package filepresenter

import (
	"bytes"
	"errors"
	"fmt"
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	exportscoredetail "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/export_score_detail"
)

type ExportScoreDetailPresenter types.Presenter[exportscoredetail.Output, api.ExportContestProjectScoreDetailResponseObject]

type exportScoreDetailPresenter struct{}

func NewExportScoreDetailPresenter() ExportScoreDetailPresenter {
	return &exportScoreDetailPresenter{}
}

func (p *exportScoreDetailPresenter) Output(output *exportscoredetail.Output) api.ExportContestProjectScoreDetailResponseObject {
	return api.ExportContestProjectScoreDetail200ApplicationoctetStreamResponse{
		Headers: api.ExportContestProjectScoreDetail200ResponseHeaders{
			// Quoted per RFC 6266: the filename starts with the contest
			// project's name, and an unquoted value breaks at its first space.
			ContentDisposition: fmt.Sprintf("attachment; filename=%q", output.Filename),
		},
		ContentLength: int64(len(output.Content)),
		Body:          bytes.NewReader(output.Content),
	}
}

func (p *exportScoreDetailPresenter) Error(err error) api.ExportContestProjectScoreDetailResponseObject {
	switch {
	case errors.Is(err, exportscoredetail.ErrorContestProjectNotFound):
		return api.ExportContestProjectScoreDetail404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, exportscoredetail.ErrorInternalServer):
		fallthrough
	default:
		return api.ExportContestProjectScoreDetail500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
