package query

import (
	"context"
	"judge-backend/internal/app/judge_domain/judge/domain/query"
	sharedtypes "judge-backend/internal/app/shared/types"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JudgmentQuery struct {
	mdb    *mongo.Client
	logger sharedtypes.Logger
}

func NewJudgmentQuery(
	mdb *mongo.Client,
	logger sharedtypes.Logger,
) query.JudgmentQuery {
	logger.Infof("[JudgmentQuery] Initializing new judgment query")
	return &JudgmentQuery{
		mdb:    mdb,
		logger: logger,
	}
}

func (q *JudgmentQuery) GetPrevSubmissionTimeByJudgmentID(id string) (time.Time, error) {
	q.logger.Infof("[JudgmentQuery] Getting previous submission time for judgment ID: %s", id)
	ctx := context.Background()
	collection := q.mdb.Database("judge").Collection("judgments")
	var result struct {
		SubmissionID string `bson:"submissionID"`
	}

	if err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result); err != nil {
		q.logger.Errorf("[JudgmentQuery] Failed to find judgment with ID %s: %v", id, err)
		return time.Time{}, err
	}

	q.logger.Debugf("[JudgmentQuery] Found submission ID %s for judgment ID %s", result.SubmissionID, id)
	submissionCollection := q.mdb.Database("judge").Collection("submissions")
	var currentSubmissionResult struct {
		UserID           int32     `bson:"userID"`
		ContestProjectID int32     `bson:"contestProjectID"`
		CreatedAt        time.Time `bson:"createdAt"`
	}

	if err := submissionCollection.FindOne(ctx, bson.M{"_id": result.SubmissionID}).Decode(&currentSubmissionResult); err != nil {
		q.logger.Errorf("[JudgmentQuery] Failed to find submission with ID %s: %v", result.SubmissionID, err)
		return time.Time{}, err
	}

	q.logger.Debugf("[JudgmentQuery] Found current submission for user ID %d, contest project ID %d",
		currentSubmissionResult.UserID, currentSubmissionResult.ContestProjectID)

	var prevSubmissionResult struct {
		CreatedAt time.Time `bson:"createdAt"`
	}

	filter := bson.M{
		"userID":           currentSubmissionResult.UserID,
		"contestProjectID": currentSubmissionResult.ContestProjectID,
		"createdAt": bson.M{
			"$lt": currentSubmissionResult.CreatedAt,
		},
		"ignored": false,
	}

	opts := options.FindOne().SetSort(bson.M{"createdAt": -1})
	if err := submissionCollection.FindOne(ctx, filter, opts).Decode(&prevSubmissionResult); err != nil {
		if err == mongo.ErrNoDocuments {
			q.logger.Infof("[JudgmentQuery] No previous submission found for judgment ID: %s", id)
			return time.Time{}, nil
		}
		q.logger.Errorf("[JudgmentQuery] Failed to find previous submission for user ID %d, contest project ID %d: %v",
			currentSubmissionResult.UserID, currentSubmissionResult.ContestProjectID, err)
		return time.Time{}, err
	}

	q.logger.Infof("[JudgmentQuery] Successfully found previous submission time: %v for judgment ID: %s",
		prevSubmissionResult.CreatedAt, id)
	return prevSubmissionResult.CreatedAt, nil
}
