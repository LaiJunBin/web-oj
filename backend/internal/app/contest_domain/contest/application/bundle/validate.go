package bundle

import (
	"fmt"
	"strings"
)

type Problem struct {
	Path    string
	Message string
}

func (p Problem) Error() string { return p.Path + ": " + p.Message }

type Problems []Problem

func (p Problems) Error() string {
	parts := make([]string, len(p))
	for i, problem := range p {
		parts[i] = problem.Error()
	}
	return strings.Join(parts, "; ")
}

// Validate reports every fault in one pass, so a hand-written bundle does not
// need one upload per typo.
//
// Only what can be known without the database. Whether a projectRef resolves
// depends on this installation's library and belongs to the importer.
func (c *Contest) Validate() Problems {
	var problems Problems
	add := func(path, format string, args ...any) {
		problems = append(problems, Problem{Path: path, Message: fmt.Sprintf(format, args...)})
	}

	if c.FormatVersion <= 0 {
		add("formatVersion", "missing")
	} else if c.FormatVersion > FormatVersion {
		add("formatVersion", "此匯出檔是 %d 版，本系統只支援到 %d 版，請升級後再匯入", c.FormatVersion, FormatVersion)
	}

	if c.Kind != KindContest {
		add("kind", "必須是 %q，得到 %q", KindContest, c.Kind)
	}

	if strings.TrimSpace(c.Name) == "" {
		add("name", "不可為空")
	}

	// The activation window is what makes a contest visible at all; inverted,
	// the contest exists and never appears, with nothing to say why.
	if !c.DeactivateTime.After(c.ActivateTime) {
		add("deactivateTime", "必須晚於 activateTime")
	}

	for i, p := range c.Projects {
		at := fmt.Sprintf("projects[%d]", i)
		if strings.TrimSpace(p.ProjectRef) == "" {
			add(at+".projectRef", "不可為空 —— 這是用來對應題庫的名稱")
		}
		if strings.TrimSpace(p.Name) == "" {
			add(at+".name", "不可為空")
		}
		if !p.DeactivateTime.After(p.ActivateTime) {
			add(at+".deactivateTime", "必須晚於 activateTime")
		}
		if !p.EndTime.After(p.StartTime) {
			add(at+".endTime", "必須晚於 startTime")
		}
		if v := p.Limits.MaxJudgeDurationMinutes; v != nil && *v <= 0 {
			add(at+".limits.maxJudgeDurationMinutes", "必須為正數，或用 null 表示沿用系統預設")
		}
		// The other three take 0 as "no limit", so only a negative is wrong.
		for name, v := range map[string]*int32{
			"submissionCooldownSeconds": p.Limits.SubmissionCooldownSeconds,
			"pidsLimit":                 p.Limits.PidsLimit,
			"memoryLimitMB":             p.Limits.MemoryLimitMB,
		} {
			if v != nil && *v < 0 {
				add(at+".limits."+name, "不可為負")
			}
		}
	}

	return problems
}
