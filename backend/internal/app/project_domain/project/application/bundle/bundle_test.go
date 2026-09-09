package bundle_test

import (
	"strings"
	"testing"

	"judge-backend/internal/app/project_domain/project/application/bundle"
)

// allParsersValid stands in for the deployment's registry. The real one is
// injected because the parsers are registered by the judge domain.
type allParsersValid struct{}

func (allParsersValid) IsValid(string) bool { return true }

type noParsersValid struct{}

func (noParsersValid) IsValid(string) bool { return false }

func valid() *bundle.Project {
	def := "value"
	return &bundle.Project{
		FormatVersion: bundle.FormatVersion,
		Kind:          bundle.KindProject,
		Name:          "靜態網站部署",
		Description:   "說明",
		ImageTag:      "test-nginx:latest",
		TestImageTag:  "test-nginx:latest",
		Env:           "FOO=bar",
		TestEnv:       "",
		ResultParser:  "auto",
		DynamicFields: []bundle.DynamicField{
			{Type: "text", Name: "repoUrl", Label: "Repo", DefaultValue: &def, IsRequired: true},
		},
		Lifecycles: []bundle.Lifecycle{
			{Name: "安裝", Description: "", Type: "before_all", RunsOn: "project", Command: "npm ci"},
		},
		GradingItems: []bundle.GradingItem{{
			Label: "1", Name: "設定檔", Description: "", Score: 10,
			ScoringStrategy: "decrease", OutputFormat: "text",
			DependsOnTestcases: []string{},
			Testcases: []bundle.Testcase{
				{Key: "a", Name: "測資 A", Command: "true", Score: 4, IsSample: true, DependsOnTestcases: []string{}},
				{Key: "b", Name: "測資 B", Command: "true", Score: 6, DependsOnTestcases: []string{"a"}},
			},
		}},
		Assets: []bundle.Asset{
			{Name: "規格書.pdf", IsEnabled: true, File: "assets/spec.pdf"},
		},
		HasGradingFiles: true,
	}
}

func paths(ps bundle.Problems) string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.Path
	}
	return strings.Join(out, ",")
}

func TestValidBundleHasNoProblems(t *testing.T) {
	if got := valid().Validate(allParsersValid{}); len(got) != 0 {
		t.Fatalf("expected a clean bundle, but got: %s", got.Error())
	}
}

// A dependency may name a testcase declared later in the file. Requiring
// declaration order would make the format depend on how the exporter happened
// to walk the grading items.
func TestForwardDependencyIsAllowed(t *testing.T) {
	b := valid()
	b.GradingItems[0].Testcases[0].DependsOnTestcases = []string{"b"}
	b.GradingItems[0].Testcases[1].DependsOnTestcases = nil

	if got := b.Validate(allParsersValid{}); len(got) != 0 {
		t.Fatalf("expected forward references to be allowed, but got: %s", got.Error())
	}
}

func TestValidateReportsEveryFaultAtOnce(t *testing.T) {
	b := valid()
	b.Name = ""
	b.ImageTag = ""
	b.GradingItems[0].Label = ""
	b.GradingItems[0].Testcases[0].Command = ""

	got := b.Validate(allParsersValid{})

	// Four independent faults. Reporting one per upload would turn fixing a
	// hand-written bundle into four round trips.
	if len(got) != 4 {
		t.Fatalf("expected all 4 faults in one pass, but got %d: %s", len(got), got.Error())
	}
}

