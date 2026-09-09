package judge

import (
	"time"

	adapterhandler "judge-backend/internal/app/judge_domain/judge/adapter/handler"
	executejudgment "judge-backend/internal/app/judge_domain/judge/application/usecase/execute_judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/factory"
	domainrepository "judge-backend/internal/app/judge_domain/judge/domain/repository"
	domainservice "judge-backend/internal/app/judge_domain/judge/domain/service"
	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	infrahandler "judge-backend/internal/app/judge_domain/judge/infrastructure/handler"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/provider"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/query"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/queue"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/repository"
	dockerruntime "judge-backend/internal/app/judge_domain/judge/infrastructure/runtime/docker"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/service"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/worker"
	sharedtypes "judge-backend/internal/app/shared/types"
	coreprovider "judge-backend/internal/app/shared/types/provider"
	"judge-backend/internal/config"
)

func NewProvider() *coreprovider.Provider {
	p := coreprovider.New(coreprovider.Config{
		Infrastructures: []interface{}{
			query.NewPriorityStrategyQuery,
			query.NewMaxJudgeDurationQuery,
			queue.NewJudgeQueueMap,
			// Publishes the live queues' load through the judge domain's own
			// contract (domain/provider.QueueStats), same arrangement as the
			// worker pool below.
			queue.NewQueueStatsProvider,
			query.NewJudgmentQuery,

			// The only place Docker is wired into the judge domain.
			dockerruntime.NewRuntime,

			worker.NewJudgeWorkerPool,
			// Publishes the pool through the judge domain's own contract
			// (domain/provider.WorkerPool); downstream contexts adapt it in
			// their own ACLs.
			worker.NewWorkerPoolProvider,

			repository.NewProjectRepository,
			repository.NewJudgmentRepository,
			repository.NewGradingItemRepository,
			repository.NewLifecycleRepository,
			repository.NewSubmissionRepository,
			repository.NewRejudgeRepository,
			repository.NewUserRepository,
		},
		Usecases: []interface{}{
			// The composition root is the one place that knows the config
			// source; the usecase itself only sees plain durations.
			func(
				provisioner domainservice.ExecutionEnvironmentProvisioner,
				submissionService domainservice.SubmissionService,
				submissionRepo domainrepository.SubmissionRepository,
				userRepo domainrepository.UserRepository,
				judgmentRepo domainrepository.JudgmentRepository,
				eventBus sharedtypes.EventBus,
				cfg *config.Config,
				logger sharedtypes.Logger,
			) executejudgment.Usecase {
				return executejudgment.NewUsecase(
					provisioner,
					submissionService,
					submissionRepo,
					userRepo,
					judgmentRepo,
					eventBus,
					time.Duration(cfg.Judge.TestcaseTimeoutSeconds)*time.Second,
					time.Duration(cfg.Judge.LifecycleTimeoutSeconds)*time.Second,
					cfg.Judge.ProvisionMaxAttempts,
					executejudgment.DefaultProvisionRetryBackoff,
					logger,
				)
			},
		},
		Handlers: []interface{}{
			// Adapter handlers: shared events → judge-internal events
			adapterhandler.RegisterProjectSubmittedHandler,
			adapterhandler.RegisterRejudgeRequestedHandler,
			adapterhandler.RegisterCancelJudgeRequestedHandler,
			adapterhandler.RegisterStaffNoteUpdateRequestedHandler,
			adapterhandler.RegisterRejudgeApprovalRequestedHandler,
			adapterhandler.RegisterRejudgeRejectionRequestedHandler,
			adapterhandler.RegisterCancelRejudgeRequestedHandler,
			// Adapter handlers: judge-internal events → shared events
			adapterhandler.RegisterJudgedHandler,
			adapterhandler.RegisterJudgeUpdatedHandler,
			// Infra handlers: judge-internal events
			infrahandler.RegisterProjectSubmittedHandler,
			infrahandler.RegisterSubmissionCreatedHandler,
			infrahandler.RegisterJudgmentCreatedHandler,
			infrahandler.RegisterRejudgeRequestedHandler,
			infrahandler.RegisterCancelJudgeRequestedHandler,
			infrahandler.RegisterStaffNoteUpdateRequestedHandler,
			infrahandler.RegisterJudgeCancelledHandler,
			infrahandler.RegisterRejudgeApprovalRequestedHandler,
			infrahandler.RegisterRejudgeRejectionRequestedHandler,
			infrahandler.RegisterJudgeUpdatedHandler,
			infrahandler.RegisterCancelRejudgeRequestedHandler,
		},
		Invokes: []interface{}{
			// Must run after this domain's Handlers (see
			// LoadQueuedJudgmentsOnStartup's doc comment for why) — Invokes
			// are wired after Handlers by app/provider/provider.go, so this
			// placement is safe regardless of position within this list.
			queue.LoadQueuedJudgmentsOnStartup,
			service.InitOrphanSubmissionRepairService,
			service.InitOrphanJudgmentRepairService,
			service.InitRejudgeCommitRepairService,
			service.InitOrphanContainerCleanupService, // must run before InitJudgeWorkerPool
			worker.InitJudgeWorkerPool,
		},
		Factories: []interface{}{
			factory.NewJudgmentFactory,
		},
		Services: []interface{}{
			service.NewSubmissionService,
			service.NewOrphanSubmissionRepairService,
			service.NewOrphanJudgmentRepairService,
			service.NewOrphanContainerCleanupService,
			service.NewRejudgeCommitRepairService,
		},
		Providers: []interface{}{
			prioritystrategies.NewProvider,
			provider.NewStoragePathProvider,
		},
	})
	return p
}
