package contest_project_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestClarification(t *testing.T) {
	testutils.TestFeature(t, "list_clarifications", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_clarification", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "send_clarification", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "reply_clarification", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "mark_clarification_to_answered", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "mark_clarification_to_unanswered", testutils.NewDefaultScenarioInitializer())
}
