package repository

import (
	"context"
	"fmt"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem/scoringstrategy"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase/resultparser"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"judge-backend/internal/app/shared/infrastructure/mongodb"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type GradingItemRepository struct {
	db     *gorm.DB
	mdb    *mongo.Client
	logger sharedtypes.Logger
}

func NewGradingItemRepository(
	db *gorm.DB,
	mdb *mongo.Client,
	logger sharedtypes.Logger,
) repository.GradingItemRepository {
	return &GradingItemRepository{
		db:     db,
		mdb:    mdb,
		logger: logger,
	}
}

func resolveParser(giResultParser, projectResultParser string, logger sharedtypes.Logger) resultparser.ResultParser {
	key := giResultParser
	if key == "" {
		key = projectResultParser
	}
	if p, ok := resultparser.Parsers[key]; ok {
		return p
	}
	logger.Warnf("[GradingItemRepository] Unknown result parser %q, falling back to AutoParser", key)
	return resultparser.AutoParser
}

func (r *GradingItemRepository) toDomainModel(gs []*model.GradingItem, projectResultParser string) []*gradingitem.GradingItem {
	r.logger.Infof("[GradingItemRepository] Converting %d grading items to domain model", len(gs))
	gradingItems := []*gradingitem.GradingItem{}
	for _, g := range gs {
		dependsIDs := []int32{}
		for _, depend := range g.Depends {
			dependsIDs = append(dependsIDs, depend.ID)
		}

		parser := resolveParser(g.ResultParser, projectResultParser, r.logger)

		testcases := []*testcase.Testcase{}
		for _, tc := range g.Testcases {
			dependIDs := []int32{}
			for _, depend := range tc.Depends {
				dependIDs = append(dependIDs, depend.ID)
			}
			testcases = append(testcases, testcase.NewTestcase(&testcase.TestcaseConfig{
				ID:              tc.ID,
				Name:            tc.Name,
				Command:         tc.Command,
				Score:           tc.Score,
				DependTestcases: dependIDs,
			}))
		}

		scoreStrategy := scoringstrategy.DecreaseScoreStrategy
		switch g.ScoringStrategy {
		case "increase":
			scoreStrategy = scoringstrategy.IncreaseScoreStrategy
		case "decrease":
			scoreStrategy = scoringstrategy.DecreaseScoreStrategy
		default:
			r.logger.Warnf("[GradingItemRepository] Unknown scoring strategy '%s' for grading item %d, using decrease strategy", g.ScoringStrategy, g.ID)
		}

		gradingItems = append(gradingItems, gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
			ID:              g.ID,
			Label:           g.Label,
			Name:            g.Name,
			Score:           g.Score,
			ScoringStrategy: scoreStrategy,
			DependTestcases: dependsIDs,
			Testcases:       testcases,
			ResultParser:    parser,
		}))
	}

	r.logger.Infof("[GradingItemRepository] Successfully converted %d grading items to domain model", len(gradingItems))
	return gradingItems
}

