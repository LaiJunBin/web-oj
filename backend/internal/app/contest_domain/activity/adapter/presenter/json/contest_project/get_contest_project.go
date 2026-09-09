package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	getcontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_contest_project"
)

type GetContestProjectPresenter types.Presenter[getcontestproject.Output, api.GetContestProjectResponseObject]

type getContestProjectPresenter struct {
}

func NewGetContestProjectPresenter() GetContestProjectPresenter {
	return &getContestProjectPresenter{}
}

func (p *getContestProjectPresenter) Output(output *getcontestproject.Output) api.GetContestProjectResponseObject {
	return api.GetContestProject200JSONResponse{
		Success: true,
		Data: api.AdminContestProjectSpecContestProject{
			Id:                        output.Project.ID,
			ProjectId:                 output.Project.ProjectID,
			Name:                      output.Project.Name,
			Description:               output.Project.Description,
			IsPublic:                  output.Project.IsPublic,
			JudgePriorityStrategy:     output.Project.JudgePriorityStrategy,
			Image:                     output.Project.Image,
			TestImage:                 output.Project.TestImage,
			ActivateTime:              output.Project.ActivateTime,
			DeactivateTime:            output.Project.DeactivateTime,
			StartTime:                 output.Project.StartTime,
			EndTime:                   output.Project.EndTime,
			MaxJudgeDurationMinutes:   output.Project.MaxJudgeDurationMinutes,
			SubmissionCooldownSeconds: output.Project.SubmissionCooldownSeconds,
			PidsLimit:                 output.Project.PidsLimit,
			MemoryLimitMB:             output.Project.MemoryLimitMB,
			ScoreboardCompact:         output.Project.ScoreboardCompact,
			ScoreboardMaskNames:       output.Project.ScoreboardMaskNames,
		},
	}
}

func (p *getContestProjectPresenter) Error(err error) api.GetContestProjectResponseObject {
	return httperror.From(err).Response()
}
