package application

import (
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
)

type JudgeWorkerApplication struct {
	listJudgeWorker             listjudgeworker.Usecase
	listJudgePriorityStrategies listjudgeprioritystrategies.Usecase
	getJudgeWorker              getjudgeworker.Usecase
	addJudgeWorker              addjudgeworker.Usecase
	deleteJudgeWorker           deletejudgeworker.Usecase
	startJudgeWorker            startjudgeworker.Usecase
	stopJudgeWorker             stopjudgeworker.Usecase
	streamJudgeWorkerLogs       streamjudgeworkerlogs.Usecase
	getJudgeQueue               getjudgequeue.Usecase
	getSystemDefaults           getsystemdefaults.Usecase
}

func NewJudgeWorkerApplication(
	listJudgeWorker listjudgeworker.Usecase,
	listJudgePriorityStrategies listjudgeprioritystrategies.Usecase,
	getJudgeWorker getjudgeworker.Usecase,
	addJudgeWorker addjudgeworker.Usecase,
	deleteJudgeWorker deletejudgeworker.Usecase,
	startJudgeWorker startjudgeworker.Usecase,
	stopJudgeWorker stopjudgeworker.Usecase,
	streamJudgeWorkerLogs streamjudgeworkerlogs.Usecase,
	getJudgeQueue getjudgequeue.Usecase,
	getSystemDefaults getsystemdefaults.Usecase,
) *JudgeWorkerApplication {
	return &JudgeWorkerApplication{
		listJudgeWorker:             listJudgeWorker,
		listJudgePriorityStrategies: listJudgePriorityStrategies,
		getJudgeWorker:              getJudgeWorker,
		addJudgeWorker:              addJudgeWorker,
		deleteJudgeWorker:           deleteJudgeWorker,
		startJudgeWorker:            startJudgeWorker,
		stopJudgeWorker:             stopJudgeWorker,
		streamJudgeWorkerLogs:       streamJudgeWorkerLogs,
		getJudgeQueue:               getJudgeQueue,
		getSystemDefaults:           getSystemDefaults,
	}
}

func (a *JudgeWorkerApplication) ListJudgeWorker(input *listjudgeworker.Input) (*listjudgeworker.Output, error) {
	return a.listJudgeWorker.Execute(input)
}

func (a *JudgeWorkerApplication) ListJudgePriorityStrategies(input *listjudgeprioritystrategies.Input) (*listjudgeprioritystrategies.Output, error) {
	return a.listJudgePriorityStrategies.Execute(input)
}

func (a *JudgeWorkerApplication) GetJudgeWorker(input *getjudgeworker.Input) (*getjudgeworker.Output, error) {
	return a.getJudgeWorker.Execute(input)
}

func (a *JudgeWorkerApplication) AddJudgeWorker(input *addjudgeworker.Input) (*addjudgeworker.Output, error) {
	return a.addJudgeWorker.Execute(input)
}

func (a *JudgeWorkerApplication) DeleteJudgeWorker(input *deletejudgeworker.Input) (*deletejudgeworker.Output, error) {
	return a.deleteJudgeWorker.Execute(input)
}

func (a *JudgeWorkerApplication) StartJudgeWorker(input *startjudgeworker.Input) (*startjudgeworker.Output, error) {
	return a.startJudgeWorker.Execute(input)
}

func (a *JudgeWorkerApplication) StopJudgeWorker(input *stopjudgeworker.Input) (*stopjudgeworker.Output, error) {
	return a.stopJudgeWorker.Execute(input)
}

func (a *JudgeWorkerApplication) StreamJudgeWorkerLogs(input *streamjudgeworkerlogs.Input) (*streamjudgeworkerlogs.Output, error) {
	return a.streamJudgeWorkerLogs.Execute(input)
}

func (a *JudgeWorkerApplication) GetJudgeQueue(input *getjudgequeue.Input) (*getjudgequeue.Output, error) {
	return a.getJudgeQueue.Execute(input)
}

func (a *JudgeWorkerApplication) GetSystemDefaults(input *getsystemdefaults.Input) (*getsystemdefaults.Output, error) {
	return a.getSystemDefaults.Execute(input)
}
