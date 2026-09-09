package repository

import (
	"context"
	"errors"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/judgment"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JudgmentRepository struct {
	mdb *mongo.Client
}

func NewJudgmentRepository(
	mdb *mongo.Client,
) repository.JudgmentRepository {
	return &JudgmentRepository{
		mdb: mdb,
	}
}

type Judgment struct {
	ID           string `bson:"_id"`
	SubmissionID string `bson:"submissionID"`
	Status       string `bson:"status"`
}

func (r *JudgmentRepository) FindByID(id string) (*judgment.Judgment, error) {
	ctx := context.Background()
	collection := r.mdb.Database("judge").Collection("judgments")

	filter := bson.M{"_id": id}
	project := bson.M{"_id": 1, "submissionID": 1, "status": 1}
	var result Judgment
	if err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(project)).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrJudgmentNotFound
		}
		return nil, err
	}

	return judgment.New(result.ID, result.SubmissionID, result.Status), nil
}
