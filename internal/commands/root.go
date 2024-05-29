// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"runtime/debug"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/networkables/mason/internal/config"
	"github.com/networkables/mason/internal/server"
)

var (
	flagDebug bool
	cmdRoot   = &cobra.Command{
		Use:   "mason",
		Short: "cli for mason",
		// Long:  `cli for mason`,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if flagDebug {
				log.SetLevel(log.DebugLevel)
				log.Debug("debug log activated")
			}
			return nil
		},
	}

	cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "print version",
		// Long:  `print version`,
		RunE: func(*cobra.Command, []string) error {
			version := "dev_unknown"
			bi, ok := debug.ReadBuildInfo()
			if ok {
				version = bi.Main.Version
			}
			m := server.NewTool()
			log.Print("mason", "version", version, "useragent", m.GetUserAgent())
			return nil
		},
	}
)

func RootExecute() {
	if err := cmdRoot.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cmdRoot.AddCommand(cmdVersion, cmdServer, cmdTool, cmdSys, cmdDebug)

	cmdRoot.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Activate debug logging")

	cfg := config.GetConfig()
	config.CfgDiscoBuildAndBindFlags(cmdRoot.PersistentFlags(), cfg.Discovery)
	config.CfgStoreBuildAndBindFlags(cmdRoot.PersistentFlags(), cfg.Store)
	config.CfgBusBuildAndBindFlags(cmdRoot.PersistentFlags(), cfg.Bus)
	config.CfgServerBuildAndBindFlags(cmdRoot.PersistentFlags(), cfg.Server)
	config.CfgPerfPingeBuildAndBindFlags(cmdRoot.PersistentFlags(), cfg.PerformancePinger)
	config.CfgEnrichBuildAndBindFlags(cmdRoot.PersistentFlags(), cfg.Enrichment)

	// viper.WriteConfigAs("test.yaml")

	// log.Info(
	// 	"disco enable",
	// 	"val",
	// 	*cfg.Discovery.Enabled,
	// 	"viper",
	// 	viper.GetBool(cfgDiscoEnabledKey),
	// )
	//
	// err := cfg.Save()
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
