package docker

import (
	"bytes"
	"context"
	"io"
	"testing"

	judgmentproject "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	domainservice "judge-backend/internal/app/judge_domain/judge/domain/service"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	dockermock "judge-backend/internal/mocks/github.com/docker/docker/client"

	types "github.com/docker/docker/api/types"
	container "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestParseContainerLabel(t *testing.T) {
	assert.Equal(t, map[string]string{"judge-managed": "true"}, parseContainerLabel("judge-managed=true"))
	assert.Equal(t, map[string]string{}, parseContainerLabel(""))
	assert.Equal(t, map[string]string{}, parseContainerLabel("no-equals-sign"))
	assert.Equal(t, map[string]string{}, parseContainerLabel("=value"))
}

// runProjectContainerEnv drives runProjectContainer far enough to capture the
// Env the project container is created with.
func runProjectContainerEnv(t *testing.T, spec *domainservice.EnvironmentSpec) []string {
	dockerClient := dockermock.NewMockAPIClient(t)
	fss := sharedmocks.NewFileSystemService(t)
	r := &Runtime{
		storagePath:  "/storage",
		dockerClient: dockerClient,
		fss:          fss,
		logger:       sharedlogger.NewMockLogger(),
		networks:     map[string]string{},
	}

	var capturedEnv []string
	dockerClient.On("ContainerCreate", mock.Anything, mock.MatchedBy(func(cfg *container.Config) bool {
		capturedEnv = cfg.Env
		return true
	}), mock.Anything, mock.Anything, mock.Anything, "").
		Return(container.CreateResponse{ID: "project-container"}, nil).Once()
	dockerClient.On("ContainerStart", mock.Anything, "project-container", container.StartOptions{}).Return(nil).Once()
	fss.On("TarPath", "/storage/submissions/sub-1/project").Return(&bytes.Buffer{}, nil).Once()
	dockerClient.On("CopyToContainer", mock.Anything, "project-container", spec.Project.MountPath(), mock.Anything, mock.Anything).Return(nil).Once()
	dockerClient.On("ContainerInspect", mock.Anything, "project-container").
		Return(types.ContainerJSON{ContainerJSONBase: &types.ContainerJSONBase{Name: "/project-container-name"}}, nil).Once()

	_, name, err := r.runProjectContainer(context.Background(), spec)
	assert.NoError(t, err)
	assert.Equal(t, "project-container-name", name)
	return capturedEnv
}

func TestRuntime_RunProjectContainer_InjectsSubmitterUsername(t *testing.T) {
	env := runProjectContainerEnv(t, &domainservice.EnvironmentSpec{
		JudgmentID:   "judgment-1",
		SubmissionID: "sub-1",
		Project: judgmentproject.NewProject(&judgmentproject.ProjectConfig{
			ID:        1,
			ImageTag:  "project-web:latest",
			MountPath: "/var/www/html",
			Env:       "FOO=bar",
		}),
		SubmitterUsername: "web07",
	})

	assert.Contains(t, env, "FOO=bar", "project's own env must be preserved")
	assert.Contains(t, env, "SUBMIT_USERNAME=web07")
}

func TestRuntime_RunProjectContainer_NoSubmitterUsername_OmitsVariable(t *testing.T) {
	env := runProjectContainerEnv(t, &domainservice.EnvironmentSpec{
		JudgmentID:   "judgment-1",
		SubmissionID: "sub-1",
		Project: judgmentproject.NewProject(&judgmentproject.ProjectConfig{
			ID:        1,
			ImageTag:  "project-web:latest",
			MountPath: "/var/www/html",
		}),
	})

	for _, e := range env {
		assert.NotContains(t, e, "SUBMIT_USERNAME", "must not inject an empty submitter variable")
	}
}

func ptr(v int32) *int32 { return &v }

// Baseline hardening applies to every container regardless of the per-project
// caps: privilege escalation is blocked, container logs are bounded, and the
// CPU cap is honoured when configured.
func TestRuntime_HostConfigFor_BaselineHardening(t *testing.T) {
	spec := &domainservice.EnvironmentSpec{
		Project: judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
	}

	t.Run("no-new-privileges and bounded logs always", func(t *testing.T) {
		hc := (&Runtime{}).hostConfigFor(spec)
		assert.Contains(t, hc.SecurityOpt, "no-new-privileges")
		assert.Equal(t, "json-file", hc.LogConfig.Type)
		assert.Equal(t, "10m", hc.LogConfig.Config["max-size"])
		assert.Zero(t, hc.NanoCPUs, "no CPU cap unless configured")
	})

	t.Run("cpu cap maps whole cores to NanoCPUs when set", func(t *testing.T) {
		hc := (&Runtime{defaultCpuLimit: 2}).hostConfigFor(spec)
		assert.Equal(t, int64(2)*1_000_000_000, hc.NanoCPUs)
	})
}

