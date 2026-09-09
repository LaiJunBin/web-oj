package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	domainprovider "judge-backend/internal/app/judge_domain/judge/domain/provider"
	domainservice "judge-backend/internal/app/judge_domain/judge/domain/service"
	"judge-backend/internal/app/shared/infrastructure/metrics"
	sharedservice "judge-backend/internal/app/shared/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Runtime is the Docker implementation of the judge domain's
// ExecutionEnvironmentProvisioner port. All Docker SDK usage of the judge
// domain lives in this package.
type Runtime struct {
	storagePath        string
	containerLabel     map[string]string
	rawContainerLabel  string
	healthcheckTimeout time.Duration
	// Deployment-wide fallbacks for the caps a contest project leaves unset,
	// already normalised so 0 means no cap.
	defaultPidsLimit     int64
	defaultMemoryLimitMB int64
	defaultCpuLimit      int64
	// maxOutputBytes bounds captured stdout+stderr per exec; 0 means no cap.
	maxOutputBytes int64
	// isolateProjectNetwork denies the contestant's project container external
	// network access (see JudgeConfig.IsolateProjectNetwork).
	isolateProjectNetwork bool
	fss                   sharedservice.FileSystemService
	dockerClient          client.APIClient
	logger                sharedtypes.Logger

	networkMu sync.Mutex
	networks  map[string]string // workerID → networkID
}

func NewRuntime(
	storagePathProvider domainprovider.StoragePathProvider,
	cfg *config.Config,
	fss sharedservice.FileSystemService,
	dockerClient client.APIClient,
	logger sharedtypes.Logger,
) domainservice.ExecutionEnvironmentProvisioner {
	return &Runtime{
		storagePath:        storagePathProvider.GetStoragePath(),
		containerLabel:     parseContainerLabel(cfg.Judge.ContainerLabel),
		rawContainerLabel:  cfg.Judge.ContainerLabel,
		healthcheckTimeout: time.Duration(cfg.Judge.HealthcheckTimeoutSeconds) * time.Second,

		defaultPidsLimit:     configuredLimit(cfg.Judge.PidsLimit),
		defaultMemoryLimitMB: configuredLimit(cfg.Judge.MemoryLimitMB),
		defaultCpuLimit:      configuredLimit(cfg.Judge.CpuLimit),
		maxOutputBytes:       configuredLimit(cfg.Judge.MaxOutputBytesPerExec),

		isolateProjectNetwork: cfg.Judge.IsolateProjectNetwork == nil || *cfg.Judge.IsolateProjectNetwork,

		fss:          fss,
		dockerClient: dockerClient,
		logger:       logger,
		networks:     map[string]string{},
	}
}

// parseContainerLabel parses a "key=value" label string into a Docker labels map.
func parseContainerLabel(label string) map[string]string {
	parts := strings.SplitN(label, "=", 2)
	if len(parts) != 2 || parts[0] == "" {
		return map[string]string{}
	}
	return map[string]string{parts[0]: parts[1]}
}

// configuredLimit normalises a deployment default to "0 means no cap".
//
// A nil pids limit cannot occur in practice — config.applyDefaults fills it —
// but a Runtime built from a config that skipped that step must not silently
// invent a cap, so nil resolves to no cap rather than to 1024. A negative
// value is accepted as no cap too: it is what an earlier revision of the
// config documentation told operators to write.
func configuredLimit(configured *int) int64 {
	if configured == nil || *configured <= 0 {
		return 0
	}
	return int64(*configured)
}

// resolveLimit picks the cap in force for one container, in units of
// whatever is being capped. 0 means no cap, at every layer: the contest
// project column, the config key, and here.
//
// nil is the one thing only the contest project column can say, and it means
// "no opinion, use the deployment default" — which is why the column is
// nullable rather than defaulted at the database.
func resolveLimit(override *int32, deploymentDefault int64) int64 {
	if override != nil {
		if *override <= 0 {
			return 0
		}
		return int64(*override)
	}
	return deploymentDefault
}