func TestValidateRejects(t *testing.T) {
	cases := map[string]struct {
		mutate   func(*bundle.Project)
		wantPath string
	}{
		"a version this build cannot read": {
			func(b *bundle.Project) { b.FormatVersion = bundle.FormatVersion + 1 },
			"formatVersion",
		},
		"the wrong kind of bundle": {
			func(b *bundle.Project) { b.Kind = "contest" },
			"kind",
		},
		"an unknown scoring strategy": {
			func(b *bundle.Project) { b.GradingItems[0].ScoringStrategy = "average" },
			"gradingItems[0].scoringStrategy",
		},
		"an unknown output format": {
			func(b *bundle.Project) { b.GradingItems[0].OutputFormat = "markdown" },
			"gradingItems[0].outputFormat",
		},
		"an unknown lifecycle moment": {
			func(b *bundle.Project) { b.Lifecycles[0].Type = "during" },
			"lifecycles[0].type",
		},
		"an unknown execution environment": {
			func(b *bundle.Project) { b.Lifecycles[0].RunsOn = "host" },
			"lifecycles[0].runsOn",
		},
		"an unknown dynamic field type": {
			func(b *bundle.Project) { b.DynamicFields[0].Type = "date" },
			"dynamicFields[0].type",
		},
		"a testcase with no key": {
			func(b *bundle.Project) { b.GradingItems[0].Testcases[0].Key = "" },
			"gradingItems[0].testcases[0].key",
		},
		"two testcases sharing a key": {
			func(b *bundle.Project) { b.GradingItems[0].Testcases[1].Key = "a" },
			"gradingItems[0].testcases[1].key",
		},
		"a dependency on a key nothing declares": {
			func(b *bundle.Project) { b.GradingItems[0].Testcases[1].DependsOnTestcases = []string{"ghost"} },
			"gradingItems[0].testcases[1].dependsOnTestcases[0]",
		},
		"a grading item depending on a key nothing declares": {
			func(b *bundle.Project) { b.GradingItems[0].DependsOnTestcases = []string{"ghost"} },
			"gradingItems[0].dependsOnTestcases[0]",
		},
		"a testcase depending on itself": {
			func(b *bundle.Project) { b.GradingItems[0].Testcases[0].DependsOnTestcases = []string{"a"} },
			"gradingItems[0].testcases[0].dependsOnTestcases[0]",
		},
		"a negative score": {
			func(b *bundle.Project) { b.GradingItems[0].Score = -1 },
			"gradingItems[0].score",
		},
		// The file name is joined onto a directory and written, so a bundle
		// must not be able to aim it outside.
		"an asset escaping the assets directory": {
			func(b *bundle.Project) { b.Assets[0].File = "assets/../../etc/passwd" },
			"assets[0].file",
		},
		"an asset outside the assets directory": {
			func(b *bundle.Project) { b.Assets[0].File = "spec.pdf" },
			"assets[0].file",
		},
		"two assets claiming one file": {
			func(b *bundle.Project) {
				b.Assets = append(b.Assets, bundle.Asset{Name: "另一份", File: "assets/spec.pdf"})
			},
			"assets[1].file",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			b := valid()
			tc.mutate(b)

			got := b.Validate(allParsersValid{})
			for _, p := range got {
				if p.Path == tc.wantPath {
					return
				}
			}
			t.Errorf("expected a problem at %s, but got: %s", tc.wantPath, paths(got))
		})
	}
}

// The parser registry is a property of the deployment, so the same bundle is
// valid on one installation and not on another. That has to be reported, not
// assumed away.
func TestValidateUsesTheInjectedParserRegistry(t *testing.T) {
	b := valid()

	if got := b.Validate(noParsersValid{}); len(got) == 0 {
		t.Fatal("expected the unknown parser to be rejected")
	}

	b.ResultParser = ""
	if got := b.Validate(noParsersValid{}); len(got) != 0 {
		t.Fatalf("an empty parser means 'use the default' and must be accepted, but got: %s", got.Error())
	}
}

// A cycle deadlocks the scheduler: nothing in it ever becomes runnable, and
// the judgment burns its whole wall-clock cap before failing.
func TestValidateRejectsDependencyCycles(t *testing.T) {
	b := valid()
	b.GradingItems[0].Testcases = []bundle.Testcase{
		{Key: "a", Name: "A", Command: "true", DependsOnTestcases: []string{"c"}},
		{Key: "b", Name: "B", Command: "true", DependsOnTestcases: []string{"a"}},
		{Key: "c", Name: "C", Command: "true", DependsOnTestcases: []string{"b"}},
	}

	got := b.Validate(allParsersValid{})
	if len(got) == 0 {
		t.Fatal("expected the three-testcase cycle to be rejected")
	}
	if !strings.Contains(got.Error(), "循環") {
		t.Errorf("expected the message to name the cycle, but got: %s", got.Error())
	}
}

// The counterpart, and the one a naive cycle check gets wrong: a diamond is
// not a cycle. Rejecting it would forbid a perfectly ordinary shape — two
// testcases that both need the same setup step.
func TestValidateAcceptsADiamond(t *testing.T) {
	b := valid()
	b.GradingItems[0].Testcases = []bundle.Testcase{
		{Key: "setup", Name: "設定", Command: "true"},
		{Key: "left", Name: "左", Command: "true", DependsOnTestcases: []string{"setup"}},
		{Key: "right", Name: "右", Command: "true", DependsOnTestcases: []string{"setup"}},
		{Key: "join", Name: "合", Command: "true", DependsOnTestcases: []string{"left", "right"}},
	}

	if got := b.Validate(allParsersValid{}); len(got) != 0 {
		t.Fatalf("a diamond is not a cycle, but got: %s", got.Error())
	}
}

