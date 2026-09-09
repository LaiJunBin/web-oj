package exportbundle

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"judge-backend/internal/app/portability_domain/bundle/application/bundle"
	previewbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/preview_bundle"
	sharedservice "judge-backend/internal/app/shared/service"
	"judge-backend/internal/app/shared/types/zipfile"
)

var (
	ErrorInvalidContestID = errors.New("invalid contest id, must be greater than 0")
	ErrorInvalidMode      = errors.New("mode must be source or archive")
	ErrorContestNotFound  = errors.New("contest not found")
	ErrorInternalServer   = errors.New("internal server error")
)

// NotBuiltError lists every image an archive-mode export would need built
// first. All of them at once: an operator told one tag per attempt would build
// it, retry, and meet the next.
type NotBuiltError struct{ Tags []string }

func (e *NotBuiltError) Error() string {
	return fmt.Sprintf("以下映像尚未建置完成，無法以 archive 模式匯出：%s；請先建置，或改用 source 模式",
		strings.Join(e.Tags, "、"))
}

type Input struct {
	ContestID int32
	Mode      string
}

type Output struct {
	Filename string
	Content  []byte
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

// The parts are produced by the SAME exporters their own endpoints use, called
// through these ports. The composite package never serialises a project or an
// image itself — if it did, the standalone export and the composite export
// would be two writers of one format, and they would drift.

type ContestExporter interface {
	// Export returns the contest bundle zip, or ErrNotFound.
	Export(contestID int32) ([]byte, error)
}

// ErrNotFound is the port-level "no such thing", shared by the exporters so
// the usecase does not depend on three domains' sentinel errors.
var ErrNotFound = errors.New("not found")

// ErrImageNotBuilt marks a tag an archive export cannot include yet.
var ErrImageNotBuilt = errors.New("image not built")

type ContestProjectLister interface {
	// ProjectIDs returns the ids of the problems the contest references, in
	// slot order.
	ProjectIDs(contestID int32) ([]int32, error)
}

type ProjectExporter interface {
	Export(projectID int32) ([]byte, error)
}

type ImageExporter interface {
	// Export returns the image bundle zip, ErrImageNotBuilt for an archive of
	// something the daemon never built.
	Export(tag, mode string) ([]byte, error)
}

type Clock interface {
	NowRFC3339() string
}

type usecase struct {
	contests ContestExporter
	lister   ContestProjectLister
	projects ProjectExporter
	images   ImageExporter
	parts    previewbundle.PartReader
	zip      sharedservice.ZipService
	clock    Clock
}

func NewUsecase(
	contests ContestExporter,
	lister ContestProjectLister,
	projects ProjectExporter,
	images ImageExporter,
	parts previewbundle.PartReader,
	zipService sharedservice.ZipService,
	clock Clock,
) Usecase {
	return &usecase{
		contests: contests, lister: lister, projects: projects,
		images: images, parts: parts, zip: zipService, clock: clock,
	}
}

// Execute assembles the whole sitting: the contest, every problem it
// references, and every image those problems need, each part being exactly
// the bytes its own export endpoint would produce.
func (u *usecase) Execute(input *Input) (*Output, error) {
	if input.ContestID <= 0 {
		return nil, ErrorInvalidContestID
	}
	if input.Mode != "source" && input.Mode != "archive" {
		return nil, ErrorInvalidMode
	}

	contestZip, err := u.contests.Export(input.ContestID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrorContestNotFound
		}
		return nil, ErrorInternalServer
	}

	files := make([]zipfile.File, 0, 16)
	manifest := &bundle.Manifest{
		FormatVersion: bundle.FormatVersion,
		Kind:          bundle.KindComposite,
		HasContest:    true,
		ExportedAt:    u.clock.NowRFC3339(),
		Images:        []string{},
		Projects:      []string{},
	}

	contestJSON, err := entryOf(contestZip, "contest.json")
	if err != nil {
		return nil, ErrorInternalServer
	}
	files = append(files, zipfile.File{Path: bundle.ContestFile, Body: string(contestJSON)})

	ids, err := u.lister.ProjectIDs(input.ContestID)
	if err != nil {
		return nil, ErrorInternalServer
	}

	// Tags are collected across projects and exported once: two problems
	// sharing an image is the ordinary case, and a bundle carrying the same
	// hundreds-of-megabytes archive twice would be a silly bundle.
	seenTags := map[string]bool{}
	tagOrder := []string{}

	for _, id := range ids {
		projectZip, err := u.projects.Export(id)
		if err != nil {
			return nil, ErrorInternalServer
		}

		slug := fmt.Sprintf("p%d", id)
		manifest.Projects = append(manifest.Projects, slug)

		exploded, err := explode(projectZip, path.Join(bundle.ProjectsDir, slug))
		if err != nil {
			return nil, ErrorInternalServer
		}
		files = append(files, exploded...)

		definition, err := entryOf(projectZip, "project.json")
		if err != nil {
			return nil, ErrorInternalServer
		}
		_, tags, err := u.parts.ReadProject(definition)
		if err != nil {
			return nil, ErrorInternalServer
		}
		for _, tag := range tags {
			if !seenTags[tag] {
				seenTags[tag] = true
				tagOrder = append(tagOrder, tag)
			}
		}
	}

	// Every unbuilt tag is collected before failing, so the operator gets the
	// whole build list in one refusal instead of one tag per attempt.
	notBuilt := []string{}
	for _, tag := range tagOrder {
		imageZip, err := u.images.Export(tag, input.Mode)
		if errors.Is(err, ErrImageNotBuilt) {
			notBuilt = append(notBuilt, tag)
			continue
		}
		if err != nil {
			return nil, ErrorInternalServer
		}
		if len(notBuilt) > 0 {
			// Already failing; keep collecting tags, skip the payload work.
			continue
		}

		slug := slugOfTag(tag)
		manifest.Images = append(manifest.Images, slug)

		exploded, err := explode(imageZip, path.Join(bundle.ImagesDir, slug))
		if err != nil {
			return nil, ErrorInternalServer
		}
		files = append(files, exploded...)
	}
	if len(notBuilt) > 0 {
		return nil, &NotBuiltError{Tags: notBuilt}
	}

	manifestJSON, err := bundle.MarshalManifest(manifest)
	if err != nil {
		return nil, ErrorInternalServer
	}
	files = append(files, zipfile.File{Path: bundle.ManifestFile, Body: string(manifestJSON)})

	archive, err := u.zip.ZipFromFiles(&files)
	if err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{
		Filename: fmt.Sprintf("bundle-contest-%d-%s.zip", input.ContestID, input.Mode),
		Content:  archive.Bytes(),
	}, nil
}

// explode re-roots a part bundle's entries under a directory, byte for byte.
func explode(partZip []byte, prefix string) ([]zipfile.File, error) {
	reader, err := zip.NewReader(bytes.NewReader(partZip), int64(len(partZip)))
	if err != nil {
		return nil, err
	}

	out := make([]zipfile.File, 0, len(reader.File))
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		out = append(out, zipfile.File{Path: path.Join(prefix, f.Name), Body: string(body)})
	}
	return out, nil
}

func entryOf(partZip []byte, name string) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(partZip), int64(len(partZip)))
	if err != nil {
		return nil, err
	}
	for _, f := range reader.File {
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
	return nil, fmt.Errorf("part bundle has no %s", name)
}

// slugOfTag turns an image tag into a directory name. Colons and slashes are
// legal in tags and not in every filesystem a bundle gets unpacked on.
func slugOfTag(tag string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			return r
		default:
			return '-'
		}
	}, tag)
}
