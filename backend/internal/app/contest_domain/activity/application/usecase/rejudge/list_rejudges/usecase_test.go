package listrejudges_test

import (
	listrejudges "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/list_rejudges"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should list rejudges successfully", func(t *testing.T) {
		q := mocks.NewRejudgeQuery(t)
		uc := listrejudges.NewUsecase(q)

		rejudges := []*dto.RejudgeInfo{
			{ID: "rejudge-1", Status: "Finished"},
			{ID: "rejudge-2", Status: "Pending"},
		}
		q.On("ListRejudges").Return(rejudges, nil).Once()

		output, err := uc.Execute(&listrejudges.Input{})

		assert.NoError(t, err)
		assert.Len(t, output.Rejudges, 2)
	})

	t.Run("should return empty list when no rejudges", func(t *testing.T) {
		q := mocks.NewRejudgeQuery(t)
		uc := listrejudges.NewUsecase(q)

		q.On("ListRejudges").Return([]*dto.RejudgeInfo{}, nil).Once()

		output, err := uc.Execute(&listrejudges.Input{})

		assert.NoError(t, err)
		assert.Empty(t, output.Rejudges)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		q := mocks.NewRejudgeQuery(t)
		uc := listrejudges.NewUsecase(q)

		q.On("ListRejudges").Return(([]*dto.RejudgeInfo)(nil), assert.AnError).Once()

		_, err := uc.Execute(&listrejudges.Input{})

		assert.Error(t, err)
	})
}
