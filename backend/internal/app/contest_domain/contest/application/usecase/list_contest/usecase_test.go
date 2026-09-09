package listcontest_test

import (
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	listcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/list_contest"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	query *mocks.AdminContestQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query: mocks.NewAdminContestQuery(t),
	}
}

func NewUsecase(t *testing.T) (listcontest.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return listcontest.NewUsecase(m.query), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return contest list", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		contests := []*dto.AdminContest{{ID: 1, Name: "Contest 1"}, {ID: 2, Name: "Contest 2"}}
		m.query.On("ListContest").Return(contests, nil).Once()

		output, err := usecase.Execute(&listcontest.Input{})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Len(t, output.Contests, 2)
	})

	t.Run("should return empty list when no contests", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("ListContest").Return([]*dto.AdminContest{}, nil).Once()

		output, err := usecase.Execute(&listcontest.Input{})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Empty(t, output.Contests)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("ListContest").Return(nil, assert.AnError).Once()

		_, err := usecase.Execute(&listcontest.Input{})

		assert.ErrorIs(t, err, listcontest.ErrorInternalServer)
	})
}
