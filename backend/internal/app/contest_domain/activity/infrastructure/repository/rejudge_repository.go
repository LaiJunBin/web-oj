package repository

import (
	"context"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/rejudge"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RejudgeRepository struct {
	mdb *mongo.Client
}

func NewRejudgeRepository(
	mdb *mongo.Client,
) repository.RejudgeRepository {
	return &RejudgeRepository{
		mdb: mdb,
	}
}

type Rejudge struct {
	ID        string   `bson:"_id"`
	Status    string   `bson:"status"`
	Reason    string   `bson:"reason"`
	Judgments []string `bson:"judgments"`
}

func (r *RejudgeRepository) FindByID(id string) (*rejudge.Rejudge, error) {
	ctx := context.Background()
	collection := r.mdb.Database("judge").Collection("rejudges")

	filter := bson.M{"_id": id}
	project := bson.M{"_id": 1, "status": 1, "reason": 1, "judgments": 1}
	var result Rejudge
	if err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(project)).Decode(&result); err != nil {
		return nil, err
	}

	return rejudge.New(&rejudge.RejudgeConfig{
		ID:        result.ID,
		Reason:    result.Reason,
		Status:    result.Status,
		Judgments: result.Judgments,
	})
}

func (r *RejudgeRepository) FindAwaitingDecisionBySubmissionIDs(submissionIDs []string) ([]*rejudge.Rejudge, error) {
	if len(submissionIDs) == 0 {
		return []*rejudge.Rejudge{}, nil
	}

	ctx := context.Background()

	// A rejudge does not record which submissions it covers; its judgments do, via
	// rejudgeID. So the link is walked from the judgment side.
	judgments := r.mdb.Database("judge").Collection("judgments")
	cursor, err := judgments.Find(ctx,
		bson.M{
			"submissionID": bson.M{"$in": submissionIDs},
			"rejudgeID":    bson.M{"$nin": []interface{}{nil, ""}},
		},
		options.Find().SetProjection(bson.M{"rejudgeID": 1}),
	)
	if err != nil {
		return nil, err
	}
	var owners []struct {
		RejudgeID string `bson:"rejudgeID"`
	}
	if err := cursor.All(ctx, &owners); err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	rejudgeIDs := []string{}
	for _, owner := range owners {
		if _, dup := seen[owner.RejudgeID]; dup {
			continue
		}
		seen[owner.RejudgeID] = struct{}{}
		rejudgeIDs = append(rejudgeIDs, owner.RejudgeID)
	}
	if len(rejudgeIDs) == 0 {
		return []*rejudge.Rejudge{}, nil
	}

	collection := r.mdb.Database("judge").Collection("rejudges")
	cursor, err = collection.Find(ctx,
		bson.M{
			"_id":    bson.M{"$in": rejudgeIDs},
			"status": bson.M{"$in": rejudge.AwaitingDecisionStatuses()},
		},
		options.Find().SetProjection(bson.M{"_id": 1, "status": 1, "reason": 1, "judgments": 1}),
	)
	if err != nil {
		return nil, err
	}
	var rows []Rejudge
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	pending := make([]*rejudge.Rejudge, 0, len(rows))
	for _, row := range rows {
		entity, err := rejudge.New(&rejudge.RejudgeConfig{
			ID:        row.ID,
			Reason:    row.Reason,
			Status:    row.Status,
			Judgments: row.Judgments,
		})
		if err != nil {
			return nil, err
		}
		pending = append(pending, entity)
	}

	return pending, nil
}
