package bundle_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestBundle(t *testing.T) {
	testutils.TestFeature(t, "composite_bundle", testutils.NewDefaultScenarioInitializer())
}
