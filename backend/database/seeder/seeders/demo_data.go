package seeders

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// demoDataJSON is embedded rather than read from disk: every seed entrypoint
// (dev server boot, cmd/seed, test harness) runs from a different working
// directory, and the demo data is committed alongside the code anyway. An
// embedded file cannot go missing at runtime, so there is no "roster not
// found" fallback to keep in sync.
//
//go:embed demo_data.json
var demoDataJSON []byte

// DemoData is the whole built-in demonstration dataset — the one thing a
// freshly cloned checkout shows on first login. Editing demo_data.json is the
// supported way to change it; these structs only mirror that file's shape.
type DemoData struct {
	// Password is shared by every demo account on purpose: it is documented
	// in the README so a first-time installer can log straight in.
	Password    string      `json:"password"`
	Admin       DemoUser    `json:"admin"`
	Contestants []DemoUser  `json:"contestants"`
	Image       DemoImage   `json:"image"`
	Project     DemoProject `json:"project"`
	Contest     DemoContest `json:"contest"`
}

type DemoUser struct {
	Name     string   `json:"name"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

type DemoImage struct {
	Tag         string `json:"tag"`
	MountPath   string `json:"mountPath"`
	Description string `json:"description"`
	IsReady     bool   `json:"isReady"`
}

type DemoProject struct {
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	ResultParser  string             `json:"resultParser"`
	DynamicFields []DemoDynamicField `json:"dynamicFields"`
	Lifecycles    []DemoLifecycle    `json:"lifecycles"`
	GradingItems  []DemoGradingItem  `json:"gradingItems"`
}

type DemoDynamicField struct {
	Type         string  `json:"type"`
	Name         string  `json:"name"`
	Label        string  `json:"label"`
	DefaultValue *string `json:"defaultValue"`
	IsRequired   bool    `json:"isRequired"`
}

type DemoLifecycle struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	RunsOn      string `json:"runsOn"`
	Command     string `json:"command"`
}

type DemoGradingItem struct {
	Label           string         `json:"label"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	Score           float64        `json:"score"`
	ScoringStrategy string         `json:"scoringStrategy"`
	OutputFormat    string         `json:"outputFormat"`
	Testcases       []DemoTestcase `json:"testcases"`
}

type DemoTestcase struct {
	Name     string  `json:"name"`
	Command  string  `json:"command"`
	Score    float64 `json:"score"`
	IsSample bool    `json:"isSample"`
}

// DemoContest carries offsets instead of timestamps so the demo contest is
// live whenever it happens to be seeded, rather than expiring at whatever
// wall-clock time the dataset was authored.
type DemoContest struct {
	Name                   string             `json:"name"`
	Description            string             `json:"description"`
	IsPublic               bool               `json:"isPublic"`
	ActivateAfterMinutes   int                `json:"activateAfterMinutes"`
	DeactivateAfterMinutes int                `json:"deactivateAfterMinutes"`
	Project                DemoContestProject `json:"project"`
}

type DemoContestProject struct {
	Name                   string `json:"name"`
	IsPublic               bool   `json:"isPublic"`
	ActivateAfterMinutes   int    `json:"activateAfterMinutes"`
	DeactivateAfterMinutes int    `json:"deactivateAfterMinutes"`
	StartAfterMinutes      int    `json:"startAfterMinutes"`
	EndAfterMinutes        int    `json:"endAfterMinutes"`
	JudgePriorityStrategy  string `json:"judgePriorityStrategy"`
}

// Users lists the admin first and then the contestants, which is the order
// they are created in and therefore the order of their auto-increment IDs.
func (d *DemoData) Users() []DemoUser {
	users := make([]DemoUser, 0, len(d.Contestants)+1)
	users = append(users, d.Admin)
	return append(users, d.Contestants...)
}

// Usernames returns every demo login in the same order as Users.
func (d *DemoData) Usernames() []string {
	users := d.Users()
	names := make([]string, 0, len(users))
	for _, u := range users {
		names = append(names, u.Username)
	}
	return names
}

// After turns a "minutes from seed time" offset into a timestamp. Passing the
// same base to every field keeps one contest's window internally consistent
// even if the seed run straddles a minute boundary.
func After(base time.Time, minutes int) time.Time {
	return base.Add(time.Duration(minutes) * time.Minute)
}

var (
	demoDataOnce sync.Once
	demoData     *DemoData
	demoDataErr  error
)

// LoadDemoData parses and validates the embedded dataset once per process.
// Every seeder shares the result, so a malformed file fails the first seeder
// rather than half-seeding a broken contest.
func LoadDemoData() (*DemoData, error) {
	demoDataOnce.Do(func() {
		demoData, demoDataErr = parseDemoData(demoDataJSON)
	})
	return demoData, demoDataErr
}

func parseDemoData(raw []byte) (*DemoData, error) {
	var data DemoData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse demo_data.json: %w", err)
	}

	if err := data.validate(); err != nil {
		return nil, fmt.Errorf("demo_data.json: %w", err)
	}

	return &data, nil
}

// validate rejects the mistakes that would otherwise surface much later as a
// confusing empty scoreboard or an un-runnable judgment: a missing password,
// duplicate logins, a grading item with no testcases, an empty command.
func (d *DemoData) validate() error {
	if d.Password == "" {
		return fmt.Errorf("password must not be empty")
	}

	if len(d.Contestants) == 0 {
		return fmt.Errorf("at least one contestant is required for the scoreboard to show anything")
	}

	seen := make(map[string]struct{}, len(d.Contestants)+1)
	for _, u := range d.Users() {
		if u.Name == "" || u.Username == "" {
			return fmt.Errorf("user %q must have both name and username", u.Username)
		}
		if len(u.Roles) == 0 {
			return fmt.Errorf("user %q must have at least one role", u.Username)
		}
		if _, dup := seen[u.Username]; dup {
			return fmt.Errorf("duplicate username %q", u.Username)
		}
		seen[u.Username] = struct{}{}
	}

	if d.Image.Tag == "" || d.Image.MountPath == "" {
		return fmt.Errorf("image needs both tag and mountPath")
	}

	if d.Project.Name == "" {
		return fmt.Errorf("project name must not be empty")
	}

	if len(d.Project.GradingItems) == 0 {
		return fmt.Errorf("project %q needs at least one grading item", d.Project.Name)
	}

	for _, item := range d.Project.GradingItems {
		if item.Name == "" {
			return fmt.Errorf("every grading item needs a name")
		}
		if len(item.Testcases) == 0 {
			return fmt.Errorf("grading item %q needs at least one testcase", item.Name)
		}
		for _, tc := range item.Testcases {
			if tc.Name == "" || tc.Command == "" {
				return fmt.Errorf("testcase of grading item %q needs both name and command", item.Name)
			}
		}
	}

	for _, lc := range d.Project.Lifecycles {
		if lc.Name == "" || lc.Command == "" {
			return fmt.Errorf("every lifecycle needs both name and command")
		}
	}

	if d.Contest.Name == "" {
		return fmt.Errorf("contest name must not be empty")
	}

	if d.Contest.Project.Name == "" {
		return fmt.Errorf("contest %q needs a contest project name", d.Contest.Name)
	}

	// An already-ended demo contest is the exact papercut this dataset exists
	// to avoid, so it is a load error rather than something to discover after
	// installing.
	if d.Contest.Project.EndAfterMinutes <= d.Contest.Project.StartAfterMinutes {
		return fmt.Errorf("contest project %q must end after it starts", d.Contest.Project.Name)
	}

	return nil
}
