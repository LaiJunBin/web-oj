package repository

import (
	"context"
	"judge-backend/internal/app/contest_domain/contest/domain/entity/submission"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SubmissionRepository struct {
	mdb    *mongo.Client
	logger sharedtypes.Logger
}

func NewSubmissionRepository(mdb *mongo.Client, logger sharedtypes.Logger) repository.SubmissionRepository {
	return &SubmissionRepository{
		mdb:    mdb,
		logger: logger,
	}
}

type Submission struct {
	ID         string `bson:"_id"`
	UserID     int32  `bson:"userID"`
	JudgmentID string `bson:"judgment_id"`
	Status     string `bson:"status"`
}

func (r *SubmissionRepository) FindByID(id string) (*submission.Submission, error) {
	r.logger.Debugf("Starting to find submission by ID", "id", id)

	ctx := context.Background()
	collection := r.mdb.Database("judge").Collection("submissions")
	pipeline := mongo.Pipeline{
		{
			{
				Key: "$match",
				Value: bson.M{
					"_id": id,
				},
			},
		},
		{
			{
				Key: "$lookup",
				Value: bson.M{
					"from": "judgments",
					"as":   "judgments",
					"let":  bson.M{"submissionID": "$_id"},
					"pipeline": mongo.Pipeline{
						{
							{
								Key: "$match",
								Value: bson.M{
									"$expr": bson.M{
										"$and": []interface{}{
											bson.M{"$eq": []interface{}{"$submissionID", "$$submissionID"}},
											bson.M{"$or": []interface{}{
												bson.M{"$ne": []interface{}{"$judgeType", "REJUDGE"}},
												bson.M{"$eq": []interface{}{"$status", "Judged"}},
											}},
										},
									},
								},
							},
						},
						{
							{
								Key: "$sort",
								Value: bson.M{
									"createdAt": -1,
								},
							},
						},
						{
							{
								Key:   "$limit",
								Value: 1,
							},
						},
					},
				},
			},
		},
		{
			{
				Key: "$set",
				Value: bson.M{
					"judgment": bson.M{"$arrayElemAt": []interface{}{"$judgments", 0}},
				},
			},
		},
		{
			{
				Key: "$project",
				Value: bson.M{
					"_id":         1,
					"userID":      1,
					"judgment_id": "$judgment._id",
					"status":      "$judgment.status",
				},
			},
		},
		{
			{
				Key:   "$limit",
				Value: 1,
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Errorf("Failed to execute aggregation pipeline", "id", id, "error", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var result *Submission
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			r.logger.Errorf("Failed to decode submission result", "id", id, "error", err)
			return nil, err
		}
	} else {
		r.logger.Debugf("Submission not found", "id", id)
		return nil, httperror.New(http.StatusNotFound, "submission not found")
	}

	submission, err := submission.NewSubmission(&submission.SubmissionConfig{
		ID:         result.ID,
		UserID:     result.UserID,
		JudgmentID: result.JudgmentID,
		Status:     result.Status,
	})
	if err != nil {
		r.logger.Errorf("Failed to create submission domain model", "id", id, "error", err)
		return nil, err
	}

	r.logger.Debugf("Successfully found submission", "id", id)
	return submission, nil
}
