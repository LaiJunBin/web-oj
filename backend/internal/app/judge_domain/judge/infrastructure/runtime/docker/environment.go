package docker

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/command"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	domainservice "judge-backend/internal/app/judge_domain/judge/domain/service"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

// environment is one provisioned Docker environment: a project container and
// a test container joined on the worker's network.
type environment struct {
	runtime            *Runtime
	judgmentID         string
	projectContainerID string
	testContainerID    string
	testEnv            []string
	// outputDrainTimeout overrides defaultOutputDrainTimeout; zero means use
	// the default. A field (not a bare constant) so tests can shrink it
	// instead of waiting out a real 30s timeout to exercise that branch.
	outputDrainTimeout time.Duration
}

const defaultOutputDrainTimeout = 30 * time.Second

func (e *environment) drainTimeout() time.Duration {
	if e.outputDrainTimeout > 0 {
		return e.outputDrainTimeout
	}
	return defaultOutputDrainTimeout
}

func (e *environment) containerFor(target domainservice.ExecTarget) (string, []string) {
	if target.IsProject() {
		return e.projectContainerID, nil
	}
	return e.testContainerID, e.testEnv
}

// Exec runs a command in the target container, polling until it exits. On
// timeout the process is killed and the report is marked TimedOut. This is
// the single exec engine shared by lifecycle and testcase execution (they
// used to have two divergent copies of this logic).
func (e *environment) Exec(ctx context.Context, target domainservice.ExecTarget, cmd command.Command, timeout time.Duration) (*domainservice.ExecReport, error) {
	logger := e.runtime.logger
	dockerClient := e.runtime.dockerClient
	containerID, env := e.containerFor(target)

	exec, err := dockerClient.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          cmd.Split(),
		Env:          env,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		logger.Errorf("[DockerRuntime] Failed to create exec in container %s (judgment %s): %v", containerID[:12], e.judgmentID, err)
		return nil, e.explainExecError(containerID, err)
	}

	resp, err := dockerClient.ContainerExecAttach(ctx, exec.ID, container.ExecStartOptions{})
	if err != nil {
		logger.Errorf("[DockerRuntime] Failed to attach exec %s (judgment %s): %v", exec.ID, e.judgmentID, err)
		return nil, e.explainExecError(containerID, err)
	}
	defer resp.Close()

	// Read stdout/stderr in the background; the pipe may stay open past the
	// command's exit when it spawned children, so completion is detected by
	// polling the exec state, not by pipe EOF. The result is delivered over
	// a buffered channel rather than shared closure variables: below, the
	// outputDrainTimeout branch stops waiting on the goroutine without
	// receiving from outputCh, and if stdout/stderr/readErr were plain
	// shared variables, the main goroutine reading them at that point would
	// race the background goroutine still writing them. The channel means
	// the only way to observe the output is to receive it, which is
	// inherently synchronized; the buffer of 1 lets the late send land
	// without blocking the (by-then long gone) background goroutine.
	type outputResult struct {
		stdout, stderr string
		err            error
	}
	outputCh := make(chan outputResult, 1)
	intentionalClose := make(chan struct{})
	var intentionalCloseOnce sync.Once
	signalClose := func() {
		intentionalCloseOnce.Do(func() { close(intentionalClose) })
	}
	go func() {
		stdout, stderr, readErr := readStdOutput(resp.Reader, e.runtime.maxOutputBytes)
		if readErr != nil {
			select {
			case <-intentionalClose:
				logger.Debugf("[DockerRuntime] Output read interrupted for exec %s (connection closed): %v", exec.ID, readErr)
			default:
				logger.Errorf("[DockerRuntime] Failed to read output for exec %s: %v", exec.ID, readErr)
			}
		}
		outputCh <- outputResult{stdout: stdout, stderr: stderr, err: readErr}
	}()

	startTime := time.Now()
	info, err := dockerClient.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		signalClose()
		logger.Errorf("[DockerRuntime] Failed to inspect exec %s (judgment %s): %v", exec.ID, e.judgmentID, err)
		return nil, e.explainExecError(containerID, err)
	}

	timedOut := false
	timeoutCh := time.After(timeout)
	for info.Running {
		select {
		case <-timeoutCh:
			signalClose()
			resp.Close() // close exec attachment to unblock the output reader (child processes may keep the pipe open)
			e.killByPid(containerID, info.Pid)
			timedOut = true
		case <-ctx.Done():
			signalClose()
			resp.Close()
			return nil, ctx.Err()
		default:
			time.Sleep(50 * time.Millisecond)
			info, err = dockerClient.ContainerExecInspect(ctx, exec.ID)
			if err != nil {
				signalClose()
				logger.Errorf("[DockerRuntime] Failed to inspect exec %s (judgment %s): %v", exec.ID, e.judgmentID, err)
				return nil, e.explainExecError(containerID, err)
			}
		}
		if timedOut {
			break
		}
	}

	// If ctx was cancelled right after the loop exited naturally (race),
	// treat the reader's error as expected.
	if ctx.Err() != nil {
		signalClose()
		resp.Close()
		return nil, ctx.Err()
	}

	// Wait for the output reader to finish; cap the wait to avoid hanging if
	// the Docker daemon becomes unresponsive after resp.Close(). stdout/
	// stderr/readErr are ONLY ever assigned here, from the channel receive —
	// on the timeout branch they simply stay at their zero values, so there
	// is nothing left for the (possibly still-running) background goroutine
	// to race against.
	var stdout, stderr string
	var readErr error
	outputLost := false
	select {
	case out := <-outputCh:
		stdout, stderr, readErr = out.stdout, out.stderr, out.err
	case <-time.After(e.drainTimeout()):
		logger.Errorf("[DockerRuntime] Timed out draining output for exec %s (judgment %s)", exec.ID, e.judgmentID)
		outputLost = true
	}
	if readErr != nil && !timedOut {
		outputLost = true
	}

	execDuration := time.Since(startTime)
	return &domainservice.ExecReport{
		Result: execresult.NewExecResult(&execresult.ExecResultConfig{
			ExitCode: info.ExitCode,
			ExecTime: execDuration.Milliseconds(),
			Stdout:   stdout,
			Stderr:   stderr,
		}),
		TimedOut:   timedOut,
		OutputLost: outputLost,
	}, nil
}

