package getrejudge_test

import (
	getrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/get_rejudge"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return rejudge successfully", func(t *testing.T) {
		q := mocks.NewRejudgeQuery(t)
		uc := getrejudge.NewUsecase(q)

		rejudge := &dto.Rejudge{ID: "rejudge-1", Status: "Finished"}
		q.On("GetRejudgeByID", "rejudge-1").Return(rejudge, nil).Once()

		output, err := uc.Execute(&getrejudge.Input{RejudgeID: "rejudge-1"})

		assert.NoError(t, err)
		assert.Equal(t, rejudge, output.Rejudge)
	})

	t.Run("should return error when rejudge ID is empty", func(t *testing.T) {
		q := mocks.NewRejudgeQuery(t)
		uc := getrejudge.NewUsecase(q)

		_, err := uc.Execute(&getrejudge.Input{RejudgeID: ""})

		assert.ErrorIs(t, err, getrejudge.ErrRejudgeIDIsRequired)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		q := mocks.NewRejudgeQuery(t)
		uc := getrejudge.NewUsecase(q)

		q.On("GetRejudgeByID", "rejudge-1").Return((*dto.Rejudge)(nil), assert.AnError).Once()

		_, err := uc.Execute(&getrejudge.Input{RejudgeID: "rejudge-1"})

		assert.Error(t, err)
	})
}
