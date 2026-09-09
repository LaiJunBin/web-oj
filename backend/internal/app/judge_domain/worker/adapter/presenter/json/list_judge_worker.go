package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/judge_domain/worker/adapter/presenter/types"
	listjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/list_judge_worker"
)

type ListJudgeWorkerPresenter types.Presenter[listjudgeworker.Output, api.ListJudgeWorkerResponseObject]

type listJudgeWorkerPresenter struct {
}

func NewListJudgeWorkerPresenter() ListJudgeWorkerPresenter {
	return &listJudgeWorkerPresenter{}
}

func (p *listJudgeWorkerPresenter) Output(output *listjudgeworker.Output) api.ListJudgeWorkerResponseObject {
	judgeWorkers := make([]api.AdminJudgeWorkerSpecJudgeWorker, 0)
	for _, j := range output.JudgeWorkers {
		judgeWorkers = append(judgeWorkers, api.AdminJudgeWorkerSpecJudgeWorker{
			Id:     j.ID,
			Status: api.AdminJudgeWorkerSpecJudgeWorkerStatus(j.Status),
			JudgeOrder: func() []api.AdminJudgeWorkerSpecJudgeOrder {
				judgeOrders := make([]api.AdminJudgeWorkerSpecJudgeOrder, 0)
				for _, o := range j.JudgeOrder {
					judgeOrders = append(judgeOrders, api.AdminJudgeWorkerSpecJudgeOrder(o))
				}
				return judgeOrders
			}(),
			CreatedAt:       j.CreatedAt,
			LastExecutionAt: j.LastExecutionAt,
		})
	}

	return api.ListJudgeWorker200JSONResponse{
		Success: true,
		Data:    judgeWorkers,
	}
}

func (p *listJudgeWorkerPresenter) Error(err error) api.ListJudgeWorkerResponseObject {
	return httperror.From(err).Response()
}
