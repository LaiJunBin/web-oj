package service

import (
	"errors"
	"sync"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/factory"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
)

// The reason is contestant-facing (the user API serialises it) and shares
// its wording with the judgment-repair sweep — from the contestant's side
// the two are the same event; the operator's detail stays in the error
// message, which only the admin API exposes.
const (
	orphanRepairReason       = "評測流程中斷，系統已標記為錯誤，請聯絡現場工作人員"
	orphanRepairErrorMessage = "submission pipeline interrupted before judgment creation"
)

type OrphanSubmissionRepairService struct {
	submissionRepo  repository.SubmissionRepository
	judgmentRepo    repository.JudgmentRepository
	judgmentFactory *factory.JudgmentFactory
	config          *config.Config
	logger          sharedtypes.Logger
	mu              sync.Mutex
}

func NewOrphanSubmissionRepairService(
	submissionRepo repository.SubmissionRepository,
	judgmentRepo repository.JudgmentRepository,
	judgmentFactory *factory.JudgmentFactory,
	config *config.Config,
	logger sharedtypes.Logger,
) *OrphanSubmissionRepairService {
	service := &OrphanSubmissionRepairService{
		submissionRepo:  submissionRepo,
		judgmentRepo:    judgmentRepo,
		judgmentFactory: judgmentFactory,
		config:          config,
		logger:          logger,
	}

	go service.start()
	return service
}

func InitOrphanSubmissionRepairService(*OrphanSubmissionRepairService) {}

func (s *OrphanSubmissionRepairService) start() {
	s.RepairOnce()

	// Pointer three-state (applyDefaults fills nil with 300): a configured
	// value <= 0 disables the periodic sweep — previously that branch was
	// unreachable because a non-positive value silently fell back to the
	// default, so "set 0 to disable" did not actually exist.
	interval := 5 * time.Minute
	if s.config != nil && s.config.Judge != nil && s.config.Judge.OrphanSubmissionRepairInterval != nil {
		configured := *s.config.Judge.OrphanSubmissionRepairInterval
		if configured <= 0 {
			s.logger.Infof("[OrphanSubmissionRepairService] Periodic repair disabled")
			return
		}
		interval = time.Duration(configured) * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		s.RepairOnce()
	}
}

// graceCutoff decides how long a submission may sit judgment-less before the
// sweep treats it as orphaned. Judgment creation is asynchronous, so a fresh
// submission is ALWAYS judgment-less for a moment; sweeping inside that window
// raced a terminal ERROR judgment against the real one (the two writes are
// independent). Twice the sweep interval is comfortably beyond anything a
// healthy creation path can take — its Mongo calls run under 10s context
// timeouts — while keeping crash-orphan repair within a few sweep cycles.
// A non-positive configured interval keeps today's semantics for the one
// startup sweep that still runs: no grace, exactly the previous behaviour
// (the BDD environment relies on this).
func (s *OrphanSubmissionRepairService) graceCutoff(now time.Time) time.Time {
	interval := 300
	if s.config != nil && s.config.Judge != nil && s.config.Judge.OrphanSubmissionRepairInterval != nil {
		interval = *s.config.Judge.OrphanSubmissionRepairInterval
	}
	if interval <= 0 {
		return now
	}
	return now.Add(-2 * time.Duration(interval) * time.Second)
}

func (s *OrphanSubmissionRepairService) RepairOnce() {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.submissionRepo.FindWithoutJudgment(s.graceCutoff(time.Now()))
	if err != nil {
		s.logger.Errorf("[OrphanSubmissionRepairService] Failed to find orphan submissions: %v", err)
		return
	}
	if len(rows) == 0 {
		s.logger.Infof("[OrphanSubmissionRepairService] No orphan submissions found")
		return
	}

	s.logger.Infof("[OrphanSubmissionRepairService] Found %d orphan submissions", len(rows))
	for _, sub := range rows {
		judgeType := judgetype.JUDGE
		if sub.IsPartial() {
			judgeType = judgetype.JUDGE_PARTIAL
		}

		judgment, err := s.judgmentFactory.CreateJudgment(sub.ID(), judgeType)
		if err != nil {
			// Deterministic failure: retrying every sweep can never succeed,
			// so record the terminal ERROR judgment (skipping the failing
			// grading-item load) and let the submission leave the orphan set.
			if errors.Is(err, repository.ErrGradingItemsNotOwned) {
				failed, ferr := s.judgmentFactory.CreateFailedJudgment(sub.ID(), judgeType, factory.GradingItemMismatchReason)
				if ferr != nil {
					s.logger.Errorf("[OrphanSubmissionRepairService] Failed to create error judgment for submission %s: %v", sub.ID(), ferr)
					continue
				}
				if err := s.judgmentRepo.Save(failed); err != nil {
					s.logger.Errorf("[OrphanSubmissionRepairService] Failed to save error judgment for submission %s: %v", sub.ID(), err)
					continue
				}
				s.logger.Infof("[OrphanSubmissionRepairService] Repaired submission %s with ERROR judgment %s: unresolvable grading items", sub.ID(), failed.ID())
				continue
			}
			s.logger.Errorf("[OrphanSubmissionRepairService] Failed to create repair judgment for submission %s: %v", sub.ID(), err)
			continue
		}

		if err := judgment.Fail(); err != nil {
			s.logger.Errorf("[OrphanSubmissionRepairService] Cannot mark judgment for submission %s as failed: %v", sub.ID(), err)
			continue
		}
		judgment.SetReason(orphanRepairReason)
		judgment.SetErrorMessage(orphanRepairErrorMessage)

		if err := s.judgmentRepo.Save(judgment); err != nil {
			s.logger.Errorf("[OrphanSubmissionRepairService] Failed to save repair judgment for submission %s: %v", sub.ID(), err)
			continue
		}

		s.logger.Infof("[OrphanSubmissionRepairService] Repaired orphan submission %s with ERROR judgment %s", sub.ID(), judgment.ID())
	}
}