type capturedCreate struct {
	mode       container.NetworkMode
	networking *network.NetworkingConfig
}

// captureContainerCreate drives one of the container-run methods and returns the
// NetworkMode (HostConfig) and NetworkingConfig the container was created with.
func captureContainerCreate(t *testing.T, isolate bool, run func(*Runtime, *domainservice.EnvironmentSpec) error) capturedCreate {
	spec := &domainservice.EnvironmentSpec{
		JudgmentID:   "judgment-1",
		SubmissionID: "sub-1",
		WorkerID:     "w1",
		Project: judgmentproject.NewProject(&judgmentproject.ProjectConfig{
			ID:           1,
			ImageTag:     "project-web:latest",
			TestImageTag: "project-test:latest",
			MountPath:    "/var/www/html",
		}),
	}

	dockerClient := dockermock.NewMockAPIClient(t)
	fss := sharedmocks.NewFileSystemService(t)
	r := &Runtime{
		storagePath:           "/storage",
		dockerClient:          dockerClient,
		fss:                   fss,
		logger:                sharedlogger.NewMockLogger(),
		networks:              map[string]string{},
		isolateProjectNetwork: isolate,
	}

	var captured capturedCreate
	dockerClient.On("ContainerCreate",
		mock.Anything,
		mock.Anything,
		mock.MatchedBy(func(hc *container.HostConfig) bool { captured.mode = hc.NetworkMode; return true }),
		mock.MatchedBy(func(nc *network.NetworkingConfig) bool { captured.networking = nc; return true }),
		mock.Anything, "").Return(container.CreateResponse{ID: "c"}, nil).Once()
	dockerClient.On("ContainerStart", mock.Anything, "c", container.StartOptions{}).Return(nil).Once()
	dockerClient.On("ContainerInspect", mock.Anything, "c").
		Return(types.ContainerJSON{ContainerJSONBase: &types.ContainerJSONBase{Name: "/c"}}, nil).Maybe()
	fss.On("Exists", mock.Anything).Return(false).Maybe()
	fss.On("TarPath", mock.Anything).Return(&bytes.Buffer{}, nil)
	dockerClient.On("CopyToContainer", mock.Anything, "c", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	assert.NoError(t, run(r, spec))
	return captured
}

// The contestant's project container must be ATTACHED TO THE INTERNAL WORKER
// NETWORK AT CREATION when isolation is on — never NetworkMode "none". none is
// an exclusive mode, so a container started with it cannot be connected to any
// other network afterwards, which broke every judgment's provisioning. The
// grading/test container keeps its own network for outbound needs (Sonar, etc.).
func TestRuntime_ProjectNetworkIsolation(t *testing.T) {
	runProject := func(r *Runtime, spec *domainservice.EnvironmentSpec) error {
		_, _, err := r.runProjectContainer(context.Background(), spec)
		return err
	}
	runTest := func(r *Runtime, spec *domainservice.EnvironmentSpec) error {
		_, err := r.runTestContainer(context.Background(), spec, nil)
		return err
	}

	iso := captureContainerCreate(t, true, runProject)
	assert.NotEqual(t, container.NetworkMode("none"), iso.mode,
		"isolated project must NOT use exclusive none mode — the daemon then forbids every later NetworkConnect")
	assert.Equal(t, container.NetworkMode(""), iso.mode)
	if assert.NotNil(t, iso.networking, "isolated project must be created already attached to a network") {
		_, ok := iso.networking.EndpointsConfig["judge-worker-w1"]
		assert.True(t, ok, "isolated project must be attached to its worker network at creation")
	}

	off := captureContainerCreate(t, false, runProject)
	assert.Equal(t, container.NetworkMode(""), off.mode,
		"with isolation off the project keeps the default bridge")
	assert.Nil(t, off.networking, "with isolation off the project joins the worker network later, not at creation")

	test := captureContainerCreate(t, true, runTest)
	assert.Equal(t, container.NetworkMode(""), test.mode,
		"the test container keeps its network even when the project is isolated")
	assert.Nil(t, test.networking)
}

// The worker bridge is only internal (no off-host route) when project isolation
// is on; otherwise it stays a normal NAT bridge.
func TestRuntime_WorkerNetworkInternalFollowsIsolation(t *testing.T) {
	for _, isolate := range []bool{true, false} {
		dockerClient := dockermock.NewMockAPIClient(t)
		r := &Runtime{
			dockerClient:          dockerClient,
			logger:                sharedlogger.NewMockLogger(),
			networks:              map[string]string{},
			isolateProjectNetwork: isolate,
		}

		var gotInternal bool
		dockerClient.On("NetworkCreate", mock.Anything, "judge-worker-w1", mock.MatchedBy(func(opts network.CreateOptions) bool {
			gotInternal = opts.Internal
			return true
		})).Return(network.CreateResponse{ID: "net-0123456789ab"}, nil).Once()

		id, err := r.ensureWorkerNetwork("w1")
		assert.NoError(t, err)
		assert.Equal(t, "net-0123456789ab", id)
		assert.Equal(t, isolate, gotInternal)
	}
}

// Every branch here fails silently in production: the container starts either
// way, and an uncapped one only reveals itself when a submission takes the
// host down. So the mapping is pinned rather than eyeballed.
func TestRuntime_HostConfigFor(t *testing.T) {
	specWith := func(pids, memoryMB *int32) *domainservice.EnvironmentSpec {
		return &domainservice.EnvironmentSpec{
			Project: judgmentproject.NewProject(&judgmentproject.ProjectConfig{
				ID:            1,
				PidsLimit:     pids,
				MemoryLimitMB: memoryMB,
			}),
		}
	}
	// The deployment defaults arrive here already normalised by
	// configuredLimit, so 0 is "no cap" rather than "unset".
	runtimeWith := func(pids, memoryMB int64) *Runtime {
		return &Runtime{defaultPidsLimit: pids, defaultMemoryLimitMB: memoryMB}
	}

	t.Run("contest project overrides the deployment default", func(t *testing.T) {
		hc := runtimeWith(1024, 0).hostConfigFor(specWith(ptr(256), ptr(512)))

		assert.Equal(t, int64(256), *hc.PidsLimit)
		assert.Equal(t, int64(512*1024*1024), hc.Memory)
	})

	t.Run("an unset contest project falls back to the deployment default", func(t *testing.T) {
		hc := runtimeWith(1024, 2048).hostConfigFor(specWith(nil, nil))

		assert.Equal(t, int64(1024), *hc.PidsLimit)
		assert.Equal(t, int64(2048*1024*1024), hc.Memory)
	})

	// The reason the columns are nullable at all. If 0 collapsed into "unset"
	// the admin UI would have no way to express "this problem legitimately
	// needs more than the house limit" short of editing config.yaml.
	t.Run("zero on the contest project removes the cap", func(t *testing.T) {
		hc := runtimeWith(1024, 2048).hostConfigFor(specWith(ptr(0), ptr(0)))

		assert.Nil(t, hc.PidsLimit)
		assert.Zero(t, hc.Memory)
	})

	// The shipped state: pids capped, memory not.
	t.Run("a zero deployment default leaves that cap off", func(t *testing.T) {
		hc := runtimeWith(1024, 0).hostConfigFor(specWith(nil, nil))

		assert.Equal(t, int64(1024), *hc.PidsLimit)
		assert.Zero(t, hc.Memory)
	})

	t.Run("a zeroed deployment default means no cap", func(t *testing.T) {
		hc := runtimeWith(0, 0).hostConfigFor(specWith(nil, nil))

		assert.Nil(t, hc.PidsLimit)
		assert.Zero(t, hc.Memory)
	})

	// Without this the cgroup gets swap equal to Memory on top, and a
	// container over its cap keeps running instead of being stopped — the
	// limit would show up in docker inspect and enforce nothing.
	t.Run("swap is pinned to the memory cap so the cap is real", func(t *testing.T) {
		hc := runtimeWith(0, 0).hostConfigFor(specWith(nil, ptr(512)))

		assert.Equal(t, hc.Memory, hc.MemorySwap)
	})
}

// Both containers execute contestant-controlled commands, so a cap on only
// one of them is a cap on neither.
func TestRuntime_BothContainersGetTheLimits(t *testing.T) {
	spec := &domainservice.EnvironmentSpec{
		JudgmentID:   "judgment-1",
		SubmissionID: "sub-1",
		Project: judgmentproject.NewProject(&judgmentproject.ProjectConfig{
			ID:            1,
			ImageTag:      "project-web:latest",
			TestImageTag:  "project-test:latest",
			MountPath:     "/var/www/html",
			PidsLimit:     ptr(256),
			MemoryLimitMB: ptr(512),
		}),
	}

	dockerClient := dockermock.NewMockAPIClient(t)
	fss := sharedmocks.NewFileSystemService(t)
	r := &Runtime{
		storagePath:  "/storage",
		dockerClient: dockerClient,
		fss:          fss,
		logger:       sharedlogger.NewMockLogger(),
		networks:     map[string]string{},
	}

	var captured []*container.HostConfig
	dockerClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.MatchedBy(func(hc *container.HostConfig) bool {
		captured = append(captured, hc)
		return true
	}), mock.Anything, mock.Anything, "").
		Return(container.CreateResponse{ID: "c"}, nil).Twice()
	dockerClient.On("ContainerStart", mock.Anything, "c", container.StartOptions{}).Return(nil).Twice()
	dockerClient.On("ContainerInspect", mock.Anything, "c").
		Return(types.ContainerJSON{ContainerJSONBase: &types.ContainerJSONBase{Name: "/c-name"}}, nil).Once()
	fss.On("TarPath", "/storage/submissions/sub-1/project").Return(&bytes.Buffer{}, nil).Twice()
	fss.On("TarPath", "/storage/submissions/sub-1/files").Return(&bytes.Buffer{}, nil).Once()
	dockerClient.On("CopyToContainer", mock.Anything, "c", "/var/www/html", mock.Anything, mock.Anything).Return(nil).Twice()
	dockerClient.On("CopyToContainer", mock.Anything, "c", "/files", mock.Anything, mock.Anything).Return(nil).Once()
	fss.On("Exists", "/storage/projects/1/files").Return(false).Once()

	_, _, err := r.runProjectContainer(context.Background(), spec)
	assert.NoError(t, err)
	_, err = r.runTestContainer(context.Background(), spec, nil)
	assert.NoError(t, err)

	assert.Len(t, captured, 2)
	for i, hc := range captured {
		// A nil host config is how this regresses — someone reverts one call
		// site to the `nil` it used to pass. Say so instead of panicking.
		if !assert.NotNil(t, hc, "container %d was created with no host config, so it is uncapped", i) {
			continue
		}
		if assert.NotNil(t, hc.PidsLimit, "container %d has no pids limit", i) {
			assert.Equal(t, int64(256), *hc.PidsLimit)
		}
		assert.Equal(t, int64(512*1024*1024), hc.Memory)
	}
}

