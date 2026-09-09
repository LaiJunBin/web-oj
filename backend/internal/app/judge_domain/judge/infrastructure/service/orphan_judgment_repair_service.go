package service

import (
	"sync"
	"time"

	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/worker"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
)

// The reason is contestant-facing (the user API serialises it), so it says
// only what happened from their point of view and who to talk to; the
// operator's detail stays in the error message, which the admin API alone
// exposes.
const (
	orphanJudgmentRepairReason       = "評測流程中斷，系統已標記為錯誤，請聯絡現場工作人員"
	orphanJudgmentRepairErrorMessage = "server restarted or worker crashed while judgment was in progress"
)

// OrphanJudgmentRepairService fails judgments stuck in Judging with no live
// worker. It is safe on a SINGLE API instance only: activeWorkerJudgmentIDs
// sees this process's worker pool, so a second instance's in-flight judgments
// would look orphaned from here and be wrongly failed. The deferred fix is
// ownership fencing: have TryStart write a per-attempt execution id and
// require it in the filter of every later write, so "is this judgment still
// mine?" has a real answer instead of being inferred from status.
type OrphanJudgmentRepairService struct {
	judgmentRepo repository.JudgmentRepository
	workerPool   *worker.JudgeWorkerPool
	eventbus     sharedtypes.EventBus
	config       *config.Config
	logger       sharedtypes.Logger
	// mu serializes RepairOnce: the periodic tick and a slow in-flight sweep
	// must not interleave over the same Judging set.
	mu sync.Mutex
}

func NewOrphanJudgmentRepairService(
	judgmentRepo repository.JudgmentRepository,
	workerPool *worker.JudgeWorkerPool,
	eventbus sharedtypes.EventBus,
	config *config.Config,
	logger sharedtypes.Logger,
) *OrphanJudgmentRepairService {
	return &OrphanJudgmentRepairService{
		judgmentRepo: judgmentRepo,
		workerPool:   workerPool,
		eventbus:     eventbus,
		config:       config,
		logger:       logger,
	}
}

// InitOrphanJudgmentRepairService starts the sweep goroutine. Unlike
// OrphanSubmissionRepairService (which starts from its constructor), the
// goroutine deliberately starts HERE: the provider Invoke order runs this
// before InitJudgeWorkerPool, so the first sweep sees an empty pool and fails
// every leftover Judging document — that IS the startup-recovery semantic —
// and the struct-literal construction used by tests never starts a ticker.
// fx lifecycle hooks are not an option; they never fire in this app (see
// orphan_container_cleanup_service.go).
func InitOrphanJudgmentRepairService(s *OrphanJudgmentRepairService) {
	go s.start()
}

func (s *OrphanJudgmentRepairService) start() {
	s.RepairOnce()

	interval, enabled := s.repairInterval()
	if !enabled {
		s.logger.Infof("[OrphanJudgmentRepairService] Periodic repair disabled")
		return
	}

	// No graceful shutdown: background goroutines in this app live until the
	// process exits (fx hooks never fire — see orphan_container_cleanup_service.go).
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		s.RepairOnce()
	}
}

// repairInterval resolves the three-state config knob: nil-safe default of
// 5 minutes, explicit <= 0 disables, positive is seconds.
func (s *OrphanJudgmentRepairService) repairInterval() (time.Duration, bool) {
	if s.config == nil || s.config.Judge == nil || s.config.Judge.OrphanJudgmentRepairInterval == nil {
		return 5 * time.Minute, true
	}
	configured := *s.config.Judge.OrphanJudgmentRepairInterval
	if configured <= 0 {
		return 0, false
	}
	return time.Duration(configured) * time.Second, true
}

func (s *OrphanJudgmentRepairService) RepairOnce() {
	s.mu.Lock()
	defer s.mu.Unlock()

	judgings, err := s.judgmentRepo.FindJudgings()
	if err != nil {
		s.logger.Errorf("[OrphanJudgmentRepairService] Failed to find judging judgments: %v", err)
		return
	}
	if len(judgings) == 0 {
		s.logger.Infof("[OrphanJudgmentRepairService] No judging judgments found")
		return
	}

	activeIDs := s.activeWorkerJudgmentIDs()
	s.logger.Infof("[OrphanJudgmentRepairService] Found %d judging judgments, %d actively processed by workers", len(judgings), len(activeIDs))

	for _, j := range judgings {
		if _, active := activeIDs[j.ID()]; active {
			s.logger.Debugf("[OrphanJudgmentRepairService] Judgment %s is actively processed, skipping", j.ID())
			continue
		}

		if err := j.Fail(); err != nil {
			s.logger.Errorf("[OrphanJudgmentRepairService] Cannot mark judgment %s as failed: %v", j.ID(), err)
			continue
		}
		j.SetReason(orphanJudgmentRepairReason)
		j.SetErrorMessage(orphanJudgmentRepairErrorMessage)

		// SaveTerminalIfJudging, not Save: this runs on its own goroutine, and
		// j is a snapshot from FindJudgings — an API cancel (TryCancel) or,
		// in a multi-instance deployment, another instance's worker can move
		// this document out of Judging between that read and this write. An
		// unconditional Save would overwrite their accepted outcome with a
		// stale ERROR.
		applied, err := s.judgmentRepo.SaveTerminalIfJudging(j)
		if err != nil {
			s.logger.Errorf("[OrphanJudgmentRepairService] Failed to save repaired judgment %s: %v", j.ID(), err)
			continue
		}
		if !applied {
			s.logger.Infof("[OrphanJudgmentRepairService] Judgment %s was no longer Judging (another writer already decided it), skipping repair", j.ID())
			continue
		}

		// Only after a confirmed applied=true: j's payload is exactly what
		// the conditional save just persisted, so publishing it needs no
		// re-fetch. Without this event the rejudge rollup never learns the
		// member reached ERROR (the rejudge would hang in RUNNING forever)
		// and the contest stream never shows the repair outcome.
		s.eventbus.Publish(judgeupdated.NewJudgeUpdatedEvent(j))

		s.logger.Infof("[OrphanJudgmentRepairService] Repaired orphan judging judgment %s", j.ID())
	}
}

func (s *OrphanJudgmentRepairService) activeWorkerJudgmentIDs() map[string]struct{} {
	ids := make(map[string]struct{})
	for _, w := range s.workerPool.ListJudgeWorkers() {
		if id := w.CurrentJudgeID(); id != "" {
			ids[id] = struct{}{}
		}
	}
	return ids
}
