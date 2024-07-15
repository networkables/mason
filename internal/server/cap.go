// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

func appLocation() (path string, isgorun bool) {
	var err error
	path, err = os.Executable()
	if err != nil {
		log.Fatal("appLocation", "error", err)
	}
	isgorun = strings.Contains(path, "go-build")
	return path, isgorun
}
