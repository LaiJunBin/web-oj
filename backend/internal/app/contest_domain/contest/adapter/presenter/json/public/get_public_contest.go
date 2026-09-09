package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	getpubliccontest "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/get_public_contest"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetPublicContestPresenter types.Presenter[getpubliccontest.Output, api.GetPublicContestResponseObject]

type getPublicContestsPresenter struct{}

func NewGetPublicContestPresenter() GetPublicContestPresenter {
	return &getPublicContestsPresenter{}
}

func (p *getPublicContestsPresenter) Output(output *getpubliccontest.Output) api.GetPublicContestResponseObject {
	projects := make([]api.CommonContestSpecContestProject, 0, len(output.Contest.Projects))
	for _, project := range output.Contest.Projects {
		projects = append(projects, api.CommonContestSpecContestProject{
			Id:     project.ID,
			Name:   project.Name,
			Status: api.CommonContestSpecProjectStatus(project.Status),
		})
	}

	return api.GetPublicContest200JSONResponse{
		Success: true,
		Data: api.CommonContestSpecContestDetail{
			Id:       output.Contest.ID,
			Name:     output.Contest.Name,
			Projects: projects,
		},
	}
}

func (p *getPublicContestsPresenter) Error(err error) api.GetPublicContestResponseObject {
	return httperror.From(err).Response()
}
