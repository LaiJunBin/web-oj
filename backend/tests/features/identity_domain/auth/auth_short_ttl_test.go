package auth_test

import (
	"judge-backend/internal/config"
	"judge-backend/tests/testutils"
	"testing"
)

// shortTTLThreshold is the longest access lifetime the short-TTL feature can
// meaningfully run under. Its scenario waits a few seconds for the token to
// expire, so anything above this would let the request succeed on the
// original token and the scenario would pass without ever exercising
// renewal — worse than not running at all.
const shortTTLThreshold = 10

// TestAuthShortTTL proves the whole lifetime chain: config supplies the
// access lifetime, a real login issues a token with it, the token expires,
// and the middleware renews it.
//
// It gates on the configured lifetime rather than on a flag of its own, so
// it can only run when it is actually meaningful. Enable it by overriding
// the lifetime for the process — viper's AutomaticEnv covers keys present in
// the config file, so AUTH_ACCESSTOKENTTLSECONDS=2 is enough — and run it
// alone via `make test:feature:short-ttl`. Running it alongside the rest of
// the suite would leave every other scenario holding two-second tokens.
func TestAuthShortTTL(t *testing.T) {
	cfg := config.NewFeatureConfig()
	if ttl := cfg.Auth.AccessTokenTTLSeconds; ttl > shortTTLThreshold {
		t.Skipf("access token lifetime is %ds; set AUTH_ACCESSTOKENTTLSECONDS to %d or less to run this (see make test:feature:short-ttl)", ttl, shortTTLThreshold)
	}

	testutils.TestFeature(t, "short_ttl_renewal", testutils.NewDefaultScenarioInitializer())
}
