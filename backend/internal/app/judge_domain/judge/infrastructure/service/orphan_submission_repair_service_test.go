package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	judgmentproject "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/submission"
	"judge-backend/internal/app/judge_domain/judge/domain/factory"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	logger "judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/config"
)

func TestOrphanSubmissionRepairService_RepairOnce(t *testing.T) {
	submissionRepo := judgemocks.NewSubmissionRepository(t)
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	projectRepo := judgemocks.NewProjectRepository(t)
	gradingItemRepo := judgemocks.NewGradingItemRepository(t)
	lifecycleRepo := judgemocks.NewLifecycleRepository(t)
	judgmentFactory := factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo)

	normalSubmission, err := submission.New(&submission.SubmissionConfig{
		UserID:           1,
		ContestProjectID: 10,
		IsPartial:        false,
	})
	assert.NoError(t, err)
	normalSubmission.SetID("sub-normal")

	partialSubmission, err := submission.New(&submission.SubmissionConfig{
		UserID:           2,
		ContestProjectID: 10,
		IsPartial:        true,
	})
	assert.NoError(t, err)
	partialSubmission.SetID("sub-partial")

	submissionRepo.On("FindWithoutJudgment", mock.AnythingOfType("time.Time")).Return([]*submission.Submission{normalSubmission, partialSubmission}, nil).Once()
	submissionRepo.On("FindByID", "sub-normal").Return(normalSubmission, nil).Once()
	submissionRepo.On("FindByID", "sub-partial").Return(partialSubmission, nil).Once()

	projectRepo.On("FindBySubmissionID", mock.Anything).Return(judgmentproject.NewProject(&judgmentproject.ProjectConfig{
		ID:           10,
		ImageTag:     "nginx:latest",
		TestImageTag: "test-nginx:latest",
		MountPath:    "/var/www/html",
	}), nil).Twice()
	gradingItemRepo.On("FindBySubmissionID", mock.Anything).Return([]*gradingitem.GradingItem{}, nil).Twice()
	lifecycleRepo.On("FindBySubmissionID", mock.Anything).Return([]*lifecycle.Lifecycle{}, nil).Twice()

	saved := []*judgment.Judgment{}
	judgmentRepo.On("Save", mock.AnythingOfType("*judgment.Judgment")).Return(nil).Run(func(args mock.Arguments) {
		saved = append(saved, args.Get(0).(*judgment.Judgment))
	}).Twice()

	service := &OrphanSubmissionRepairService{
		submissionRepo:  submissionRepo,
		judgmentRepo:    judgmentRepo,
		judgmentFactory: judgmentFactory,
		logger:          logger.NewMockLogger(),
	}

	service.RepairOnce()

	if assert.Len(t, saved, 2) {
		assert.Equal(t, judgestatus.ERROR.String(), saved[0].Status().String())
		assert.Equal(t, judgetype.JUDGE.String(), saved[0].JudgeType().String())
		assert.Equal(t, orphanRepairReason, *saved[0].Reason())
		assert.Equal(t, orphanRepairErrorMessage, *saved[0].ErrorMessage())
		assert.Equal(t, "sub-normal", saved[0].SubmissionID())

		assert.Equal(t, judgestatus.ERROR.String(), saved[1].Status().String())
		assert.Equal(t, judgetype.JUDGE_PARTIAL.String(), saved[1].JudgeType().String())
		assert.Equal(t, orphanRepairReason, *saved[1].Reason())
		assert.Equal(t, orphanRepairErrorMessage, *saved[1].ErrorMessage())
		assert.Equal(t, "sub-partial", saved[1].SubmissionID())
	}
}

