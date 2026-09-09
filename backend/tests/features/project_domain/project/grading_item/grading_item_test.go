package system_configuration_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestGradingItem(t *testing.T) {
	testutils.TestFeature(t, "list_grading_item", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_grading_item", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "create_grading_item", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "update_grading_item", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "delete_grading_item", testutils.NewDefaultScenarioInitializer())
}
