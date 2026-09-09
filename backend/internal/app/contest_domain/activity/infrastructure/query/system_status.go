package query

import (
	"context"
	"fmt"
	"time"

	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// errorStatus is the judgment status meaning the run itself failed. Spelled out
// here for the same reason judgedStatus is (see rejudge.go): the judge domain
// owns that entity, and reaching into it from this bounded context is the
// coupling this codebase already decided against.
const errorStatus = "Error"

// minuteBucketFormat buckets a timestamp to the minute, in UTC.
//
// $dateToString rather than the far more direct $dateTrunc: the deployed
// MongoDB is 4.4, and $dateTrunc did not arrive until 5.0. The same version
// ceiling is why this codebase avoids multi-document transactions — see
// CommitJudgmentsByRejudgeID.
const minuteBucketFormat = "2006-01-02T15:04:05Z"

type SystemStatusQuery struct {
	mdb *mongo.Client
	// now is injected so the window boundaries are testable without waiting on
	// the wall clock.
	now func() time.Time
}

func NewSystemStatusQuery(
	mdb *mongo.Client,
) query.SystemStatusQuery {
	return &SystemStatusQuery{
		mdb: mdb,
		now: time.Now,
	}
}

type submissionRateBucket struct {
	Minute string `bson:"_id"`
	Count  int32  `bson:"count"`
}

func (q *SystemStatusQuery) GetSubmissionRate(minutes int32, contestProjectID *int32) ([]*dto.SubmissionRatePoint, error) {
	ctx := context.Background()
	collection := q.mdb.Database("judge").Collection("submissions")

	// The window ends at the current, still-filling minute and spans `minutes`
	// buckets back from there, so the newest bucket is always the one being
	// written to right now.
	end := q.now().UTC().Truncate(time.Minute)
	start := end.Add(-time.Duration(minutes-1) * time.Minute)

	match := bson.M{"createdAt": bson.M{"$gte": start}}
	if contestProjectID != nil {
		match["contestProjectID"] = *contestProjectID
	}

	// Partial and ignored submissions are counted like any other: this is a
	// measure of incoming load, and both of those were still received, parsed
	// and queued. The analytics endpoint excludes them because it measures
	// scores, which is a different question.
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{
			{
				Key: "$group",
				Value: bson.M{
					"_id": bson.M{
						"$dateToString": bson.M{
							"format":   "%Y-%m-%dT%H:%M:00Z",
							"date":     "$createdAt",
							"timezone": "UTC",
						},
					},
					"count": bson.M{"$sum": 1},
				},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var buckets []submissionRateBucket
	if err := cursor.All(ctx, &buckets); err != nil {
		return nil, err
	}

	counts := make(map[time.Time]int32, len(buckets))
	for _, bucket := range buckets {
		minute, err := time.Parse(minuteBucketFormat, bucket.Minute)
		if err != nil {
			return nil, fmt.Errorf("failed to parse submission rate bucket %q: %w", bucket.Minute, err)
		}
		counts[minute.UTC()] = bucket.Count
	}

	// Every minute in the window is emitted, including the empty ones. A caller
	// plotting this cannot tell a missing bucket from a zero one, and a series
	// with holes renders as a line that skips over the quiet periods rather
	// than dropping to zero across them.
	points := make([]*dto.SubmissionRatePoint, 0, minutes)
	for minute := start; !minute.After(end); minute = minute.Add(time.Minute) {
		points = append(points, &dto.SubmissionRatePoint{
			Minute: minute,
			Count:  counts[minute],
		})
	}

	return points, nil
}

type judgmentErrorRow struct {
	ID               string    `bson:"_id"`
	SubmissionID     string    `bson:"submissionID"`
	ContestProjectID int32     `bson:"contestProjectID"`
	ErrorMessage     *string   `bson:"errorMessage"`
	CreatedAt        time.Time `bson:"createdAt"`
}

func (q *SystemStatusQuery) ListJudgmentErrors(limit int32) ([]*dto.JudgmentError, error) {
	ctx := context.Background()
	collection := q.mdb.Database("judge").Collection("judgments")

	cursor, err := collection.Find(ctx,
		bson.M{"status": errorStatus},
		options.Find().
			SetSort(bson.D{{Key: "createdAt", Value: -1}}).
			SetLimit(int64(limit)).
			SetProjection(bson.M{
				"_id":              1,
				"submissionID":     1,
				"contestProjectID": 1,
				"errorMessage":     1,
				"createdAt":        1,
			}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	rows := []judgmentErrorRow{}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	errors := make([]*dto.JudgmentError, 0, len(rows))
	for _, row := range rows {
		errors = append(errors, &dto.JudgmentError{
			JudgmentID:       row.ID,
			SubmissionID:     row.SubmissionID,
			ContestProjectID: row.ContestProjectID,
			ErrorMessage:     row.ErrorMessage,
			CreatedAt:        row.CreatedAt,
		})
	}

	return errors, nil
}
