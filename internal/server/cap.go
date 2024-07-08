// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"
	"syscall"

	"github.com/charmbracelet/log"
	"kernel.org/pub/linux/libs/security/libcap/cap"
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

func HasCapabilities(cfg *Config) bool {
	// do not test if no privileges are required
	if !cfg.Discovery.Icmp.Privileged && !cfg.Discovery.Arp.Enabled &&
		!cfg.Pinger.Privileged {
		return true
	}

	// TODO: below does not work in a container
	return true

	loc, isgorun := appLocation()
	c, err := cap.GetFile(loc)
	if err != nil {
		if isgorun {
			log.Warn("run via 'go run' is not supported when escalated privileges are required")
		}
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
	// log.Printf(" Priv: %t  hasRawNet: %t", *cfg.Discovery.IcmpPing.Privileged, hasRawNet)
	if (cfg.Discovery.Icmp.Privileged || cfg.Discovery.Arp.Enabled || cfg.Pinger.Privileged) &&
		!hasRawNet {
		return false
	}
	return true
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
	loc, isgorun := appLocation()
	if isgorun {
		return errors.New("cannot set capabilities when run with 'go run'")
	}
	err = todo.SetFile(loc)
	if err != nil {
		log.Error("setCapabilities", "file", loc, "error", err)
	}
	return err
}
