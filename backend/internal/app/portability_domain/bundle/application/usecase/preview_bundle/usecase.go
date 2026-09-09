package previewbundle

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"

	"judge-backend/internal/app/portability_domain/bundle/application/bundle"
	sharedservice "judge-backend/internal/app/shared/service"
)

var (
	ErrorInternalServer = errors.New("internal server error")
	ErrorInvalidBundle  = errors.New("這不是有效的完整匯出檔")
)

type InvalidBundleError struct{ Details []string }

func (e *InvalidBundleError) Error() string { return ErrorInvalidBundle.Error() }
func (e *InvalidBundleError) Unwrap() error { return ErrorInvalidBundle }

// Item statuses, matching the API's enum.
const (
	StatusNew       = "new"
	StatusIdentical = "identical"
	StatusConflict  = "conflict"
)

type Input struct {
	Bundle *bytes.Buffer
}

type Output struct {
	FormatVersion int
	Images        []ImageItem
	Projects      []ProjectItem
	Contest       *ContestItem
	// Blockers is what stops the import. Collected separately from the items
	// so the caller decides whether to offer "run" by looking at one list,
	// rather than by re-deriving it from three.
	Blockers []string
}

type ImageItem struct {
	Tag            string
	Mode           string
	Status         string
	UsedByProjects []string
}

type ProjectItem struct {
	Name             string
	Status           string
	MissingImageTags []string
}

