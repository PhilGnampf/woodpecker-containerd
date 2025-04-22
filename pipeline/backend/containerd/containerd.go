package containerd

import (
   "context"
   "fmt"
   "io"

   "github.com/containerd/containerd"
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
   return Flags
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
   // No-op: workflows run in specified namespace
   return nil
}

// StartStep creates and starts a containerd task for the step.
func (b *containerdBackend) StartStep(ctx context.Context, step *backend.Step, taskUUID string) error {
   return fmt.Errorf("StartStep not implemented for containerd backend")
}

// WaitStep waits for the step's task to finish.
func (b *containerdBackend) WaitStep(ctx context.Context, step *backend.Step, taskUUID string) (*backend.State, error) {
   return nil, fmt.Errorf("WaitStep not implemented for containerd backend")
}

// TailStep streams logs for the given step.
func (b *containerdBackend) TailStep(ctx context.Context, step *backend.Step, taskUUID string) (io.ReadCloser, error) {
   return nil, fmt.Errorf("TailStep not implemented for containerd backend")
}

// DestroyStep stops and removes the step's task.
func (b *containerdBackend) DestroyStep(ctx context.Context, step *backend.Step, taskUUID string) error {
   return fmt.Errorf("DestroyStep not implemented for containerd backend")
}

// DestroyWorkflow cleans up any resources for the workflow.
func (b *containerdBackend) DestroyWorkflow(ctx context.Context, conf *backend.Config, taskUUID string) error {
   if b.client != nil {
       return b.client.Close()
   }
   return nil
}