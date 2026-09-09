// Package bundle is the composite bundle: one zip carrying a whole sitting —
// its images, its problems and the contest that arranges them.
//
// It does not define a fourth format. Inside, each part is exactly the bundle
// that part's own endpoint produces, laid out in directories:
//
//	manifest.json
//	contest.json                     (optional)
//	projects/<slug>/project.json     + assets/, grading-files.zip
//	images/<slug>/image.json         + context.zip or image.tar
//
// So projects/<slug>/ IS a project bundle and images/<slug>/ IS an image
// bundle — the same bytes those endpoints emit and accept. That is what keeps
// this package about orchestration rather than about serialisation: it decides
// what order things happen in and what to tell the operator, and delegates
// every question of "what does this field mean" to the domain that owns it.
package bundle

import bundlecodec "judge-backend/internal/app/shared/types/bundle_codec"

// FormatVersion is the composite manifest's own version. It is independent of
// the part formats: this one changes when the LAYOUT changes, and a part's
// version changes when that part's fields do.
const FormatVersion = 1

const KindComposite = "composite"

const (
	ManifestFile = "manifest.json"
	ContestFile  = "contest.json"
	ProjectsDir  = "projects"
	ImagesDir    = "images"
)

// Manifest is the index. It exists so a reader knows what is inside without
// walking the archive, and so the importer can refuse a bundle whose layout it
// does not understand before reading anything else.
type Manifest struct {
	FormatVersion int    `json:"formatVersion"`
	Kind          string `json:"kind"`

	// Images and Projects are the directory names under ImagesDir and
	// ProjectsDir. Listed rather than discovered by scanning: a directory the
	// manifest does not mention is a bundle that was assembled wrongly, and
	// silently importing it would import something nobody described.
	Images   []string `json:"images"`
	Projects []string `json:"projects"`

	// HasContest says whether ContestFile is present. A composite bundle
	// without one is a way to ship a set of problems and their environments
	// without dictating a sitting.
	HasContest bool `json:"hasContest"`

	// ExportedAt is provenance and nothing reads it back. Unlike the part
	// bundles, whose definitions are kept free of timestamps so they stay
	// diffable, the manifest is an index rather than a definition — a changed
	// timestamp here says the archive was rebuilt, which is true.
	ExportedAt string `json:"exportedAt"`
}

func MarshalManifest(m *Manifest) ([]byte, error) { return bundlecodec.Marshal(m) }

func UnmarshalManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := bundlecodec.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
