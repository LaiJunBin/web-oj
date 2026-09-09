package dto

// ProjectExport is a project loaded whole, for writing to a bundle.
//
// It exists alongside ProjectDetail rather than reusing it because the two
// have opposite requirements. ProjectDetail is shaped for a screen: it groups
// lifecycles by moment, drops an asset's description, and omits a grading
// item's scoring strategy — all reasonable, because the screen that consumes
// it does not show them. An export that dropped the same fields would produce
// a bundle that imports cleanly into a subtly different problem.
//
// Ids are carried because dependencies are stored as ids and the exporter has
// to turn them into bundle keys. They do not reach the bundle itself.
type ProjectExport struct {
	ID            int32
	Name          string
	Description   string
	ImageTag      string
	TestImageTag  string
	Env           string
	TestEnv       string
	ResultParser  string
	DynamicFields []ExportDynamicField
	Lifecycles    []ExportLifecycle
	GradingItems  []ExportGradingItem
	Assets        []ExportAsset
}

type ExportDynamicField struct {
	Type         string
	Name         string
	Label        string
	DefaultValue *string
	IsRequired   bool
}

type ExportLifecycle struct {
	Name        string
	Description string
	Type        string
	RunsOn      string
	Command     string
}

type ExportGradingItem struct {
	ID              int32
	Label           string
	Name            string
	Description     string
	Score           float64
	ScoringStrategy string
	OutputFormat    string
	ResultParser    string
	// DependTestcaseIDs may name testcases belonging to any grading item in
	// the project, not only this one.
	DependTestcaseIDs []int32
	Testcases         []ExportTestcase
}

type ExportTestcase struct {
	ID                int32
	Name              string
	Command           string
	Score             float64
	IsSample          bool
	DependTestcaseIDs []int32
}

type ExportAsset struct {
	// ID is also the file name under the asset storage directory.
	ID          string
	Name        string
	Description string
	IsEnabled   bool
}
