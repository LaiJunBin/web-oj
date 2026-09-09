package system_configuration_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestLifecycle(t *testing.T) {
	testutils.TestFeature(t, "get_lifecycle", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "create_lifecycle", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "update_lifecycle", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "delete_lifecycle", testutils.NewDefaultScenarioInitializer())
}
