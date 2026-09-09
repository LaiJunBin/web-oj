// Package bundle is the on-disk form of a contest.
//
// It carries the contest and the per-problem settings that make up a sitting —
// the four times, the resource caps, the scoreboard switches — but never the
// problems themselves. Each slot names its problem, and the importer resolves
// that name against the library already installed.
//
// Naming rather than embedding is what keeps importing a contest from also
// creating problems, which is what would put the two in an order that has to
// be got right. A slot whose name resolves to nothing is reported, not
// invented: the operator imports that problem's own bundle and tries again.
package bundle

import (
	"time"

	bundlecodec "judge-backend/internal/app/shared/types/bundle_codec"
)

// FormatVersion is the schema version. Any schema change bumps it, additive
// ones included — see bundlecodec.
const FormatVersion = 1

const KindContest = "contest"

// ContestFile is the definition inside the bundle's zip.
const ContestFile = "contest.json"

// Contest is the definition of one sitting.
type Contest struct {
	FormatVersion int    `json:"formatVersion"`
	Kind          string `json:"kind"`

	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"isPublic"`

	ActivateTime   time.Time `json:"activateTime"`
	DeactivateTime time.Time `json:"deactivateTime"`

	Projects []Project `json:"projects"`
}

// Project is one problem's slot in the contest.
type Project struct {
	// ProjectRef is the name of the problem in the library. Names are not
	// unique, so this can match more than one; the importer takes that as
	// unresolved rather than guessing, because picking the wrong problem
	// produces a contest that looks configured and grades something else.
	ProjectRef string `json:"projectRef"`

	Name     string `json:"name"`
	IsPublic bool   `json:"isPublic"`

	JudgePriorityStrategy string `json:"judgePriorityStrategy"`

	ActivateTime   time.Time `json:"activateTime"`
	DeactivateTime time.Time `json:"deactivateTime"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`

	Limits     Limits     `json:"limits"`
	Scoreboard Scoreboard `json:"scoreboard"`
}

// Limits are the per-slot caps. Every field is nullable and null means "use
// the deployment default" — the same three states the API and the admin form
// use, carried across unchanged so an import does not quietly pin a value the
// original left free.
type Limits struct {
	MaxJudgeDurationMinutes   *int32 `json:"maxJudgeDurationMinutes"`
	SubmissionCooldownSeconds *int32 `json:"submissionCooldownSeconds"`
	PidsLimit                 *int32 `json:"pidsLimit"`
	MemoryLimitMB             *int32 `json:"memoryLimitMB"`
}

type Scoreboard struct {
	Compact   bool `json:"compact"`
	MaskNames bool `json:"maskNames"`
}

// Participants are deliberately absent. They are personal data, and a contest
// moved to another venue is run by different people anyway — an import that
// carried them would be both a privacy leak and wrong.

func Marshal(c *Contest) ([]byte, error) { return bundlecodec.Marshal(c) }

func Unmarshal(data []byte) (*Contest, error) {
	var c Contest
	if err := bundlecodec.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
