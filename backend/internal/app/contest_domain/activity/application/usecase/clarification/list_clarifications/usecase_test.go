package listclarifications_test

import (
	listclarifications "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/list_clarifications"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should list clarifications successfully", func(t *testing.T) {
		q := mocks.NewClarificationQuery(t)
		uc := listclarifications.NewUsecase(q)

		clarifications := []*dto.Clarification{
			{ID: 1, Subject: "First"},
			{ID: 2, Subject: "Second"},
		}
		q.On("ListClarifications", int32(10)).Return(clarifications, nil).Once()

		output, err := uc.Execute(&listclarifications.Input{ContestProjectID: 10})

		assert.NoError(t, err)
		assert.Len(t, output.Clarifications, 2)
	})

	t.Run("should return empty list when no clarifications", func(t *testing.T) {
		q := mocks.NewClarificationQuery(t)
		uc := listclarifications.NewUsecase(q)

		q.On("ListClarifications", int32(10)).Return([]*dto.Clarification{}, nil).Once()

		output, err := uc.Execute(&listclarifications.Input{ContestProjectID: 10})

		assert.NoError(t, err)
		assert.Empty(t, output.Clarifications)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		q := mocks.NewClarificationQuery(t)
		uc := listclarifications.NewUsecase(q)

		q.On("ListClarifications", int32(10)).Return(([]*dto.Clarification)(nil), assert.AnError).Once()

		_, err := uc.Execute(&listclarifications.Input{ContestProjectID: 10})

		assert.Error(t, err)
	})
}
