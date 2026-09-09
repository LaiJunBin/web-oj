package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	getusercontestschedule "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_contest_schedule"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetUserContestSchedulePresenter types.Presenter[getusercontestschedule.Output, api.GetUserContestScheduleResponseObject]

type getUserContestSchedulesPresenter struct{}

func NewGetUserContestSchedulePresenter() GetUserContestSchedulePresenter {
	return &getUserContestSchedulesPresenter{}
}

func (p *getUserContestSchedulesPresenter) Output(output *getusercontestschedule.Output) api.GetUserContestScheduleResponseObject {
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

	return api.GetUserContestSchedule200JSONResponse{
		Success: true,
		Data: api.CommonContestSpecContestSchedule{
			Projects: projects,
		},
	}
}

func (p *getUserContestSchedulesPresenter) Error(err error) api.GetUserContestScheduleResponseObject {
	return httperror.From(err).Response()
}
