package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	listusercontests "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_contests"
	"judge-backend/internal/app/shared/types/httperror"
)

type ListUserContestsPresenter types.Presenter[listusercontests.Output, api.ListUserContestsResponseObject]

type listUserContestsPresenter struct{}

func NewListUserContestsPresenter() ListUserContestsPresenter {
	return &listUserContestsPresenter{}
}

func (p *listUserContestsPresenter) Output(output *listusercontests.Output) api.ListUserContestsResponseObject {
	contests := make([]api.CommonContestSpecContest, 0, len(output.Contests))
	for _, contest := range output.Contests {
		contests = append(contests, api.CommonContestSpecContest{
			Id:   contest.ID,
			Name: contest.Name,
		})
	}

	return api.ListUserContests200JSONResponse{
		Success: true,
		Data:    contests,
	}
}

func (p *listUserContestsPresenter) Error(err error) api.ListUserContestsResponseObject {
	return httperror.From(err).Response()
}