// explainExecError adds the reason to a Docker error from an exec, when the
// daemon's own message hides it.
//
// A memory cap makes a failure mode reachable that could not happen before:
// the kernel kills the container's main process, and from then on every exec
// comes back as "container is not running". That message names the daemon's
// symptom, and it is what lands in the judgment's error message, so the
// operator sees a broken judge rather than a submission that asked for more
// memory than the contest project allows.
//
// Best-effort by construction: any failure to diagnose returns the original
// error untouched, because the error being explained matters more than the
// explanation.
func (e *environment) explainExecError(containerID string, err error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// State lives behind an embedded *ContainerJSONBase, so the nil check on
	// the pointer has to come first — reaching straight for info.State would
	// panic on a response missing it, turning a reportable failure into a
	// crashed worker.
	info, inspectErr := e.runtime.dockerClient.ContainerInspect(ctx, containerID)
	if inspectErr != nil || info.ContainerJSONBase == nil || info.State == nil || !info.State.OOMKilled {
		return err
	}
	return fmt.Errorf("container exceeded its memory limit and was killed by the kernel: %w", err)
}

// killByPid best-effort kills a process inside a container. Used when an exec
// exceeds its timeout.
func (e *environment) killByPid(containerID string, pid int) {
	logger := e.runtime.logger
	dockerClient := e.runtime.dockerClient

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kill, err := dockerClient.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          []string{"kill", "-9", fmt.Sprintf("%d", pid)},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		logger.Errorf("[DockerRuntime] Failed to create kill exec for pid %d in container %s: %v", pid, containerID[:12], err)
		return
	}

	resp, err := dockerClient.ContainerExecAttach(ctx, kill.ID, container.ExecStartOptions{})
	if err != nil {
		logger.Errorf("[DockerRuntime] Failed to attach kill exec %s for pid %d in container %s: %v", kill.ID, pid, containerID[:12], err)
		return
	}
	defer resp.Close()
	io.Copy(io.Discard, resp.Reader)

	// Inspect the kill exec instead of unconditionally claiming success: the
	// pid comes from ContainerExecInspect and is a HOST-namespace PID, which
	// inside the container's own PID namespace usually maps to nothing — so
	// this kill is frequently a silent no-op (`kill` may not even exist in
	// the image). Saying so in the log is what makes the gap visible; the
	// runaway process is otherwise cleaned up at container teardown, which
	// means it keeps consuming the container's limits (and, unless the
	// operator set memory/CPU caps, the host's) until then. A correct fix
	// has to learn the process's in-namespace PID — by wrapping the exec'd
	// command, or by matching on process name — which changes real timeout
	// behaviour, cannot assume the image ships `sh` or `kill`, and must not
	// pollute stdout because that is what gets graded. It is deferred until
	// it can be verified against real containers; the Docker API offers no
	// "terminate this exec" call.
	inspect, inspectErr := dockerClient.ContainerExecInspect(ctx, kill.ID)
	switch {
	case inspectErr != nil || inspect.Running:
		logger.Warnf("[DockerRuntime] Sent kill -9 to pid %d in container %s but could not confirm the outcome (inspect err: %v) — a host-namespace PID is usually invalid inside the container, so the timed-out process may still be running until the container is torn down", pid, containerID[:12], inspectErr)
	case inspect.ExitCode != 0:
		logger.Warnf("[DockerRuntime] kill -9 %d in container %s exited with code %d — the process was likely NOT killed (a host-namespace PID is usually invalid inside the container) and may run until the container is torn down", pid, containerID[:12], inspect.ExitCode)
	default:
		logger.Debugf("[DockerRuntime] Killed pid %d in container %s", pid, containerID[:12])
	}
}

