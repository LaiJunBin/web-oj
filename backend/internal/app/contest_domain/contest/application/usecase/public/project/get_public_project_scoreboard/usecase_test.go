package getpublicprojectscoreboard_test

import (
	"errors"
	"testing"
	"time"

	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	getpublicprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/get_public_project_scoreboard"
	"judge-backend/internal/app/contest_domain/contest/domain/entity/project"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	sharedtypes "judge-backend/internal/app/shared/types"

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

type usecaseMocks struct {
	repository *mocks.ProjectRepository
	query      *mocks.ProjectQuery
}

func newUsecaseMocks(t *testing.T) *usecaseMocks {
	return &usecaseMocks{
		repository: mocks.NewProjectRepository(t),
		query:      mocks.NewProjectQuery(t),
	}
}

func newUsecase(t *testing.T) (getpublicprojectscoreboard.Usecase, *usecaseMocks) {
	m := newUsecaseMocks(t)
	return getpublicprojectscoreboard.NewUsecase(m.repository, m.query, &noopLogger{}), m
}

func newPublicProject(t *testing.T) *project.Project {
	p, _ := project.New(&project.ProjectConfig{
		ID:             1,
		Name:           "Public Project",
		IsPublic:       true,
		ActivateTime:   time.Now(),
		DeactivateTime: time.Now().Add(24 * time.Hour),
		StartTime:      time.Now(),
		EndTime:        time.Now().Add(24 * time.Hour),
	})
	return p
}

func newPrivateProject(t *testing.T) *project.Project {
	p, _ := project.New(&project.ProjectConfig{
		ID:             2,
		Name:           "Private Project",
		IsPublic:       false,
		ActivateTime:   time.Now(),
		DeactivateTime: time.Now().Add(24 * time.Hour),
		StartTime:      time.Now(),
		EndTime:        time.Now().Add(24 * time.Hour),
	})
	return p
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return scoreboard for public project", func(t *testing.T) {
		usecase, m := newUsecase(t)

		pub := newPublicProject(t)
		expected := &dto.Scoreboard{
			Heads: []*dto.ScoreboardHead{{ID: 1, Name: "Problem 1"}},
			Rows:  []*dto.ScoreboardRow{{Name: "Alice", Username: "alice", Statuses: []string{"AC"}}},
		}
		m.repository.On("FindByID", int32(1)).Return(pub, nil).Once()
		m.query.On("GetProjectScoreboardByID", int32(1)).Return(expected, nil).Once()

		output, err := usecase.Execute(&getpublicprojectscoreboard.Input{ContestProjectID: 1})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, expected, output.Scoreboard)
	})

	t.Run("should return error when project not found", func(t *testing.T) {
		usecase, m := newUsecase(t)

		repoErr := errors.New("project not found")
		m.repository.On("FindByID", int32(1)).Return(nil, repoErr).Once()

		output, err := usecase.Execute(&getpublicprojectscoreboard.Input{ContestProjectID: 1})

		assert.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Nil(t, output)
	})

	t.Run("should return error when project is not public", func(t *testing.T) {
		usecase, m := newUsecase(t)

		priv := newPrivateProject(t)
		m.repository.On("FindByID", int32(2)).Return(priv, nil).Once()

		output, err := usecase.Execute(&getpublicprojectscoreboard.Input{ContestProjectID: 2})

		assert.Error(t, err)
		assert.ErrorIs(t, err, getpublicprojectscoreboard.ErrorProjectNotFound)
		assert.Nil(t, output)
	})

	t.Run("should return error when scoreboard query fails", func(t *testing.T) {
		usecase, m := newUsecase(t)

		pub := newPublicProject(t)
		queryErr := errors.New("scoreboard query error")
		m.repository.On("FindByID", int32(1)).Return(pub, nil).Once()
		m.query.On("GetProjectScoreboardByID", int32(1)).Return(nil, queryErr).Once()

		output, err := usecase.Execute(&getpublicprojectscoreboard.Input{ContestProjectID: 1})

		assert.Error(t, err)
		assert.ErrorIs(t, err, queryErr)
		assert.Nil(t, output)
	})
}
