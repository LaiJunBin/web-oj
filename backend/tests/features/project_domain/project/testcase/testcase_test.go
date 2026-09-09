package system_configuration_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestTestcase(t *testing.T) {
	testutils.TestFeature(t, "create_testcase", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "update_testcase", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "delete_testcase", testutils.NewDefaultScenarioInitializer())
}
