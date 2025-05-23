// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package user

import (
	"context"
	"os"
	"text/template"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

var userListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list all users",
	ArgsUsage: " ",
	Action:    userList,
	Flags:     []cli.Flag{common.FormatFlag(tmplUserList, false)},
}

func userList(ctx context.Context, c *cli.Command) error {
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	opt := woodpecker.UserListOptions{}

	users, err := client.UserList(opt)
	if err != nil || len(users) == 0 {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}
	for _, user := range users {
		if err := tmpl.Execute(os.Stdout, user); err != nil {
			return err
		}
	}
	return nil
}

// Template for user list items.
var tmplUserList = `{{ .Login }}`
