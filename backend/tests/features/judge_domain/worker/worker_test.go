package worker_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestWorker(t *testing.T) {
	testutils.TestFeature(t, "get_judge_queue", NewWorkerScenarioInitializer())
	testutils.TestFeature(t, "stream_judge_queue", NewWorkerScenarioInitializer())
	// The default initializer, not the worker one: this endpoint only reads
	// configuration, so resetting judge state around it would cost minutes to
	// prove nothing.
	testutils.TestFeature(t, "get_system_defaults", testutils.NewDefaultScenarioInitializer())
}
