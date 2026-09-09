package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/judge_domain/worker/adapter/presenter/types"
	getjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/get_judge_worker"
)

type GetJudgeWorkerPresenter types.Presenter[getjudgeworker.Output, api.GetJudgeWorkerResponseObject]

type getJudgeWorkerPresenter struct {
}

func NewGetJudgeWorkerPresenter() GetJudgeWorkerPresenter {
	return &getJudgeWorkerPresenter{}
}

func (p *getJudgeWorkerPresenter) Output(output *getjudgeworker.Output) api.GetJudgeWorkerResponseObject {
	return api.GetJudgeWorker200JSONResponse{
		Success: true,
		Data: api.AdminJudgeWorkerSpecJudgeWorker{
			Id:     output.JudgeWorker.ID,
			Status: api.AdminJudgeWorkerSpecJudgeWorkerStatus(output.JudgeWorker.Status),
			JudgeOrder: func() []api.AdminJudgeWorkerSpecJudgeOrder {
				judgeOrders := make([]api.AdminJudgeWorkerSpecJudgeOrder, 0)
				for _, o := range output.JudgeWorker.JudgeOrder {
					judgeOrders = append(judgeOrders, api.AdminJudgeWorkerSpecJudgeOrder(o))
				}
				return judgeOrders
			}(),
			CreatedAt:       output.JudgeWorker.CreatedAt,
			LastExecutionAt: output.JudgeWorker.LastExecutionAt,
		},
	}
}

func (p *getJudgeWorkerPresenter) Error(err error) api.GetJudgeWorkerResponseObject {
	return httperror.From(err).Response()
}
