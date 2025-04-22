package containerd

import (
   "fmt"
   "os"
   "path/filepath"

   "github.com/urfave/cli/v3"
)

// defaultContainerdEndpoint returns a sensible default endpoint for containerd
// honoring user environment for rootless/containerd socket locations.
func defaultContainerdEndpoint() string {
   // respect explicit environment variable
   if env := os.Getenv("CONTAINERD_ENDPOINT"); env != "" {
       return env
   }
   // check for user-level containerd socket in XDG_RUNTIME_DIR
   if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
       path1 := filepath.Join(dir, "containerd", "containerd.sock")
       if fi, err := os.Stat(path1); err == nil && !fi.IsDir() {
           return path1
       }
       path2 := filepath.Join(dir, "containerd.sock")
       if fi, err := os.Stat(path2); err == nil && !fi.IsDir() {
           return path2
       }
   }
   // check for systemd user socket location
   if uid := os.Geteuid(); uid != 0 {
       path3 := fmt.Sprintf("/run/user/%d/containerd/containerd.sock", uid)
       if fi, err := os.Stat(path3); err == nil && !fi.IsDir() {
           return path3
       }
   }
   // fallback to system socket
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
}
