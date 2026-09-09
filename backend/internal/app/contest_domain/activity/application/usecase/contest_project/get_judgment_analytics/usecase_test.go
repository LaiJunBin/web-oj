package getjudgmentanalytics_test

import (
	getjudgmentanalytics "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_judgment_analytics"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return analytics successfully", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		uc := getjudgmentanalytics.NewUsecase(q)

		avgDuration := float64(1500)
		avgWait := float64(200)
		analytics := &dto.JudgmentAnalytics{
			TotalSubmissions: 3,
			TotalJudgments:   5,
			StatusCounts: []*dto.JudgmentStatusCount{
				{Status: "Judged", Count: 4},
				{Status: "CompileError", Count: 1},
			},
			AvgJudgingDurationMs: &avgDuration,
			AvgQueueWaitMs:       &avgWait,
		}
		q.On("GetJudgmentAnalytics", int32(1)).Return(analytics, nil).Once()

		output, err := uc.Execute(&getjudgmentanalytics.Input{ContestProjectID: 1})

		assert.NoError(t, err)
		assert.Equal(t, int32(3), output.Analytics.TotalSubmissions)
		assert.Equal(t, int32(5), output.Analytics.TotalJudgments)
		assert.Len(t, output.Analytics.StatusCounts, 2)
		assert.Equal(t, float64(1500), *output.Analytics.AvgJudgingDurationMs)
		assert.Equal(t, float64(200), *output.Analytics.AvgQueueWaitMs)
	})

	t.Run("should return analytics with null averages when no completed judgments", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		uc := getjudgmentanalytics.NewUsecase(q)

		analytics := &dto.JudgmentAnalytics{
			TotalSubmissions:     0,
			TotalJudgments:       0,
			StatusCounts:         []*dto.JudgmentStatusCount{},
			AvgJudgingDurationMs: nil,
			AvgQueueWaitMs:       nil,
		}
		q.On("GetJudgmentAnalytics", int32(1)).Return(analytics, nil).Once()

		output, err := uc.Execute(&getjudgmentanalytics.Input{ContestProjectID: 1})

		assert.NoError(t, err)
		assert.Equal(t, int32(0), output.Analytics.TotalSubmissions)
		assert.Empty(t, output.Analytics.StatusCounts)
		assert.Nil(t, output.Analytics.AvgJudgingDurationMs)
		assert.Nil(t, output.Analytics.AvgQueueWaitMs)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		uc := getjudgmentanalytics.NewUsecase(q)

		q.On("GetJudgmentAnalytics", int32(1)).Return((*dto.JudgmentAnalytics)(nil), assert.AnError).Once()

		_, err := uc.Execute(&getjudgmentanalytics.Input{ContestProjectID: 1})

		assert.Error(t, err)
	})
}