// A self-dependency is already reported against the exact dependency that
// caused it. Letting the cycle finder report it again, under a path naming no
// particular testcase, gives the reader a second vaguer copy of a message they
// have already acted on.
func TestSelfDependencyIsReportedOnce(t *testing.T) {
	b := valid()
	b.GradingItems[0].Testcases[0].DependsOnTestcases = []string{"a"}

	got := b.Validate(allParsersValid{})

	if len(got) != 1 {
		t.Fatalf("expected exactly one problem, but got %d: %s", len(got), got.Error())
	}
	if got[0].Path != "gradingItems[0].testcases[0].dependsOnTestcases[0]" {
		t.Errorf("expected the precise path, but got %s", got[0].Path)
	}
}

// project.json goes in version control and gets reviewed as a diff, so two
// exports of an unchanged project must produce identical bytes. Anything that
// varies run to run — a map, a timestamp, a database id — breaks that, and
// breaks it quietly: the file still imports, it just shows spurious changes in
// every review.
func TestMarshalIsDeterministic(t *testing.T) {
	first, err := bundle.Marshal(valid())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for i := 0; i < 20; i++ {
		again, err := bundle.Marshal(valid())
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if string(again) != string(first) {
			t.Fatalf("run %d produced different bytes:\n%s\n---\n%s", i, first, again)
		}
	}
}

// The property that catches a dropped field. A field missing from Marshal, or
// from the struct, survives every hand-written assertion but cannot survive a
// round trip.
func TestMarshalUnmarshalRoundTripsToIdenticalBytes(t *testing.T) {
	first, err := bundle.Marshal(valid())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	parsed, err := bundle.Unmarshal(first)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	second, err := bundle.Marshal(parsed)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}

	if string(second) != string(first) {
		t.Errorf("round trip changed the bytes:\n%s\n---\n%s", first, second)
	}
}

// The typos that hurt are on optional fields: "isSampel": true leaves isSample
// false, the import succeeds, and the mistake surfaces later as a testcase
// mysteriously hidden from contestants.
func TestUnmarshalRejectsUnknownFields(t *testing.T) {
	first, err := bundle.Marshal(valid())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	typoed := strings.Replace(string(first), `"isSample"`, `"isSampel"`, 1)
	if typoed == string(first) {
		t.Fatal("the fixture no longer contains isSample; this test is not exercising anything")
	}

	if _, err := bundle.Unmarshal([]byte(typoed)); err == nil {
		t.Error("expected a misspelled field to be rejected, but it parsed")
	}
}

// Commands are shell, and shell is full of < > &. Escaping them leaves a file
// that is technically correct and unreadable to whoever has to edit it.
func TestMarshalDoesNotEscapeShellSyntax(t *testing.T) {
	b := valid()
	b.GradingItems[0].Testcases[0].Command = `sh -c "grep -q 'a<b && c>d' /tmp/x"`

	out, err := bundle.Marshal(b)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Go escapes < > & by default, which would render this command as
	// 'a\u003cb \u0026\u0026 c\u003ed' — still correct, still importable,
	// and no longer something a person can read or edit.
	if !strings.Contains(string(out), `a<b && c>d`) {
		t.Errorf("expected the command to survive verbatim, but got:\n%s", out)
	}
	if strings.Contains(string(out), `\u003c`) || strings.Contains(string(out), `\u0026`) {
		t.Errorf("expected no HTML escaping, but got:\n%s", out)
	}
}

// null and "" are different answers to "what does this field default to", and
// the difference reaches the contestant's submission form.
func TestNullDefaultValueSurvivesTheRoundTrip(t *testing.T) {
	b := valid()
	b.DynamicFields[0].DefaultValue = nil

	out, err := bundle.Marshal(b)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	parsed, err := bundle.Unmarshal(out)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if parsed.DynamicFields[0].DefaultValue != nil {
		t.Errorf("expected null to stay null, but got %q", *parsed.DynamicFields[0].DefaultValue)
	}
}
