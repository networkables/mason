// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"github.com/networkables/mason/internal/commands"
)

var (
	Version    string
	Commit     string
	CommitDate string
	TreeState  string
)

func main() {
	commands.RootExecute()
}