// hostConfigFor builds the resource limits both of a judgment's containers run
// under. Both get the same caps: a fork bomb in a test command is no less
// effective than one in the submitted project.
func (r *Runtime) hostConfigFor(spec *domainservice.EnvironmentSpec) *container.HostConfig {
	limits := spec.Project.Limits()
	hostConfig := &container.HostConfig{
		// A submission cannot gain privileges beyond what its image already
		// grants — setuid binaries inside the container can no longer escalate.
		SecurityOpt: []string{"no-new-privileges"},
		// Cap the container's own log so a chatty (or hostile) submission cannot
		// fill the host disk through the json-file driver. The judge reads output
		// through the exec/attach stream, not these files, so one small rotation
		// is plenty. json-file also keeps ContainerLogs readable for diagnostics.
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{"max-size": "10m", "max-file": "1"},
		},
	}

	if pids := resolveLimit(limits.Pids, r.defaultPidsLimit); pids > 0 {
		hostConfig.PidsLimit = &pids
	}

	if memoryMB := resolveLimit(limits.MemoryMB, r.defaultMemoryLimitMB); memoryMB > 0 {
		memoryBytes := memoryMB * 1024 * 1024
		hostConfig.Memory = memoryBytes
		// MemorySwap must be set with Memory, not left alone. Its default
		// grants the cgroup swap equal to Memory on top, so a container past
		// its cap keeps running — slowly — instead of being stopped. The
		// limit would read as enforced in every inspect output and enforce
		// nothing. Equal values disable swap for the cgroup.
		hostConfig.MemorySwap = memoryBytes
	}

	if r.defaultCpuLimit > 0 {
		// NanoCPUs is whole CPUs times 1e9; the config is expressed in whole
		// CPUs to match the pids/memory knobs.
		hostConfig.NanoCPUs = r.defaultCpuLimit * 1_000_000_000
	}

	return hostConfig
}

// labelsForJudgment returns Docker labels for a container belonging to a specific judgment.
func (r *Runtime) labelsForJudgment(judgmentID string) map[string]string {
	labels := make(map[string]string, len(r.containerLabel)+1)
	for k, v := range r.containerLabel {
		labels[k] = v
	}
	labels["judgment-id"] = judgmentID
	return labels
}

func (r *Runtime) Provision(ctx context.Context, spec *domainservice.EnvironmentSpec) (domainservice.ExecutionEnvironment, error) {
	env, err := r.provisionEnvironment(ctx, spec)
	metrics.ObserveProvisionAttempt(err == nil)
	return env, err
}

