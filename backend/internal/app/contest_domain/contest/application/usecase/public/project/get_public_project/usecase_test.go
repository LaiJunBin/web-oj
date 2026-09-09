package getpublicproject_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	getpublicproject "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/get_public_project"
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
	query *mocks.ProjectQuery
}

func newUsecaseMocks(t *testing.T) *usecaseMocks {
	return &usecaseMocks{
		query: mocks.NewProjectQuery(t),
	}
}

func newUsecase(t *testing.T) (getpublicproject.Usecase, *usecaseMocks) {
	m := newUsecaseMocks(t)
	return getpublicproject.NewUsecase(m.query, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return public project successfully", func(t *testing.T) {
		usecase, m := newUsecase(t)

		expected := dto.ProjectDetail{
			Project: dto.Project{ID: 1, Name: "Public Project"},
		}
		m.query.On("GetPublicProjectByID", int32(1)).Return(expected, nil).Once()

		output, err := usecase.Execute(&getpublicproject.Input{ContestProjectID: 1})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, expected, output.Project)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := newUsecase(t)

		queryErr := errors.New("project not found")
		m.query.On("GetPublicProjectByID", int32(1)).Return(dto.ProjectDetail{}, queryErr).Once()

		output, err := usecase.Execute(&getpublicproject.Input{ContestProjectID: 1})

		assert.Error(t, err)
		assert.ErrorIs(t, err, queryErr)
		assert.Nil(t, output)
	})
}
