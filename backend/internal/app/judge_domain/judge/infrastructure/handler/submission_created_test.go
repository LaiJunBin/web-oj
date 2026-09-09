package handler

import (
	"fmt"
	"testing"

	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/events/submissioncreated"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	judgmentproject "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/submission"
	"judge-backend/internal/app/judge_domain/judge/domain/factory"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// A submission whose recorded grading items cannot be resolved for its own
// problem must surface as a terminal ERROR judgment, not vanish: without one,
// every list and detail query hides the submission, and the orphan repair
// sweep retries the same deterministic failure forever.
func TestCreateJudgment_UnresolvableGradingItems_RecordsErrorJudgment(t *testing.T) {
	eventbus := sharedmocks.NewEventBus(t)
	eventRepo := sharedmocks.NewEventRepository(t)
	submissionRepo := judgemocks.NewSubmissionRepository(t)
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	projectRepo := judgemocks.NewProjectRepository(t)
	gradingItemRepo := judgemocks.NewGradingItemRepository(t)
	lifecycleRepo := judgemocks.NewLifecycleRepository(t)

	sub, err := submission.New(&submission.SubmissionConfig{
		UserID:           1,
		ContestProjectID: 10,
		IsPartial:        true,
		GradingItemIds:   []int32{9},
	})
	assert.NoError(t, err)
	sub.SetID("sub-1")

	// CreateJudgment loads submission/project/lifecycles before failing on
	// the grading items; CreateFailedJudgment loads the first three again.
	submissionRepo.On("FindByID", "sub-1").Return(sub, nil).Twice()
	projectRepo.On("FindBySubmissionID", "sub-1").Return(judgmentproject.NewProject(&judgmentproject.ProjectConfig{
		ID:           10,
		ImageTag:     "nginx:latest",
		TestImageTag: "test-nginx:latest",
		MountPath:    "/var/www/html",
	}), nil).Twice()
	lifecycleRepo.On("FindBySubmissionID", "sub-1").Return([]*lifecycle.Lifecycle{}, nil).Twice()
	// Exactly once: the fallback must NOT retry the failing load.
	gradingItemRepo.On("FindBySubmissionID", "sub-1").
		Return(nil, fmt.Errorf("submission sub-1: %w", repository.ErrGradingItemsNotOwned)).Once()

	eventRepo.On("TryClaimEvent", mock.AnythingOfType("string")).Return(true, nil).Once()

	var saved *judgment.Judgment
	judgmentRepo.On("Save", mock.AnythingOfType("*judgment.Judgment")).Return(nil).
		Run(func(args mock.Arguments) { saved = args.Get(0).(*judgment.Judgment) }).Once()

	// Exactly ONE publish — the judge-updated visibility event. A second
	// publish (a JudgmentCreated would enqueue the ERROR judgment) trips the
	// strict mock.
	var published interface{}
	eventbus.On("Publish", mock.Anything).
		Run(func(args mock.Arguments) { published = args.Get(0) }).Once()

	h := &SubmissionCreatedHandler{
		eventbus:        eventbus,
		eventRepo:       eventRepo,
		judgmentFactory: factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo),
		judgmentRepo:    judgmentRepo,
		logger:          sharedlogger.NewMockLogger(),
	}

	h.CreateJudgment(submissioncreated.NewSubmissionCreatedEvent("sub-1", 10, true))

	if assert.NotNil(t, saved) {
		assert.Equal(t, judgestatus.ERROR.String(), saved.Status().String())
		assert.Equal(t, judgetype.JUDGE_PARTIAL.String(), saved.JudgeType().String())
		assert.Equal(t, factory.GradingItemMismatchReason, *saved.Reason())
		assert.Equal(t, factory.GradingItemMismatchReason, *saved.ErrorMessage())
		assert.Equal(t, "sub-1", saved.SubmissionID())
		assert.Empty(t, saved.GradingItems())
	}
	assert.IsType(t, &judgeupdated.JudgeUpdatedEvent{}, published)
}
