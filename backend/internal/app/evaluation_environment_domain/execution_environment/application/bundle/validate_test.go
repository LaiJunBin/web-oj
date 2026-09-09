package bundle

import (
	"strings"
	"testing"
)

// The rename-on-delete namespace is reserved: a bundle declaring a marker
// tag is refused as a 400 problem here, before the factory's plainer error
// deeper in the import flow could turn it into a 500.
func TestValidate_RejectsReservedDeletedMarkerTag(t *testing.T) {
	i := &Image{
		FormatVersion: FormatVersion,
		Kind:          KindImage,
		Tag:           "nginx:latest__deleted_12",
		MountPath:     "/var/www/html",
		Mode:          ModeSource,
		Fingerprint:   "abc",
	}

	problems := i.Validate()

	found := false
	for _, p := range problems {
		if p.Path == "tag" && strings.Contains(p.Message, "__deleted_") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a tag problem naming the reserved marker, got %v", problems)
	}
}
