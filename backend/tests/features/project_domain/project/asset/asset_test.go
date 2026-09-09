package system_configuration_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestAsset(t *testing.T) {
	testutils.TestFeature(t, "get_asset", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "upload_asset", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "update_asset", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "delete_asset", testutils.NewDefaultScenarioInitializer())
}
