package handler

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/events/projectsubmitted"
	gradingitem "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	judgmentproject "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/submission"
	"judge-backend/internal/app/judge_domain/judge/domain/factory"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Regression tests for the submissionRepo.Save failure branch of
// CreateSubmission: it used to clean up and return WITHOUT recording anything,
// so the submission silently vanished while the API had already answered
// success — the one branch that skipped recordFailedSubmission.

type projectSubmittedFixture struct {
	handler        *ProjectSubmittedHandler
	eventRepo      *sharedmocks.EventRepository
	submissionRepo *judgemocks.SubmissionRepository
	service        *judgemocks.SubmissionService
	judgmentRepo   *judgemocks.JudgmentRepository
	projectRepo    *judgemocks.ProjectRepository
	eventbus       *sharedmocks.EventBus
}

func newProjectSubmittedFixture(t *testing.T) *projectSubmittedFixture {
	f := &projectSubmittedFixture{
		eventRepo:      sharedmocks.NewEventRepository(t),
		submissionRepo: judgemocks.NewSubmissionRepository(t),
		service:        judgemocks.NewSubmissionService(t),
		judgmentRepo:   judgemocks.NewJudgmentRepository(t),
		projectRepo:    judgemocks.NewProjectRepository(t),
		eventbus:       sharedmocks.NewEventBus(t),
	}
	gradingItemRepo := judgemocks.NewGradingItemRepository(t)
	lifecycleRepo := judgemocks.NewLifecycleRepository(t)
	// The factory's FindBySubmissionID calls only happen when the record
	// path reaches CreateJudgment; register them leniently (Maybe) so the
	// "record also fails" scenario does not trip unmet expectations.
	f.projectRepo.On("FindBySubmissionID", mock.Anything).Return(judgmentproject.NewProject(&judgmentproject.ProjectConfig{
		ID:           10,
		ImageTag:     "nginx:latest",
		TestImageTag: "test-nginx:latest",
		MountPath:    "/var/www/html",
	}), nil).Maybe()
	gradingItemRepo.On("FindBySubmissionID", mock.Anything).Return([]*gradingitem.GradingItem{}, nil).Maybe()
	lifecycleRepo.On("FindBySubmissionID", mock.Anything).Return([]*lifecycle.Lifecycle{}, nil).Maybe()

	f.handler = &ProjectSubmittedHandler{
		eventbus:          f.eventbus,
		eventRepo:         f.eventRepo,
		submissionRepo:    f.submissionRepo,
		submissionService: f.service,
		judgmentRepo:      f.judgmentRepo,
		judgmentFactory:   factory.NewJudgmentFactory(f.projectRepo, f.submissionRepo, gradingItemRepo, lifecycleRepo),
		logger:            sharedlogger.NewMockLogger(),
	}
	return f
}

func projectSubmittedEvent() *projectsubmitted.Event {
	return projectsubmitted.New(
		"evt-1", 1, 10, false, nil, bytes.NewBufferString("zip-bytes"), nil, time.Now(),
	)
}

