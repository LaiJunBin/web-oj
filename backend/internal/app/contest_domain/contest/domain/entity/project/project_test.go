package project_test

import (
	"testing"
	"time"

	"judge-backend/internal/app/contest_domain/contest/domain/entity/project"
)

func mustProject(t *testing.T, start, end time.Time) *project.Project {
	t.Helper()
	p, err := project.New(&project.ProjectConfig{
		ID:             1,
		Name:           "Project1",
		ActivateTime:   start.Add(-time.Hour),
		DeactivateTime: end.Add(time.Hour),
		StartTime:      start,
		EndTime:        end,
	})
	if err != nil {
		t.Fatalf("failed to build project: %v", err)
	}
	return p
}

// TestProject_SubmissionBoundaries pins the submission-window boundary
// semantics to the read side's model.ContestProject (strict inequalities on
// the server clock): exactly at StartTime the window has NOT opened yet
// (IsStarted = StartTime.Before(now)), and exactly at EndTime it is STILL
// open (IsEnded = EndTime.Before(now)). The submit gate must agree with what
// contestants see, so these four instants are pinned to the nanosecond.
func TestProject_SubmissionBoundaries(t *testing.T) {
	start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 1, 17, 0, 0, 0, time.UTC)
	p := mustProject(t, start, end)

	cases := []struct {
		name         string
		now          time.Time
		wantNotStart bool
		wantEnded    bool
	}{
		{"well before the window", start.Add(-time.Hour), true, false},
		{"exactly at StartTime the window is not yet open", start, true, false},
		{"one instant after StartTime it is open", start.Add(time.Nanosecond), false, false},
		{"mid-window", start.Add(4 * time.Hour), false, false},
		{"exactly at EndTime it is still open", end, false, false},
		{"one instant after EndTime it is closed", end.Add(time.Nanosecond), false, true},
		{"well after the window", end.Add(time.Hour), false, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := p.IsSubmissionNotStarted(c.now); got != c.wantNotStart {
				t.Errorf("IsSubmissionNotStarted(%v) = %v, want %v", c.now, got, c.wantNotStart)
			}
			if got := p.IsSubmissionEnded(c.now); got != c.wantEnded {
				t.Errorf("IsSubmissionEnded(%v) = %v, want %v", c.now, got, c.wantEnded)
			}
		})
	}
}
