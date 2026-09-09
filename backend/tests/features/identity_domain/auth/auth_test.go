package auth_test

import (
	"judge-backend/tests/testutils"
	"testing"
)

func TestAuth(t *testing.T) {
	testutils.TestFeature(t, "login", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "get_profile", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "logout", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "refresh_token", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "renew_access_token", testutils.NewDefaultScenarioInitializer())
	testutils.TestFeature(t, "password_change_revokes_sessions", testutils.NewDefaultScenarioInitializer())
}
