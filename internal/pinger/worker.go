// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package pinger

import (
	"context"

	"github.com/charmbracelet/log"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/workerpool"
)

type Worker struct {
	In chan model.Device
	*workerpool.Pool[model.Device, PerformancePingResponseEvent]
}

func NewWorker(cfg *Config) *Worker {
	input := make(chan model.Device)
	return &Worker{
		In:   input,
		Pool: workerpool.New("pinger", input, BuildPingDevice(cfg)),
	}
}

func (w *Worker) Run(ctx context.Context, max int) {
	w.Pool.Run(ctx, max)
}

func (w *Worker) Close() {
	log.Info("pinger workerpool shutdown")
	close(w.In)
}
