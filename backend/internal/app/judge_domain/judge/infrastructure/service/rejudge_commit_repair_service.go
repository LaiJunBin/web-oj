package service

import (
	"sync"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
)

// RejudgeCommitRepairService closes CommitJudgmentsByRejudgeID's crash window:
// an approval whose judgments flip was lost (rejudge APPROVED, judgments
// still Rejudged) is replayed so the approved results become visible. See
// RejudgeRepository.RepairUncommittedApprovals.
//
// It used to run only at startup — but nobody restarts the server
// mid-contest, so an approval broken during the event stayed broken for the
// audience that matters most. It now also sweeps on a ticker
// (judge.rejudgeCommitRepairInterval, default 300s, explicit <= 0 disables),
// same shape as OrphanJudgmentRepairService.
type RejudgeCommitRepairService struct {
	rejudgeRepo repository.RejudgeRepository
	config      *config.Config
	logger      sharedtypes.Logger
	// mu serializes RepairOnce: a periodic tick must not interleave with a
	// slow in-flight sweep.
	mu sync.Mutex
}

func NewRejudgeCommitRepairService(
	rejudgeRepo repository.RejudgeRepository,
	config *config.Config,
	logger sharedtypes.Logger,
) *RejudgeCommitRepairService {
	return &RejudgeCommitRepairService{
		rejudgeRepo: rejudgeRepo,
		config:      config,
		logger:      logger,
	}
}

func InitRejudgeCommitRepairService(s *RejudgeCommitRepairService) {
	go s.start()
}

func (s *RejudgeCommitRepairService) start() {
	s.RepairOnce()

	interval, enabled := s.repairInterval()
	if !enabled {
		s.logger.Infof("[RejudgeCommitRepairService] Periodic repair disabled")
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
func (s *RejudgeCommitRepairService) repairInterval() (time.Duration, bool) {
	if s.config == nil || s.config.Judge == nil || s.config.Judge.RejudgeCommitRepairInterval == nil {
		return 5 * time.Minute, true
	}
	configured := *s.config.Judge.RejudgeCommitRepairInterval
	if configured <= 0 {
		return 0, false
	}
	return time.Duration(configured) * time.Second, true
}

func (s *RejudgeCommitRepairService) RepairOnce() {
	s.mu.Lock()
	defer s.mu.Unlock()

	repaired, err := s.rejudgeRepo.RepairUncommittedApprovals()
	if err != nil {
		s.logger.Errorf("[RejudgeCommitRepairService] Failed to repair uncommitted approvals: %v", err)
		return
	}
	if repaired == 0 {
		s.logger.Infof("[RejudgeCommitRepairService] No uncommitted approvals found")
		return
	}
	s.logger.Infof("[RejudgeCommitRepairService] Repaired %d judgments left behind by interrupted approvals", repaired)
}
