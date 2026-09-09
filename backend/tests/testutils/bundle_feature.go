package testutils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/projectpath"

	"github.com/cucumber/godog"
)

// Steps for the export/import bundles.
//
// They live apart from api_feature.go because they share one thing nothing
// else in the suite needs: the bytes of a downloaded zip, kept between steps
// so a scenario can export a project and then import that exact archive back.
// That round trip is the assertion worth making — it exercises both directions
// against each other, and a field dropped by either one cannot survive it.

// exportedBundle is the last archive downloaded by IExportTheProjectBundle.
var exportedBundle []byte

// ResetExportedBundle clears the carried archive between scenarios, so a
// scenario that forgets to export cannot silently assert against the previous
// one's bytes.
func ResetExportedBundle() { exportedBundle = nil }

func (a *APIFeature) IExportTheProjectBundle(method, endpoint string) error {
	if err := a.ISendrequestTo(method, endpoint); err != nil {
		return err
	}

	if a.resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(a.resp.Body)
		return fmt.Errorf("expected 200, but got %d: %s", a.resp.StatusCode, body)
	}

	if cd := a.resp.Header.Get("Content-Disposition"); !strings.HasPrefix(cd, "attachment") {
		return fmt.Errorf("expected an attachment, but Content-Disposition was %q", cd)
	}

	defer a.resp.Body.Close()
	body, err := io.ReadAll(a.resp.Body)
	if err != nil {
		return err
	}

	if _, err := zip.NewReader(bytes.NewReader(body), int64(len(body))); err != nil {
		return fmt.Errorf("the exported bundle is not a readable zip: %w", err)
	}

	exportedBundle = body
	return nil
}

// TheExportedBundleShouldContain asserts the archive's entries EXACTLY.
//
// Exactly, not "at least": an export that quietly stops writing assets still
// produces a bundle that imports without error, just into a project missing
// its attachments. A subset check would pass throughout.
func (a *APIFeature) TheExportedBundleShouldContain(table *godog.Table) error {
	r, err := readExportedBundle()
	if err != nil {
		return err
	}

	got := make([]string, 0, len(r.File))
	for _, f := range r.File {
		got = append(got, f.Name)
	}

	want := make([]string, 0, len(table.Rows)-1)
	for i := 1; i < len(table.Rows); i++ {
		want = append(want, table.Rows[i].Cells[0].Value)
	}

	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		return fmt.Errorf("bundle entries differ\nwant: %v\ngot:  %v", want, got)
	}
	return nil
}

func (a *APIFeature) TheExportedBundleFileShouldMatchJSON(name string, body *godog.DocString) error {
	content, err := readBundleEntry(name)
	if err != nil {
		return err
	}
	return MatchJSON(content, body.Content)
}

// TheExportedBundleFileShouldBe asserts a non-JSON entry byte for byte, which
// is how an attachment's content is checked to have survived the round trip.
func (a *APIFeature) TheExportedBundleFileShouldBe(name string, body *godog.DocString) error {
	content, err := readBundleEntry(name)
	if err != nil {
		return err
	}
	if string(content) != body.Content {
		return fmt.Errorf("%s differs\nwant: %q\ngot:  %q", name, body.Content, content)
	}
	return nil
}

// IImportTheExportedBundle re-uploads the archive the scenario just exported.
func (a *APIFeature) IImportTheExportedBundle(method, endpoint string) error {
	if exportedBundle == nil {
		return fmt.Errorf("no bundle has been exported in this scenario")
	}
	return a.postBundle(method, endpoint, "bundle.zip", exportedBundle)
}

// IImportTheExportedBundleWithForce is the same upload with force=true — the
// override an operator reaches for after a 409 told them what it would affect.
func (a *APIFeature) IImportTheExportedBundleWithForce(method, endpoint string) error {
	if exportedBundle == nil {
		return fmt.Errorf("no bundle has been exported in this scenario")
	}
	return a.postBundleWithFields(method, endpoint, "bundle.zip", exportedBundle, map[string]string{"force": "true"})
}

