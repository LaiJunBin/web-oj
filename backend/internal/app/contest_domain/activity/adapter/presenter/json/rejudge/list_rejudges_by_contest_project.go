package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	listrejudgesbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/list_rejudges_by_contest_project"
	"judge-backend/internal/app/shared/types/httperror"
)

type ListRejudgesByContestProjectPresenter types.Presenter[listrejudgesbycontestproject.Output, api.ListRejudgesByContestProjectResponseObject]

type listRejudgesByContestProjectPresenter struct {
}

func NewListRejudgesByContestProjectPresenter() ListRejudgesByContestProjectPresenter {
	return &listRejudgesByContestProjectPresenter{}
}

func (p *listRejudgesByContestProjectPresenter) Output(output *listrejudgesbycontestproject.Output) api.ListRejudgesByContestProjectResponseObject {
	return api.ListRejudgesByContestProject200JSONResponse{
		Success: true,
		Data:    rejudgeList(output.Rejudges),
	}
}

func (p *listRejudgesByContestProjectPresenter) Error(err error) api.ListRejudgesByContestProjectResponseObject {
	return httperror.From(err).Response()
}
