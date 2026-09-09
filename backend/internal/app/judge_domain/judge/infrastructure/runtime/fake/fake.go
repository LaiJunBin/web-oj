// Package fake provides an in-memory ExecutionEnvironmentProvisioner used by
// feature tests: every command "succeeds" instantly with exit code 0, so the
// full submission → judged pipeline can be exercised deterministically
// without a Docker daemon.
package fake

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/command"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	domainservice "judge-backend/internal/app/judge_domain/judge/domain/service"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type Runtime struct {
	logger sharedtypes.Logger
}

func NewRuntime(logger sharedtypes.Logger) domainservice.ExecutionEnvironmentProvisioner {
	return &Runtime{logger: logger}
}

func (r *Runtime) Provision(ctx context.Context, spec *domainservice.EnvironmentSpec) (domainservice.ExecutionEnvironment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.logger.Infof("[FakeRuntime] Provisioned fake environment for judgment %s", spec.JudgmentID)
	return &environment{judgmentID: spec.JudgmentID, logger: r.logger}, nil
}

func (r *Runtime) ReleaseWorker(workerID string) {}

func (r *Runtime) CleanupOrphans(activeJudgmentIDs map[string]struct{}, activeWorkerIDs map[string]struct{}) {
	r.logger.Infof("[FakeRuntime] CleanupOrphans is a no-op")
}

type environment struct {
	judgmentID string
	logger     sharedtypes.Logger
}

func (e *environment) Exec(ctx context.Context, target domainservice.ExecTarget, cmd command.Command, timeout time.Duration) (*domainservice.ExecReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	e.logger.Debugf("[FakeRuntime] Exec %q on %s for judgment %s", cmd.String(), target.String(), e.judgmentID)
	exitCode, stderr := fakeVerdict(cmd.String())
	return &domainservice.ExecReport{
		Result: execresult.NewExecResult(&execresult.ExecResultConfig{
			ExitCode: exitCode,
			ExecTime: 1,
			Stdout:   "",
			Stderr:   stderr,
		}),
	}, nil
}

// fakeVerdict lets a scenario ask for a failing step without a real container:
// the literal command `false` exits 1 and `exit N` exits N (with a stderr line
// naming it); everything else succeeds as before. Scenarios that exercise the
// lifecycle fail-fast path rely on this.
func fakeVerdict(cmd string) (int, string) {
	if cmd == "false" {
		return 1, "fake runtime: command 'false' failed"
	}
	var code int
	if n, _ := fmt.Sscanf(cmd, "exit %d", &code); n == 1 && !strings.ContainsAny(cmd[len("exit"):], "&|;") {
		return code, fmt.Sprintf("fake runtime: command exited with %d", code)
	}
	return 0, ""
}

func (e *environment) StreamLogs(ctx context.Context, w io.Writer) error {
	<-ctx.Done()
	return ctx.Err()
}

func (e *environment) Teardown() {}
