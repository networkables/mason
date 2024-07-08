// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"errors"
	"runtime/debug"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/networkables/mason/internal/asn"
	"github.com/networkables/mason/internal/bus"
	"github.com/networkables/mason/internal/combostore"
	"github.com/networkables/mason/internal/discovery"
	"github.com/networkables/mason/internal/enrichment"
	"github.com/networkables/mason/internal/netflows"
	"github.com/networkables/mason/internal/oui"
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/internal/server"
	"github.com/networkables/mason/internal/sqlitestore"
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
			m := server.New()
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

	// TODO: set all flags is probably only useful for server commands, should get rid of this and only apply flags needed at the command level (or sub command)
	setAllFlags(cmdRoot.PersistentFlags(), server.GetConfig())

	// viper.WriteConfigAs("config.yaml")

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

func setAllFlags(f *pflag.FlagSet, c *server.Config) {
	// Sane defaults
	server.SetFlags(f, c)
	discovery.SetFlags(f, c.Discovery)
	combostore.SetFlags(f, c.Store.Combo)
	sqlitestore.SetFlags(f, c.Store.Sqlite)
	bus.SetFlags(f, c.Bus)
	pinger.SetFlags(f, c.Pinger)
	enrichment.SetFlags(f, c.Enrichment)
	netflows.SetFlags(f, c.NetFlows)
	asn.SetFlags(f, c.Asn)
	oui.SetFlags(f, c.Oui)

	// Env
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("MASON")
	viper.AutomaticEnv()

	// Config file
	viper.AddConfigPath(".")
	viper.AddConfigPath(c.ConfigDirectory)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil && !errors.Is(err, viper.ConfigFileNotFoundError{}) {
		log.Warn("did not find a config file to read")
	}

	err = viper.Unmarshal(c)
	if err != nil {
		log.Fatal(err)
	}
}
