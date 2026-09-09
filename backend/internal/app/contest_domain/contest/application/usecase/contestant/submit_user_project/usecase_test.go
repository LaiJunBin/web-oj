package submituserproject_test

import (
	"bytes"
	submituserproject "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/submit_user_project"
	"judge-backend/internal/app/contest_domain/contest/domain/entity/project"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/config"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

type noopLogger struct{}

func (l *noopLogger) Debugf(msg string, args ...interface{}) {}
func (l *noopLogger) Infof(msg string, args ...interface{})  {}
func (l *noopLogger) Warnf(msg string, args ...interface{})  {}
func (l *noopLogger) Errorf(msg string, args ...interface{}) {}
func (l *noopLogger) Fatalf(msg string, args ...interface{}) {}
func (l *noopLogger) Panicf(msg string, args ...interface{}) {}

var _ sharedtypes.Logger = (*noopLogger)(nil)

type UsecaseMocks struct {
	query       *mocks.ContestProjectQuery
	projectRepo *mocks.ProjectRepository
	cooldown    *mocks.SubmissionCooldown
	publisher   *mocks.ProjectSubmitPublisher
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query:       mocks.NewContestProjectQuery(t),
		projectRepo: mocks.NewProjectRepository(t),
		cooldown:    mocks.NewSubmissionCooldown(t),
		publisher:   mocks.NewProjectSubmitPublisher(t),
	}
}

// testConfig carries the system-default cooldown the contest project falls
// back to when it has none of its own.
func testConfig(defaultCooldownSeconds int) *config.Config {
	return &config.Config{Judge: &config.JudgeConfig{SubmissionCooldownSeconds: defaultCooldownSeconds}}
}

func NewUsecase(t *testing.T) (submituserproject.Usecase, *UsecaseMocks) {
	return newUsecaseWithConfig(t, testConfig(0))
}

func newUsecaseWithConfig(t *testing.T, cfg *config.Config) (submituserproject.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return submituserproject.NewUsecase(m.query, m.projectRepo, m.cooldown, m.publisher, cfg, &noopLogger{}), m
}

// projectWithWindow builds the domain entity the submission-window gate reads.
// Boundary tests belong to the entity's own test; here the windows are wide.
func projectWithWindow(t *testing.T, start, end time.Time) *project.Project {
	t.Helper()
	p, err := project.New(&project.ProjectConfig{
		ID:             10,
		Name:           "Project1",
		ActivateTime:   start.Add(-time.Hour),
		DeactivateTime: end.Add(time.Hour),
		StartTime:      start,
		EndTime:        end,
	})
	if err != nil {
		t.Fatalf("failed to build project entity: %v", err)
	}
	return p
}

