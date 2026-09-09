package service

import (
	domainservice "judge-backend/internal/app/judge_domain/judge/domain/service"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/worker"
	sharedtypes "judge-backend/internal/app/shared/types"
)

// OrphanContainerCleanupService removes runtime resources (containers,
// networks) left behind by a previous process. The runtime-specific work
// lives behind the ExecutionEnvironmentProvisioner port; this service only
// collects which judgments/workers are still active.
type OrphanContainerCleanupService struct {
	provisioner domainservice.ExecutionEnvironmentProvisioner
	workerPool  *worker.JudgeWorkerPool
	logger      sharedtypes.Logger
}

func NewOrphanContainerCleanupService(
	provisioner domainservice.ExecutionEnvironmentProvisioner,
	workerPool *worker.JudgeWorkerPool,
	logger sharedtypes.Logger,
) *OrphanContainerCleanupService {
	return &OrphanContainerCleanupService{
		provisioner: provisioner,
		workerPool:  workerPool,
		logger:      logger,
	}
}

// InitOrphanContainerCleanupService is called as an fx Invoke.
// It runs synchronously so that orphan runtime resources are freed before
// InitJudgeWorkerPool creates workers.
//
// NOTE: fx lifecycle OnStart/OnStop hooks are never reached in this codebase
// because SetupHandler (also an Invoke) blocks in gin's r.Run(), which causes
// fx.Start() to hang before it can fire any OnStart hooks. We therefore follow
// the same pattern as OrphanSubmissionRepairService and OrphanJudgmentRepairService:
// run work directly from the Invoke function.
func InitOrphanContainerCleanupService(s *OrphanContainerCleanupService) {
	s.CleanupOnce()
}

func (s *OrphanContainerCleanupService) CleanupOnce() {
	// Collect active judgment IDs and active worker IDs from the pool.
	// On a fresh startup both sets are empty, but checking defensively ensures
	// we never remove resources belonging to an in-progress judgment or a live worker.
	activeJudgmentIDs := map[string]struct{}{}
	activeWorkerIDs := map[string]struct{}{}
	for _, w := range s.workerPool.ListJudgeWorkers() {
		activeWorkerIDs[w.ID()] = struct{}{}
		if id := w.CurrentJudgeID(); id != "" {
			activeJudgmentIDs[id] = struct{}{}
		}
	}

	s.logger.Infof("[OrphanContainerCleanupService] Starting cleanup (active judgments: %d, active workers: %d)", len(activeJudgmentIDs), len(activeWorkerIDs))
	s.provisioner.CleanupOrphans(activeJudgmentIDs, activeWorkerIDs)
}
