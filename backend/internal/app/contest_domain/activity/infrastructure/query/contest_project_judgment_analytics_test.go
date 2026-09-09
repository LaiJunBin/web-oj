package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBuildJudgmentAnalytics_SeparatesJudgingAndQueueAverages(t *testing.T) {
	base := time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC)

	// Background:
	// A previous bug returned identical avgJudgingDurationMs and avgQueueWaitMs
	// because both API fields pointed to the same temporary variable.
	// This fixture keeps judging duration and queue wait intentionally different
	// so a regression becomes obvious immediately.
	submissions := []*analyticsSubmission{
		{ID: "sub-1", CreatedAt: base},
	}
	started := base.Add(10 * time.Second)
	finished := base.Add(25 * time.Second)
	judgments := []*analyticsJudgment{
		{
			SubmissionID:      "sub-1",
			JudgeType:         "JUDGE",
			Status:            "Judged",
			CreatedAt:         base.Add(1 * time.Second),
			JudgingStartedAt:  &started,
			JudgingFinishedAt: &finished,
		},
	}

	analytics := buildJudgmentAnalytics(submissions, judgments)

	if assert.NotNil(t, analytics.AvgJudgingDurationMs) {
		assert.Equal(t, 15000.0, *analytics.AvgJudgingDurationMs)
	}
	if assert.NotNil(t, analytics.AvgQueueWaitMs) {
		assert.Equal(t, 10000.0, *analytics.AvgQueueWaitMs)
	}
	if analytics.AvgJudgingDurationMs != nil && analytics.AvgQueueWaitMs != nil {
		assert.NotEqual(t, *analytics.AvgJudgingDurationMs, *analytics.AvgQueueWaitMs)
	}
}

func TestBuildJudgmentAnalytics_UsesJudgmentCreatedAtForRejudges(t *testing.T) {
	base := time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC)

	// Background:
	// Rejudges can be triggered long after the original submission was created.
	// Queue wait must therefore start from judgment.createdAt for REJUDGE,
	// not from submission.createdAt, otherwise the admin analytics page inflates
	// queue time by hours or days.
	submissions := []*analyticsSubmission{
		{ID: "sub-1", CreatedAt: base},
	}
	rejudgeCreated := base.Add(2 * time.Hour)
	started := rejudgeCreated.Add(30 * time.Second)
	finished := started.Add(45 * time.Second)
	judgments := []*analyticsJudgment{
		{
			SubmissionID:      "sub-1",
			JudgeType:         "REJUDGE",
			Status:            "Rejudged",
			CreatedAt:         rejudgeCreated,
			JudgingStartedAt:  &started,
			JudgingFinishedAt: &finished,
		},
	}

	analytics := buildJudgmentAnalytics(submissions, judgments)

	if assert.NotNil(t, analytics.AvgQueueWaitMs) {
		assert.Equal(t, 30000.0, *analytics.AvgQueueWaitMs)
	}
	if assert.NotNil(t, analytics.AvgJudgingDurationMs) {
		assert.Equal(t, 45000.0, *analytics.AvgJudgingDurationMs)
	}
}

func TestBuildJudgmentAnalytics_IgnoresIncompleteAndFailedJudgmentsForAverages(t *testing.T) {
	base := time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC)

	// Design intent:
	// Status counts should still reflect all judgments shown on the admin page,
	// but averages should only use successfully completed work. This keeps the
	// page informative without letting failed or half-written rows skew timing.
	submissions := []*analyticsSubmission{
		{ID: "sub-1", CreatedAt: base},
		{ID: "sub-2", CreatedAt: base.Add(1 * time.Minute)},
		{ID: "sub-3", CreatedAt: base.Add(2 * time.Minute)},
	}
	started1 := base.Add(5 * time.Second)
	finished1 := started1.Add(20 * time.Second)
	started2 := base.Add(70 * time.Second)
	judgments := []*analyticsJudgment{
		{
			SubmissionID:      "sub-1",
			JudgeType:         "JUDGE",
			Status:            "Judged",
			CreatedAt:         base.Add(1 * time.Second),
			JudgingStartedAt:  &started1,
			JudgingFinishedAt: &finished1,
		},
		{
			SubmissionID:      "sub-2",
			JudgeType:         "JUDGE",
			Status:            "Error",
			CreatedAt:         base.Add(61 * time.Second),
			JudgingStartedAt:  &started2,
			JudgingFinishedAt: nil,
		},
		{
			SubmissionID:      "sub-3",
			JudgeType:         "JUDGE",
			Status:            "Queued",
			CreatedAt:         base.Add(121 * time.Second),
			JudgingStartedAt:  nil,
			JudgingFinishedAt: nil,
		},
	}

	analytics := buildJudgmentAnalytics(submissions, judgments)

	assert.Equal(t, int32(3), analytics.TotalSubmissions)
	assert.Equal(t, int32(3), analytics.TotalJudgments)
	assert.Len(t, analytics.StatusCounts, 3)
	if assert.NotNil(t, analytics.AvgJudgingDurationMs) {
		assert.Equal(t, 20000.0, *analytics.AvgJudgingDurationMs)
	}
	if assert.NotNil(t, analytics.AvgQueueWaitMs) {
		assert.Equal(t, 5000.0, *analytics.AvgQueueWaitMs)
	}
}

func TestAnalyticsStructs_BsonDecodingMatchesMongoFieldNames(t *testing.T) {
	base := time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC)
	started := base.Add(10 * time.Second)
	finished := base.Add(25 * time.Second)

	// Regression background:
	// The analytics helper was once extracted without BSON tags. Mongo queries
	// still returned documents, but `_id`, `submissionID`, and `judgeType`
	// silently decoded into zero values. The admin page then showed
	// totalSubmissions > 0 while totalJudgments/statusCounts/averages collapsed
	// to zero or null. This test ensures our helper structs stay aligned with
	// Mongo field names.
	submissionRaw := bson.M{
		"_id":       "sub-1",
		"createdAt": base,
	}
	judgmentRaw := bson.M{
		"submissionID":      "sub-1",
		"judgeType":         "JUDGE",
		"status":            "Judged",
		"createdAt":         base.Add(1 * time.Second),
		"judgingStartedAt":  started,
		"judgingFinishedAt": finished,
	}

	submissionBytes, err := bson.Marshal(submissionRaw)
	assert.NoError(t, err)
	judgmentBytes, err := bson.Marshal(judgmentRaw)
	assert.NoError(t, err)

	var submission analyticsSubmission
	var judgment analyticsJudgment
	assert.NoError(t, bson.Unmarshal(submissionBytes, &submission))
	assert.NoError(t, bson.Unmarshal(judgmentBytes, &judgment))

	assert.Equal(t, "sub-1", submission.ID)
	assert.Equal(t, base, submission.CreatedAt)
	assert.Equal(t, "sub-1", judgment.SubmissionID)
	assert.Equal(t, "JUDGE", judgment.JudgeType)
	assert.Equal(t, "Judged", judgment.Status)
	if assert.NotNil(t, judgment.JudgingStartedAt) {
		assert.Equal(t, started, *judgment.JudgingStartedAt)
	}
	if assert.NotNil(t, judgment.JudgingFinishedAt) {
		assert.Equal(t, finished, *judgment.JudgingFinishedAt)
	}
}
