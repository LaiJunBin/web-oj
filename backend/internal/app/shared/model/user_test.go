package model

import (
	"strings"
	"testing"
)

func TestDeletedUsername(t *testing.T) {
	t.Run("appends the marker and row id", func(t *testing.T) {
		got := DeletedUsername("alice", 17)
		if got != "alice__deleted_17" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("truncates the base so the result honours size:191", func(t *testing.T) {
		long := strings.Repeat("a", 190)
		got := DeletedUsername(long, 12345)
		if len(got) > 191 {
			t.Fatalf("length %d exceeds the column size", len(got))
		}
		if !strings.HasSuffix(got, "__deleted_12345") {
			t.Fatalf("suffix lost after truncation: %q", got)
		}
	})
}

func TestOriginalUsername(t *testing.T) {
	cases := []struct {
		name   string
		stored string
		want   string
	}{
		{"round-trips a renamed username", DeletedUsername("alice", 17), "alice"},
		{"no marker passes through", "alice", "alice"},
		{"non-digit suffix passes through", "alice__deleted_x1", "alice__deleted_x1"},
		{"empty suffix passes through", "alice__deleted_", "alice__deleted_"},
		// A live user really named "x__deleted_9" who is then deleted:
		// only the OUTER (last) suffix is stripped, restoring the odd but
		// legitimate original name.
		{"double suffix strips only the outer layer", DeletedUsername("x__deleted_9", 42), "x__deleted_9"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := OriginalUsername(c.stored); got != c.want {
				t.Fatalf("OriginalUsername(%q) = %q, want %q", c.stored, got, c.want)
			}
		})
	}

	t.Run("a truncated base is not fully recoverable — accepted", func(t *testing.T) {
		long := strings.Repeat("a", 190)
		stripped := OriginalUsername(DeletedUsername(long, 1))
		if stripped == long {
			t.Fatal("expected lossy recovery for a truncated base (if this ever passes, the truncation logic changed)")
		}
		if !strings.HasPrefix(long, stripped) {
			t.Fatalf("recovered base %q is not a prefix of the original", stripped)
		}
	})
}
