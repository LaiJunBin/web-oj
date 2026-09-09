package contest_project_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestContestProject(t *testing.T) {
	testutils.TestFeature(t, "get_contest_project", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_submissions_by_contest_project", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "get_submission", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "get_submissions_by_user", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "ignore_submission", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "unignore_submission", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "list_rejudges", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "list_rejudges_by_contest_project", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "get_rejudge", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "get_rejudge_submission_comparison", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "approve_rejudge", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "reject_rejudge", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "cancel_rejudge", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "get_user_scores", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "rejudge_contest_project", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "get_submission_rate", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "get_judgment_analytics", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "list_judgment_errors", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "stream_judgment_errors", testutils.NewMongoScenarioInitializer())
	testutils.TestFeature(t, "stream_submission_rate", testutils.NewMongoScenarioInitializer())
}
