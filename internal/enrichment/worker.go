// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package enrichment

import (
	"context"

	"github.com/charmbracelet/log"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/workerpool"
)

type Worker struct {
	In chan EnrichDeviceRequest
	*workerpool.Pool[EnrichDeviceRequest, model.Device]
}

func NewWorker() *Worker {
	input := make(chan EnrichDeviceRequest)
	return &Worker{
		In:   input,
		Pool: workerpool.New("enrichment", input, EnrichDevice),
	}
}

func (w *Worker) Run(ctx context.Context, max int) {
	w.Pool.Run(ctx, max)
}

func (w *Worker) Close() {
	log.Info("enrichment workerpool shutdown")
	close(w.In)
}
