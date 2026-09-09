package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	getrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/get_rejudge"
)

type GetRejudgePresenter types.Presenter[getrejudge.Output, api.GetRejudgeResponseObject]

type getRejudgePresenter struct {
}

func NewGetRejudgePresenter() GetRejudgePresenter {
	return &getRejudgePresenter{}
}

func transformLifecycle(lifecycles []*dto.Lifecycle) []api.AdminJudgmentSpecLifecycle {
	result := make([]api.AdminJudgmentSpecLifecycle, 0)
	for _, lifecycle := range lifecycles {
		result = append(result, api.AdminJudgmentSpecLifecycle{
			Id:                   lifecycle.ID,
			Command:              lifecycle.Command,
			ExecutionEnvironment: lifecycle.ExecutionEnvironment,
			Result: api.AdminJudgmentSpecResult{
				ExitCode: lifecycle.Result.ExitCode,
				ExecTime: lifecycle.Result.ExecTime,
				Stdout:   lifecycle.Result.Stdout,
				Stderr:   lifecycle.Result.Stderr,
			},
		})
	}
	return result
}

func (p *getRejudgePresenter) Output(output *getrejudge.Output) api.GetRejudgeResponseObject {
	submissions := []api.AdminJudgmentSpecSubmissionComparison{}
	for _, submission := range output.Rejudge.Submissions {
		// Left nil when the submission had no adopted judgment before the rejudge,
		// so the response carries a null rather than a zero-scored object the UI
		// would render as "0 分".
		var before *api.AdminJudgmentSpecJudgmentOutcomeChange
		if submission.Before != nil {
			before = &api.AdminJudgmentSpecJudgmentOutcomeChange{
				Score:      submission.Before.Score,
				TotalScore: submission.Before.TotalScore,
				Status:     submission.Before.Status,
			}
		}

		submissions = append(submissions, api.AdminJudgmentSpecSubmissionComparison{
			Id:     submission.ID,
			Before: before,
			After: api.AdminJudgmentSpecJudgmentOutcomeChange{
				Score:      submission.After.Score,
				TotalScore: submission.After.TotalScore,
				Status:     submission.After.Status,
			},
		})
	}

	return api.GetRejudge200JSONResponse{
		Success: true,
		Data: api.AdminJudgmentSpecRejudgeDetail{
			Id:          output.Rejudge.ID,
			Reason:      output.Rejudge.Reason,
			Status:      api.AdminJudgmentSpecRejudgeStatus(output.Rejudge.Status),
			CreatedAt:   output.Rejudge.CreatedAt,
			Submissions: submissions,
		},
	}
}

func (p *getRejudgePresenter) Error(err error) api.GetRejudgeResponseObject {
	return httperror.From(err).Response()
}
