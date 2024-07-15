// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

//go:build netbsd

package server

import (
	"errors"
)

func HasCapabilities(cfg *Config) bool {
	return true
}

func SetCapabilities() error {
	return errors.New("cannot set capabilities on this platform")
}
