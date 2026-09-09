package system_configuration_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestContest(t *testing.T) {
	testutils.TestFeature(t, "list_contest", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_contest", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "create_contest", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "update_contest", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "delete_contest", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "list_contest_participants", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "add_users_to_contest", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "remove_users_from_contest", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "add_groups_to_contest", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "remove_groups_from_contest", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "contest_bundle", testutils.NewDefaultScenarioInitializer())
}