// An orphan whose grading items cannot be resolved is a DETERMINISTIC
// failure: retrying it every sweep can never succeed. It must leave the
// orphan set through an item-less ERROR judgment instead.
func TestOrphanSubmissionRepairService_RepairOnce_UnresolvableGradingItems(t *testing.T) {
	submissionRepo := judgemocks.NewSubmissionRepository(t)
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	projectRepo := judgemocks.NewProjectRepository(t)
	gradingItemRepo := judgemocks.NewGradingItemRepository(t)
	lifecycleRepo := judgemocks.NewLifecycleRepository(t)
	judgmentFactory := factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo)

	orphan, err := submission.New(&submission.SubmissionConfig{
		UserID:           1,
		ContestProjectID: 10,
		IsPartial:        true,
		GradingItemIds:   []int32{9},
	})
	assert.NoError(t, err)
	orphan.SetID("sub-unresolvable")

	submissionRepo.On("FindWithoutJudgment", mock.AnythingOfType("time.Time")).Return([]*submission.Submission{orphan}, nil).Once()
	submissionRepo.On("FindByID", "sub-unresolvable").Return(orphan, nil).Twice()
	projectRepo.On("FindBySubmissionID", "sub-unresolvable").Return(judgmentproject.NewProject(&judgmentproject.ProjectConfig{
		ID:           10,
		ImageTag:     "nginx:latest",
		TestImageTag: "test-nginx:latest",
		MountPath:    "/var/www/html",
	}), nil).Twice()
	lifecycleRepo.On("FindBySubmissionID", "sub-unresolvable").Return([]*lifecycle.Lifecycle{}, nil).Twice()
	// Once: the fallback path must not retry the failing load.
	gradingItemRepo.On("FindBySubmissionID", "sub-unresolvable").
		Return(nil, fmt.Errorf("submission sub-unresolvable: %w", repository.ErrGradingItemsNotOwned)).Once()

	var saved *judgment.Judgment
	judgmentRepo.On("Save", mock.AnythingOfType("*judgment.Judgment")).Return(nil).
		Run(func(args mock.Arguments) { saved = args.Get(0).(*judgment.Judgment) }).Once()

	service := &OrphanSubmissionRepairService{
		submissionRepo:  submissionRepo,
		judgmentRepo:    judgmentRepo,
		judgmentFactory: judgmentFactory,
		logger:          logger.NewMockLogger(),
	}

	service.RepairOnce()

	if assert.NotNil(t, saved) {
		assert.Equal(t, judgestatus.ERROR.String(), saved.Status().String())
		assert.Equal(t, judgetype.JUDGE_PARTIAL.String(), saved.JudgeType().String())
		assert.Equal(t, factory.GradingItemMismatchReason, *saved.Reason())
		assert.Equal(t, "sub-unresolvable", saved.SubmissionID())
		assert.Empty(t, saved.GradingItems())
	}
}

func TestOrphanSubmissionRepairService_RepairOnce_NoOrphans(t *testing.T) {
	submissionRepo := judgemocks.NewSubmissionRepository(t)
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	submissionRepo.On("FindWithoutJudgment", mock.AnythingOfType("time.Time")).Return([]*submission.Submission{}, nil).Once()

	service := &OrphanSubmissionRepairService{
		submissionRepo: submissionRepo,
		judgmentRepo:   judgmentRepo,
		logger:         logger.NewMockLogger(),
	}

	service.RepairOnce()
}

// The cutoff IS the safety property, so its three states are pinned: the
// config-less default, the "disabled ticker keeps the startup sweep graceless"
// escape hatch, and a configured interval.
func TestOrphanSubmissionRepairService_GraceCutoff(t *testing.T) {
	now := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	intervalPtr := func(v int) *int { return &v }

	t.Run("defaults to twice the default interval without config", func(t *testing.T) {
		s := &OrphanSubmissionRepairService{}
		assert.Equal(t, now.Add(-600*time.Second), s.graceCutoff(now))
	})

	t.Run("a non-positive interval means no grace", func(t *testing.T) {
		s := &OrphanSubmissionRepairService{config: &config.Config{Judge: &config.JudgeConfig{OrphanSubmissionRepairInterval: intervalPtr(0)}}}
		assert.Equal(t, now, s.graceCutoff(now))
	})

	t.Run("a configured interval doubles into the grace", func(t *testing.T) {
		s := &OrphanSubmissionRepairService{config: &config.Config{Judge: &config.JudgeConfig{OrphanSubmissionRepairInterval: intervalPtr(120)}}}
		assert.Equal(t, now.Add(-240*time.Second), s.graceCutoff(now))
	})
}

// Pins the wiring: RepairOnce must hand the GRACED cutoff to the repository.
// Each link is pinned on its own (graceCutoff above, the $match in the
// repository's mongo test), but without this a RepairOnce that passed a bare
// time.Now() would still turn every other pin green.
func TestOrphanSubmissionRepairService_RepairOnce_PassesGracedCutoff(t *testing.T) {
	submissionRepo := judgemocks.NewSubmissionRepository(t)
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	interval := 300

	submissionRepo.On("FindWithoutJudgment", mock.MatchedBy(func(cutoff time.Time) bool {
		age := time.Since(cutoff)
		return age > 9*time.Minute && age < 11*time.Minute
	})).Return([]*submission.Submission{}, nil).Once()

	service := &OrphanSubmissionRepairService{
		submissionRepo: submissionRepo,
		judgmentRepo:   judgmentRepo,
		config:         &config.Config{Judge: &config.JudgeConfig{OrphanSubmissionRepairInterval: &interval}},
		logger:         logger.NewMockLogger(),
	}

	service.RepairOnce()
}
