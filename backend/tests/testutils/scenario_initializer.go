package testutils

import (
	"context"
	"fmt"

	"judge-backend/internal/config"

	"github.com/cucumber/godog"
)

func InitContextSteps(ctx *Context) {
	api := ctx.API

	ctx.Context.Step(`^I have users:$`, api.IHaveUsers)
	ctx.Context.Step(`^I have images:$`, api.IHaveImages)
	ctx.Context.Step(`^I have builds:$`, api.IHaveBuilds)
	ctx.Context.Step(`^I have build logs:$`, api.IHaveBuildLogs)
	ctx.Context.Step(`^I have projects:$`, api.IHaveProjects)
	ctx.Context.Step(`^I have dynamic fields:$`, api.IHaveDynamicFields)
	ctx.Context.Step(`^I have grading-items:$`, api.IHaveGradingItems)
	ctx.Context.Step(`^I have grading-item depends:$`, api.IHaveGradingItemDepends)
	ctx.Context.Step(`^I have testcases:$`, api.IHaveTestcases)
	ctx.Context.Step(`^I have testcase depends:$`, api.IHaveTestcaseDepends)
	ctx.Context.Step(`^I have assets:$`, api.IHaveAssets)
	ctx.Context.Step(`^I have lifecycles:$`, api.IHaveProjectsLifecycles)
	ctx.Context.Step(`^I have contests:$`, api.IHaveContests)
	ctx.Context.Step(`^I have contest project associations:$`, api.IHaveContestProjectAssociations)
	ctx.Context.Step(`^I have groups:$`, api.IHaveGroups)
	ctx.Context.Step(`^I have user groups:$`, api.IHaveUserGroups)
	ctx.Context.Step(`^I have contest participants:$`, api.IHaveContestParticipants)
	ctx.Context.Step(`^I have clarifications:$`, api.IHaveClarifications)
	ctx.Context.Step(`^I have refresh tokens for the following users:$`, api.IHaveRefreshTokensForTheFollowingUsers)
	ctx.Context.Step(`^I have submissions:$`, api.IHaveSubmissions)
	ctx.Context.Step(`^I have judgments:$`, api.IHaveJudgments)
	ctx.Context.Step(`^I have rejudges:$`, api.IHaveRejudges)
	ctx.Context.Step(`^a rejudge for contest project (\d+) should exist within (\d+) seconds$`, api.ARejudgeForContestProjectShouldExistWithin)
	ctx.Context.Step(`^all judge workers are stopped$`, api.AllJudgeWorkersAreStopped)
	ctx.Context.Step(`^all judge workers are started$`, api.AllJudgeWorkersAreStarted)
	ctx.Context.Step(`^the "([^"]*)" judge queue depth should reach (\d+) within (\d+) seconds$`, api.TheJudgeQueueDepthShouldReachWithin)
	ctx.Context.Step(`^the judge queue should drain within (\d+) seconds$`, api.TheJudgeQueueShouldDrainWithin)
	ctx.Context.Step(`^the submission rate over (\d+) minutes should have (\d+) points totalling (\d+)$`, api.TheSubmissionRateShouldHavePointsTotalling)
	ctx.Context.Step(`^the submission rate over (\d+) minutes for contest project (\d+) should have (\d+) points totalling (\d+)$`, api.TheSubmissionRateForContestProjectShouldHavePointsTotalling)
	ctx.Context.Step(`^the streamed submission rate should total (\d+) within (\d+) seconds$`, api.TheStreamedSubmissionRateShouldTotalWithin)
	ctx.Context.Step(`^I have no images$`, api.IHaveNoImages)
	ctx.Context.Step(`^I am logged in as "([^"]*)"$`, api.IAmLoggedInAs)
	ctx.Context.Step(`^I am logged in as "([^"]*)" with an expired access token$`, api.IAmLoggedInAsWithAnExpiredAccessToken)
	ctx.Context.Step(`^I log in via the API as "([^"]*)" with password "([^"]*)"$`, api.ILogInViaTheAPI)
	ctx.Context.Step(`^I wait for (\d+) seconds$`, api.IWaitForSeconds)
	ctx.Context.Step(`^setup "([^"]*)" project into "([^"]*)" image$`, api.SetupProjectIntoImage)

	ctx.Context.Step(`^I send a "(GET|POST|PUT|PATCH|DELETE)" request to "([^"]*)":$`, api.ISendrequestTo)
	ctx.Context.Step(`^I send a "(GET|POST|PUT|PATCH|DELETE)" request to "([^"]*)" with the following data:$`, api.ISendrequestToWithTheFollowingData)
	ctx.Context.Step(`^I send a "(GET|POST|PUT|PATCH|DELETE)" request to "([^"]*)" with the following cookies:$`, api.ISendrequestToWithTheFollowingCookies)
	ctx.Context.Step(`^I send a "(GET|POST|PUT|PATCH|DELETE)" request to "([^"]*)" with the following formdata:$`, api.ISendrequestToWithTheFollowingFormData)

	ctx.Context.Step(`^the response code should be (\d+)$`, api.TheResponseCodeShouldBe)
	ctx.Context.Step(`^the response should match json:$`, api.TheResponseShouldMatchJSON)
	ctx.Context.Step(`^the response should set cookie "([^"]*)"$`, api.TheResponseShouldSetCookie)
	ctx.Context.Step(`^set cookie "([^"]*)"`, api.SetCookie)

	ctx.Context.Step(`^when start streaming logs via request to "([^"]*)":$`, api.StartStreamingViaRequestTo)
	ctx.Context.Step(`^wait for (\d+) seconds for the image "([^"]*)" build to complete$`, api.WaitForSecondsForTheImageBuildToComplete)
	ctx.Context.Step(`^wait for (\d+) seconds for the image "([^"]*)" build to error$`, api.WaitForSecondsForTheImageBuildToError)
	ctx.Context.Step(`^streaming logs should have "([^"]*)" message within (\d+) seconds$`, api.TheStreamShouldHaveMessage)
	ctx.Context.Step(`^streaming logs should not have "([^"]*)" message$`, api.TheStreamShouldNotHaveMessage)
	ctx.Context.Step(`^the user "([^"]*)" should not exist$`, api.TheUserShouldNotExist)
	ctx.Context.Step(`^can login with the new user:$`, api.CanLoginWithTheUser)
	ctx.Context.Step(`^the user should be created with the following data:$`, api.TheUserShouldExistWithTheFollowingData)
	ctx.Context.Step(`^can login with the updated user:$`, api.CanLoginWithTheUser)
	ctx.Context.Step(`^the user should be updated with the following data:$`, api.TheUserShouldExistWithTheFollowingData)
	ctx.Context.Step(`^the image "([^"]*)" should not exist$`, api.TheImageShouldNotExist)
	ctx.Context.Step(`^the build "([^"]*)" status should be "([^"]*)" within (\d+) seconds$`, api.TheBuildStatusShouldBe)

	ctx.Context.Step(`^I can download the asset "([^"]*)" with "([^"]*)" request to "([^"]*)"$`, api.ICanDownloadAssetFileWithRequestTo)
	ctx.Context.Step(`^I can download the grading-project with "([^"]*)" request to "([^"]*)"$`, api.ICanDownloadGradingProjectWithRequestTo)

	ctx.Context.Step(`^I export the project bundle with a "(GET)" request to "([^"]*)"$`, api.IExportTheProjectBundle)
	ctx.Context.Step(`^the exported bundle should contain:$`, api.TheExportedBundleShouldContain)
	ctx.Context.Step(`^the exported bundle file "([^"]*)" should match json:$`, api.TheExportedBundleFileShouldMatchJSON)
	ctx.Context.Step(`^the exported bundle file "([^"]*)" should be:$`, api.TheExportedBundleFileShouldBe)
	ctx.Context.Step(`^I import the exported bundle with a "(POST)" request to "([^"]*)"$`, api.IImportTheExportedBundle)
	ctx.Context.Step(`^I import the exported bundle with force with a "(POST)" request to "([^"]*)"$`, api.IImportTheExportedBundleWithForce)
	ctx.Context.Step(`^I import a bundle with a "(POST)" request to "([^"]*)" whose project.json is:$`, api.IImportABundleWhoseProjectJSONIs)
	ctx.Context.Step(`^I import a bundle with a "(POST)" request to "([^"]*)" whose content is:$`, api.IImportARawBundle)
	ctx.Context.Step(`^there should be (\d+) projects$`, api.ThereShouldBeProjects)
	ctx.Context.Step(`^I have no grading project files$`, api.IHaveNoGradingProjectFiles)
	ctx.Context.Step(`^I export the bundle with a "(GET)" request to "([^"]*)"$`, api.IExportTheBundle)
	ctx.Context.Step(`^I import a bundle with a "(POST)" request to "([^"]*)" whose "([^"]*)" is:$`, api.IImportABundleWhoseJSONIs)
	ctx.Context.Step(`^there should be (\d+) contests$`, api.ThereShouldBeContests)
}

