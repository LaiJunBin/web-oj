package docker

import (
	"bufio"
	"errors"
	"net"
	"testing"

	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	dockermock "judge-backend/internal/mocks/github.com/docker/docker/client"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// hijackedWith fakes the attach stream of an exec: the payload is readable
// until EOF, and Close() works. killByPid only drains and closes it.
func hijackedWith(payload []byte) types.HijackedResponse {
	server, client := net.Pipe()
	go func() {
		if len(payload) > 0 {
			_, _ = server.Write(payload)
		}
		server.Close()
	}()
	return types.HijackedResponse{Conn: client, Reader: bufio.NewReader(client)}
}

func newKillTestEnvironment(dockerClient *dockermock.MockAPIClient) *environment {
	return &environment{
		runtime: &Runtime{
			dockerClient: dockerClient,
			logger:       sharedlogger.NewMockLogger(),
		},
		projectContainerID: "project-container-full-id",
	}
}

// Observability regression tests for the timeout kill: killByPid must inspect the kill exec's
// outcome rather than unconditionally claiming success — the pid is a
// host-namespace pid and the kill is frequently a silent no-op. The
// MockLogger is a no-op, so the assertions here are the mock expectations:
// which docker calls happen, with which arguments, on each branch.

func TestKillByPid_SendsKillCommandAndInspectsOutcome(t *testing.T) {
	dockerClient := dockermock.NewMockAPIClient(t)
	env := newKillTestEnvironment(dockerClient)

	dockerClient.On("ContainerExecCreate", mock.Anything, "project-container-full-id",
		mock.MatchedBy(func(opts container.ExecOptions) bool {
			return len(opts.Cmd) == 3 && opts.Cmd[0] == "kill" && opts.Cmd[1] == "-9" && opts.Cmd[2] == "1234"
		})).Return(types.IDResponse{ID: "kill-exec-id"}, nil).Once()
	dockerClient.On("ContainerExecAttach", mock.Anything, "kill-exec-id", mock.Anything).
		Return(hijackedWith(nil), nil).Once()
	// The whole point of the observability fix: the outcome is inspected.
	dockerClient.On("ContainerExecInspect", mock.Anything, "kill-exec-id").
		Return(container.ExecInspect{ExecID: "kill-exec-id", Running: false, ExitCode: 0}, nil).Once()

	require.NotPanics(t, func() { env.killByPid("project-container-full-id", 1234) })
}

func TestKillByPid_NonZeroExitStillInspectsAndDoesNotPanic(t *testing.T) {
	dockerClient := dockermock.NewMockAPIClient(t)
	env := newKillTestEnvironment(dockerClient)

	dockerClient.On("ContainerExecCreate", mock.Anything, mock.Anything, mock.Anything).
		Return(types.IDResponse{ID: "kill-exec-id"}, nil).Once()
	dockerClient.On("ContainerExecAttach", mock.Anything, "kill-exec-id", mock.Anything).
		Return(hijackedWith([]byte("kill: (1234) - No such process\n")), nil).Once()
	// 137/1: the kill ran but did not do what it claims — the Warnf branch.
	dockerClient.On("ContainerExecInspect", mock.Anything, "kill-exec-id").
		Return(container.ExecInspect{ExecID: "kill-exec-id", Running: false, ExitCode: 1}, nil).Once()

	require.NotPanics(t, func() { env.killByPid("project-container-full-id", 1234) })
}

func TestKillByPid_InspectFailureDoesNotPanic(t *testing.T) {
	dockerClient := dockermock.NewMockAPIClient(t)
	env := newKillTestEnvironment(dockerClient)

	dockerClient.On("ContainerExecCreate", mock.Anything, mock.Anything, mock.Anything).
		Return(types.IDResponse{ID: "kill-exec-id"}, nil).Once()
	dockerClient.On("ContainerExecAttach", mock.Anything, "kill-exec-id", mock.Anything).
		Return(hijackedWith(nil), nil).Once()
	dockerClient.On("ContainerExecInspect", mock.Anything, "kill-exec-id").
		Return(container.ExecInspect{}, errors.New("daemon went away")).Once()

	require.NotPanics(t, func() { env.killByPid("project-container-full-id", 1234) })
}

func TestKillByPid_ExecCreateFailureStopsEarly(t *testing.T) {
	dockerClient := dockermock.NewMockAPIClient(t)
	env := newKillTestEnvironment(dockerClient)

	// No attach/inspect expectations: create fails, so nothing else may run.
	dockerClient.On("ContainerExecCreate", mock.Anything, mock.Anything, mock.Anything).
		Return(types.IDResponse{}, errors.New("no such container")).Once()

	require.NotPanics(t, func() { env.killByPid("project-container-full-id", 1234) })
}
