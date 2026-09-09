package getuserscores_test

import (
	getuserscores "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_user_scores"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return user scores successfully", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		uc := getuserscores.NewUsecase(q)

		scores := []*dto.UserScore{
			{Score: 95.0, SubmissionID: "sub-1"},
		}
		q.On("GetUserScores", int32(1)).Return(scores, nil).Once()

		output, err := uc.Execute(&getuserscores.Input{ID: 1})

		assert.NoError(t, err)
		assert.Len(t, output.UserScores, 1)
	})

	t.Run("should return empty scores when none exist", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		uc := getuserscores.NewUsecase(q)

		q.On("GetUserScores", int32(1)).Return([]*dto.UserScore{}, nil).Once()

		output, err := uc.Execute(&getuserscores.Input{ID: 1})

		assert.NoError(t, err)
		assert.Empty(t, output.UserScores)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		uc := getuserscores.NewUsecase(q)

		q.On("GetUserScores", int32(1)).Return(([]*dto.UserScore)(nil), assert.AnError).Once()

		_, err := uc.Execute(&getuserscores.Input{ID: 1})

		assert.Error(t, err)
	})
}
