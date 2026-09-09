package importimage

import (
	archivetar "archive/tar"
	"bytes"
	"errors"
	"testing"
)

func buildImageArchive(t *testing.T, manifest *string) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := archivetar.NewWriter(&buf)
	write := func(name, body string) {
		if err := tw.WriteHeader(&archivetar.Header{Name: name, Size: int64(len(body)), Mode: 0644}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	// A layer file so the archive resembles a real docker save output.
	write("layer/layer.tar", "x")
	if manifest != nil {
		write("manifest.json", *manifest)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func strptr(s string) *string { return &s }

func TestVerifyArchiveContainsOnlyTag(t *testing.T) {
	const allowed = "img:latest"

	cases := []struct {
		name     string
		manifest *string
		wantErr  bool
	}{
		{"single entry with only the allowed tag", strptr(`[{"RepoTags":["img:latest"]}]`), false},
		{"untagged image is allowed (post-load digest check catches a missing tag)", strptr(`[{"RepoTags":null}]`), false},
		{"an extra tag is rejected", strptr(`[{"RepoTags":["img:latest","evil:latest"]}]`), true},
		{"a different tag is rejected", strptr(`[{"RepoTags":["other:latest"]}]`), true},
		{"multiple images are rejected", strptr(`[{"RepoTags":["img:latest"]},{"RepoTags":["x:1"]}]`), true},
		{"missing manifest is rejected", nil, true},
		{"unparseable manifest is rejected", strptr(`not json`), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyArchiveContainsOnlyTag(buildImageArchive(t, tc.manifest), allowed)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected rejection, got nil")
				}
				var invalid *InvalidBundleError
				if !errors.As(err, &invalid) {
					t.Fatalf("expected InvalidBundleError, got %T: %v", err, err)
				}
			} else if err != nil {
				t.Fatalf("expected acceptance, got %v", err)
			}
		})
	}
}
