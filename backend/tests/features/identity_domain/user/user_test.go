package system_configuration_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestUser(t *testing.T) {
	testutils.TestFeature(t, "list_role", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "list_user", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_user", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "create_user", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "update_user", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "delete_user", testutils.NewDefaultScenarioInitializer())

	testutils.TestFeature(t, "import_users", testutils.NewDefaultScenarioInitializer())

	testutils.TestFeature(t, "group_management", testutils.NewDefaultScenarioInitializer())
}
