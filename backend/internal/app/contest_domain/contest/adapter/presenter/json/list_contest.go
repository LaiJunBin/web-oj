package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	listcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/list_contest"
	"judge-backend/internal/app/shared/types/httperror"
)

type ListContestPresenter types.Presenter[listcontest.Output, api.ListContestResponseObject]

type listContestPresenter struct {
}

func NewListContestPresenter() ListContestPresenter {
	return &listContestPresenter{}
}

func (p *listContestPresenter) Output(output *listcontest.Output) api.ListContestResponseObject {
	contests := make([]api.AdminContestSpecContest, 0)
	for _, contest := range output.Contests {
		contests = append(contests, api.AdminContestSpecContest{
			Id:             contest.ID,
			Name:           contest.Name,
			Description:    contest.Description,
			IsPublic:       contest.IsPublic,
			TotalProjects:  contest.TotalProjects,
			ActivateTime:   contest.ActivationTime,
			DeactivateTime: contest.DeactivationTime,
			CreatedAt:      contest.CreatedAt,
			UpdatedAt:      contest.UpdatedAt,
		})
	}

	return api.ListContest200JSONResponse{
		Success: true,
		Data:    contests,
	}
}

func (p *listContestPresenter) Error(err error) api.ListContestResponseObject {
	return httperror.From(err).Response()
}
