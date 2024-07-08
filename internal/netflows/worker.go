// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package netflows

import (
	"context"

	"github.com/charmbracelet/log"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/workerpool"
)

type Worker struct {
	In chan []byte
	*workerpool.Pool[[]byte, []model.IpFlow]
}

func NewWorker(cfg *Config, input chan []byte) *Worker {
	return &Worker{
		In:   input,
		Pool: workerpool.New("netflows", input, parser),
	}
}

func (w *Worker) Run(ctx context.Context, max int) {
	w.Pool.Run(ctx, max)
}

func (w *Worker) Close() {
	log.Info("netflows workerpool shutdown")
	close(w.In)
}
