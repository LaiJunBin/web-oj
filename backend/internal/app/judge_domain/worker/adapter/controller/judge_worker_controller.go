package judgeworkercontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/judge_domain/worker/adapter/presenter"
	"judge-backend/internal/app/judge_domain/worker/application"
	addjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/add_judge_worker"
	deletejudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/delete_judge_worker"
	getjudgequeue "judge-backend/internal/app/judge_domain/worker/application/usecase/get_judge_queue"
	getjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/get_judge_worker"
	getsystemdefaults "judge-backend/internal/app/judge_domain/worker/application/usecase/get_system_defaults"
	listjudgeprioritystrategies "judge-backend/internal/app/judge_domain/worker/application/usecase/list_judge_priority_strategies"
	listjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/list_judge_worker"
	startjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/start_judge_worker"
	stopjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/stop_judge_worker"
	streamjudgeworkerlogs "judge-backend/internal/app/judge_domain/worker/application/usecase/stream_judge_worker_logs"

	"github.com/gin-gonic/gin"
)

type JudgeWorkerController struct {
	app       *application.JudgeWorkerApplication
	presenter *presenter.JudgeWorkerPresenter
}

func NewJudgeWorkerController(
	app *application.JudgeWorkerApplication,
	presenter *presenter.JudgeWorkerPresenter,
) *JudgeWorkerController {
	return &JudgeWorkerController{
		app:       app,
		presenter: presenter,
	}
}

func (ctl *JudgeWorkerController) ListJudgeWorker(ctx *gin.Context, request api.ListJudgeWorkerRequestObject) (api.ListJudgeWorkerResponseObject, error) {
	input := &listjudgeworker.Input{}

	output, err := ctl.app.ListJudgeWorker(input)
	if err != nil {
		return ctl.presenter.ListJudgeWorker.Error(err), nil
	}

	return ctl.presenter.ListJudgeWorker.Output(output), nil
}

func (ctl *JudgeWorkerController) ListJudgePriorityStrategies(ctx *gin.Context, request api.ListJudgePriorityStrategiesRequestObject) (api.ListJudgePriorityStrategiesResponseObject, error) {
	input := &listjudgeprioritystrategies.Input{}

	output, err := ctl.app.ListJudgePriorityStrategies(input)
	if err != nil {
		return ctl.presenter.ListJudgePriorityStrategies.Error(err), nil
	}

	return ctl.presenter.ListJudgePriorityStrategies.Output(output), nil
}

// GetSystemDefaults reports what a contest project's blank fields resolve to.
// Without it an administrator sees an empty box with no way to know the value
// behind it, which also makes "set it to the system default" unactionable.
func (ctl *JudgeWorkerController) GetSystemDefaults(ctx *gin.Context, request api.GetSystemDefaultsRequestObject) (api.GetSystemDefaultsResponseObject, error) {
	input := &getsystemdefaults.Input{}

	output, err := ctl.app.GetSystemDefaults(input)
	if err != nil {
		return ctl.presenter.GetSystemDefaults.Error(err), nil
	}

	return ctl.presenter.GetSystemDefaults.Output(output), nil
}

func (ctl *JudgeWorkerController) GetJudgeWorker(ctx *gin.Context, request api.GetJudgeWorkerRequestObject) (api.GetJudgeWorkerResponseObject, error) {
	input := &getjudgeworker.Input{
		JudgeWorkerID: request.Id,
	}

	output, err := ctl.app.GetJudgeWorker(input)
	if err != nil {
		return ctl.presenter.GetJudgeWorker.Error(err), nil
	}

	return ctl.presenter.GetJudgeWorker.Output(output), nil
}

func (ctl *JudgeWorkerController) AddJudgeWorker(ctx *gin.Context, request api.AddJudgeWorkerRequestObject) (api.AddJudgeWorkerResponseObject, error) {
	judgeorder := []string{}
	for _, j := range request.Body.JudgeOrder {
		judgeorder = append(judgeorder, string(j))
	}
	input := &addjudgeworker.Input{
		JudgeOrder: judgeorder,
	}

	output, err := ctl.app.AddJudgeWorker(input)
	if err != nil {
		return ctl.presenter.AddJudgeWorker.Error(err), nil
	}

	return ctl.presenter.AddJudgeWorker.Output(output), nil
}

func (ctl *JudgeWorkerController) DeleteJudgeWorker(ctx *gin.Context, request api.DeleteJudgeWorkerRequestObject) (api.DeleteJudgeWorkerResponseObject, error) {
	input := &deletejudgeworker.Input{
		JudgeWorkerID: request.Id,
	}

	output, err := ctl.app.DeleteJudgeWorker(input)
	if err != nil {
		return ctl.presenter.DeleteJudgeWorker.Error(err), nil
	}

	return ctl.presenter.DeleteJudgeWorker.Output(output), nil
}

func (ctl *JudgeWorkerController) StartJudgeWorker(ctx *gin.Context, request api.StartJudgeWorkerRequestObject) (api.StartJudgeWorkerResponseObject, error) {
	input := &startjudgeworker.Input{
		JudgeWorkerID: request.Id,
	}

	output, err := ctl.app.StartJudgeWorker(input)
	if err != nil {
		return ctl.presenter.StartJudgeWorker.Error(err), nil
	}

	return ctl.presenter.StartJudgeWorker.Output(output), nil
}

func (ctl *JudgeWorkerController) StopJudgeWorker(ctx *gin.Context, request api.StopJudgeWorkerRequestObject) (api.StopJudgeWorkerResponseObject, error) {
	input := &stopjudgeworker.Input{
		JudgeWorkerID: request.Id,
	}

	output, err := ctl.app.StopJudgeWorker(input)
	if err != nil {
		return ctl.presenter.StopJudgeWorker.Error(err), nil
	}

	return ctl.presenter.StopJudgeWorker.Output(output), nil
}

func (ctl *JudgeWorkerController) StreamJudgeWorkers(ctx *gin.Context, request api.StreamJudgeWorkersRequestObject) (api.StreamJudgeWorkersResponseObject, error) {
	return ctl.presenter.StreamJudgeWorkers.Output(ctx.Request.Context(), nil), nil
}

func (ctl *JudgeWorkerController) GetJudgeQueue(ctx *gin.Context, request api.GetJudgeQueueRequestObject) (api.GetJudgeQueueResponseObject, error) {
	output, err := ctl.app.GetJudgeQueue(&getjudgequeue.Input{})
	if err != nil {
		return ctl.presenter.GetJudgeQueue.Error(err), nil
	}

	return ctl.presenter.GetJudgeQueue.Output(output), nil
}

func (ctl *JudgeWorkerController) StreamJudgeQueue(ctx *gin.Context, request api.StreamJudgeQueueRequestObject) (api.StreamJudgeQueueResponseObject, error) {
	return ctl.presenter.StreamJudgeQueue.Output(ctx.Request.Context(), nil), nil
}

func (ctl *JudgeWorkerController) StreamJudgeWorkerLogs(ctx *gin.Context, request api.StreamJudgeWorkerLogsRequestObject) (api.StreamJudgeWorkerLogsResponseObject, error) {
	input := &streamjudgeworkerlogs.Input{
		JudgeWorkerID: request.Id,
	}

	output, err := ctl.app.StreamJudgeWorkerLogs(input)
	if err != nil {
		return ctl.presenter.StreamJudgeWorkerLogs.Error(err), nil
	}

	return ctl.presenter.StreamJudgeWorkerLogs.Output(ctx.Request.Context(), output), nil
}
