package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	getuserprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_project_scoreboard"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetUserProjectScoreboardPresenter types.Presenter[getuserprojectscoreboard.Output, api.GetUserProjectScoreboardResponseObject]

type getUserProjectScoreboardPresenter struct{}

func NewGetUserProjectScoreboardPresenter() GetUserProjectScoreboardPresenter {
	return &getUserProjectScoreboardPresenter{}
}

func (p *getUserProjectScoreboardPresenter) Output(output *getuserprojectscoreboard.Output) api.GetUserProjectScoreboardResponseObject {
	heads := make([]api.UserProjectSpecProjectScoreboardHead, 0, len(output.Scoreboard.Heads))
	for _, head := range output.Scoreboard.Heads {
		heads = append(heads, api.UserProjectSpecProjectScoreboardHead{
			Id:   head.ID,
			Name: head.Name,
		})
	}

	rows := make([]api.UserProjectSpecProjectScoreboardData, 0, len(output.Scoreboard.Rows))
	for _, row := range output.Scoreboard.Rows {
		statuses := make([]api.SharedSubmissionGradingItemStatus, 0, len(row.Statuses))
		for _, status := range row.Statuses {
			statuses = append(statuses, api.SharedSubmissionGradingItemStatus(status))
		}

		rows = append(rows, api.UserProjectSpecProjectScoreboardData{
			Name:     row.Name,
			Username: row.Username,
			Statuses: statuses,
		})
	}

	return api.GetUserProjectScoreboard200JSONResponse{
		Success: true,
		Data: api.UserProjectSpecProjectScoreboard{
			Heads:   heads,
			Rows:    rows,
			Compact: output.Scoreboard.Compact,
		},
	}
}

func (p *getUserProjectScoreboardPresenter) Error(err error) api.GetUserProjectScoreboardResponseObject {
	return httperror.From(err).Response()
}
