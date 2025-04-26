package containerd

import (
   "os"
   "github.com/urfave/cli/v3"
)

// defaultContainerdEndpoint returns a sensible default endpoint for containerd
// honoring user environment for rootless/containerd socket locations.
func defaultContainerdEndpoint() string {
   // Respect explicit environment variable
   if env := os.Getenv("CONTAINERD_ENDPOINT"); env != "" {
       return env
   }
   // Default to the system containerd socket
   return "/run/containerd/containerd.sock"
}

// Flags defines the CLI flags for configuring the containerd backend.
var Flags = []cli.Flag{
   &cli.StringFlag{
       Sources: cli.EnvVars("CONTAINERD_ENDPOINT"),
       Name:    "backend-containerd-endpoint",
       Usage:   "containerd endpoint address",
       Value:   defaultContainerdEndpoint(),
   },
   &cli.StringFlag{
       Sources: cli.EnvVars("CONTAINERD_NAMESPACE"),
       Name:    "backend-containerd-namespace",
       Usage:   "containerd namespace to use",
       Value:   "default",
   },
   &cli.StringFlag{
       Sources: cli.EnvVars("CONTAINERD_SNAPSHOTTER"),
       Name:    "backend-containerd-snapshotter",
       Usage:   "containerd snapshotter to use (e.g. overlayfs,fuse-overlayfs)",
       Value:   "overlayfs",
   },
}
