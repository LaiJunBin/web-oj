package resultparser

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
)

var ExitCodeParser = newResultParser("exitcode", func(result *execresult.ExecResult, configuredScore float64) (status.Status, float64, string) {
	if result.ExitCode() == 0 {
		return status.PASSED, configuredScore, result.Stdout()
	}
	return status.FAILED, configuredScore, result.Stdout()
})