// muxFrame builds one stdcopy-multiplexed frame as the Docker daemon would
// emit it: 8-byte header {stream, 0, 0, 0, len(BE32)} followed by payload.
func muxFrame(stream byte, payload string) []byte {
	frame := make([]byte, 8+len(payload))
	frame[0] = stream
	frame[4] = byte(len(payload) >> 24)
	frame[5] = byte(len(payload) >> 16)
	frame[6] = byte(len(payload) >> 8)
	frame[7] = byte(len(payload))
	copy(frame[8:], payload)
	return frame
}

func TestRuntime_DumpContainerLogs_WritesDemuxedTailToSink(t *testing.T) {
	dockerClient := dockermock.NewMockAPIClient(t)
	r := &Runtime{dockerClient: dockerClient, logger: sharedlogger.NewMockLogger()}

	payload := append(muxFrame(1, "Starting MariaDB...\n"), muxFrame(2, "MariaDB did not become available in time\n")...)
	dockerClient.On("ContainerLogs", mock.Anything, "container-1234567890", mock.MatchedBy(func(opts container.LogsOptions) bool {
		return opts.ShowStdout && opts.ShowStderr && !opts.Follow && opts.Tail == provisionDiagnosticsTailLines
	})).Return(io.NopCloser(bytes.NewReader(payload)), nil).Once()

	var sink bytes.Buffer
	r.dumpContainerLogs(&sink, "container-1234567890", "project")

	out := sink.String()
	assert.Contains(t, out, "--- project container logs (last 50 lines) ---")
	assert.Contains(t, out, "Starting MariaDB...", "stdout frames must be demuxed into the sink")
	assert.Contains(t, out, "MariaDB did not become available in time", "stderr frames must be demuxed into the sink")
	assert.Contains(t, out, "--- end of project container logs ---")
}

