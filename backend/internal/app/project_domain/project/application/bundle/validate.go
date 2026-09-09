package bundle

import (
	"fmt"
	"sort"
	"strings"

	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem/outputformater"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem/scoringstrategy"
	"judge-backend/internal/app/project_domain/project/domain/entity/lifecycle"
	"judge-backend/internal/app/project_domain/project/domain/entity/project/dynamicfield"
)

// ParserValidator reports whether a result parser name is one this deployment
// knows. Declared here as a one-method interface rather than imported: the
// parsers are registered by the judge domain, and project code reaches them
// only through a binding in the composition root. The domain's own
// ResultParserValidator satisfies this.
type ParserValidator interface {
	IsValid(parser string) bool
}

// Problem is one thing wrong with a bundle, located by its JSON path.
//
// The path matters more than it looks. A bundle can be hand-written, and
// "invalid scoring strategy" tells its author nothing about which of forty
// grading items to look at.
type Problem struct {
	Path    string
	Message string
}

func (p Problem) Error() string {
	return p.Path + ": " + p.Message
}

// Problems is every fault found in one pass.
type Problems []Problem

func (p Problems) Error() string {
	parts := make([]string, len(p))
	for i, problem := range p {
		parts[i] = problem.Error()
	}
	return strings.Join(parts, "; ")
}

// Validate checks everything that can be known without touching the database,
// and reports ALL of it at once.
//
// Stopping at the first fault would be cheaper and much worse to use: fixing a
// hand-written bundle would become one upload per typo. Anything that needs
// the database — whether a referenced image exists, whether the name collides
// — belongs to the importer, not here.
func (p *Project) Validate(parsers ParserValidator) Problems {
	var problems Problems
	add := func(path, format string, args ...any) {
		problems = append(problems, Problem{Path: path, Message: fmt.Sprintf(format, args...)})
	}

	// A bundle from a newer version may parse cleanly and still mean
	// something else, so refuse it rather than guess.
	if p.FormatVersion <= 0 {
		add("formatVersion", "missing")
	} else if p.FormatVersion > FormatVersion {
		add("formatVersion", "此匯出檔是 %d 版，本系統只支援到 %d 版，請升級後再匯入", p.FormatVersion, FormatVersion)
	}

	if p.Kind != KindProject {
		add("kind", "必須是 %q，得到 %q", KindProject, p.Kind)
	}

	if strings.TrimSpace(p.Name) == "" {
		add("name", "不可為空")
	}
	if strings.TrimSpace(p.ImageTag) == "" {
		add("imageTag", "不可為空")
	}
	if strings.TrimSpace(p.TestImageTag) == "" {
		add("testImageTag", "不可為空")
	}
	if p.ResultParser != "" && !parsers.IsValid(p.ResultParser) {
		add("resultParser", "%q 不是本系統支援的 result parser", p.ResultParser)
	}

	for i, f := range p.DynamicFields {
		at := fmt.Sprintf("dynamicFields[%d]", i)
		if strings.TrimSpace(f.Name) == "" {
			add(at+".name", "不可為空")
		}
		if _, err := dynamicfield.GetTypeByValue(f.Type); err != nil {
			add(at+".type", "%q 不是合法的欄位型別", f.Type)
		}
	}

	for i, l := range p.Lifecycles {
		at := fmt.Sprintf("lifecycles[%d]", i)
		if strings.TrimSpace(l.Name) == "" {
			add(at+".name", "不可為空")
		}
		if strings.TrimSpace(l.Command) == "" {
			add(at+".command", "不可為空")
		}
		if !isLifecycleType(l.Type) {
			add(at+".type", "%q 不是合法的生命週期時機", l.Type)
		}
		if !isRunsOn(l.RunsOn) {
			add(at+".runsOn", "%q 不是合法的執行環境", l.RunsOn)
		}
	}

	// Collected first so a dependency may point forwards, at a testcase
	// declared under a later grading item.
	declared := map[string]string{} // key → the path that declared it
	for i, g := range p.GradingItems {
		for j, t := range g.Testcases {
			at := fmt.Sprintf("gradingItems[%d].testcases[%d]", i, j)
			if strings.TrimSpace(t.Key) == "" {
				add(at+".key", "不可為空 —— 相依關係要靠它指認")
				continue
			}
			if where, dup := declared[t.Key]; dup {
				add(at+".key", "與 %s 重複；key 必須在整個匯出檔內唯一", where)
				continue
			}
			declared[t.Key] = at
		}
	}

	for i, g := range p.GradingItems {
		at := fmt.Sprintf("gradingItems[%d]", i)
		if strings.TrimSpace(g.Name) == "" {
			add(at+".name", "不可為空")
		}
		if strings.TrimSpace(g.Label) == "" {
			add(at+".label", "不可為空")
		}
		if g.Score < 0 {
			add(at+".score", "不可為負")
		}
		if _, ok := scoringstrategy.Strategies[g.ScoringStrategy]; !ok {
			add(at+".scoringStrategy", "%q 不是合法的計分方式", g.ScoringStrategy)
		}
		if _, ok := outputformater.Formaters[g.OutputFormat]; !ok {
			add(at+".outputFormat", "%q 不是合法的輸出格式", g.OutputFormat)
		}
		if g.ResultParser != "" && !parsers.IsValid(g.ResultParser) {
			add(at+".resultParser", "%q 不是本系統支援的 result parser", g.ResultParser)
		}
		for k, dep := range g.DependsOnTestcases {
			if _, ok := declared[dep]; !ok {
				add(fmt.Sprintf("%s.dependsOnTestcases[%d]", at, k), "找不到 key 為 %q 的測資", dep)
			}
		}

		for j, t := range g.Testcases {
			tat := fmt.Sprintf("%s.testcases[%d]", at, j)
			if strings.TrimSpace(t.Name) == "" {
				add(tat+".name", "不可為空")
			}
			if strings.TrimSpace(t.Command) == "" {
				add(tat+".command", "不可為空")
			}
			if t.Score < 0 {
				add(tat+".score", "不可為負")
			}
			for k, dep := range t.DependsOnTestcases {
				if dep == t.Key {
					add(fmt.Sprintf("%s.dependsOnTestcases[%d]", tat, k), "測資不可相依自己")
					continue
				}
				if _, ok := declared[dep]; !ok {
					add(fmt.Sprintf("%s.dependsOnTestcases[%d]", tat, k), "找不到 key 為 %q 的測資", dep)
				}
			}
		}
	}

	// A cycle deadlocks the judgment scheduler: no testcase in the cycle ever
	// becomes runnable, and the judgment sits until its wall-clock cap fires.
	// Cheaper to refuse the bundle than to debug that during a contest.
	for _, cycle := range findCycles(p) {
		add("gradingItems[].testcases[].dependsOnTestcases", "測資相依形成循環：%s", strings.Join(cycle, " → "))
	}

	seenAssetFiles := map[string]bool{}
	for i, a := range p.Assets {
		at := fmt.Sprintf("assets[%d]", i)
		if strings.TrimSpace(a.Name) == "" {
			add(at+".name", "不可為空")
		}
		if strings.TrimSpace(a.File) == "" {
			add(at+".file", "不可為空")
			continue
		}
		// The file name is joined onto a directory and then written, so a
		// bundle must not be able to aim it anywhere else.
		if strings.Contains(a.File, "..") || strings.HasPrefix(a.File, "/") {
			add(at+".file", "路徑不合法：%q", a.File)
		}
		if !strings.HasPrefix(a.File, AssetsDir+"/") {
			add(at+".file", "必須位於 %s/ 之下，得到 %q", AssetsDir, a.File)
		}
		if seenAssetFiles[a.File] {
			add(at+".file", "與前面的附件指向同一個檔案：%q", a.File)
		}
		seenAssetFiles[a.File] = true
	}

	return problems
}

