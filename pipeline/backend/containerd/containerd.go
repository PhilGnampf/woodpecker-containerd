package containerd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	"github.com/urfave/cli/v3"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// containerdBackend implements the Backend interface using containerd.
type containerdBackend struct {
	client    *containerd.Client
	endpoint  string
	namespace string
}

// New returns a new containerd backend.
func New() backend.Backend {
	return &containerdBackend{}
}

// Name returns the backend name.
func (b *containerdBackend) Name() string {
	return "containerd"
}

// Flags returns CLI flags for this backend.
func (b *containerdBackend) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "backend-containerd-endpoint",
			Usage: "containerd socket endpoint",
			Value: "/run/containerd/containerd.sock",
		},
		&cli.StringFlag{
			Name:  "backend-containerd-namespace",
			Usage: "containerd namespace to use",
			Value: "woodpecker",
		},
	}
}

// IsAvailable checks if containerd is reachable.
func (b *containerdBackend) IsAvailable(ctx context.Context) bool {
	endpoint := "/run/containerd/containerd.sock"
	if c, ok := ctx.Value(backend.CliCommand).(*cli.Command); ok {
		endpoint = c.String("backend-containerd-endpoint")
	}
	client, err := containerd.New(endpoint)
	if err != nil {
		return false
	}
	client.Close()
	return true
}

// Load initializes the containerd client and reports backend info.
func (b *containerdBackend) Load(ctx context.Context) (*backend.BackendInfo, error) {
	c, ok := ctx.Value(backend.CliCommand).(*cli.Command)
	if !ok {
		return nil, fmt.Errorf("no CLI context found for containerd backend")
	}
	b.endpoint = c.String("backend-containerd-endpoint")
	b.namespace = c.String("backend-containerd-namespace")
	client, err := containerd.New(b.endpoint)
	if err != nil {
		return nil, err
	}
	b.client = client
	return &backend.BackendInfo{Platform: "containerd"}, nil
}

// SetupWorkflow performs any initialization before running steps.
func (b *containerdBackend) SetupWorkflow(ctx context.Context, conf *backend.Config, taskUUID string) error {
	return nil
}

// StartStep creates and starts a containerd task for the step.
func (b *containerdBackend) StartStep(ctx context.Context, step *backend.Step, taskUUID string) error {
	ctx = namespaces.WithNamespace(ctx, b.namespace)

	image, err := b.client.Pull(ctx, step.Image, containerd.WithPullUnpack)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	id := taskUUID + "-" + step.UUID
	snapshotName := id + "-snapshot"

	specOpts := []oci.SpecOpts{
		oci.WithImageConfig(image),
	}
	if len(step.Commands) > 0 {
		cmdStr := strings.Join(step.Commands, " && ")
		specOpts = append(specOpts, oci.WithProcessArgs("/bin/sh", "-c", cmdStr))
	}

	container, err := b.client.NewContainer(
		ctx,
		id,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(snapshotName, image),
		containerd.WithNewSpec(specOpts...),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	if err := task.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	return nil
}

// WaitStep waits for the step's task to finish.
func (b *containerdBackend) WaitStep(ctx context.Context, step *backend.Step, taskUUID string) (*backend.State, error) {
	ctx = namespaces.WithNamespace(ctx, b.namespace)
	id := taskUUID + "-" + step.UUID

	container, err := b.client.LoadContainer(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load container: %w", err)
	}
	task, err := container.Task(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("load task: %w", err)
	}
	exitStatusC, err := task.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("task wait: %w", err)
	}
	status := <-exitStatusC
	code, _, err := status.Result()
	if err != nil {
		return nil, fmt.Errorf("exit result: %w", err)
	}

	return &backend.State{
		Exited:   true,
		ExitCode: int(code),
	}, nil
}

func (b *containerdBackend) TailStep(ctx context.Context, step *backend.Step, taskUUID string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}


func (b *containerdBackend) DestroyStep(ctx context.Context, step *backend.Step, taskUUID string) error {
	ctx = namespaces.WithNamespace(ctx, b.namespace)
	id := taskUUID + "-" + step.UUID

	container, err := b.client.LoadContainer(ctx, id)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("load container: %w", err)
	}

	task, err := container.Task(ctx, nil)
	if err == nil {
		if killErr := task.Kill(ctx, syscall.SIGKILL); killErr != nil && !strings.Contains(killErr.Error(), "not found") {
			return fmt.Errorf("kill task: %w", killErr)
		}

		waitC, waitErr := task.Wait(ctx)
		if waitErr == nil {
			select {
			case <-waitC:
			case <-ctx.Done():
			}
		}

		if _, delErr := task.Delete(ctx); delErr != nil && !strings.Contains(delErr.Error(), "not found") {
			return fmt.Errorf("delete task: %w", delErr)
		}
	}

	if delErr := container.Delete(ctx, containerd.WithSnapshotCleanup); delErr != nil && !strings.Contains(delErr.Error(), "not found") {
		return fmt.Errorf("delete container: %w", delErr)
	}

	return nil
}


// DestroyWorkflow cleans up any resources for the workflow.
func (b *containerdBackend) DestroyWorkflow(ctx context.Context, conf *backend.Config, taskUUID string) error {
	if b.client != nil {
		return b.client.Close()
	}
	return nil
}
