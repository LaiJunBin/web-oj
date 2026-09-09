package imagetag

import (
	"testing"

	"judge-backend/internal/app/shared/model"
)

func TestContainsDeletedMarker(t *testing.T) {
	cases := []struct {
		tag  string
		want bool
	}{
		{"nginx:latest", false},
		{"evil__deleted_1:latest", true},
		{"nginx:latest__deleted_12", true},
		{"nginx:deleted", false},    // the word alone is not the marker
		{"nginx:_deleted_1", false}, // single underscore is not the marker
	}
	for _, c := range cases {
		if got := ContainsDeletedMarker(c.tag); got != c.want {
			t.Errorf("ContainsDeletedMarker(%q) = %v, want %v", c.tag, got, c.want)
		}
	}
}

// The domain layer cannot import shared/model, so the marker constant is
// duplicated — this pin is what keeps the two values one value.
func TestDeletedTagMarker_MatchesModelConstant(t *testing.T) {
	if DeletedTagMarker != model.DeletedImageTagMarker {
		t.Fatalf("imagetag.DeletedTagMarker (%q) != model.DeletedImageTagMarker (%q) — the duplicated constants drifted",
			DeletedTagMarker, model.DeletedImageTagMarker)
	}
}

// Validate deliberately accepts marker tags: build-history hydration reads a
// soft-deleted row's renamed tag back through New, and rejecting it there
// would break every dead image's build history. The rule lives at the
// creation choke point (ImageFactory) instead.
func TestValidate_StillAcceptsMarkerTags(t *testing.T) {
	if _, err := New("nginx:latest__deleted_12"); err != nil {
		t.Fatalf("hydration of a renamed dead-row tag must keep working, got %v", err)
	}
}
