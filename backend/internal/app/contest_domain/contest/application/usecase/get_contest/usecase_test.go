package getcontest_test

import (
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	getcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/get_contest"
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

func NewUsecase(t *testing.T) (getcontest.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return getcontest.NewUsecase(m.query), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return contest detail", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("GetContestByID", int32(1)).Return(&dto.AdminContestDetail{}, nil).Once()

		output, err := usecase.Execute(&getcontest.Input{ID: 1})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.NotNil(t, output.Contest)
	})

	t.Run("should return error when id is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)

		_, err := usecase.Execute(&getcontest.Input{ID: 0})

		assert.ErrorIs(t, err, getcontest.ErrorInvalidID)
	})

	t.Run("should return error when query returns error", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("GetContestByID", int32(1)).Return(nil, assert.AnError).Once()

		_, err := usecase.Execute(&getcontest.Input{ID: 1})

		assert.Error(t, err)
	})
}
