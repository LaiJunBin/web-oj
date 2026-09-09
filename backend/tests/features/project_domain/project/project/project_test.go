package system_configuration_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestProject(t *testing.T) {
	testutils.TestFeature(t, "list_project", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_project", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "create_project", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "update_project", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "delete_project", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "project_bundle", testutils.NewDefaultScenarioInitializer())
}
