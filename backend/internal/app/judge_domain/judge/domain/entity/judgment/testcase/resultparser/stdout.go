package resultparser

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
	"regexp"
	"strconv"
	"strings"
)

/*
The verdict must be the ENTIRE first non-blank line of stdout, in the format
"pass [score]" or "fail [score]" (case-insensitive). The score is optional; if
omitted, configuredScore is used as the fallback.
Examples:
  - "pass 100"  -> keyword="pass", score=100
  - "FAIL 85.5" -> keyword="fail", score=85.5
  - "pass"      -> keyword="pass", score=<configuredScore>

Anchored on both ends on purpose: a prefix-only match would let unrelated
output like "passing tests..." or a malformed "pass 10 extra" line silently
score as PASSED (and would make the auto parser hijack such output into the
stdout branch instead of falling back to the exit code).
*/
var stdoutVerdictRegex = regexp.MustCompile(`(?i)^(pass|fail)(?:\s+([\d]+(?:\.[\d]+)?))?\s*$`)

var StdoutParser = newResultParser("stdout", func(result *execresult.ExecResult, configuredScore float64) (status.Status, float64, string) {
	keyword, score, rest, ok := parseOutputScore(result.Stdout(), configuredScore)
	if !ok {
		// Nothing was consumed — keep the raw stdout so whatever the grading
		// script actually printed stays available for diagnosis.
		return status.ERROR, 0, result.Stdout()
	}
	if keyword == "pass" {
		return status.PASSED, score, rest
	}
	return status.FAILED, score, rest
})

// splitFirstContentLine returns the first non-blank line of stdout (trimmed)
// and the offset where the remaining output begins — the byte after that
// line's newline, or len(stdout) when it is the last line.
func splitFirstContentLine(stdout string) (line string, restStart int) {
	offset := 0
	for offset < len(stdout) {
		lineEnd := len(stdout)
		next := len(stdout)
		if i := strings.IndexByte(stdout[offset:], '\n'); i >= 0 {
			lineEnd = offset + i
			next = offset + i + 1
		}
		if content := strings.TrimSpace(stdout[offset:lineEnd]); content != "" {
			return content, next
		}
		offset = next
	}
	return "", len(stdout)
}

// parseOutputScore extracts the verdict from stdout's first non-blank line.
// rest is stdout with that line removed — the part safe to show afterwards.
func parseOutputScore(stdout string, configuredScore float64) (keyword string, score float64, rest string, ok bool) {
	line, restStart := splitFirstContentLine(stdout)
	m := stdoutVerdictRegex.FindStringSubmatch(line)
	if m == nil {
		return "", 0, "", false
	}
	rest = stdout[restStart:]
	if m[2] == "" {
		return strings.ToLower(m[1]), configuredScore, rest, true
	}
	s, err := strconv.ParseFloat(m[2], 64)
	if err != nil {
		return "", 0, "", false
	}
	return strings.ToLower(m[1]), s, rest, true
}