func (e *environment) StreamLogs(ctx context.Context, w io.Writer) error {
	logger := e.runtime.logger
	logs, err := e.runtime.dockerClient.ContainerLogs(ctx, e.testContainerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	})
	if err != nil {
		logger.Errorf("[DockerRuntime] Failed to get logs for container %s: %v", e.testContainerID[:12], err)
		return err
	}
	defer logs.Close()

	if _, err := stdcopy.StdCopy(w, w, bufio.NewReader(logs)); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("failed to stream container logs: %w", err)
	}
	return nil
}

// Teardown force-removes both containers. It intentionally uses a background
// context so cleanup still happens after the judgment's context is cancelled.
func (e *environment) Teardown() {
	e.runtime.logger.Debugf("[DockerRuntime] Tearing down environment for judgment %s", e.judgmentID)
	e.runtime.removeContainer(e.testContainerID)
	e.runtime.removeContainer(e.projectContainerID)
}

// boundedWriter appends to buf until a shared byte budget is exhausted, then
// discards the rest while still reporting a full write so the caller keeps
// draining the reader (the container is throttled by the exec timeout, not by a
// short write). stdout and stderr share one budget so their combined size — the
// thing that lands in memory and in Mongo — is what is capped.
type boundedWriter struct {
	buf       *bytes.Buffer
	remaining *int64
	truncated *bool
}

func (w *boundedWriter) Write(p []byte) (int, error) {
	if *w.remaining <= 0 {
		*w.truncated = true
		return len(p), nil
	}
	if int64(len(p)) > *w.remaining {
		w.buf.Write(p[:*w.remaining])
		*w.remaining = 0
		*w.truncated = true
		return len(p), nil
	}
	w.buf.Write(p)
	*w.remaining -= int64(len(p))
	return len(p), nil
}

// readStdOutput demultiplexes the exec stream into stdout and stderr, capping
// the combined captured bytes at maxBytes (0 means no cap). A submission that
// prints without bound is truncated with a marker instead of buffering
// everything into memory and the persisted judgment.
func readStdOutput(reader *bufio.Reader, maxBytes int64) (string, string, error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	remaining := int64(math.MaxInt64)
	if maxBytes > 0 {
		remaining = maxBytes
	}
	truncated := false
	sw := &boundedWriter{buf: stdoutBuf, remaining: &remaining, truncated: &truncated}
	ew := &boundedWriter{buf: stderrBuf, remaining: &remaining, truncated: &truncated}

	if _, err := stdcopy.StdCopy(sw, ew, reader); err != nil {
		return "", "", fmt.Errorf("failed to copy stdout/stderr: %w", err)
	}

	stdout, err := normalizeLines(stdoutBuf)
	if err != nil {
		return "", "", err
	}
	stderr, err := normalizeLines(stderrBuf)
	if err != nil {
		return "", "", err
	}

	if truncated {
		stdout += fmt.Sprintf("\n[output truncated: exceeded %d bytes]\n", maxBytes)
	}

	return stdout, stderr, nil
}

// normalizeLines re-emits buf line by line with a trailing newline per line,
// using a strings.Builder (not O(n^2) string concatenation) and a scan buffer
// large enough that a long line is not silently dropped — the previous version
// hit bufio.Scanner's 64 KiB default and truncated the rest without reporting.
func normalizeLines(buf *bytes.Buffer) (string, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(buf)
	scanner.Buffer(make([]byte, 0, 64*1024), buf.Len()+1)
	for scanner.Scan() {
		sb.Write(scanner.Bytes())
		sb.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to scan output: %w", err)
	}
	return sb.String(), nil
}
