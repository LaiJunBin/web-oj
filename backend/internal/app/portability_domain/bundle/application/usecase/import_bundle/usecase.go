package importbundle

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"strings"

	"judge-backend/internal/app/portability_domain/bundle/application/bundle"
	previewbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/preview_bundle"
	sharedservice "judge-backend/internal/app/shared/service"
	"judge-backend/internal/app/shared/types/zipfile"
)

var (
	ErrorInternalServer = errors.New("internal server error")
	ErrorConflict       = errors.New("匯出檔與系統既有的內容衝突")
)

// ConflictError is a refusal the operator can override with force. Its
// details are the preview's blockers, so what the API refuses over is exactly
// what the preview showed.
type ConflictError struct{ Details []string }

func (e *ConflictError) Error() string { return ErrorConflict.Error() }
func (e *ConflictError) Unwrap() error { return ErrorConflict }

type Input struct {
	Bundle *bytes.Buffer
	Force  bool
}

type Output struct {
	CreatedImages []string
	ReusedImages  []string
	QueuedBuilds  []string

	CreatedProjectIDs []int32
	ContestID         *int32

	UnresolvedProjects []string

	// Warnings carries non-fatal notes on an otherwise successful import. A
	// midway failure no longer lands here: it returns a PartialImportError so
	// the response is an honest non-2xx rather than a success with caveats.
	Warnings []string
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

// The three importers, each the same usecase the part's own endpoint runs.
// The orchestrator adds nothing to what they do — it only decides the order
// and assembles the report.

type ImageImporter interface {
	Import(partZip []byte, force bool) (buildQueued bool, err error)
}

type ProjectImporter interface {
	Import(partZip []byte) (projectID int32, name string, err error)
}

type ContestImporter interface {
	Import(partZip []byte) (contestID int32, created int32, unresolved []string, err error)
}

type usecase struct {
	zip      sharedservice.ZipService
	preview  previewbundle.Usecase
	images   ImageImporter
	projects ProjectImporter
	contests ContestImporter
}

func NewUsecase(
	zipService sharedservice.ZipService,
	preview previewbundle.Usecase,
	images ImageImporter,
	projects ProjectImporter,
	contests ContestImporter,
) Usecase {
	return &usecase{zip: zipService, preview: preview, images: images, projects: projects, contests: contests}
}

// Execute runs the preview first and imports only what it approved, in the
// one order that works: images, then the projects that reference them, then
// the contest that references the projects.
func (u *usecase) Execute(input *Input) (*Output, error) {
	// The preview is the gate, not a convenience. Whatever it lists as a
	// blocker is refused here, so the operator can trust that a clean preview
	// means a clean import — the two cannot drift because one runs the other.
	seen, err := u.preview.Execute(&previewbundle.Input{Bundle: input.Bundle})
	if err != nil {
		return nil, err
	}

	if len(seen.Blockers) > 0 {
		conflictsOnly := true
		for _, p := range seen.Projects {
			if len(p.MissingImageTags) > 0 {
				conflictsOnly = false
			}
		}
		// A missing image is not overridable: force means "replace what is
		// there", and no amount of replacing conjures an image nobody has.
		if !conflictsOnly || !input.Force {
			return nil, &ConflictError{Details: seen.Blockers}
		}
	}

	contents, _, err := previewbundle.Read(u.zip, input.Bundle)
	if err != nil {
		return nil, err
	}

	out := &Output{
		CreatedImages:      []string{},
		ReusedImages:       []string{},
		QueuedBuilds:       []string{},
		CreatedProjectIDs:  []int32{},
		UnresolvedProjects: []string{},
		Warnings:           []string{},
	}

	statusByTag := map[string]string{}
	for _, item := range seen.Images {
		statusByTag[item.Tag] = item.Status
	}

	for i, slug := range contents.Manifest.Images {
		tag := seen.Images[i].Tag

		// Identical content is reused without touching anything — the
		// ordinary case when two bundles share a base image.
		if statusByTag[tag] == previewbundle.StatusIdentical {
			out.ReusedImages = append(out.ReusedImages, tag)
			continue
		}

		partZip, err := rezip(u.zip, contents.Files, path.Join(bundle.ImagesDir, slug))
		if err != nil {
			return nil, ErrorInternalServer
		}

		queued, err := u.images.Import(partZip, input.Force)
		if err != nil {
			return u.abort(out, err, fmt.Sprintf("匯入映像 %q 失敗：%v", tag, err))
		}
		out.CreatedImages = append(out.CreatedImages, tag)
		if queued {
			out.QueuedBuilds = append(out.QueuedBuilds, tag)
		}
	}

	for _, slug := range contents.Manifest.Projects {
		partZip, err := rezip(u.zip, contents.Files, path.Join(bundle.ProjectsDir, slug))
		if err != nil {
			return nil, ErrorInternalServer
		}

		id, name, err := u.projects.Import(partZip)
		if err != nil {
			return u.abort(out, err, fmt.Sprintf("匯入題目 %q 失敗：%v", slug, err))
		}
		_ = name
		out.CreatedProjectIDs = append(out.CreatedProjectIDs, id)
	}

	if contents.Manifest.HasContest {
		partZip, err := singleFileZip(u.zip, "contest.json", contents.Files[bundle.ContestFile])
		if err != nil {
			return nil, ErrorInternalServer
		}

		contestID, _, unresolved, err := u.contests.Import(partZip)
		if err != nil {
			return u.abort(out, err, fmt.Sprintf("匯入競賽失敗：%v", err))
		}
		out.ContestID = &contestID
		out.UnresolvedProjects = unresolved
	}

	return out, nil
}

// PartialImportError reports that a composite import failed partway. The parts
// that succeeded were created by their own usecases, each with its own
// compensation for the one that failed; this level owes the operator an honest
// report — never a success — listing the failure and what already exists so
// they can finish or remove it by hand. The presenter maps it to a single 5xx
// (there is nothing to "force" past), with Details() carrying the failure and
// the created-item list. cause is retained via Unwrap for logging/inspection,
// not to vary the status code.
type PartialImportError struct {
	details []string
	cause   error
}

func (e *PartialImportError) Error() string     { return strings.Join(e.details, "; ") }
func (e *PartialImportError) Details() []string { return e.details }
func (e *PartialImportError) Unwrap() error     { return e.cause }

// abort stops at the first failure and returns it — no pretend rollback across
// three domains, and crucially not a success. The returned error carries both
// the failure and the list of what already exists.
func (u *usecase) abort(out *Output, cause error, warning string) (*Output, error) {
	details := []string{warning}
	if len(out.CreatedImages) > 0 {
		details = append(details, fmt.Sprintf("已建立的映像：%s", strings.Join(out.CreatedImages, "、")))
	}
	if len(out.CreatedProjectIDs) > 0 {
		ids := make([]string, len(out.CreatedProjectIDs))
		for i, id := range out.CreatedProjectIDs {
			ids[i] = fmt.Sprintf("#%d", id)
		}
		details = append(details, fmt.Sprintf("已建立的題目：%s；不需要的話請手動刪除", strings.Join(ids, "、")))
	}
	return nil, &PartialImportError{details: details, cause: cause}
}

// rezip reassembles one part's files into the standalone bundle its importer
// accepts — the exact inverse of the export's explode.
func rezip(zipService sharedservice.ZipService, files map[string][]byte, prefix string) ([]byte, error) {
	entries := make([]zipfile.File, 0, 4)
	for name, body := range files {
		if !strings.HasPrefix(name, prefix+"/") {
			continue
		}
		entries = append(entries, zipfile.File{
			Path: strings.TrimPrefix(name, prefix+"/"),
			Body: string(body),
		})
	}

	buf, err := zipService.ZipFromFiles(&entries)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func singleFileZip(zipService sharedservice.ZipService, name string, body []byte) ([]byte, error) {
	entries := []zipfile.File{{Path: name, Body: string(body)}}
	buf, err := zipService.ZipFromFiles(&entries)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