type defaultScenarioInitializer struct{}

func (s *defaultScenarioInitializer) InitializeScenario(ctx *Context) {
	api := ctx.API
	db := ctx.DB
	ctx.Context.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		fmt.Printf("[feature][hook][before] scenario=%q id=%s initializer=default\n", sc.Name, sc.Id)
		ResetDB(db)
		RunMigration(db)
		if err := ClearSubmissionCooldowns(); err != nil {
			return ctx, err
		}
		api.ResetResponse(sc)
		ResetExportedBundle()
		return ctx, nil
	})
	ctx.Context.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		fmt.Printf("[feature][hook][after] scenario=%q id=%s initializer=default err=%v\n", sc.Name, sc.Id, err)
		api.CloseStreamForScenario(sc.Id)
		return ctx, err
	})

	InitContextSteps(ctx)
}

func NewDefaultScenarioInitializer() ScenarioInitializer {
	return &defaultScenarioInitializer{}
}

type seederedScenarioInitializer struct{}

func (s *seederedScenarioInitializer) InitializeScenario(ctx *Context) {
	api := ctx.API
	db := ctx.DB
	ctx.Context.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		fmt.Printf("[feature][hook][before] scenario=%q id=%s initializer=seedered\n", sc.Name, sc.Id)
		ResetDB(db)
		RunMigrationsAndSeeders(db)
		if err := ClearSubmissionCooldowns(); err != nil {
			return ctx, err
		}
		api.ResetResponse(sc)
		ResetExportedBundle()
		return ctx, nil
	})
	ctx.Context.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		fmt.Printf("[feature][hook][after] scenario=%q id=%s initializer=seedered err=%v\n", sc.Name, sc.Id, err)
		api.CloseStreamForScenario(sc.Id)
		return ctx, err
	})

	InitContextSteps(ctx)
}

