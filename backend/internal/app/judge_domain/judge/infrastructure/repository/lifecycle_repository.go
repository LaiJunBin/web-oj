package repository

import (
	"context"
	"time"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"judge-backend/internal/app/shared/infrastructure/mongodb"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type LifecycleRepository struct {
	db     *gorm.DB
	mdb    *mongo.Client
	logger sharedtypes.Logger
}

func NewLifecycleRepository(
	db *gorm.DB,
	mdb *mongo.Client,
	logger sharedtypes.Logger,
) repository.LifecycleRepository {
	return &LifecycleRepository{
		db:     db,
		mdb:    mdb,
		logger: logger,
	}
}

func (r *LifecycleRepository) toDomainModel(ls []*model.Lifecycle) []*lifecycle.Lifecycle {
	r.logger.Infof("[LifecycleRepository] Converting %d lifecycles to domain model", len(ls))
	Lifecycles := []*lifecycle.Lifecycle{}
	for _, l := range ls {
		Lifecycles = append(Lifecycles, lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
			ID:                   l.ID,
			ExecutionEnvironment: l.RunsOn,
			Type:                 l.Type,
			Command:              l.Command,
		}))
	}

	r.logger.Infof("[LifecycleRepository] Successfully converted %d lifecycles to domain model", len(Lifecycles))
	return Lifecycles
}

func (r *LifecycleRepository) FindBySubmissionID(submissionID string) ([]*lifecycle.Lifecycle, error) {
	r.logger.Infof("[LifecycleRepository] Finding lifecycles for submission %s", submissionID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("submissions"), r.logger)
	filter := bson.M{"_id": submissionID}
	projection := bson.M{"contestProjectID": 1}
	var result bson.M
	if err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result); err != nil {
		r.logger.Errorf("[LifecycleRepository] Failed to find submission %s: %v", submissionID, err)
		return nil, err
	}

	cpid := result["contestProjectID"].(int32)
	cp := model.ContestProject{}
	if err := r.db.Where("id = ?", cpid).First(&cp).Error; err != nil {
		r.logger.Errorf("[LifecycleRepository] Failed to find contest project %d: %v", cpid, err)
		return nil, err
	}

	Lifecycles := []*model.Lifecycle{}
	if err := r.db.Where("project_id = ?", cp.ProjectID).Find(&Lifecycles).Error; err != nil {
		r.logger.Errorf("[LifecycleRepository] Failed to find lifecycles for project %d: %v", cp.ProjectID, err)
		return nil, err
	}

	r.logger.Infof("[LifecycleRepository] Found %d lifecycles for submission %s", len(Lifecycles), submissionID)
	return r.toDomainModel(Lifecycles), nil
}
