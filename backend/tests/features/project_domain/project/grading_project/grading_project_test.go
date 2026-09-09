package system_configuration_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestGradingProject(t *testing.T) {
	testutils.TestFeature(t, "grading_project", testutils.NewDefaultScenarioInitializer())
}
