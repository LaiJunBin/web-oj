package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	listpubliccontests "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/list_public_contests"
	"judge-backend/internal/app/shared/types/httperror"
)

type ListPublicContestsPresenter types.Presenter[listpubliccontests.Output, api.ListPublicContestsResponseObject]

type listPublicContestsPresenter struct{}

func NewListPublicContestsPresenter() ListPublicContestsPresenter {
	return &listPublicContestsPresenter{}
}

func (p *listPublicContestsPresenter) Output(output *listpubliccontests.Output) api.ListPublicContestsResponseObject {
	contests := make([]api.CommonContestSpecContest, 0, len(output.Contests))
	for _, contest := range output.Contests {
		contests = append(contests, api.CommonContestSpecContest{
			Id:   contest.ID,
			Name: contest.Name,
		})
	}

	return api.ListPublicContests200JSONResponse{
		Success: true,
		Data:    contests,
	}
}

func (p *listPublicContestsPresenter) Error(err error) api.ListPublicContestsResponseObject {
	return httperror.From(err).Response()
}
