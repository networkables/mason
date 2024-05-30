// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
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
	"github.com/networkables/mason/internal/config"
	"github.com/networkables/mason/internal/server"
	"github.com/networkables/mason/internal/store"
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

	cfg := config.GetConfig()

	masonServer, err := startMason(ctx, cfg)
	if err != nil {
		return err
	}

	sshServer, err := startSSHServer(cfg.Server, masonServer)
	if err != nil {
		log.Error("ssh server start", "error", err)
		normalcancel()
		return err
	}

	httpServer := wui.New(masonServer, cfg.Server)
	go httpServer.Start()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
	log.Debug("done")
	normalcancel()

	// Shutdown sequence starting
	shutdownctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
	}()

	// Shutdown SSH
	if err := sshServer.Shutdown(shutdownctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("ssh shutdown", "error", err)
	}
	// Shutdown HTTP
	if err := httpServer.Shutdown(shutdownctx); err != nil {
		log.Error("http shutdown", "error", err)
	}

	return nil
}

func startMason(ctx context.Context, cfg *config.Config) (*server.Mason, error) {
	if !cfg.Server.IgnoreCap && !server.HasCapabilities() {
		return nil, errors.New("not all capabilities are present, run sudo ./mason sys setcap")
	}

	store, err := store.NewCombo(cfg.Store)
	if err != nil {
		return nil, err
	}

	m := server.New(
		cfg,
		bus.New(cfg.Bus),
		store,
	)
	go m.Run(ctx)
	return m, nil
}

func startSSHServer(cfg *config.Server, masonServer *server.Mason) (*ssh.Server, error) {
	host := ""
	tuiport := *cfg.TUIPort
	tui := tui.New(masonServer)

	sshServer, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, tuiport)),
		wish.WithHostKeyPath("mason_ssh_ed25519"),
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
		log.Info("sshserver startup", "addr", sshServer.Addr)
		if err = sshServer.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("could not start server", "error", err)
		}
	}()

	return sshServer, nil
}
