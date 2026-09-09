package judge_test

import (
	"context"

	"judge-backend/internal/config"
	"judge-backend/tests/testutils"

	"github.com/cucumber/godog"
)

type judgeScenarioInitializer struct{}

func NewJudgeScenarioInitializer() testutils.ScenarioInitializer {
	return &judgeScenarioInitializer{}
}

func (s *judgeScenarioInitializer) InitializeScenario(ctx *testutils.Context) {
	api := ctx.API
	db := ctx.DB
	cfg := config.NewFeatureConfig()
	judge := newJudgeFeature(ctx.MDB, cfg)

	ctx.Context.Before(func(gCtx context.Context, sc *godog.Scenario) (context.Context, error) {
		if err := testutils.ResetJudgeState(cfg); err != nil {
			return gCtx, err
		}

		testutils.ResetDB(db)
		testutils.RunMigration(db)
		testutils.ResetRDB(ctx.RDB)
		testutils.ResetMongoDB()
		api.ResetResponse(sc)
		return gCtx, nil
	})

	// Register default steps (users, projects, grading items, etc.)
	testutils.InitContextSteps(ctx)

	// Register judge-specific steps
	initJudgeSteps(ctx, judge)
}
