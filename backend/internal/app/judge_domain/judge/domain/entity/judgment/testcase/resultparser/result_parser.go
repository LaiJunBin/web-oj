package resultparser

import (
	"errors"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
)

var (
	Parsers = map[string]ResultParser{}

	ErrorInvalidResultParser = errors.New("invalid result parser")
)

type ResultParser interface {
	String() string
	// Parse decides the testcase outcome and returns the stdout that should
	// be persisted afterwards (publicStdout). A parser that consumed a
	// verdict from the output strips that line — sample-testcase stdout is
	// contestant-visible, and the grader's control line must not reach them.
	// Parsers that read nothing from stdout return it unchanged.
	Parse(result *execresult.ExecResult, configuredScore float64) (s status.Status, score float64, publicStdout string)
}

type resultParser struct {
	value string
	parse func(result *execresult.ExecResult, configuredScore float64) (status.Status, float64, string)
}

func newResultParser(value string, parse func(*execresult.ExecResult, float64) (status.Status, float64, string)) ResultParser {
	p := &resultParser{value: value, parse: parse}
	Parsers[value] = p
	return p
}

func (r *resultParser) String() string {
	return r.value
}

func (r *resultParser) Parse(result *execresult.ExecResult, configuredScore float64) (status.Status, float64, string) {
	return r.parse(result, configuredScore)
}

func GetValidParserNames() []string {
	names := make([]string, 0, len(Parsers))
	for name := range Parsers {
		names = append(names, name)
	}
	return names
}