// openProject is the "gate passes" default for subtests that only care about
// what happens beyond the gate. Registered per-subtest with .Maybe() rather
// than in a shared helper: testify matches expectations in registration
// order, and a shared catch-all would shadow the negative cases' overrides.
func openProject(t *testing.T) *project.Project {
	return projectWithWindow(t, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
}

// int32Ptr marks a contest project as having its own cooldown, as opposed to
// inheriting the system default.
func int32Ptr(v int32) *int32 { return &v }

func validInput() *submituserproject.Input {
	return &submituserproject.Input{
		UserID:           1,
		ContestProjectID: 10,
		Project:          bytes.NewBufferString("project data"),
		GradingItemIDs:   []int32{},
		DynamicFields:    map[string]string{},
	}
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should submit project successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("CheckUserIsParticipant", int32(10), int32(1)).Return(nil).Once()
		m.query.On("CheckProjectAndGradingItemsExists", int32(10), []int32{}).Return(nil).Once()
		m.projectRepo.On("FindByID", int32(10)).Return(openProject(t), nil).Maybe()
		m.query.On("GetSubmissionCooldownSeconds", int32(10)).Return(int32Ptr(0), nil).Maybe()
		m.cooldown.On("Reserve", int32(1), int32(10), int32(0)).Return(true, int32(0), nil).Maybe()
		m.publisher.On("Publish", testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil).Once()

		output, err := usecase.Execute(validInput())

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when user id is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := validInput()
		input.UserID = 0

		_, err := usecase.Execute(input)

		assert.ErrorIs(t, err, submituserproject.ErrorInvalidUserID)
	})

	t.Run("should return error when contest project id is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := validInput()
		input.ContestProjectID = 0

		_, err := usecase.Execute(input)

		assert.ErrorIs(t, err, submituserproject.ErrorInvalidContestProjectID)
	})

	t.Run("should return error when project is nil", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := validInput()
		input.Project = nil

		_, err := usecase.Execute(input)

		assert.ErrorIs(t, err, submituserproject.ErrorProjectRequired)
	})

	t.Run("should return error when project or grading items do not exist", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("CheckUserIsParticipant", int32(10), int32(1)).Return(nil).Once()
		m.query.On("CheckProjectAndGradingItemsExists", int32(10), []int32{}).Return(assert.AnError).Once()

		_, err := usecase.Execute(validInput())

		assert.Error(t, err)
	})

	t.Run("should submit partial project with specific grading items", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		gradingItemIDs := []int32{1, 2, 3}
		m.query.On("CheckUserIsParticipant", int32(10), int32(1)).Return(nil).Once()
		m.query.On("CheckProjectAndGradingItemsExists", int32(10), gradingItemIDs).Return(nil).Once()
		m.projectRepo.On("FindByID", int32(10)).Return(openProject(t), nil).Maybe()
		m.query.On("GetSubmissionCooldownSeconds", int32(10)).Return(int32Ptr(0), nil).Maybe()
		m.cooldown.On("Reserve", int32(1), int32(10), int32(0)).Return(true, int32(0), nil).Maybe()
		m.publisher.On("Publish", testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil).Once()

		input := validInput()
		input.GradingItemIDs = gradingItemIDs

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when user is not a participant", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("CheckUserIsParticipant", int32(10), int32(1)).Return(assert.AnError).Once()

		_, err := usecase.Execute(validInput())

		assert.Error(t, err)
	})

	// The gate is consulted with whatever the contest project resolves to,
	// and its verdict is what the caller sees.
	t.Run("rejects a submission the gate refuses and reports the wait", func(t *testing.T) {
		usecase, m := newUsecaseWithConfig(t, testConfig(0))
		m.query.On("CheckUserIsParticipant", int32(10), int32(1)).Return(nil).Once()
		m.query.On("CheckProjectAndGradingItemsExists", int32(10), []int32{}).Return(nil).Once()
		m.projectRepo.On("FindByID", int32(10)).Return(openProject(t), nil).Maybe()
		m.query.On("GetSubmissionCooldownSeconds", int32(10)).Return(int32Ptr(30), nil).Once()
		m.cooldown.On("Reserve", int32(1), int32(10), int32(30)).Return(false, int32(20), nil).Once()

		_, err := usecase.Execute(validInput())

		var cooldownErr *submituserproject.SubmissionCooldownError
		assert.ErrorAs(t, err, &cooldownErr)
		assert.EqualValues(t, 20, cooldownErr.RetryAfterSeconds)
		m.publisher.AssertNotCalled(t, "Publish")
	})

	t.Run("submits when the gate admits", func(t *testing.T) {
		usecase, m := newUsecaseWithConfig(t, testConfig(0))
		m.query.On("CheckUserIsParticipant", int32(10), int32(1)).Return(nil).Once()
		m.query.On("CheckProjectAndGradingItemsExists", int32(10), []int32{}).Return(nil).Once()
		m.projectRepo.On("FindByID", int32(10)).Return(openProject(t), nil).Maybe()
		m.query.On("GetSubmissionCooldownSeconds", int32(10)).Return(int32Ptr(30), nil).Once()
		m.cooldown.On("Reserve", int32(1), int32(10), int32(30)).Return(true, int32(0), nil).Once()
		m.publisher.On("Publish", testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil).Once()

		_, err := usecase.Execute(validInput())

		assert.NoError(t, err)
	})

	// A contest project storing 0 has deliberately turned the cooldown off
	// and must not inherit the system default.
	t.Run("an explicit zero overrides a non-zero system default", func(t *testing.T) {
		usecase, m := newUsecaseWithConfig(t, testConfig(30))
		m.query.On("CheckUserIsParticipant", int32(10), int32(1)).Return(nil).Once()
		m.query.On("CheckProjectAndGradingItemsExists", int32(10), []int32{}).Return(nil).Once()
		m.projectRepo.On("FindByID", int32(10)).Return(openProject(t), nil).Maybe()
		m.query.On("GetSubmissionCooldownSeconds", int32(10)).Return(int32Ptr(0), nil).Once()
		m.cooldown.On("Reserve", int32(1), int32(10), int32(0)).Return(true, int32(0), nil).Once()
		m.publisher.On("Publish", testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil).Once()

		_, err := usecase.Execute(validInput())

		assert.NoError(t, err)
	})

	// nil is "not configured", which is what makes the system default reachable.
	t.Run("falls back to the system default when the contest project has none", func(t *testing.T) {
		usecase, m := newUsecaseWithConfig(t, testConfig(30))
		m.query.On("CheckUserIsParticipant", int32(10), int32(1)).Return(nil).Once()
		m.query.On("CheckProjectAndGradingItemsExists", int32(10), []int32{}).Return(nil).Once()
		m.projectRepo.On("FindByID", int32(10)).Return(openProject(t), nil).Maybe()
		m.query.On("GetSubmissionCooldownSeconds", int32(10)).Return(nil, nil).Once()
		m.cooldown.On("Reserve", int32(1), int32(10), int32(30)).Return(false, int32(25), nil).Once()

		_, err := usecase.Execute(validInput())

		var cooldownErr *submituserproject.SubmissionCooldownError
		assert.ErrorAs(t, err, &cooldownErr)
		assert.EqualValues(t, 25, cooldownErr.RetryAfterSeconds)
	})

	// The three rejections below also pin the gate's position in the chain:
	// after the participant/existence checks (their subtests above deliberately
	// register no FindByID expectation, so moving the gate ahead of them makes
	// this file panic), and before the cooldown (a submission rejected on time
	// must not consume the contestant's cooldown).
	t.Run("rejects a submission before the submission window opens", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("CheckUserIsParticipant", int32(10), int32(1)).Return(nil).Once()
		m.query.On("CheckProjectAndGradingItemsExists", int32(10), []int32{}).Return(nil).Once()
		m.projectRepo.On("FindByID", int32(10)).
			Return(projectWithWindow(t, time.Now().Add(time.Hour), time.Now().Add(2*time.Hour)), nil).Once()

		_, err := usecase.Execute(validInput())

		assert.ErrorIs(t, err, submituserproject.ErrorSubmissionNotStarted)
		m.cooldown.AssertNotCalled(t, "Reserve")
		m.publisher.AssertNotCalled(t, "Publish")
	})

	t.Run("rejects a submission after the submission window closes", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("CheckUserIsParticipant", int32(10), int32(1)).Return(nil).Once()
		m.query.On("CheckProjectAndGradingItemsExists", int32(10), []int32{}).Return(nil).Once()
		m.projectRepo.On("FindByID", int32(10)).
			Return(projectWithWindow(t, time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour)), nil).Once()

		_, err := usecase.Execute(validInput())

		assert.ErrorIs(t, err, submituserproject.ErrorSubmissionEnded)
		m.cooldown.AssertNotCalled(t, "Reserve")
		m.publisher.AssertNotCalled(t, "Publish")
	})

	// A non-activated contest project answers exactly like a missing one, so
	// probing the API reveals nothing beyond what the read side already shows.
	t.Run("rejects a submission when the contest project is not activated", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("CheckUserIsParticipant", int32(10), int32(1)).Return(nil).Once()
		m.query.On("CheckProjectAndGradingItemsExists", int32(10), []int32{}).Return(nil).Once()
		m.projectRepo.On("FindByID", int32(10)).
			Return(nil, httperror.New(http.StatusNotFound, "project not found")).Once()

		_, err := usecase.Execute(validInput())

		assert.ErrorIs(t, err, submituserproject.ErrorContestProjectNotFound)
		m.cooldown.AssertNotCalled(t, "Reserve")
		m.publisher.AssertNotCalled(t, "Publish")
	})
}