func (r *Runtime) provisionEnvironment(ctx context.Context, spec *domainservice.EnvironmentSpec) (domainservice.ExecutionEnvironment, error) {
	r.logger.Infof("[DockerRuntime] Provisioning environment for judgment %s (submission %s)", spec.JudgmentID, spec.SubmissionID)

	networkID, err := r.ensureWorkerNetwork(spec.WorkerID)
	if err != nil {
		return nil, err
	}

	projectContainerID, projectContainerName, err := r.runProjectContainer(ctx, spec)
	if err != nil {
		return nil, err
	}

	testEnv := spec.Project.TestEnv()
	for k, v := range spec.Metadata {
		testEnv = append(testEnv, fmt.Sprintf("PAYLOAD_%s=%s", k, v))
	}
	testEnv = append(testEnv, fmt.Sprintf("PROJECT_CONTAINER_NAME=%s", projectContainerName))
	if spec.SubmitterUsername != "" {
		testEnv = append(testEnv, fmt.Sprintf("SUBMIT_USERNAME=%s", spec.SubmitterUsername))
	}

	testContainerID, err := r.runTestContainer(ctx, spec, testEnv)
	if err != nil {
		r.removeContainer(projectContainerID)
		return nil, err
	}

	env := &environment{
		runtime:            r,
		judgmentID:         spec.JudgmentID,
		projectContainerID: projectContainerID,
		testContainerID:    testContainerID,
		testEnv:            testEnv,
	}

	// Both containers are now starting. Wait for the project container to be
	// healthy before proceeding — test container file-copy time overlaps with
	// application (e.g. MariaDB) startup.
	if err := r.waitForProjectHealthy(ctx, projectContainerID, spec.SubmissionID); err != nil {
		// The container's final words are the only evidence of WHY it died
		// or went unhealthy — capture them before the teardown erases them.
		r.dumpContainerLogs(spec.DiagnosticsSink, projectContainerID, "project")
		env.Teardown()
		return nil, err
	}

	// When isolation is on, the project container was already attached to the
	// worker network at creation (none-mode would forbid connecting afterwards),
	// and on a user-defined network it is reachable by its container name via
	// Docker's embedded DNS, so no explicit alias is needed. Otherwise it started
	// on the default bridge and joins the worker network now, once healthy.
	if !r.isolateProjectNetwork {
		if err := r.dockerClient.NetworkConnect(ctx, networkID, projectContainerID, &dockernetwork.EndpointSettings{
			Aliases: []string{projectContainerName},
		}); err != nil {
			r.logger.Errorf("[DockerRuntime] Failed to connect project container %s to network %s: %v", projectContainerID, networkID, err)
			env.Teardown()
			return nil, err
		}
	}
	if err := r.dockerClient.NetworkConnect(ctx, networkID, testContainerID, nil); err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to connect test container %s to network %s: %v", testContainerID, networkID, err)
		env.Teardown()
		return nil, err
	}

	r.logger.Infof("[DockerRuntime] Provisioned environment for judgment %s (project %s, test %s)", spec.JudgmentID, projectContainerID[:12], testContainerID[:12])
	return env, nil
}

// workerNetworkName is the stable name of a worker's bridge network. Using the
// name (not the id) as the endpoint key when attaching a container survives a
// daemon-side recreation of the network, since the name is what stays constant.
func workerNetworkName(workerID string) string {
	return fmt.Sprintf("judge-worker-%s", workerID)
}