func TestRuntime_DumpContainerLogs_FetchFailure_IsSilent(t *testing.T) {
	dockerClient := dockermock.NewMockAPIClient(t)
	r := &Runtime{dockerClient: dockerClient, logger: sharedlogger.NewMockLogger()}

	dockerClient.On("ContainerLogs", mock.Anything, "container-1234567890", mock.Anything).
		Return(nil, assert.AnError).Once()

	var sink bytes.Buffer
	r.dumpContainerLogs(&sink, "container-1234567890", "project")

	assert.Empty(t, sink.String(), "a diagnostics fetch failure must not write partial output")
}

func TestRuntime_DumpContainerLogs_NilSink_SkipsDockerCall(t *testing.T) {
	// no ContainerLogs expectation registered — any call would fail the test
	dockerClient := dockermock.NewMockAPIClient(t)
	r := &Runtime{dockerClient: dockerClient, logger: sharedlogger.NewMockLogger()}

	r.dumpContainerLogs(nil, "container-1234567890", "project")
}

func TestRuntime_CleanupOrphans_RemovesOrphanContainersAndNetworks(t *testing.T) {
	dockerClient := dockermock.NewMockAPIClient(t)
	r := &Runtime{
		rawContainerLabel: "judge-managed=true",
		dockerClient:      dockerClient,
		logger:            sharedlogger.NewMockLogger(),
		networks:          map[string]string{},
	}

	dockerClient.On("ContainerList", mock.Anything, mock.AnythingOfType("container.ListOptions")).
		Return([]types.Container{
			{ID: "container-1234567890", Labels: map[string]string{"judgment-id": "judgment-1"}},
			{ID: "container-active-9876", Labels: map[string]string{"judgment-id": "judgment-active"}},
		}, nil).Once()
	dockerClient.On("ContainerRemove", mock.Anything, "container-1234567890", container.RemoveOptions{Force: true}).Return(nil).Once()
	dockerClient.On("NetworkList", mock.Anything, mock.AnythingOfType("network.ListOptions")).
		Return([]network.Inspect{
			{ID: "network-1234567890", Name: "judge-worker-orphan", Labels: map[string]string{"worker-id": "orphan-worker-id"}},
			{ID: "network-active-9876", Name: "judge-worker-active", Labels: map[string]string{"worker-id": "active-worker-id"}},
		}, nil).Once()
	dockerClient.On("NetworkRemove", mock.Anything, "network-1234567890").Return(nil).Once()

	r.CleanupOrphans(
		map[string]struct{}{"judgment-active": {}},
		map[string]struct{}{"active-worker-id": {}},
	)
}

