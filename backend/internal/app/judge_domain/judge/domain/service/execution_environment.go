package service

import (
	"context"
	"io"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/command"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
)

// ExecTarget selects which side of the provisioned environment a command runs in.
type ExecTarget struct {
	value string
}

var (
	TargetProject = ExecTarget{"project"}
	TargetTest    = ExecTarget{"test"}
)

func (t ExecTarget) IsProject() bool { return t == TargetProject }
func (t ExecTarget) IsTest() bool    { return t == TargetTest }
func (t ExecTarget) String() string  { return t.value }

// EnvironmentSpec describes what a judgment run needs from the runtime,
// in judging terms only — no container/network concepts leak in here.
type EnvironmentSpec struct {
	JudgmentID   string
	SubmissionID string
	// WorkerID scopes runtime resources (e.g. an isolated network) to the
	// worker executing this judgment.
	WorkerID string
	Project  *project.Project
	// Metadata are the submission's dynamic fields; the runtime exposes them
	// to test-side commands (PAYLOAD_* variables in the Docker implementation).
	Metadata map[string]string
	// SubmitterUsername is the username of the user who owns the submission;
	// the runtime exposes it to both sides (SUBMIT_USERNAME in the Docker
	// implementation) — the project side provisions per-user resources with
	// it (e.g. seat-specific database names and accounts), the test side uses
	// it to address those resources from lifecycle commands.
	SubmitterUsername string
	// DiagnosticsSink, when set, receives human-readable provisioning
	// diagnostics — e.g. the project container's final log lines when it
	// dies or never becomes healthy, which would otherwise vanish with the
	// teardown. Best-effort: runtimes ignore it when nil and must never let
	// a diagnostics failure mask the provisioning error being diagnosed.
	DiagnosticsSink io.Writer
}

// ExecReport is the outcome of a single command execution.
type ExecReport struct {
	Result *execresult.ExecResult
	// TimedOut is true when the command was killed because it exceeded the
	// given timeout; Result still carries whatever output was captured.
	TimedOut bool
	// OutputLost is true when the command finished but its output could not
	// be (fully) captured; the exit code in Result is still trustworthy.
	OutputLost bool
}

// ExecutionEnvironment is a provisioned, ready-to-use environment for one
// judgment run.
type ExecutionEnvironment interface {
	// Exec runs a command in the selected target and waits for completion,
	// killing it after timeout. Returns ctx.Err() if ctx is cancelled.
	Exec(ctx context.Context, target ExecTarget, cmd command.Command, timeout time.Duration) (*ExecReport, error)
	// StreamLogs copies the environment's test-side output into w until ctx
	// is cancelled or the environment is torn down.
	StreamLogs(ctx context.Context, w io.Writer) error
	// Teardown releases every runtime resource belonging to this environment.
	// Safe to call in a deferred cleanup path; must not depend on ctx state.
	Teardown()
}

// ExecutionEnvironmentProvisioner is the judge domain's runtime abstraction.
// Swapping the container backend (Docker today; Podman, Kubernetes, a fake for
// tests tomorrow) means providing another implementation of this interface.
type ExecutionEnvironmentProvisioner interface {
	Provision(ctx context.Context, spec *EnvironmentSpec) (ExecutionEnvironment, error)
	// ReleaseWorker frees runtime resources scoped to a worker (e.g. its
	// isolated network). Called when a worker is deleted.
	ReleaseWorker(workerID string)
	// CleanupOrphans removes runtime resources whose judgment or worker is no
	// longer active. Called once at startup before workers are created.
	CleanupOrphans(activeJudgmentIDs map[string]struct{}, activeWorkerIDs map[string]struct{})
}
