package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	rejudgecontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/rejudge_contest_project"
)

type RejudgeContestProjectPresenter types.Presenter[rejudgecontestproject.Output, api.RejudgeContestProjectResponseObject]

type rejudgeContestProjectPresenter struct {
}

func NewRejudgeContestProjectPresenter() RejudgeContestProjectPresenter {
	return &rejudgeContestProjectPresenter{}
}

func (p *rejudgeContestProjectPresenter) Output(output *rejudgecontestproject.Output) api.RejudgeContestProjectResponseObject {
	return api.RejudgeContestProject200JSONResponse{
		Success: true,
	}
}

func (p *rejudgeContestProjectPresenter) Error(err error) api.RejudgeContestProjectResponseObject {
	return httperror.From(err).Response()
}
