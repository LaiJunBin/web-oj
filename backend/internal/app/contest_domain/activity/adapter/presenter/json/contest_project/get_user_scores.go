package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	getuserscores "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_user_scores"
)

type GetUserScoresPresenter types.Presenter[getuserscores.Output, api.GetUserScoresResponseObject]

type getUserScoresPresenter struct {
}

func NewGetUserScoresPresenter() GetUserScoresPresenter {
	return &getUserScoresPresenter{}
}

func (p *getUserScoresPresenter) Output(output *getuserscores.Output) api.GetUserScoresResponseObject {
	userScores := make([]api.AdminContestProjectSpecUserScore, 0, len(output.UserScores))
	for _, userScore := range output.UserScores {
		userScores = append(userScores, api.AdminContestProjectSpecUserScore{
			User: api.AdminContestProjectSpecUser{
				Id:       userScore.User.ID,
				Name:     userScore.User.Name,
				Username: userScore.User.Username,
			},
			Score:        userScore.Score,
			SubmissionId: userScore.SubmissionID,
		})
	}

	return api.GetUserScores200JSONResponse{
		Success: true,
		Data:    userScores,
	}
}

func (p *getUserScoresPresenter) Error(err error) api.GetUserScoresResponseObject {
	return httperror.From(err).Response()
}
