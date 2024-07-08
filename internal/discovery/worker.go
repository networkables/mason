// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package discovery

import (
	"context"

	"github.com/charmbracelet/log"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/workerpool"
)

type Worker struct {
	In chan model.Addr
	*workerpool.Pool[model.Addr, model.EventDeviceDiscovered]
}

func NewWorker(cfg *Config) *Worker {
	input := make(chan model.Addr)
	return &Worker{
		In:   input,
		Pool: workerpool.New("discovery", input, BuildAddrScannerFunc(BuildAddrScanners(cfg))),
	}
}

func (w *Worker) Run(ctx context.Context, max int) {
	w.Pool.Run(ctx, max)
}

func (w *Worker) Close() {
	log.Info("discovery workerpool shutdown")
	close(w.In)
}

type NetworkScannerWorker struct {
	In chan model.Network
	*workerpool.Pool[model.Network, string]
}

func NewNetworkScannerWorker(status *string, devin chan model.Addr) *NetworkScannerWorker {
	input := make(chan model.Network)
	return &NetworkScannerWorker{
		In:   input,
		Pool: workerpool.New("networkscan", input, BuildNetworkScanFunc(devin, status)),
	}
}

func (w *NetworkScannerWorker) Run(ctx context.Context) {
	w.Pool.Run(ctx, 1)
}

func (w *NetworkScannerWorker) Close() {
	log.Info("networkscanner workerpool shutdown")
	close(w.In)
}