type ContestItem struct {
	Name               string
	ProjectSlots       int32
	UnresolvedProjects []string
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

// The three domains this reads from. Declared here as ports and bound in the
// composition root, so the composite bundle can ask each domain about its own
// world without any of them learning that composite bundles exist.

// ImageInspector answers what an image bundle would do to this installation.
type ImageInspector interface {
	// Inspect reports the status of a tag against the installed image, and
	// which projects would be affected if it were overwritten.
	Inspect(tag, mode, fingerprint string) (status string, usedBy []string, err error)
}

// ProjectInspector answers the same for a project bundle.
type ProjectInspector interface {
	// NameExists reports whether the library already holds a project with
	// this name. Imports always create, so this is advisory — it tells the
	// operator they are about to end up with two.
	NameExists(name string) (bool, error)
	// TagsPresent reports which of the given image tags this installation
	// already has, so the preview can subtract the ones the bundle carries.
	TagsPresent(tags []string) (map[string]bool, error)
}

// PartReader parses the part bundles. The composite package deliberately does
// not parse project.json or image.json itself: those belong to the domains
// that own them, and a second parser would drift from the first.
type PartReader interface {
	ReadProject(definition []byte) (name string, imageTags []string, err error)
	ReadImage(definition []byte) (tag, mode, fingerprint string, err error)
	ReadContest(definition []byte) (name string, projectRefs []string, err error)
}

type usecase struct {
	zip      sharedservice.ZipService
	images   ImageInspector
	projects ProjectInspector
	parts    PartReader
}

func NewUsecase(
	zip sharedservice.ZipService,
	images ImageInspector,
	projects ProjectInspector,
	parts PartReader,
) Usecase {
	return &usecase{zip: zip, images: images, projects: projects, parts: parts}
}

// Execute reads the archive and reports what importing it would do. It writes
// nothing, which is the whole point: the operator sees the consequences —
// including which live problems an overwritten image would affect — before
// deciding, rather than being handed an automatic judgement to trust.
func (u *usecase) Execute(input *Input) (*Output, error) {
	contents, manifest, err := Read(u.zip, input.Bundle)
	if err != nil {
		return nil, err
	}
	return u.describe(contents, manifest)
}

// Contents is a composite archive, already read and indexed.
type Contents struct {
	Files    map[string][]byte
	Manifest *bundle.Manifest
}

// Read opens and validates the archive's shape. Shared with the import
// usecase so both agree on what a well-formed composite bundle is.
func Read(zipService sharedservice.ZipService, archive *bytes.Buffer) (*Contents, *bundle.Manifest, error) {
	if archive == nil || archive.Len() == 0 {
		return nil, nil, &InvalidBundleError{Details: []string{"沒有收到檔案"}}
	}
	// ReadZipToMap bounds total size and entry count, so a composite bundle —
	// which fans out into several child importers — cannot be a zip bomb read
	// straight into memory here.
	files, err := zipService.ReadZipToMap(archive)
	if err != nil {
		return nil, nil, &InvalidBundleError{Details: []string{"上傳的檔案不是有效或過大的 zip"}}
	}

	manifestJSON, ok := files[bundle.ManifestFile]
	if !ok {
		return nil, nil, &InvalidBundleError{
			Details: []string{"匯出檔裡沒有 " + bundle.ManifestFile + "，這看起來不是「完整匯出」產生的檔案"},
		}
	}

	manifest, err := bundle.UnmarshalManifest(manifestJSON)
	if err != nil {
		return nil, nil, &InvalidBundleError{Details: []string{bundle.ManifestFile + " 無法解析：" + err.Error()}}
	}

	details := make([]string, 0)
	if manifest.FormatVersion <= 0 {
		details = append(details, "manifest.formatVersion: missing")
	} else if manifest.FormatVersion > bundle.FormatVersion {
		details = append(details, fmt.Sprintf(
			"manifest.formatVersion: 此匯出檔是 %d 版，本系統只支援到 %d 版，請升級後再匯入",
			manifest.FormatVersion, bundle.FormatVersion))
	}
	if manifest.Kind != bundle.KindComposite {
		details = append(details, fmt.Sprintf("manifest.kind: 必須是 %q，得到 %q", bundle.KindComposite, manifest.Kind))
	}

	// Every part the manifest names has to be there. A manifest that promises
	// a problem the archive does not contain would otherwise be discovered
	// half way through the import, with images already created.
	for _, slug := range manifest.Images {
		if _, ok := files[path.Join(bundle.ImagesDir, slug, "image.json")]; !ok {
			details = append(details, fmt.Sprintf("manifest 列出映像 %q，但匯出檔裡沒有 %s/%s/image.json", slug, bundle.ImagesDir, slug))
		}
	}
	for _, slug := range manifest.Projects {
		if _, ok := files[path.Join(bundle.ProjectsDir, slug, "project.json")]; !ok {
			details = append(details, fmt.Sprintf("manifest 列出題目 %q，但匯出檔裡沒有 %s/%s/project.json", slug, bundle.ProjectsDir, slug))
		}
	}
	if manifest.HasContest {
		if _, ok := files[bundle.ContestFile]; !ok {
			details = append(details, "manifest 說有競賽，但匯出檔裡沒有 "+bundle.ContestFile)
		}
	}

	if len(details) > 0 {
		return nil, nil, &InvalidBundleError{Details: details}
	}

	return &Contents{Files: files, Manifest: manifest}, manifest, nil
}

func (u *usecase) describe(contents *Contents, manifest *bundle.Manifest) (*Output, error) {
	out := &Output{
		FormatVersion: manifest.FormatVersion,
		Images:        make([]ImageItem, 0, len(manifest.Images)),
		Projects:      make([]ProjectItem, 0, len(manifest.Projects)),
		Blockers:      make([]string, 0),
	}

	// Images first, and their tags remembered: a project naming an image this
	// installation lacks is fine as long as the same bundle brings it.
	carried := map[string]bool{}
	for _, slug := range manifest.Images {
		definition := contents.Files[path.Join(bundle.ImagesDir, slug, "image.json")]
		tag, mode, print, err := u.parts.ReadImage(definition)
		if err != nil {
			return nil, &InvalidBundleError{Details: []string{fmt.Sprintf("%s/%s/image.json 無法解析：%v", bundle.ImagesDir, slug, err)}}
		}

		status, usedBy, err := u.images.Inspect(tag, mode, print)
		if err != nil {
			return nil, ErrorInternalServer
		}
		// The API declares this a list, and a nil slice serialises as null —
		// which reads as "unknown" where the truth is "nobody". Normalised
		// here because this usecase owns the output's shape; leaving it to
		// the presenter would leave the next presenter to forget.
		if usedBy == nil {
			usedBy = []string{}
		}

		carried[tag] = true
		out.Images = append(out.Images, ImageItem{Tag: tag, Mode: mode, Status: status, UsedByProjects: usedBy})

		if status == StatusConflict {
			blocker := fmt.Sprintf("映像 %q 已存在且內容不同", tag)
			if len(usedBy) > 0 {
				blocker += fmt.Sprintf("，覆蓋會影響 %d 個題目：%s", len(usedBy), strings.Join(usedBy, "、"))
			}
			out.Blockers = append(out.Blockers, blocker)
		}
	}

	importedNames := map[string]bool{}
	for _, slug := range manifest.Projects {
		definition := contents.Files[path.Join(bundle.ProjectsDir, slug, "project.json")]
		name, tags, err := u.parts.ReadProject(definition)
		if err != nil {
			return nil, &InvalidBundleError{Details: []string{fmt.Sprintf("%s/%s/project.json 無法解析：%v", bundle.ProjectsDir, slug, err)}}
		}

		present, err := u.projects.TagsPresent(tags)
		if err != nil {
			return nil, ErrorInternalServer
		}

		missing := make([]string, 0)
		for _, tag := range tags {
			if !present[tag] && !carried[tag] {
				missing = append(missing, tag)
			}
		}
		sort.Strings(missing)

		// Always "new": a project import always creates (a colliding name gets
		// a suffix), so no other status would be true. "identical" is reserved
		// for things that genuinely will not change — an image whose content
		// matches — and using it for a name match here would tell the operator
		// "nothing will happen" about an import that adds a second copy.
		importedNames[name] = true
		out.Projects = append(out.Projects, ProjectItem{Name: name, Status: StatusNew, MissingImageTags: missing})

		if len(missing) > 0 {
			out.Blockers = append(out.Blockers, fmt.Sprintf(
				"題目 %q 需要映像 %s，但匯出檔與系統都沒有", name, strings.Join(missing, "、")))
		}
	}

	if manifest.HasContest {
		name, refs, err := u.parts.ReadContest(contents.Files[bundle.ContestFile])
		if err != nil {
			return nil, &InvalidBundleError{Details: []string{bundle.ContestFile + " 無法解析：" + err.Error()}}
		}

		unresolved := make([]string, 0)
		for _, ref := range refs {
			// A slot resolves if the bundle brings that problem, or the
			// library already holds exactly one by that name. The second half
			// is the importer's to decide — here it is enough to say the
			// bundle does not supply it.
			if importedNames[ref] {
				continue
			}
			exists, err := u.projects.NameExists(ref)
			if err != nil {
				return nil, ErrorInternalServer
			}
			if !exists {
				unresolved = append(unresolved, ref)
			}
		}

		out.Contest = &ContestItem{
			Name:               name,
			ProjectSlots:       int32(len(refs)),
			UnresolvedProjects: unresolved,
		}
		// Unresolved slots are NOT a blocker: the contest is still worth
		// creating with the slots that do resolve, and the operator gets a
		// list of what to import next.
	}

	return out, nil
}