func TestCreateSubmission_SaveFailsTwice_CleansUpWithoutJudgmentOrPublish(t *testing.T) {
	f := newProjectSubmittedFixture(t)

	f.eventRepo.On("TryClaimEvent", "evt-1").Return(true, nil).Once()
	f.service.On("SaveProject", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()
	f.service.On("SaveDynamicFields", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()
	// Main-path Save fails, and the best-effort retry inside
	// recordFailedSubmission fails too.
	f.submissionRepo.On("Save", mock.Anything).Return(errors.New("db down")).Twice()
	f.service.On("CleanupSubmissionFiles", mock.AnythingOfType("string")).Return(nil).Once()
	// No judgmentRepo.Save, no eventbus.Publish: mockery's NewX(t) asserts
	// no unexpected calls automatically.

	f.handler.CreateSubmission(projectSubmittedEvent())
}

func TestCreateSubmission_SaveFailsOnceThenRecordSucceeds_ErrorJudgmentSaved(t *testing.T) {
	f := newProjectSubmittedFixture(t)

	f.eventRepo.On("TryClaimEvent", "evt-1").Return(true, nil).Once()
	f.service.On("SaveProject", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()
	f.service.On("SaveDynamicFields", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()

	var savedID string
	// Main-path Save fails; the retry inside recordFailedSubmission succeeds.
	f.submissionRepo.On("Save", mock.Anything).Return(errors.New("transient")).Once()
	f.submissionRepo.On("Save", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		savedID = args.Get(0).(*submission.Submission).ID()
	}).Once()
	f.service.On("CleanupSubmissionFiles", mock.AnythingOfType("string")).Return(nil).Once()
	// CreateJudgment re-reads the submission by ID.
	f.submissionRepo.On("FindByID", mock.AnythingOfType("string")).Return(func(id string) *submission.Submission {
		s, _ := submission.New(&submission.SubmissionConfig{UserID: 1, ContestProjectID: 10})
		s.SetID(id)
		return s
	}, nil).Once()

	f.judgmentRepo.On("Save", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		j := args.Get(0).(interface {
			Status() judgestatus.JudgeStatus
			SubmissionID() string
		})
		assert.True(t, j.Status().IsError(), "the recorded judgment must be terminal ERROR, got %v", j.Status())
		assert.Equal(t, savedID, j.SubmissionID())
	}).Once()
	// Still no submissioncreated publish: the submission failed, it must not
	// enter the judging pipeline.

	f.handler.CreateSubmission(projectSubmittedEvent())
	assert.NotEmpty(t, savedID)
}

// The double failure: the project files could not be stored AND the
// grading-item set does not resolve, so recordFailedSubmission's own
// CreateJudgment fails too. The item-less fallback judgment is what keeps the
// submission visible instead of silently dropping it.
func TestCreateSubmission_RecordHitsUnresolvableGradingItems_FallsBack(t *testing.T) {
	eventbus := sharedmocks.NewEventBus(t)
	eventRepo := sharedmocks.NewEventRepository(t)
	submissionRepo := judgemocks.NewSubmissionRepository(t)
	svc := judgemocks.NewSubmissionService(t)
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	projectRepo := judgemocks.NewProjectRepository(t)
	gradingItemRepo := judgemocks.NewGradingItemRepository(t)
	lifecycleRepo := judgemocks.NewLifecycleRepository(t)

	eventRepo.On("TryClaimEvent", "evt-1").Return(true, nil).Once()
	svc.On("SaveProject", mock.AnythingOfType("string"), mock.Anything).Return(errors.New("disk full")).Once()
	svc.On("CleanupSubmissionFiles", mock.AnythingOfType("string")).Return(nil).Once()
	submissionRepo.On("Save", mock.Anything).Return(nil).Once()
	submissionRepo.On("FindByID", mock.AnythingOfType("string")).Return(func(id string) *submission.Submission {
		s, _ := submission.New(&submission.SubmissionConfig{UserID: 1, ContestProjectID: 10})
		s.SetID(id)
		return s
	}, nil).Twice()
	projectRepo.On("FindBySubmissionID", mock.AnythingOfType("string")).Return(judgmentproject.NewProject(&judgmentproject.ProjectConfig{
		ID:           10,
		ImageTag:     "nginx:latest",
		TestImageTag: "test-nginx:latest",
		MountPath:    "/var/www/html",
	}), nil).Twice()
	lifecycleRepo.On("FindBySubmissionID", mock.AnythingOfType("string")).Return([]*lifecycle.Lifecycle{}, nil).Twice()
	// Once: the fallback path must not retry the failing load.
	gradingItemRepo.On("FindBySubmissionID", mock.AnythingOfType("string")).
		Return(nil, fmt.Errorf("unresolvable: %w", repository.ErrGradingItemsNotOwned)).Once()

	judgmentRepo.On("Save", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		j := args.Get(0).(interface {
			Status() judgestatus.JudgeStatus
			Reason() *string
		})
		assert.True(t, j.Status().IsError(), "the fallback judgment must be terminal ERROR")
		// The reason stays the storage failure: the contestant sees why THEIR
		// submission failed, not the secondary loading problem.
		if assert.NotNil(t, j.Reason()) {
			assert.Equal(t, "解壓縮或儲存提交檔案失敗", *j.Reason())
		}
	}).Once()
	// eventbus has no expectations: nothing may be published on this path.

	h := &ProjectSubmittedHandler{
		eventbus:          eventbus,
		eventRepo:         eventRepo,
		submissionRepo:    submissionRepo,
		submissionService: svc,
		judgmentRepo:      judgmentRepo,
		judgmentFactory:   factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo),
		logger:            sharedlogger.NewMockLogger(),
	}

	h.CreateSubmission(projectSubmittedEvent())
}