func isLifecycleType(v string) bool {
	_, ok := lifecycle.Types[v]
	return ok
}

func isRunsOn(v string) bool {
	for _, e := range lifecycle.GetExecutionEnvironments() {
		if e.String() == v {
			return true
		}
	}
	return false
}

// findCycles returns each dependency cycle once, as the keys involved.
func findCycles(p *Project) [][]string {
	edges := map[string][]string{}
	for _, g := range p.GradingItems {
		for _, t := range g.Testcases {
			// Self-loops are excluded: they are already reported against the
			// exact dependency that caused them. Left in, they would also
			// surface here as "a → a" under a path that names no particular
			// testcase — a second, vaguer copy of a message the reader has
			// already acted on.
			for _, dep := range t.DependsOnTestcases {
				if dep != t.Key {
					edges[t.Key] = append(edges[t.Key], dep)
				}
			}
			if _, ok := edges[t.Key]; !ok {
				edges[t.Key] = nil
			}
		}
	}

	const (
		unvisited = 0
		onStack   = 1
		done      = 2
	)
	state := map[string]int{}
	var stack []string
	var cycles [][]string
	seen := map[string]bool{}

	var walk func(key string)
	walk = func(key string) {
		state[key] = onStack
		stack = append(stack, key)

		for _, next := range edges[key] {
			switch state[next] {
			case unvisited:
				if _, known := edges[next]; known {
					walk(next)
				}
			case onStack:
				// Report from where the cycle closes, so the message names
				// only the keys actually in it.
				at := len(stack) - 1
				for at >= 0 && stack[at] != next {
					at--
				}
				if at >= 0 {
					cycle := append([]string{}, stack[at:]...)
					if id := cycleID(cycle); !seen[id] {
						seen[id] = true
						cycles = append(cycles, append(cycle, next))
					}
				}
			}
		}

		stack = stack[:len(stack)-1]
		state[key] = done
	}

	keys := make([]string, 0, len(edges))
	for k := range edges {
		keys = append(keys, k)
	}
	// Sorted so the same bundle always reports the same cycles in the same
	// order; map iteration would make the error message flap between runs.
	sort.Strings(keys)

	for _, k := range keys {
		if state[k] == unvisited {
			walk(k)
		}
	}
	return cycles
}

// cycleID is a rotation-independent identity, so one cycle reached from two
// different starting points is reported once.
func cycleID(cycle []string) string {
	rotated := append([]string{}, cycle...)
	sort.Strings(rotated)
	return strings.Join(rotated, "\x00")
}