// ensureWorkerNetwork returns the worker's isolated bridge network, creating
// or recreating it when missing (e.g. after a Docker daemon restart).
func (r *Runtime) ensureWorkerNetwork(workerID string) (string, error) {
	r.networkMu.Lock()
	defer r.networkMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if networkID, ok := r.networks[workerID]; ok {
		if _, err := r.dockerClient.NetworkInspect(ctx, networkID, dockernetwork.InspectOptions{}); err == nil {
			return networkID, nil
		}
		r.logger.Warnf("[DockerRuntime] Worker %s network %s not found, recreating", workerID, networkID)
	}

	name := workerNetworkName(workerID)
	resp, err := r.dockerClient.NetworkCreate(ctx, name, dockernetwork.CreateOptions{
		Driver: "bridge",
		// Internal when project isolation is on: containers on this network can
		// talk to each other (test → project) but the network itself provides no
		// route off-host. The test container keeps a separate, non-internal
		// bridge for its outbound needs; the project container has only this one.
		Internal: r.isolateProjectNetwork,
		Labels:   map[string]string{"worker-id": workerID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create worker network: %w", err)
	}
	r.networks[workerID] = resp.ID
	r.logger.Infof("[DockerRuntime] Created network %s for worker %s", resp.ID[:12], workerID)
	return resp.ID, nil
}

func (r *Runtime) ReleaseWorker(workerID string) {
	r.networkMu.Lock()
	networkID, ok := r.networks[workerID]
	delete(r.networks, workerID)
	r.networkMu.Unlock()
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := r.dockerClient.NetworkRemove(ctx, networkID); err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to remove network %s of worker %s: %v", networkID[:12], workerID, err)
		return
	}
	r.logger.Debugf("[DockerRuntime] Removed network %s of worker %s", networkID[:12], workerID)
}

func (r *Runtime) runProjectContainer(ctx context.Context, spec *domainservice.EnvironmentSpec) (string, string, error) {
	r.logger.Debugf("[DockerRuntime] Running project container for submission %s", spec.SubmissionID)
	env := spec.Project.Env()
	if spec.SubmitterUsername != "" {
		env = append(env, fmt.Sprintf("SUBMIT_USERNAME=%s", spec.SubmitterUsername))
	}
	hostConfig := r.hostConfigFor(spec)
	var networkingConfig *dockernetwork.NetworkingConfig
	if r.isolateProjectNetwork {
		// Attach directly to the internal worker network AT CREATION. Isolation
		// is that network being Internal (no off-host route), NOT NetworkMode
		// "none": none is an exclusive mode, so a container started with it can
		// never be connected to another network afterwards — the daemon rejects
		// it with "container cannot be connected to multiple networks with one of
		// the networks in private (none) mode". The test container joins the same
		// network only after the health check, so it still cannot reach the
		// project until the project is ready. The network is keyed by name so a
		// daemon-side recreate cannot leave a stale id here.
		networkingConfig = &dockernetwork.NetworkingConfig{
			EndpointsConfig: map[string]*dockernetwork.EndpointSettings{
				workerNetworkName(spec.WorkerID): {},
			},
		}
	}
	resp, err := r.dockerClient.ContainerCreate(ctx, &container.Config{
		Image: spec.Project.ImageTag(),
		Volumes: map[string]struct{}{
			spec.Project.MountPath(): {},
		},
		Env:    env,
		Labels: r.labelsForJudgment(spec.JudgmentID),
	}, hostConfig, networkingConfig, nil, "")
	if err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to create project container for submission %s: %v", spec.SubmissionID, err)
		return "", "", err
	}

	if err := r.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to start project container %s for submission %s: %v", resp.ID, spec.SubmissionID, err)
		r.removeContainer(resp.ID)
		return "", "", err
	}

	if err := r.copyPathToContainer(ctx, resp.ID, r.submissionProjectPath(spec.SubmissionID), spec.Project.MountPath()); err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to copy project files to container %s for submission %s: %v", resp.ID, spec.SubmissionID, err)
		r.removeContainer(resp.ID)
		return "", "", err
	}

	info, err := r.dockerClient.ContainerInspect(ctx, resp.ID)
	if err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to inspect project container %s for submission %s: %v", resp.ID, spec.SubmissionID, err)
		r.removeContainer(resp.ID)
		return "", "", err
	}

	name := info.Name[1:]
	r.logger.Debugf("[DockerRuntime] Started project container %s (%s) for submission %s", resp.ID, name, spec.SubmissionID)
	return resp.ID, name, nil
}

