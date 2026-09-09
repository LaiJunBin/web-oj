package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	getpublicprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/get_public_project_scoreboard"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetPublicProjectScoreboardPresenter types.Presenter[getpublicprojectscoreboard.Output, api.GetPublicProjectScoreboardResponseObject]

type getPublicProjectScoreboardPresenter struct{}

func NewGetPublicProjectScoreboardPresenter() GetPublicProjectScoreboardPresenter {
	return &getPublicProjectScoreboardPresenter{}
}

func (p *getPublicProjectScoreboardPresenter) Output(output *getpublicprojectscoreboard.Output) api.GetPublicProjectScoreboardResponseObject {
	heads := make([]api.CommonProjectSpecProjectScoreboardHead, 0, len(output.Scoreboard.Heads))
	for _, head := range output.Scoreboard.Heads {
		heads = append(heads, api.CommonProjectSpecProjectScoreboardHead{
			Id:   head.ID,
			Name: head.Name,
		})
	}

	rows := make([]api.CommonProjectSpecProjectScoreboardData, 0, len(output.Scoreboard.Rows))
	for _, row := range output.Scoreboard.Rows {
		statuses := make([]api.SharedSubmissionGradingItemStatus, 0, len(row.Statuses))
		for _, status := range row.Statuses {
			statuses = append(statuses, api.SharedSubmissionGradingItemStatus(status))
		}

		rows = append(rows, api.CommonProjectSpecProjectScoreboardData{
			Name:     row.Name,
			Username: row.Username,
			Statuses: statuses,
		})
	}

	return api.GetPublicProjectScoreboard200JSONResponse{
		Success: true,
		Data: api.CommonProjectSpecProjectScoreboard{
			Heads:   heads,
			Rows:    rows,
			Compact: output.Scoreboard.Compact,
		},
	}
}

func (p *getPublicProjectScoreboardPresenter) Error(err error) api.GetPublicProjectScoreboardResponseObject {
	return httperror.From(err).Response()
}
