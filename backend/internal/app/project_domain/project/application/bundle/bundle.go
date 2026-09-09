// Package bundle is the on-disk form of a project: the JSON schema a project
// is exported to and imported from, plus the layout of the zip that carries
// it.
//
// It exists so a venue can move a problem between installations without
// retyping its grading items and testcases, and without anybody editing the
// seeder. That means the format has two audiences, and both constrain it:
//
//   - the exporter and importer, which need it complete and unambiguous
//   - a human, who may hand-write a bundle in a text editor and diff two
//     versions of one in review
//
// The second audience is why project.json holds only the definition and never
// a timestamp or a source database id. Export it twice and you get the same
// bytes, so a diff shows what somebody changed rather than when they ran the
// export. Provenance goes in meta.json, which the importer ignores.
package bundle

// FormatVersion is the schema version written into every bundle.
//
// EVERY schema change bumps it, additive ones included. That follows from
// Unmarshal rejecting unknown fields (see its comment for why): a bundle
// carrying a field this build has never heard of cannot be read, so it must
// not claim to be a version this build supports. The importer refuses any
// version above its own rather than guess, because a bundle from the future
// can look perfectly parseable and mean something else.
const FormatVersion = 1

// KindProject marks a bundle carrying one project definition.
const KindProject = "project"

// File names inside a project bundle's zip.
const (
	// ProjectFile is the definition. Deterministic: two exports of an
	// unchanged project produce identical bytes.
	ProjectFile = "project.json"
	// MetaFile is provenance only, and the importer ignores it. Keeping it
	// out of ProjectFile is what makes that determinism possible.
	MetaFile = "meta.json"
	// AssetsDir holds the contestant-downloadable attachments, one file per
	// asset, named by Asset.File.
	AssetsDir = "assets"
	// GradingFilesFile is the grading project directory, zipped. It is
	// copied into the test container at judge time, so it is part of the
	// problem, not an attachment.
	GradingFilesFile = "grading-files.zip"
)

// Meta is provenance. Written on export, ignored on import — nothing here may
// change what gets created, or hand-written bundles would need to fake it.
type Meta struct {
	ExportedAt string `json:"exportedAt"`
	// SourceProjectID is the id on the machine that produced the bundle. It
	// is a breadcrumb for support ("which project did this come from?"), not
	// an identifier the importer may resolve — ids are meaningless across
	// installations.
	SourceProjectID int32 `json:"sourceProjectId"`
}

// Project is the definition of one problem.
type Project struct {
	FormatVersion int    `json:"formatVersion"`
	Kind          string `json:"kind"`

	Name        string `json:"name"`
	Description string `json:"description"`

	// ImageTag and TestImageTag reference images by tag, never by content.
	// An image is built on the machine that runs the contest, so a project
	// bundle names the image it needs and the importer reports any that are
	// missing; carrying the image itself is the image bundle's job.
	ImageTag     string `json:"imageTag"`
	TestImageTag string `json:"testImageTag"`

	Env     string `json:"env"`
	TestEnv string `json:"testEnv"`

	ResultParser string `json:"resultParser"`

	DynamicFields []DynamicField `json:"dynamicFields"`
	Lifecycles    []Lifecycle    `json:"lifecycles"`
	GradingItems  []GradingItem  `json:"gradingItems"`
	Assets        []Asset        `json:"assets"`

	// HasGradingFiles says whether GradingFilesFile is present in the zip. A
	// flag rather than a path: there is exactly one such file and its name is
	// fixed, so a path would be a second place for the two to disagree.
	HasGradingFiles bool `json:"hasGradingFiles"`
}

// DynamicField is one input the contestant fills in when submitting.
type DynamicField struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Label string `json:"label"`
	// DefaultValue is null when the field has none, which is distinct from an
	// empty default.
	DefaultValue *string `json:"defaultValue"`
	IsRequired   bool    `json:"isRequired"`
}

// Lifecycle is a command run around the testcases.
type Lifecycle struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	RunsOn      string `json:"runsOn"`
	Command     string `json:"command"`
}

// GradingItem is one scored item, holding its own testcases.
//
// Order is not a field: it is the position in the array. Carrying both invites
// a bundle whose array order and order values disagree, and then every reader
// has to decide which one lies.
type GradingItem struct {
	Label           string  `json:"label"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Score           float64 `json:"score"`
	ScoringStrategy string  `json:"scoringStrategy"`
	OutputFormat    string  `json:"outputFormat"`
	ResultParser    string  `json:"resultParser"`

	// DependsOnTestcases holds Testcase.Key values, which may name testcases
	// under any grading item in this bundle.
	DependsOnTestcases []string `json:"dependsOnTestcases"`

	Testcases []Testcase `json:"testcases"`
}

// Testcase is one command whose outcome is scored.
type Testcase struct {
	// Key identifies this testcase within the bundle so dependencies can
	// point at it. Database ids cannot serve: they are assigned by the
	// machine that happens to hold the project and mean nothing anywhere
	// else. Unique across the whole bundle, not just its grading item.
	Key string `json:"key"`

	Name     string  `json:"name"`
	Command  string  `json:"command"`
	Score    float64 `json:"score"`
	IsSample bool    `json:"isSample"`

	// DependsOnTestcases holds other Testcase.Key values. Forward references
	// are allowed, which is why the importer resolves keys in a second pass.
	DependsOnTestcases []string `json:"dependsOnTestcases"`
}

// Asset is a contestant-downloadable attachment.
//
// No description field, although the assets table has that column: the
// repository writes the display name into it and nothing in the product can
// set it to anything else, so it carries no information the name does not.
// Exporting it would promise a fidelity the import cannot deliver — the
// creation path has nowhere to put it back.
type Asset struct {
	Name      string `json:"name"`
	IsEnabled bool   `json:"isEnabled"`
	// File is the path inside the zip, under AssetsDir. Stored explicitly
	// because an asset's display name can contain characters a file name
	// cannot, and two assets may share a display name.
	File string `json:"file"`
}
