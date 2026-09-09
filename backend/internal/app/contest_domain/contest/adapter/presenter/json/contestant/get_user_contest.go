package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	getusercontest "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_contest"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetUserContestPresenter types.Presenter[getusercontest.Output, api.GetUserContestResponseObject]

type getUserContestPresenter struct{}

func NewGetUserContestPresenter() GetUserContestPresenter {
	return &getUserContestPresenter{}
}

func (p *getUserContestPresenter) Output(output *getusercontest.Output) api.GetUserContestResponseObject {
	projects := make([]api.CommonContestSpecContestProject, 0, len(output.Contest.Projects))
	for _, project := range output.Contest.Projects {
		projects = append(projects, api.CommonContestSpecContestProject{
			Id:     project.ID,
			Name:   project.Name,
			Status: api.CommonContestSpecProjectStatus(project.Status),
		})
	}

	return api.GetUserContest200JSONResponse{
		Success: true,
		Data: api.CommonContestSpecContestDetail{
			Id:       output.Contest.ID,
			Name:     output.Contest.Name,
			Projects: projects,
		},
	}
}

func (p *getUserContestPresenter) Error(err error) api.GetUserContestResponseObject {
	return httperror.From(err).Response()
}