// IImportABundleWhoseProjectJSONIs builds a minimal archive around the given
// definition and uploads it.
//
// Written inline rather than kept as a fixture file on purpose: these
// scenarios exist to pin what happens to a MALFORMED bundle, and a reader
// cannot judge the expected error without seeing the input that caused it.
func (a *APIFeature) IImportABundleWhoseProjectJSONIs(method, endpoint string, body *godog.DocString) error {
	archive := &bytes.Buffer{}
	w := zip.NewWriter(archive)
	f, err := w.Create("project.json")
	if err != nil {
		return err
	}
	if _, err := f.Write([]byte(body.Content)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return a.postBundle(method, endpoint, "bundle.zip", archive.Bytes())
}

// IImportARawBundle uploads bytes that are not a zip at all.
func (a *APIFeature) IImportARawBundle(method, endpoint string, body *godog.DocString) error {
	return a.postBundle(method, endpoint, "bundle.zip", []byte(body.Content))
}

func (a *APIFeature) postBundle(method, endpoint, filename string, content []byte) error {
	return a.postBundleWithFields(method, endpoint, filename, content, nil)
}

func (a *APIFeature) postBundleWithFields(method, endpoint, filename string, content []byte, fields map[string]string) error {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	part, err := w.CreateFormFile("bundle", filename)
	if err != nil {
		return err
	}
	if _, err := part.Write(content); err != nil {
		return err
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(method, a.buildAPIURL(endpoint), buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	for k, v := range a.cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	a.resp = resp
	return nil
}

func readExportedBundle() (*zip.Reader, error) {
	if exportedBundle == nil {
		return nil, fmt.Errorf("no bundle has been exported in this scenario")
	}
	return zip.NewReader(bytes.NewReader(exportedBundle), int64(len(exportedBundle)))
}

func readBundleEntry(name string) ([]byte, error) {
	r, err := readExportedBundle()
	if err != nil {
		return nil, err
	}
	for _, f := range r.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}

	names := make([]string, 0, len(r.File))
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	return nil, fmt.Errorf("the bundle has no entry %q; it contains %v", name, names)
}

// ThereShouldBeProjects is the cleanup assertion. An import that fails after
// creating a project would otherwise leave a half-built problem in the library
// with nothing in the response to say so, and the next scenario — or the next
// operator — would find it and wonder.
func (a *APIFeature) ThereShouldBeProjects(want int) error {
	var got int64
	if err := GetDB().Model(&model.Project{}).Count(&got).Error; err != nil {
		return err
	}
	if int(got) != want {
		return fmt.Errorf("expected %d projects, but found %d", want, got)
	}
	return nil
}

// IHaveNoGradingProjectFiles clears the grading project storage directory.
//
// ResetDB restores the database between scenarios, but nothing resets the
// storage directory — a feature that uploaded grading files earlier in the
// same suite leaves them on disk for whatever runs next. Most features never
// notice, because they only read files they wrote themselves. An export reads
// whatever is on disk for the project it is handed, so it is the first thing
// here to see another feature's leftovers, and it has to say what state it
// needs rather than inherit one.
//
// Stated as a precondition rather than worked around by picking unused project
// ids: the ids would collide again the moment somebody adds a feature, and the
// scenario would go back to depending on luck.
func (a *APIFeature) IHaveNoGradingProjectFiles() error {
	// StoragePath is relative and the test process runs from the feature's own
	// directory, so it has to be resolved against the repo root — the same
	// thing IHaveAssets does with os.Chdir before writing its fixtures.
	root := path.Join(projectpath.Root, a.config.Server.StoragePath)

	// os.RemoveAll succeeds on a path that does not exist, so a step pointed
	// at the wrong place reports success having deleted nothing. That is
	// exactly how the first version of this failed, so check the storage root
	// itself resolves before trusting the removal.
	if _, err := os.Stat(root); err != nil {
		return fmt.Errorf("storage root %q is not there, so this step would silently delete nothing: %w", root, err)
	}

	return os.RemoveAll(path.Join(root, "projects"))
}

// The bundle steps above are written for the project endpoints but nothing in
// them is project-specific: an export is an export. These aliases let the image
// and contest features read naturally rather than talking about "the project
// bundle" when they mean an image.

func (a *APIFeature) IExportTheBundle(method, endpoint string) error {
	return a.IExportTheProjectBundle(method, endpoint)
}

// IImportABundleWhoseJSONIs builds an archive around one named document.
//
// Each bundle kind puts its definition in a differently named file, so the
// name is a parameter here rather than baked in as it is for projects.
func (a *APIFeature) IImportABundleWhoseJSONIs(method, endpoint, filename string, body *godog.DocString) error {
	archive := &bytes.Buffer{}
	w := zip.NewWriter(archive)
	f, err := w.Create(filename)
	if err != nil {
		return err
	}
	if _, err := f.Write([]byte(body.Content)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return a.postBundle(method, endpoint, "bundle.zip", archive.Bytes())
}

// ThereShouldBeContests mirrors ThereShouldBeProjects: proof that a rejected
// import created nothing.
func (a *APIFeature) ThereShouldBeContests(want int) error {
	var got int64
	if err := GetDB().Model(&model.Contest{}).Count(&got).Error; err != nil {
		return err
	}
	if int(got) != want {
		return fmt.Errorf("expected %d contests, but found %d", want, got)
	}
	return nil
}
