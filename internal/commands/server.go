// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/spf13/cobra"

	"github.com/networkables/mason/internal/bus"
	"github.com/networkables/mason/internal/combostore"
	"github.com/networkables/mason/internal/server"
	"github.com/networkables/mason/internal/sqlitestore"
	"github.com/networkables/mason/internal/tui"
	"github.com/networkables/mason/internal/wui"
)

var cmdServer = &cobra.Command{
	Use:   "server",
	Short: "start server",
	// Long:  `start server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCmdServer()
	},
}

func runCmdServer() error {
	ctx, normalcancel := context.WithCancel(context.Background())
	defer normalcancel()

	cfg := server.GetConfig()

	masonServer, err := startMason(ctx, cfg)
	if err != nil {
		return err
	}

	sshServer, err := startSSHServer(cfg.Tui.ListenAddress, cfg.Tui.KeyDirectory, masonServer)
	if err != nil {
		log.Error("ssh server start", "error", err)
		normalcancel()
		return err
	}

	httpServer := wui.New(masonServer, cfg.Wui.ListenAddress)
	go httpServer.Start()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
	log.Info("caught interrupt signal, starting normalcancel")
	normalcancel()
	// time.Sleep(5 * time.Second)

	// Shutdown sequence starting
	shutdownctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
	}()

	// Shutdown SSH
	if err := sshServer.Shutdown(shutdownctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("ssh shutdown", "error", err)
	}
	log.Info("ssh shutdown")
	// Shutdown HTTP
	if err := httpServer.Shutdown(shutdownctx); err != nil {
		log.Error("http shutdown", "error", err)
	}
	log.Info("http shutdown")

	return nil
}

func startMason(ctx context.Context, cfg *server.Config) (*server.Mason, error) {
	// if !cfg.IgnoreCap && !server.HasCapabilities(cfg) {
	// 	return nil, errors.New("not all capabilities are present, run sudo ./mason sys setcap")
	// }
	if !server.HasCapabilities(cfg) {
		return nil, errors.New("not all capabilities are present, run sudo ./mason sys setcap")
	}

	var (
		store     server.Storer
		flowstore server.NetflowStorer
		err       error
	)

	if cfg.Store.Combo.Enabled {
		store, err = combostore.New(cfg.Store.Combo)
		if err != nil {
			return nil, err
		}
	} else if cfg.Store.Sqlite.Enabled {
		sqls, err := sqlitestore.New(cfg.Store.Sqlite)
		if err != nil {
			return nil, err
		}
		err = sqls.MigrateUp(ctx)
		if err != nil {
			return nil, err
		}
		store = sqls
		flowstore = sqls
	}

	m := server.New(
		server.WithConfig(cfg),
		server.WithBus(bus.New(cfg.Bus)),
		server.WithStore(store),
		server.WithNetflowStorer(flowstore),
	)
	go m.Run(ctx)
	return m, nil
}

func startSSHServer(
	listenaddress string,
	keydir string,
	masonServer *server.Mason,
) (*ssh.Server, error) {
	tui := tui.New(masonServer)

	keypath := "mason_ssh_ed25519"
	if keydir != "" {
		keypath = filepath.Join(keydir, keypath)
	}

	sshServer, err := wish.NewServer(
		wish.WithAddress(listenaddress),
		wish.WithHostKeyPath(keypath),
		wish.WithMiddleware(
			bubbletea.Middleware(tui.TeaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		return nil, err
	}

	go func() {
		log.Info("starting ssh server", "addr", sshServer.Addr)
		if err = sshServer.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("could not start ssh server", "error", err)
		}
	}()

	return sshServer, nil
}
