package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	getcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/get_contest"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetContestPresenter types.Presenter[getcontest.Output, api.GetContestResponseObject]

type getContestPresenter struct {
}

func NewGetContestPresenter() GetContestPresenter {
	return &getContestPresenter{}
}

func (p *getContestPresenter) Output(output *getcontest.Output) api.GetContestResponseObject {
	projects := make([]api.AdminContestProjectSpecContestProject, 0)
	for _, project := range output.Contest.Projects {
		projects = append(projects, api.AdminContestProjectSpecContestProject{
			Id:                        project.ID,
			ProjectId:                 project.ProjectID,
			Name:                      project.Name,
			Description:               project.Description,
			IsPublic:                  project.IsPublic,
			Image:                     project.Image,
			TestImage:                 project.TestImage,
			ActivateTime:              project.ActivateTime,
			DeactivateTime:            project.DeactivateTime,
			StartTime:                 project.StartTime,
			EndTime:                   project.EndTime,
			JudgePriorityStrategy:     project.JudgePriorityStrategy,
			MaxJudgeDurationMinutes:   project.MaxJudgeDurationMinutes,
			SubmissionCooldownSeconds: project.SubmissionCooldownSeconds,
			PidsLimit:                 project.PidsLimit,
			MemoryLimitMB:             project.MemoryLimitMB,
			ScoreboardCompact:         project.ScoreboardCompact,
			ScoreboardMaskNames:       project.ScoreboardMaskNames,
		})
	}

	return api.GetContest200JSONResponse{
		Success: true,
		Data: api.AdminContestSpecContestDetail{
			Id:             output.Contest.ID,
			Name:           output.Contest.Name,
			Description:    output.Contest.Description,
			IsPublic:       output.Contest.IsPublic,
			Projects:       projects,
			ActivateTime:   output.Contest.ActivationTime,
			DeactivateTime: output.Contest.DeactivationTime,
			CreatedAt:      output.Contest.CreatedAt,
			UpdatedAt:      output.Contest.UpdatedAt,
		},
	}
}

func (p *getContestPresenter) Error(err error) api.GetContestResponseObject {
	return httperror.From(err).Response()
}
