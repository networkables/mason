// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/networkables/mason/internal/server"
)

var (
	cmdSys = &cobra.Command{
		Use:   "sys",
		Short: "mason system commands",
		// Long:  `start server`,
	}

	cmdSysHasCap = &cobra.Command{
		Use:   "cap",
		Short: "check if the mason binary has been assigned the required capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdSysHasCap(args)
		},
	}

	cmdSysSetCap = &cobra.Command{
		Use:   "setcap",
		Short: "set the required capabilities on the mason binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdSysSetCap(args)
		},
	}
)

func init() {
	cmdSys.AddCommand(cmdSysHasCap)
	cmdSys.AddCommand(cmdSysSetCap)
}

func runCmdSysHasCap([]string) error {
	good := server.HasCapabilities()
	if good {
		log.Info("mason has the required capabilities")
		return nil
	}
	log.Error("not all capabilities are present, run sudo ./mason sys setcap")
	return nil
}

func runCmdSysSetCap([]string) error {
	err := server.SetCapabilities()
	if err != nil {
		log.Error("setCapabilities", "error", err)
	}
	return nil
}
