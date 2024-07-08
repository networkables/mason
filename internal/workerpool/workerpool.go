// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package workerpool

import (
	"context"

	"github.com/charmbracelet/log"
)

type Pool[Inbound, Outbound any] struct {
	Name          string
	in            chan Inbound
	doit          func(context.Context, Inbound) (Outbound, error)
	C             chan Outbound
	E             chan error
	activeworkers chan bool
}

func New[Inbound, Outbound any](
	name string,
	in chan Inbound,
	f func(context.Context, Inbound) (Outbound, error),
) *Pool[Inbound, Outbound] {
	return &Pool[Inbound, Outbound]{
		Name: name,
		in:   in,
		doit: f,
		C:    make(chan Outbound),
		E:    make(chan error),
	}
}

func (wp *Pool[Inbound, Outbound]) Run(ctx context.Context, maxworkers int) {
	if maxworkers == 0 {
		log.Warn(
			"workerpool set to run with zero max workers, changing to 1 to prevent freezes",
			"name",
			wp.Name,
		)
		maxworkers = 1
	}
	wp.activeworkers = make(chan bool, maxworkers)
	keepworking := true
	for keepworking {
		select {
		case <-ctx.Done():
			keepworking = false
		case i, ok := <-wp.in:
			if !ok {
				keepworking = false
				continue
			}
			wp.activeworkers <- true

			go func(ctx context.Context, i Inbound) {
				out, err := wp.doit(ctx, i)
				if err != nil {
					wp.E <- err
				} else {
					wp.C <- out
				}
				<-wp.activeworkers
			}(ctx, i)

		}
	}
	for ; maxworkers > 0; maxworkers-- {
		wp.activeworkers <- true
	}
	close(wp.C)
	close(wp.E)
}

func (wp *Pool[Inbound, Outbound]) Active() int {
	return len(wp.activeworkers)
}
