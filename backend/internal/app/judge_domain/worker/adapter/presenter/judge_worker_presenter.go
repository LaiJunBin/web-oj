package presenter

import (
	commonpresenter "judge-backend/internal/app/shared/adapter/presenter/common"
	sharedtypes "judge-backend/internal/app/shared/types"
	espresenter "judge-backend/internal/app/judge_domain/worker/adapter/presenter/event_stream"
	jsonpresenter "judge-backend/internal/app/judge_domain/worker/adapter/presenter/json"
)

type JudgeWorkerPresenter struct {
	ListJudgeWorker             jsonpresenter.ListJudgeWorkerPresenter
	ListJudgePriorityStrategies jsonpresenter.ListJudgePriorityStrategiesPresenter
	GetSystemDefaults           jsonpresenter.GetSystemDefaultsPresenter
	GetJudgeWorker              jsonpresenter.GetJudgeWorkerPresenter
	AddJudgeWorker              jsonpresenter.AddJudgeWorkerPresenter
	DeleteJudgeWorker           *commonpresenter.JSONPresenter
	StartJudgeWorker            *commonpresenter.JSONPresenter
	StopJudgeWorker             *commonpresenter.JSONPresenter
	StreamJudgeWorkers          espresenter.StreamJudgeWorkersPresenter
	StreamJudgeWorkerLogs       espresenter.StreamJudgeWorkerLogsPresenter
	GetJudgeQueue               jsonpresenter.GetJudgeQueuePresenter
	StreamJudgeQueue            espresenter.StreamJudgeQueuePresenter
}

func NewJudgeWorkerPresenter(
	eventBus sharedtypes.EventBus,
	streamJudgeWorkersPresenter espresenter.StreamJudgeWorkersPresenter,
	streamJudgeQueuePresenter espresenter.StreamJudgeQueuePresenter,
) *JudgeWorkerPresenter {
	return &JudgeWorkerPresenter{
		GetJudgeQueue:               jsonpresenter.NewGetJudgeQueuePresenter(),
		StreamJudgeQueue:            streamJudgeQueuePresenter,
		ListJudgeWorker:             jsonpresenter.NewListJudgeWorkerPresenter(),
		ListJudgePriorityStrategies: jsonpresenter.NewListJudgePriorityStrategiesPresenter(),
		GetSystemDefaults:           jsonpresenter.NewGetSystemDefaultsPresenter(),
		GetJudgeWorker:              jsonpresenter.NewGetJudgeWorkerPresenter(),
		AddJudgeWorker:              jsonpresenter.NewAddJudgeWorkerPresenter(),
		DeleteJudgeWorker:           commonpresenter.NewJSONPresenter(),
		StartJudgeWorker:            commonpresenter.NewJSONPresenter(),
		StopJudgeWorker:             commonpresenter.NewJSONPresenter(),
		StreamJudgeWorkers:          streamJudgeWorkersPresenter,
		StreamJudgeWorkerLogs:       espresenter.NewStreamJudgeWorkerLogsPresenter(eventBus),
	}
}