func NewSeederedScenarioInitializer() ScenarioInitializer {
	return &seederedScenarioInitializer{}
}

// mongoScenarioInitializer resets both MySQL and MongoDB before each scenario.
// Use this for tests that rely on MongoDB data (submissions, judgments, rejudges).
type mongoScenarioInitializer struct{}

func (s *mongoScenarioInitializer) InitializeScenario(ctx *Context) {
	api := ctx.API
	db := ctx.DB
	cfg := config.NewFeatureConfig()
	ctx.Context.Before(func(gCtx context.Context, sc *godog.Scenario) (context.Context, error) {
		fmt.Printf("[feature][hook][before] scenario=%q id=%s initializer=mongo\n", sc.Name, sc.Id)

		// Scenarios on this initializer reach the judge pipeline — requesting a
		// rejudge creates judgments and queues them — so the judge state has to
		// be reset here for the same reasons the judge suite does it, and
		// before the databases are cleaned. It is also what restores workers a
		// scenario stopped to make "queued" observable, which nothing else in
		// this initializer would undo.
		if err := ResetJudgeState(cfg); err != nil {
			return gCtx, err
		}

		ResetDB(db)
		RunMigration(db)
		ResetMongoDB()
		if err := ClearSubmissionCooldowns(); err != nil {
			return gCtx, err
		}
		api.ResetResponse(sc)
		ResetExportedBundle()
		return gCtx, nil
	})
	ctx.Context.After(func(gCtx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		fmt.Printf("[feature][hook][after] scenario=%q id=%s initializer=mongo err=%v\n", sc.Name, sc.Id, err)
		api.CloseStreamForScenario(sc.Id)
		return gCtx, err
	})

	InitContextSteps(ctx)
}

func NewMongoScenarioInitializer() ScenarioInitializer {
	return &mongoScenarioInitializer{}
}
