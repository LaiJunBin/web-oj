package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/judge_domain/worker/adapter/presenter/types"
	addjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/add_judge_worker"
)

type AddJudgeWorkerPresenter types.Presenter[addjudgeworker.Output, api.AddJudgeWorkerResponseObject]

type addJudgeWorkerPresenter struct {
}

func NewAddJudgeWorkerPresenter() AddJudgeWorkerPresenter {
	return &addJudgeWorkerPresenter{}
}

func (p *addJudgeWorkerPresenter) Output(output *addjudgeworker.Output) api.AddJudgeWorkerResponseObject {
	return api.AddJudgeWorker200JSONResponse{
		Success: true,
		Data:    output.ID,
	}
}

func (p *addJudgeWorkerPresenter) Error(err error) api.AddJudgeWorkerResponseObject {
	return httperror.From(err).Response()
}