func (r *Runtime) runTestContainer(ctx context.Context, spec *domainservice.EnvironmentSpec, env []string) (string, error) {
	r.logger.Debugf("[DockerRuntime] Running test container for submission %s", spec.SubmissionID)
	resp, err := r.dockerClient.ContainerCreate(ctx, &container.Config{
		Image: spec.Project.TestImageTag(),
		Volumes: map[string]struct{}{
			"/files":                 {},
			"/project":               {},
			spec.Project.MountPath(): {},
		},
		Env:        env,
		WorkingDir: "/project",
		Labels:     r.labelsForJudgment(spec.JudgmentID),
	}, r.hostConfigFor(spec), nil, nil, "")
	if err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to create test container for submission %s: %v", spec.SubmissionID, err)
		return "", err
	}

	if err := r.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to start test container %s for submission %s: %v", resp.ID, spec.SubmissionID, err)
		r.removeContainer(resp.ID)
		return "", err
	}

	gradingProjectFilesPath := fmt.Sprintf("%s/projects/%d/files", r.storagePath, spec.Project.ID())
	if r.fss.Exists(gradingProjectFilesPath) {
		if err := r.copyPathToContainer(ctx, resp.ID, gradingProjectFilesPath, "/project"); err != nil {
			r.logger.Errorf("[DockerRuntime] Failed to copy grading project files to container %s for submission %s: %v", resp.ID, spec.SubmissionID, err)
			r.removeContainer(resp.ID)
			return "", err
		}
	}

	if err := r.copyPathToContainer(ctx, resp.ID, r.submissionFilesPath(spec.SubmissionID), "/files"); err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to copy payload files to container %s for submission %s: %v", resp.ID, spec.SubmissionID, err)
		r.removeContainer(resp.ID)
		return "", err
	}

	if err := r.copyPathToContainer(ctx, resp.ID, r.submissionProjectPath(spec.SubmissionID), spec.Project.MountPath()); err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to copy project files to container %s for submission %s: %v", resp.ID, spec.SubmissionID, err)
		r.removeContainer(resp.ID)
		return "", err
	}

	r.logger.Debugf("[DockerRuntime] Started test container %s for submission %s", resp.ID, spec.SubmissionID)
	return resp.ID, nil
}

func (r *Runtime) submissionProjectPath(submissionID string) string {
	return fmt.Sprintf("%s/submissions/%s/project", r.storagePath, submissionID)
}

func (r *Runtime) submissionFilesPath(submissionID string) string {
	return fmt.Sprintf("%s/submissions/%s/files", r.storagePath, submissionID)
}

func (r *Runtime) copyPathToContainer(ctx context.Context, containerID string, hostPath string, containerPath string) error {
	content, err := r.fss.TarPath(hostPath)
	if err != nil {
		return fmt.Errorf("failed to tar %s: %w", hostPath, err)
	}
	return r.dockerClient.CopyToContainer(ctx, containerID, containerPath, content, container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	})
}

// provisionDiagnosticsTailLines bounds each capture: the worker log ring
// holds 500 entries, and a retried provisioning failure dumps once per
// attempt — 50 lines × 3 attempts stays well clear of evicting everything
// else.
const provisionDiagnosticsTailLines = "50"

// dumpContainerLogs writes the container's last log lines into sink so a
// container that fails during provisioning leaves its final words in the
// worker log instead of vanishing with the teardown. Best-effort by
// contract: sink may be nil, and a failure to fetch logs must never mask
// the provisioning error this is diagnosing — its own context keeps a hung
// daemon from stalling the failure path.
func (r *Runtime) dumpContainerLogs(sink io.Writer, containerID string, role string) {
	if sink == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logs, err := r.dockerClient.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       provisionDiagnosticsTailLines,
	})
	if err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to fetch diagnostics logs for %s container %s: %v", role, containerID[:12], err)
		return
	}
	defer logs.Close()

	fmt.Fprintf(sink, "--- %s container logs (last %s lines) ---\n", role, provisionDiagnosticsTailLines)
	if _, err := stdcopy.StdCopy(sink, sink, logs); err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to copy diagnostics logs for %s container %s: %v", role, containerID[:12], err)
	}
	fmt.Fprintf(sink, "--- end of %s container logs ---\n", role)
}

