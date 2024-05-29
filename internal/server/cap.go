// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"syscall"

	"github.com/charmbracelet/log"
	"kernel.org/pub/linux/libs/security/libcap/cap"
)

func appLocation() string {
	loc, err := os.Executable()
	if err != nil {
		log.Error("appLocation", "error", err)
	}
	return loc
}

func HasCapabilities() bool {
	loc := appLocation()
	c, err := cap.GetFile(loc)
	if err != nil {
		if err == syscall.ENODATA {
			return false
		}
		return false
	}
	hasRawNet, err := c.GetFlag(cap.Effective&cap.Permitted, cap.NET_RAW)
	if err != nil {
		log.Error("hasCapabilities", "error", err)
		return false
	}
	return hasRawNet
}

func SetCapabilities() error {
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not check uid of app: %v", err)
	}
	if u.Uid != "0" {
		return errors.New("setcap must be run using sudo")
	}

	todo := cap.NewSet()
	err = todo.SetFlag(cap.Effective, true, cap.NET_RAW)
	if err != nil {
		return err
	}
	err = todo.SetFlag(cap.Permitted, true, cap.NET_RAW)
	if err != nil {
		return err
	}
	loc := appLocation()
	err = todo.SetFile(loc)
	if err != nil {
		log.Error("setCapabilities", "file", loc, "error", err)
	}
	return err
}
