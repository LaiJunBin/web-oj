package repository

import (
	"context"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/submission"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"judge-backend/internal/app/shared/infrastructure/mongodb"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SubmissionRepository struct {
	mdb    *mongo.Client
	logger sharedtypes.Logger
}

func NewSubmissionRepository(mdb *mongo.Client, logger sharedtypes.Logger) repository.SubmissionRepository {
	logger.Infof("[SubmissionRepository] Initializing new submission repository")
	return &SubmissionRepository{
		mdb:    mdb,
		logger: logger,
	}
}

func (r *SubmissionRepository) Save(submission *submission.Submission) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("submissions"), r.logger)

	r.logger.Infof("[SubmissionRepository] Starting to save submission %s", submission.ID())

	if submission.ID() == "" {
		submission.SetID(primitive.NewObjectID().Hex())
		r.logger.Infof("[SubmissionRepository] Generated new ID for submission: %s", submission.ID())
	}

	filter := bson.M{"_id": submission.ID()}
	payload := bson.M{
		"_id":              submission.ID(),
		"userID":           submission.UserID(),
		"contestProjectID": submission.ContestProjectID(),
		"gradingItemIds":   submission.GradingItemIds(),
		"isPartial":        submission.IsPartial(),
		"ignored":          submission.IsIgnored(),
		"createdAt":        submission.CreatedAt(),
	}

	r.logger.Debugf("[SubmissionRepository] Saving submission with payload: %+v", payload)

	if _, err := collection.UpdateOne(ctx, filter, bson.M{"$set": payload}, options.Update().SetUpsert(true)); err != nil {
		r.logger.Errorf("[SubmissionRepository] Failed to save submission %s: %v", submission.ID(), err)
		return err
	}

	r.logger.Infof("[SubmissionRepository] Successfully saved submission %s for user %s in contest project %s",
		submission.ID(),
		submission.UserID(),
		submission.ContestProjectID())
	return nil
}

func (r *SubmissionRepository) FindByID(id string) (*submission.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("submissions"), r.logger)
	filter := bson.M{"_id": id}

	var row storedSubmission
	if err := collection.FindOne(ctx, filter).Decode(&row); err != nil {
		return nil, err
	}

	sub, err := submission.New(&submission.SubmissionConfig{
		UserID:           row.UserID,
		ContestProjectID: row.ContestProjectID,
		GradingItemIds:   row.GradingItemIds,
		IsPartial:        row.IsPartial,
		Ignored:          row.Ignored,
		CreatedAt:        row.CreatedAt,
	})
	if err != nil {
		return nil, err
	}
	sub.SetID(row.ID)
	return sub, nil
}

type storedSubmission struct {
	ID               string    `bson:"_id"`
	UserID           int32     `bson:"userID"`
	ContestProjectID int32     `bson:"contestProjectID"`
	GradingItemIds   []int32   `bson:"gradingItemIds"`
	IsPartial        bool      `bson:"isPartial"`
	Ignored          bool      `bson:"ignored"`
	CreatedAt        time.Time `bson:"createdAt"`
}

func (r *SubmissionRepository) FindWithoutJudgment(olderThan time.Time) ([]*submission.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("submissions"), r.logger)

	pipeline := mongo.Pipeline{
		// The age filter sits before the $lookup on purpose: it is the grace
		// period that keeps the sweep off submissions whose asynchronous
		// judgment creation is still in flight (see the interface comment),
		// and filtering early also spares the join for every fresh row. A
		// document without createdAt counts as old enough — every production
		// write sets the field unconditionally, and a Mongo $lt would
		// otherwise exclude such a document from repair forever.
		{{Key: "$match", Value: bson.M{
			"ignored": bson.M{"$ne": true},
			"$or": bson.A{
				bson.M{"createdAt": bson.M{"$lt": olderThan}},
				bson.M{"createdAt": bson.M{"$exists": false}},
			},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from": "judgments",
			"as":   "judgments",
			"let":  bson.M{"submissionID": "$_id"},
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{"$expr": bson.M{"$eq": []interface{}{"$submissionID", "$$submissionID"}}}}},
				{{Key: "$limit", Value: 1}},
			},
		}}},
		{{Key: "$match", Value: bson.M{"judgments.0": bson.M{"$exists": false}}}},
		{{Key: "$project", Value: bson.M{
			"_id":              1,
			"userID":           1,
			"contestProjectID": 1,
			"gradingItemIds":   1,
			"isPartial":        1,
			"ignored":          1,
			"createdAt":        1,
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decoded row by row on purpose: cursor.All fails the WHOLE batch on the
	// first malformed document, and this feeds a periodic sweep — one bad
	// document would deterministically abort every future repair pass. A row
	// that does not decode is logged and skipped instead.
	var rows []*storedSubmission
	for cursor.Next(ctx) {
		var row storedSubmission
		if err := cursor.Decode(&row); err != nil {
			r.logger.Errorf("[SubmissionRepository] Skipping a submission document that does not decode: %v", err)
			continue
		}
		rows = append(rows, &row)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	result := make([]*submission.Submission, 0, len(rows))
	for _, row := range rows {
		sub, err := submission.New(&submission.SubmissionConfig{
			UserID:           row.UserID,
			ContestProjectID: row.ContestProjectID,
			GradingItemIds:   row.GradingItemIds,
			IsPartial:        row.IsPartial,
			Ignored:          row.Ignored,
			CreatedAt:        row.CreatedAt,
		})
		if err != nil {
			return nil, err
		}
		sub.SetID(row.ID)
		result = append(result, sub)
	}

	return result, nil
}
