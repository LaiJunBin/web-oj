package username

import (
	"testing"

	"judge-backend/internal/app/shared/model"
)

func TestContainsDeletedMarker(t *testing.T) {
	cases := []struct {
		username string
		want     bool
	}{
		{"alice", false},
		{"alice__deleted_7", true},
		{"__deleted_1", true},
		{"deleted", false},         // the word alone is not the marker
		{"alice_deleted_7", false}, // single underscore is not the marker
	}
	for _, c := range cases {
		if got := ContainsDeletedMarker(c.username); got != c.want {
			t.Errorf("ContainsDeletedMarker(%q) = %v, want %v", c.username, got, c.want)
		}
	}
}

// The domain layer cannot import shared/model, so the marker constant is
// duplicated — this pin is what keeps the two values one value.
func TestDeletedUsernameMarker_MatchesModelConstant(t *testing.T) {
	if DeletedUsernameMarker != model.DeletedUsernameMarker {
		t.Fatalf("username.DeletedUsernameMarker (%q) != model.DeletedUsernameMarker (%q) — the duplicated constants drifted",
			DeletedUsernameMarker, model.DeletedUsernameMarker)
	}
}

// Validate deliberately accepts marker usernames: hydration (the repository's
// toDomainUser → user.New) must never couple to the marker rule, so a future
// Unscoped read of a renamed dead row cannot blow up on it. The rule lives at
// the creation entry points (create_user, import_users) instead.
func TestValidate_StillAcceptsMarkerUsernames(t *testing.T) {
	if _, err := New("alice__deleted_7"); err != nil {
		t.Fatalf("a marker-shaped username must stay hydratable, got %v", err)
	}
}
