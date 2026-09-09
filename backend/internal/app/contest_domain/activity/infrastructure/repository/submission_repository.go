package repository

import (
	"context"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/submission"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SubmissionRepository struct {
	mdb *mongo.Client
}

func NewSubmissionRepository(mdb *mongo.Client) repository.SubmissionRepository {
	return &SubmissionRepository{
		mdb: mdb,
	}
}

type Submission struct {
	ID               string    `bson:"_id"`
	UserID           int32     `bson:"userID"`
	ContestProjectID int32     `bson:"contestProjectID"`
	GradingItemIds   []int32   `bson:"gradingItemIds"`
	IsPartial        bool      `bson:"isPartial"`
	Ignored          bool      `bson:"ignored"`
	CreatedAt        time.Time `bson:"createdAt"`
}

func (r *SubmissionRepository) FindByID(id string) (*submission.Submission, error) {
	ctx := context.Background()
	collection := r.mdb.Database("judge").Collection("submissions")
	filter := bson.M{"_id": id}
	project := bson.M{"_id": 1, "userID": 1, "contestProjectID": 1, "gradingItemIds": 1, "isPartial": 1, "ignored": 1, "createdAt": 1}
	var result Submission
	if err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(project)).Decode(&result); err != nil {
		return nil, err
	}

	s := submission.New(&submission.SubmissionConfig{
		ID:               result.ID,
		UserID:           result.UserID,
		ContestProjectID: result.ContestProjectID,
		GradingItemIds:   result.GradingItemIds,
		IsPartial:        result.IsPartial,
		Ignored:          result.Ignored,
		CreatedAt:        result.CreatedAt,
	})

	return s, nil
}

func (r *SubmissionRepository) FindByContestProjectIDAndGradingItemIDs(contestProjectID int32, gradingItemIDs []int32) ([]*submission.Submission, error) {
	ctx := context.Background()
	collection := r.mdb.Database("judge").Collection("submissions")

	var pipeline mongo.Pipeline
	if len(gradingItemIDs) == 0 {
		pipeline = mongo.Pipeline{
			{{
				Key: "$match",
				Value: bson.M{
					"$and": []interface{}{
						bson.M{"contestProjectID": contestProjectID},
					},
				},
			}},
		}
	} else {
		pipeline = mongo.Pipeline{
			{{
				Key: "$match",
				Value: bson.M{
					"$and": []interface{}{
						bson.M{"contestProjectID": contestProjectID},
						bson.M{
							"$or": []interface{}{
								bson.M{"isPartial": false},
								bson.M{
									"gradingItemIds": bson.M{"$in": gradingItemIDs},
								},
							},
						},
					},
				},
			}},
		}
	}

	pipeline = append(pipeline, bson.D{{
		Key: "$project",
		Value: bson.M{
			"_id":              1,
			"userID":           1,
			"contestProjectID": 1,
			"gradingItemIds":   1,
			"isPartial":        1,
			"ignored":          1,
			"createdAt":        1,
		},
	}})

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	var submissions []*submission.Submission
	for cursor.Next(ctx) {
		var result Submission
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}

		s := submission.New(&submission.SubmissionConfig{
			ID:               result.ID,
			UserID:           result.UserID,
			ContestProjectID: result.ContestProjectID,
			IsPartial:        result.IsPartial,
			CreatedAt:        result.CreatedAt,
		})
		submissions = append(submissions, s)
	}

	return submissions, nil
}

func (r *SubmissionRepository) Save(submission *submission.Submission) error {
	ctx := context.Background()
	collection := r.mdb.Database("judge").Collection("submissions")

	if submission.ID() == "" {
		submission.SetID(primitive.NewObjectID().Hex())
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

	if _, err := collection.UpdateOne(ctx, filter, bson.M{"$set": payload}, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}
