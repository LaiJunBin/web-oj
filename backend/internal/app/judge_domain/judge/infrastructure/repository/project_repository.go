package repository

import (
	"context"
	"time"
	"errors"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"judge-backend/internal/app/shared/infrastructure/mongodb"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db     *gorm.DB
	mdb    *mongo.Client
	logger sharedtypes.Logger
}

func NewProjectRepository(
	db *gorm.DB,
	mdb *mongo.Client,
	logger sharedtypes.Logger,
) repository.ProjectRepository {
	return &ProjectRepository{
		db:     db,
		mdb:    mdb,
		logger: logger,
	}
}

func (r *ProjectRepository) FindBySubmissionID(submissionID string) (*project.Project, error) {
	r.logger.Infof("[ProjectRepository] Finding project for submission %s", submissionID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("submissions"), r.logger)
	filter := bson.M{"_id": submissionID}
	projection := bson.M{"contestProjectID": 1}
	var result bson.M
	if err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result); err != nil {
		r.logger.Errorf("[ProjectRepository] Failed to find submission %s: %v", submissionID, err)
		return nil, err
	}

	cpid := result["contestProjectID"].(int32)
	cp := &model.ContestProject{}
	if err := r.db.Preload("Project", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Image")
	}).Where("id = ?", cpid).First(cp).Error; err != nil {
		r.logger.Errorf("[ProjectRepository] Failed to find contest project %d: %v", cpid, err)
		return nil, err
	}

	if cp.Project.ImageTag == nil || cp.Project.TestImageTag == nil {
		r.logger.Errorf("[ProjectRepository] Project %d is not ready: missing image tags", cp.ID)
		return nil, errors.New("project not ready")
	}

	proj := project.NewProject(&project.ProjectConfig{
		ID:           cp.Project.ID,
		ImageTag:     *cp.Project.ImageTag,
		TestImageTag: *cp.Project.TestImageTag,
		Env:          cp.Project.Env,
		TestEnv:      cp.Project.TestEnv,
		MountPath:    cp.Project.Image.MountPath,
		ResultParser: cp.Project.ResultParser,
		// Read off the contest project, not the project: a cap has to match
		// the machine the contest runs on, which the same problem reused at a
		// different venue does not share.
		PidsLimit:     cp.PidsLimit,
		MemoryLimitMB: cp.MemoryLimitMB,
	})

	r.logger.Infof("[ProjectRepository] Successfully found project for submission %s", submissionID)
	return proj, nil
}
