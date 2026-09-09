package contest_project_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestPublic(t *testing.T) {
	testutils.TestFeature(t, "get_public_contest_schedule", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "list_public_contests", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_public_contest", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_public_contest_project", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_public_contest_project_scoreboard", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "download_public_project_asset", testutils.NewDefaultScenarioInitializer())
	// testutils.TestFeature(t, "stream_public_project_scoreboard", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "stream_public_project_time", testutils.NewDefaultScenarioInitializer())
}
