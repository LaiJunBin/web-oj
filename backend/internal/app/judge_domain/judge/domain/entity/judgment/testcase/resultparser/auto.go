package resultparser

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
)

var AutoParser = newResultParser("auto", func(result *execresult.ExecResult, configuredScore float64) (status.Status, float64, string) {
	if _, _, _, ok := parseOutputScore(result.Stdout(), configuredScore); ok {
		return StdoutParser.Parse(result, configuredScore)
	}
	return ExitCodeParser.Parse(result, configuredScore)
})