func (r *Runtime) waitForProjectHealthy(ctx context.Context, containerID string, submissionID string) error {
	deadline := time.Now().Add(r.healthcheckTimeout)
	for {
		info, err := r.dockerClient.ContainerInspect(ctx, containerID)
		if err != nil {
			return fmt.Errorf("failed to inspect project container: %w", err)
		}

		if !info.State.Running {
			return fmt.Errorf("project container exited with code %d", info.State.ExitCode)
		}

		// No HEALTHCHECK defined — nothing to wait for
		if info.State.Health == nil {
			return nil
		}

		switch info.State.Health.Status {
		case "healthy":
			r.logger.Infof("[DockerRuntime] Project container %s is healthy for submission %s", containerID, submissionID)
			return nil
		case "unhealthy":
			return fmt.Errorf("project container %s became unhealthy", containerID)
		default:
			if time.Now().After(deadline) {
				r.logger.Errorf("[DockerRuntime] Project container %s health check timed out after %v for submission %s (last status: %s)", containerID, r.healthcheckTimeout, submissionID, info.State.Health.Status)
				return fmt.Errorf("timed out waiting for project container %s to become healthy", containerID)
			}
			r.logger.Debugf("[DockerRuntime] Project container %s health: %s, waiting...", containerID, info.State.Health.Status)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}

// removeContainer force-removes a container outside the judgment's cancellable
// context so cleanup still happens after cancellation.
func (r *Runtime) removeContainer(containerID string) {
	if containerID == "" {
		return
	}
	if err := r.dockerClient.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true}); err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to remove container %s: %v", containerID, err)
	}
}

func (r *Runtime) CleanupOrphans(activeJudgmentIDs map[string]struct{}, activeWorkerIDs map[string]struct{}) {
	if r.rawContainerLabel == "" {
		r.logger.Warnf("[DockerRuntime] ContainerLabel is not configured, skipping orphan cleanup")
		return
	}

	ctx := context.Background()

	containerFilter := filters.NewArgs(filters.Arg("label", r.rawContainerLabel))
	r.logger.Infof("[DockerRuntime] Starting orphan cleanup (label: %s, active judgments: %d, active workers: %d)", r.rawContainerLabel, len(activeJudgmentIDs), len(activeWorkerIDs))

	containers, err := r.dockerClient.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: containerFilter,
	})
	if err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to list containers: %v", err)
	} else {
		removed := 0
		for _, c := range containers {
			judgmentID := c.Labels["judgment-id"]
			if _, active := activeJudgmentIDs[judgmentID]; active {
				r.logger.Debugf("[DockerRuntime] Skipping container %s (judgment %s is active)", c.ID[:12], judgmentID)
				continue
			}
			r.logger.Infof("[DockerRuntime] Removing orphan container %s (judgment %s)", c.ID[:12], judgmentID)
			if err := r.dockerClient.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true}); err != nil {
				r.logger.Errorf("[DockerRuntime] Failed to remove container %s: %v", c.ID[:12], err)
				continue
			}
			removed++
		}
		r.logger.Infof("[DockerRuntime] Removed %d/%d orphan containers", removed, len(containers))
	}

	// Networks are worker-scoped (one persistent network per worker, labelled
	// with "worker-id"). Remove any network whose worker is no longer active.
	networkFilter := filters.NewArgs(filters.Arg("label", "worker-id"))
	networks, err := r.dockerClient.NetworkList(ctx, dockernetwork.ListOptions{
		Filters: networkFilter,
	})
	if err != nil {
		r.logger.Errorf("[DockerRuntime] Failed to list networks: %v", err)
	} else {
		removed := 0
		for _, n := range networks {
			workerID := n.Labels["worker-id"]
			if _, active := activeWorkerIDs[workerID]; active {
				r.logger.Debugf("[DockerRuntime] Skipping network %s (worker %s is active)", n.ID[:12], workerID)
				continue
			}
			r.logger.Infof("[DockerRuntime] Removing orphan network %s (%s, worker %s)", n.ID[:12], n.Name, workerID)
			if err := r.dockerClient.NetworkRemove(ctx, n.ID); err != nil {
				r.logger.Errorf("[DockerRuntime] Failed to remove network %s: %v", n.ID[:12], err)
				continue
			}
			removed++
		}
		r.logger.Infof("[DockerRuntime] Removed %d/%d orphan networks", removed, len(networks))
	}

	r.logger.Infof("[DockerRuntime] Orphan cleanup completed")
}
