package contest_project_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestContestant(t *testing.T) {
	testutils.TestFeature(t, "get_user_contest_schedule", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "list_user_contests", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_user_contest", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_user_contest_project", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "download_user_project_asset", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "list_user_project_clarifications", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_user_project_clarification", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "send_user_project_clarification", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "reply_user_clarification", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "list_user_project_submissions", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "get_user_submission", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "get_user_project_scoreboard", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "submit_user_project", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "stream_user_submission", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "stream_user_project_time", testutils.NewDefaultScenarioInitializer())
	// testutils.TestFeature(t, "stream_user_project_scoreboard", testutils.NewDefaultScenarioInitializer())
	// testutils.TestFeature(t, "stream_user_project_clarifications", testutils.NewDefaultScenarioInitializer())
	// testutils.TestFeature(t, "stream_user_submissions", testutils.NewDefaultScenarioInitializer())
}
