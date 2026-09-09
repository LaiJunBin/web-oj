package judge_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestJudge(t *testing.T) {
	testutils.TestFeature(t, "submit_user_project", NewJudgeScenarioInitializer())
	testutils.TestFeature(t, "cancel_judge", NewJudgeScenarioInitializer())
	testutils.TestFeature(t, "staff_note", NewJudgeScenarioInitializer())
	testutils.TestFeature(t, "rejudge_submission", NewJudgeScenarioInitializer())
}
