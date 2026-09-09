package system_configuration_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestImage(t *testing.T) {
	testutils.TestFeature(t, "list_image", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_image", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "create_image", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "update_image", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "delete_image", testutils.NewDefaultScenarioInitializer())

	testutils.TestFeature(t, "list_build", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_build", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "cancel_build", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "rebuild", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "image_bundle", testutils.NewDefaultScenarioInitializer())
}
