package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	getpubliccontestschedule "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/get_public_contest_schedule"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetPublicContestSchedulePresenter types.Presenter[getpubliccontestschedule.Output, api.GetPublicContestScheduleResponseObject]

type getPublicContestSchedulesPresenter struct{}

func NewGetPublicContestSchedulePresenter() GetPublicContestSchedulePresenter {
	return &getPublicContestSchedulesPresenter{}
}

func (p *getPublicContestSchedulesPresenter) Output(output *getpubliccontestschedule.Output) api.GetPublicContestScheduleResponseObject {
	projects := make([]api.CommonContestSpecContestProjectSchedule, 0, len(output.ContestSchedule.Projects))
	for _, project := range output.ContestSchedule.Projects {
		projects = append(projects, api.CommonContestSpecContestProjectSchedule{
			Name:        project.Name,
			Description: project.Description,
			Score:       project.Score,
			StartTime:   project.StartTime,
			EndTime:     project.EndTime,
		})
	}

	return api.GetPublicContestSchedule200JSONResponse{
		Success: true,
		Data: api.CommonContestSpecContestSchedule{
			Projects: projects,
		},
	}
}

func (p *getPublicContestSchedulesPresenter) Error(err error) api.GetPublicContestScheduleResponseObject {
	return httperror.From(err).Response()
}
