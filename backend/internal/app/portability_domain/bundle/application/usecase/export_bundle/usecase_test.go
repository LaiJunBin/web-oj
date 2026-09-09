package exportbundle_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"strings"
	"testing"

	exportbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/export_bundle"
	"judge-backend/internal/app/shared/infrastructure/logger"
	zipservice "judge-backend/internal/app/shared/infrastructure/services/zip_service"
)

// The fakes stand in for three domains' exporters. Each hands back a minimal
// but REAL zip, because the usecase re-reads its parts' entries — a fake that
// returned placeholder bytes would pass nothing through explode.

func partZip(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	w := zip.NewWriter(buf)
	for name, body := range entries {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

type fakeContests struct{ zip []byte }

func (f fakeContests) Export(int32) ([]byte, error) { return f.zip, nil }

type fakeLister struct{ ids []int32 }

func (f fakeLister) ProjectIDs(int32) ([]int32, error) { return f.ids, nil }

type fakeProjects struct{ zips map[int32][]byte }

func (f fakeProjects) Export(id int32) ([]byte, error) { return f.zips[id], nil }

type fakeImages struct {
	zips     map[string][]byte
	notBuilt map[string]bool
	asked    []string
}

func (f *fakeImages) Export(tag, mode string) ([]byte, error) {
	f.asked = append(f.asked, tag)
	if f.notBuilt[tag] {
		return nil, exportbundle.ErrImageNotBuilt
	}
	return f.zips[tag], nil
}

type fakeParts struct{ tagsByProject map[string][]string }

func (f fakeParts) ReadProject(definition []byte) (string, []string, error) {
	name := string(definition)
	return name, f.tagsByProject[name], nil
}
func (fakeParts) ReadImage([]byte) (string, string, string, error) { return "", "", "", nil }
func (fakeParts) ReadContest([]byte) (string, []string, error)     { return "", nil, nil }

type fixedClock struct{}

func (fixedClock) NowRFC3339() string { return "2026-08-12T00:00:00Z" }

func newUsecase(t *testing.T, lister fakeLister, projects fakeProjects, images *fakeImages, parts fakeParts) exportbundle.Usecase {
	t.Helper()
	return exportbundle.NewUsecase(
		fakeContests{zip: partZip(t, map[string]string{"contest.json": `{"name":"c"}`})},
		lister,
		projects,
		images,
		parts,
		zipservice.NewService(logger.NewMockLogger(), zipservice.Limits{}),
		fixedClock{},
	)
}

// Two problems sharing an image is the ordinary case, and a composite bundle
// that carried the same archive twice would double for nothing.
func TestSharedImagesAreExportedOnce(t *testing.T) {
	images := &fakeImages{zips: map[string][]byte{
		"shared:v1": partZip(t, map[string]string{"image.json": "{}", "context.zip": "ctx"}),
	}}
	u := newUsecase(t,
		fakeLister{ids: []int32{1, 2}},
		fakeProjects{zips: map[int32][]byte{
			1: partZip(t, map[string]string{"project.json": "P1"}),
			2: partZip(t, map[string]string{"project.json": "P2"}),
		}},
		images,
		fakeParts{tagsByProject: map[string][]string{
			"P1": {"shared:v1"},
			"P2": {"shared:v1"},
		}},
	)

	out, err := u.Execute(&exportbundle.Input{ContestID: 1, Mode: "source"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if len(images.asked) != 1 {
		t.Errorf("expected the shared image to be exported once, but it was asked for %d times", len(images.asked))
	}

	reader, err := zip.NewReader(bytes.NewReader(out.Content), int64(len(out.Content)))
	if err != nil {
		t.Fatalf("output is not a zip: %v", err)
	}
	names := make([]string, 0, len(reader.File))
	for _, f := range reader.File {
		names = append(names, f.Name)
	}
	joined := strings.Join(names, "\n")
	for _, want := range []string{"manifest.json", "contest.json", "projects/p1/project.json", "projects/p2/project.json", "images/shared-v1/image.json"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected the archive to contain %s, got:\n%s", want, joined)
		}
	}
}

// An operator told one unbuilt tag per attempt would build it, retry, and meet
// the next. The whole list has to come out of the first refusal.
func TestArchiveModeCollectsEveryUnbuiltTagBeforeFailing(t *testing.T) {
	images := &fakeImages{
		zips:     map[string][]byte{"built:v1": partZip(t, map[string]string{"image.json": "{}"})},
		notBuilt: map[string]bool{"raw-a:v1": true, "raw-b:v1": true},
	}
	u := newUsecase(t,
		fakeLister{ids: []int32{1, 2, 3}},
		fakeProjects{zips: map[int32][]byte{
			1: partZip(t, map[string]string{"project.json": "P1"}),
			2: partZip(t, map[string]string{"project.json": "P2"}),
			3: partZip(t, map[string]string{"project.json": "P3"}),
		}},
		images,
		fakeParts{tagsByProject: map[string][]string{
			"P1": {"raw-a:v1"},
			"P2": {"built:v1"},
			"P3": {"raw-b:v1"},
		}},
	)

	_, err := u.Execute(&exportbundle.Input{ContestID: 1, Mode: "archive"})

	var notBuilt *exportbundle.NotBuiltError
	if !errors.As(err, &notBuilt) {
		t.Fatalf("expected NotBuiltError, got %v", err)
	}
	if len(notBuilt.Tags) != 2 {
		t.Fatalf("expected both unbuilt tags in one refusal, got %v", notBuilt.Tags)
	}
}
