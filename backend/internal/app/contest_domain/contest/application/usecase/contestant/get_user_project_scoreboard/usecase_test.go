package getuserprojectscoreboard_test

import (
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	getuserprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_project_scoreboard"
	"judge-backend/internal/app/contest_domain/contest/domain/entity/project"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	sharedtypes "judge-backend/internal/app/shared/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	repository          *mocks.ProjectRepository
	query               *mocks.ProjectQuery
	contestProjectQuery *mocks.ContestProjectQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository:          mocks.NewProjectRepository(t),
		query:               mocks.NewProjectQuery(t),
		contestProjectQuery: mocks.NewContestProjectQuery(t),
	}
}

func NewUsecase(t *testing.T) (getuserprojectscoreboard.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return getuserprojectscoreboard.NewUsecase(m.repository, m.query, m.contestProjectQuery, &noopLogger{}), m
}

func newProject(t *testing.T) *project.Project {
	t.Helper()
	now := time.Now()
	p, err := project.New(&project.ProjectConfig{
		ID:             1,
		Name:           "Test Project",
		IsPublic:       true,
		ActivateTime:   now,
		DeactivateTime: now.Add(24 * time.Hour),
		StartTime:      now,
		EndTime:        now.Add(24 * time.Hour),
	})
	assert.NoError(t, err)
	return p
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return scoreboard successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		p := newProject(t)
		scoreboard := &dto.Scoreboard{}
		m.contestProjectQuery.On("CheckUserCanAccessProject", int32(1), int32(10)).Return(nil).Once()
		m.repository.On("FindByID", int32(1)).Return(p, nil).Once()
		m.query.On("GetProjectScoreboardByID", int32(1)).Return(scoreboard, nil).Once()

		output, err := usecase.Execute(&getuserprojectscoreboard.Input{UserID: 10, ContestProjectID: 1})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, scoreboard, output.Scoreboard)
	})

	t.Run("should return error when user cannot access project", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.contestProjectQuery.On("CheckUserCanAccessProject", int32(1), int32(10)).Return(assert.AnError).Once()

		_, err := usecase.Execute(&getuserprojectscoreboard.Input{UserID: 10, ContestProjectID: 1})

		assert.Error(t, err)
	})

	t.Run("should return error when project not found", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.contestProjectQuery.On("CheckUserCanAccessProject", int32(99), int32(10)).Return(nil).Once()
		m.repository.On("FindByID", int32(99)).Return(nil, assert.AnError).Once()

		_, err := usecase.Execute(&getuserprojectscoreboard.Input{UserID: 10, ContestProjectID: 99})

		assert.Error(t, err)
	})

	t.Run("should return error when scoreboard query fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		p := newProject(t)
		m.contestProjectQuery.On("CheckUserCanAccessProject", int32(1), int32(10)).Return(nil).Once()
		m.repository.On("FindByID", int32(1)).Return(p, nil).Once()
		m.query.On("GetProjectScoreboardByID", int32(1)).Return((*dto.Scoreboard)(nil), assert.AnError).Once()

		_, err := usecase.Execute(&getuserprojectscoreboard.Input{UserID: 10, ContestProjectID: 1})

		assert.Error(t, err)
	})
}
