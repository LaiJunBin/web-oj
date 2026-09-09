package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	createcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/create_contest"
	"judge-backend/internal/app/shared/types/httperror"
)

type CreateContestPresenter types.Presenter[createcontest.Output, api.CreateContestResponseObject]

type createContestPresenter struct {
}

func NewCreateContestPresenter() CreateContestPresenter {
	return &createContestPresenter{}
}

func (p *createContestPresenter) Output(output *createcontest.Output) api.CreateContestResponseObject {
	return api.CreateContest200JSONResponse{
		Success: true,
		Data:    output.ID,
	}
}

func (p *createContestPresenter) Error(err error) api.CreateContestResponseObject {
	return httperror.From(err).Response()
}
