package getclarification_test

import (
	getclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/get_clarification"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return clarification successfully", func(t *testing.T) {
		q := mocks.NewClarificationQuery(t)
		uc := getclarification.NewUsecase(q)

		detail := &dto.ClarificationDetail{
			Clarification: &dto.Clarification{ID: 1, Subject: "Test"},
		}
		q.On("GetClarificationByID", int32(1)).Return(detail, nil).Once()

		output, err := uc.Execute(&getclarification.Input{ClarificationID: 1})

		assert.NoError(t, err)
		assert.Equal(t, detail, output.Clarification)
	})

	t.Run("should return error when clarification not found", func(t *testing.T) {
		q := mocks.NewClarificationQuery(t)
		uc := getclarification.NewUsecase(q)

		q.On("GetClarificationByID", int32(99)).Return((*dto.ClarificationDetail)(nil), nil).Once()

		_, err := uc.Execute(&getclarification.Input{ClarificationID: 99})

		assert.ErrorIs(t, err, getclarification.ErrorClarificationNotFound)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		q := mocks.NewClarificationQuery(t)
		uc := getclarification.NewUsecase(q)

		q.On("GetClarificationByID", int32(1)).Return((*dto.ClarificationDetail)(nil), assert.AnError).Once()

		_, err := uc.Execute(&getclarification.Input{ClarificationID: 1})

		assert.Error(t, err)
	})
}
