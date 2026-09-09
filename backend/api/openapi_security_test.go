package api_test

import (
	"os"
	"sort"
	"testing"

	"gopkg.in/yaml.v2"
)

// Authorization is spec-driven end to end: the generated wrapper only calls
// c.Set(<Permission>Scopes, ...) for the schemes an operation declares, and
// every middleware self-gates on its scope key being present. An operation
// that declares no security therefore reaches its controller with no
// authentication and no permission check at all — it is fully public.
//
// That is correct for the handful of endpoints below and a silent hole for
// anything else, so the allowlist is the contract: adding a public operation
// is a deliberate edit here, not an omission in the TypeSpec.
//
// The spec is generated from typespec/, so a failure here means the fix
// belongs in the .tsp source, not in openapi.yaml.
var intentionallyPublic = map[string]bool{
	"GET /csrf-token":                                     true,
	"DELETE /csrf-token":                                  true,
	"GET /public/contests":                                true,
	"GET /public/contests/{contestId}":                    true,
	"GET /public/contests/{contestId}/schedule":           true,
	"GET /public/contest-projects/{id}":                   true,
	"GET /public/contest-projects/{id}/assets/{assetId}":  true,
	"GET /public/contest-projects/{id}/scoreboard":        true,
	"GET /public/contest-projects/{id}/stream-scoreboard": true,
	"GET /public/contest-projects/{id}/stream-time":       true,
}

type specOperation struct {
	OperationID string                `yaml:"operationId"`
	Security    []map[string][]string `yaml:"security"`
}

type specDocument struct {
	Paths map[string]map[string]specOperation `yaml:"paths"`
}

func loadSpec(t *testing.T) specDocument {
	t.Helper()

	raw, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}

	var doc specDocument
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse openapi.yaml: %v", err)
	}
	if len(doc.Paths) == 0 {
		t.Fatal("openapi.yaml has no paths — the spec failed to parse into the expected shape")
	}
	return doc
}

func TestEveryOperationDeclaresSecurity(t *testing.T) {
	doc := loadSpec(t)

	var unprotected []string
	seen := map[string]bool{}
	for path, methods := range doc.Paths {
		for method, op := range methods {
			if op.OperationID == "" {
				continue // not an operation (parameters, servers, …)
			}
			key := upper(method) + " " + path
			seen[key] = true
			if len(op.Security) == 0 && !intentionallyPublic[key] {
				unprotected = append(unprotected, key)
			}
		}
	}

	sort.Strings(unprotected)
	for _, key := range unprotected {
		t.Errorf("operation %q declares no security and is not in intentionallyPublic — "+
			"it is reachable with no authentication; add @usePermission in typespec/ "+
			"or allowlist it here on purpose", key)
	}

	// A stale allowlist is its own hazard: it silently blesses whatever moves
	// onto that path later.
	for key := range intentionallyPublic {
		if !seen[key] {
			t.Errorf("intentionallyPublic lists %q, which no longer exists in the spec — remove it", key)
		}
	}
}

// Anything that mutates state must sit behind a permission, not merely behind
// a login. Authorization alone proves who you are; it says nothing about what
// you may do.
func TestMutatingOperationsRequireAPermission(t *testing.T) {
	doc := loadSpec(t)

	mutating := map[string]bool{"post": true, "put": true, "patch": true, "delete": true}
	for path, methods := range doc.Paths {
		for method, op := range methods {
			if op.OperationID == "" || !mutating[method] {
				continue
			}
			key := upper(method) + " " + path
			if intentionallyPublic[key] {
				continue
			}

			hasPermission := false
			for _, requirement := range op.Security {
				for scheme := range requirement {
					if len(scheme) > len("Permission") && scheme[:len("Permission")] == "Permission" {
						hasPermission = true
					}
					// The auth endpoints are gated on possession of a cookie
					// rather than on a permission, which is the point of them.
					if scheme == "RefreshToken" || scheme == "CsrfToken" {
						hasPermission = true
					}
				}
			}

			if !hasPermission {
				t.Errorf("mutating operation %q declares no Permission* scheme — "+
					"any logged-in user can call it", key)
			}
		}
	}
}

// enforcedSchemes is every security scheme the middleware chain actually
// checks, one entry per file in
// internal/app/identity_domain/auth/adapter/middleware.
//
// The chain is opt-in per scheme: each middleware runs on every route and
// returns early unless it finds its own scope key on the context. A scheme
// that appears in the spec with no middleware behind it is therefore not a
// weaker check — it is *no* check, and the operation is open to anyone
// logged in. Nothing else in the build would notice, because the generated
// wrapper happily sets a scope key nobody reads.
//
// So the spec and the middleware chain have to agree exactly, in both
// directions: a new scheme in the TypeSpec needs a new middleware, and a
// retired scheme should not linger here.
var enforcedSchemes = map[string]bool{
	"Authorization":        true, // authorization.go
	"UserID":               true, // with_user_id.go
	"PermissionRead":       true, // auth_read.go
	"PermissionAdmin":      true, // auth_admin.go
	"PermissionContestant": true, // auth_contestant.go
	"RefreshToken":         true, // with_refresh_token.go
	"CsrfToken":            true, // csrf.go (also applied globally to unsafe methods)
}

func TestSpecUsesOnlyEnforcedSecuritySchemes(t *testing.T) {
	doc := loadSpec(t)

	used := map[string]bool{}
	for _, methods := range doc.Paths {
		for _, op := range methods {
			if op.OperationID == "" {
				continue
			}
			for _, requirement := range op.Security {
				for scheme := range requirement {
					used[scheme] = true
				}
			}
		}
	}

	for scheme := range used {
		if !enforcedSchemes[scheme] {
			t.Errorf("spec uses security scheme %q, which no middleware enforces — "+
				"operations declaring it are reachable by any authenticated user; "+
				"add the middleware and register it in app/provider.Middlewares", scheme)
		}
	}

	for scheme := range enforcedSchemes {
		if !used[scheme] {
			t.Errorf("middleware enforces security scheme %q, which no operation "+
				"declares — either the TypeSpec dropped it (delete the middleware) "+
				"or an endpoint lost its annotation", scheme)
		}
	}
}

func upper(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] -= 'a' - 'A'
		}
	}
	return string(b)
}
