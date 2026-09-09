package judgeworker

import (
	judgeworkercontroller "judge-backend/internal/app/judge_domain/worker/adapter/controller"
	"judge-backend/internal/app/judge_domain/worker/adapter/presenter"
	espresenter "judge-backend/internal/app/judge_domain/worker/adapter/presenter/event_stream"
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
	"judge-backend/internal/app/judge_domain/worker/infrastructure/acl"
	"judge-backend/internal/app/judge_domain/worker/infrastructure/query"
	"judge-backend/internal/app/judge_domain/worker/infrastructure/repository"
	"judge-backend/internal/app/judge_domain/worker/infrastructure/service"
	"judge-backend/internal/app/shared/types/provider"
)

func NewProvider() *provider.Provider {
	return provider.New(provider.Config{
		Infrastructures: []interface{}{
			// Anti-corruption layer: adapts the judge subdomain's published
			// contract into this subdomain's own WorkerPool port.
			acl.NewJudgeWorkerPool,
			// Same arrangement for the judge subdomain's queue-observability
			// contract.
			acl.NewJudgeQueueStats,
			query.NewJudgeWorkerQuery,
			query.NewJudgeQueueQuery,
			repository.NewJudgeWorkerRepository,
			service.NewJudgeWorkerService,
		},
		Controllers: []interface{}{
			judgeworkercontroller.NewJudgeWorkerController,
		},
		Presenters: []interface{}{
			espresenter.NewStreamJudgeWorkersPresenter,
			espresenter.NewStreamJudgeQueuePresenter,
			presenter.NewJudgeWorkerPresenter,
		},
		Usecases: []interface{}{
			listjudgeworker.NewUsecase,
			listjudgeprioritystrategies.NewUsecase,
			getsystemdefaults.NewUsecase,
			getjudgeworker.NewUsecase,
			addjudgeworker.NewUsecase,
			deletejudgeworker.NewUsecase,
			startjudgeworker.NewUsecase,
			stopjudgeworker.NewUsecase,
			streamjudgeworkerlogs.NewUsecase,
			getjudgequeue.NewUsecase,
		},
		Applications: []interface{}{
			application.NewJudgeWorkerApplication,
		},
	})
}
