// Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"

	"github.com/rs/zerolog/log"

   "go.woodpecker-ci.org/woodpecker/v3/cmd/agent/core"
   "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/containerd"
   "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/docker"
   "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/kubernetes"
   "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/local"
   backendTypes "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

var backends = []backendTypes.Backend{
   kubernetes.New(),
   docker.New(),
   local.New(),
   containerd.New(),
}

func main() {
	ctx := utils.WithContextSigtermCallback(context.Background(), func() {
		log.Info().Msg("termination signal is received, shutting down agent")
	})
	core.RunAgent(ctx, backends)
}
