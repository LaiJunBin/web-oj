package listjudgmenterrors_test

import (
	"testing"

	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	listjudgmenterrors "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/list_judgment_errors"
	"judge-backend/internal/app/contest_domain/activity/mocks"

	"github.com/stretchr/testify/assert"
)

func int32Ptr(v int32) *int32 { return &v }

func TestUsecase_Execute(t *testing.T) {
	t.Run("defaults the limit when the caller does not ask for one", func(t *testing.T) {
		q := mocks.NewSystemStatusQuery(t)
		uc := listjudgmenterrors.NewUsecase(q)

		q.On("ListJudgmentErrors", listjudgmenterrors.DefaultLimit).
			Return([]*dto.JudgmentError{}, nil).Once()

		output, err := uc.Execute(&listjudgmenterrors.Input{})

		assert.NoError(t, err)
		assert.Empty(t, output.Errors)
	})

	t.Run("passes the requested limit through", func(t *testing.T) {
		q := mocks.NewSystemStatusQuery(t)
		uc := listjudgmenterrors.NewUsecase(q)

		q.On("ListJudgmentErrors", int32(5)).
			Return([]*dto.JudgmentError{{JudgmentID: "j-1"}}, nil).Once()

		output, err := uc.Execute(&listjudgmenterrors.Input{Limit: int32Ptr(5)})

		assert.NoError(t, err)
		assert.Len(t, output.Errors, 1)
	})

	t.Run("rejects a limit outside the supported range without querying", func(t *testing.T) {
		for _, limit := range []int32{0, -1, listjudgmenterrors.MaxLimit + 1} {
			q := mocks.NewSystemStatusQuery(t)
			uc := listjudgmenterrors.NewUsecase(q)

			_, err := uc.Execute(&listjudgmenterrors.Input{Limit: int32Ptr(limit)})

			assert.ErrorIs(t, err, listjudgmenterrors.ErrorInvalidLimit, "limit=%d", limit)
			q.AssertNotCalled(t, "ListJudgmentErrors")
		}
	})

	t.Run("returns the error when the query fails", func(t *testing.T) {
		q := mocks.NewSystemStatusQuery(t)
		uc := listjudgmenterrors.NewUsecase(q)

		q.On("ListJudgmentErrors", listjudgmenterrors.DefaultLimit).
			Return(([]*dto.JudgmentError)(nil), assert.AnError).Once()

		_, err := uc.Execute(&listjudgmenterrors.Input{})

		assert.Error(t, err)
	})
}
