package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/judge_domain/worker/adapter/presenter/types"
	listjudgeprioritystrategies "judge-backend/internal/app/judge_domain/worker/application/usecase/list_judge_priority_strategies"
)

type ListJudgePriorityStrategiesPresenter types.Presenter[listjudgeprioritystrategies.Output, api.ListJudgePriorityStrategiesResponseObject]

type listJudgePriorityStrategiesPresenter struct {
}

func NewListJudgePriorityStrategiesPresenter() ListJudgePriorityStrategiesPresenter {
	return &listJudgePriorityStrategiesPresenter{}
}

func (p *listJudgePriorityStrategiesPresenter) Output(output *listjudgeprioritystrategies.Output) api.ListJudgePriorityStrategiesResponseObject {
	return api.ListJudgePriorityStrategies200JSONResponse{
		Success: true,
		Data:    output.JudgePriorityStrategies,
	}
}

func (p *listJudgePriorityStrategiesPresenter) Error(err error) api.ListJudgePriorityStrategiesResponseObject {
	return httperror.From(err).Response()
}