func TestRuntime_CleanupOrphans_SkipsWhenLabelNotConfigured(t *testing.T) {
	dockerClient := dockermock.NewMockAPIClient(t)
	r := &Runtime{
		rawContainerLabel: "",
		dockerClient:      dockerClient,
		logger:            sharedlogger.NewMockLogger(),
		networks:          map[string]string{},
	}

	// no docker calls expected at all
	r.CleanupOrphans(map[string]struct{}{}, map[string]struct{}{})
}

func TestRuntime_ReleaseWorker_RemovesTrackedNetwork(t *testing.T) {
	dockerClient := dockermock.NewMockAPIClient(t)
	r := &Runtime{
		dockerClient: dockerClient,
		logger:       sharedlogger.NewMockLogger(),
		networks:     map[string]string{"worker-1": "network-abcdef123456"},
	}

	dockerClient.On("NetworkRemove", mock.Anything, "network-abcdef123456").Return(nil).Once()

	r.ReleaseWorker("worker-1")
	assert.Empty(t, r.networks)

	// releasing an unknown worker is a no-op
	r.ReleaseWorker("worker-unknown")
}

// The config keys are pointers precisely so an operator's explicit 0 is not
// mistaken for an absent key. Normalising them is where that distinction gets
// spent, so it is worth its own test.
func TestConfiguredLimit(t *testing.T) {
	// Cannot happen after config.applyDefaults, but a Runtime built from a
	// config that skipped it must not invent a cap out of nothing.
	assert.Equal(t, int64(0), configuredLimit(nil), "an unset key is no cap, never a guessed one")

	// The case the pointer exists for.
	zero := 0
	assert.Equal(t, int64(0), configuredLimit(&zero), "an explicit 0 is no cap")

	// Still honoured: an earlier revision of the config docs said to write this.
	negative := -1
	assert.Equal(t, int64(0), configuredLimit(&negative), "a negative value is no cap")

	value := 1024
	assert.Equal(t, int64(1024), configuredLimit(&value))
}
