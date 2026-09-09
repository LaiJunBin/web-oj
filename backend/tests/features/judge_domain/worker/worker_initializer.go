package worker_test

import (
	"context"

	"judge-backend/internal/config"
	"judge-backend/tests/testutils"

	"github.com/cucumber/godog"
)

type workerScenarioInitializer struct{}

func NewWorkerScenarioInitializer() testutils.ScenarioInitializer {
	return &workerScenarioInitializer{}
}

func (s *workerScenarioInitializer) InitializeScenario(ctx *testutils.Context) {
	api := ctx.API
	db := ctx.DB
	cfg := config.NewFeatureConfig()

	// The judge reset runs BEFORE the database resets: a judgment left over
	// from the previous scenario would otherwise be picked up mid-reset and
	// write ghost documents into the freshly cleaned MongoDB. It is also what
	// restores workers a queue-depth scenario stopped.
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

	ctx.Context.After(func(gCtx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		api.CloseStreamForScenario(sc.Id)
		return gCtx, err
	})

	testutils.InitContextSteps(ctx)
}