func (r *GradingItemRepository) FindBySubmissionID(submissionID string) ([]*gradingitem.GradingItem, error) {
	r.logger.Infof("[GradingItemRepository] Finding grading items for submission %s", submissionID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("submissions"), r.logger)
	filter := bson.M{"_id": submissionID}
	projection := bson.M{"contestProjectID": 1, "gradingItemIds": 1}
	var result bson.M
	if err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result); err != nil {
		r.logger.Errorf("[GradingItemRepository] Failed to find submission %s: %v", submissionID, err)
		return nil, err
	}

	// A malformed document must surface as an error: this runs on async
	// handler goroutines, where a panic takes down the whole process.
	rawIDs, ok := result["gradingItemIds"].(primitive.A)
	if !ok {
		return nil, fmt.Errorf("submission %s: gradingItemIds is not an array", submissionID)
	}
	gradingItemIDs := []int32{}
	seen := map[int32]bool{}
	for _, rawID := range rawIDs {
		id, ok := rawID.(int32)
		if !ok {
			return nil, fmt.Errorf("submission %s: grading item id %v is not an int32", submissionID, rawID)
		}
		if !seen[id] {
			seen[id] = true
			gradingItemIDs = append(gradingItemIDs, id)
		}
	}
	cpid, ok := result["contestProjectID"].(int32)
	if !ok {
		return nil, fmt.Errorf("submission %s: contestProjectID is not an int32", submissionID)
	}

	// The contest project is the authority on WHICH problem's grading items
	// this submission may name; everything below is scoped to cp.ProjectID.
	cp := model.ContestProject{}
	if err := r.db.Where("id = ?", cpid).First(&cp).Error; err != nil {
		r.logger.Errorf("[GradingItemRepository] Failed to find contest project %d: %v", cpid, err)
		return nil, err
	}

	gradingItems := []*model.GradingItem{}
	if len(gradingItemIDs) == 0 {
		r.logger.Infof("[GradingItemRepository] No grading item IDs found for submission %s, fetching from contest project", submissionID)
		if err := r.db.Preload("Testcases", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("Depends")
		}).Preload("Depends").Where("project_id = ?", cp.ProjectID).Find(&gradingItems).Error; err != nil {
			r.logger.Errorf("[GradingItemRepository] Failed to find grading items for project %d: %v", cp.ProjectID, err)
			return nil, err
		}
	} else {
		r.logger.Infof("[GradingItemRepository] Found %d grading item IDs for submission %s", len(gradingItemIDs), submissionID)
		if err := r.db.Preload("Testcases", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("Depends")
		}).Preload("Depends").Where("project_id = ? AND id IN ?", cp.ProjectID, gradingItemIDs).Find(&gradingItems).Error; err != nil {
			r.logger.Errorf("[GradingItemRepository] Failed to find grading items by IDs: %v", err)
			return nil, err
		}

		// Ids that did not resolve are tolerable ONLY when they are this
		// problem's own soft-deleted items — an admin removed them after the
		// submission was made, and skipping them is exactly what the
		// unscoped-by-default query always did. Anything else (another
		// problem's item, an id that never existed) is a set this submission
		// was never allowed to name: fail the load rather than judge it.
		// Filtering would be worse than failing — a grading item with no
		// testcases scores as PASSED by design, so a filtered-down set could
		// award points for items that were never run.
		if len(gradingItems) != len(gradingItemIDs) {
			loaded := map[int32]bool{}
			for _, g := range gradingItems {
				loaded[g.ID] = true
			}
			missing := []int32{}
			for _, id := range gradingItemIDs {
				if !loaded[id] {
					missing = append(missing, id)
				}
			}
			var deletedOwn []int32
			if err := r.db.Unscoped().Model(&model.GradingItem{}).
				Where("project_id = ? AND id IN ?", cp.ProjectID, missing).
				Pluck("id", &deletedOwn).Error; err != nil {
				r.logger.Errorf("[GradingItemRepository] Failed to resolve missing grading items for submission %s: %v", submissionID, err)
				return nil, err
			}
			if len(deletedOwn) != len(missing) {
				r.logger.Errorf("[GradingItemRepository] Submission %s names grading items outside project %d: requested %v", submissionID, cp.ProjectID, gradingItemIDs)
				return nil, fmt.Errorf("submission %s: %w", submissionID, repository.ErrGradingItemsNotOwned)
			}
		}

		// The WHERE above already guarantees this; the check is a zero-cost
		// invariant so a future query change cannot quietly widen the scope.
		for _, g := range gradingItems {
			if g.ProjectID != cp.ProjectID {
				return nil, fmt.Errorf("submission %s: %w", submissionID, repository.ErrGradingItemsNotOwned)
			}
		}
	}

	// The result parser comes from the resolved problem, never from whichever
	// grading item happened to load first.
	var projectRecord model.Project
	if err := r.db.Select("result_parser").Where("id = ?", cp.ProjectID).First(&projectRecord).Error; err != nil {
		r.logger.Warnf("[GradingItemRepository] Failed to fetch project result_parser for project %d: %v", cp.ProjectID, err)
	}

	return r.toDomainModel(gradingItems, projectRecord.ResultParser), nil
}
