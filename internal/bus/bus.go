// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bus

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"

	"github.com/networkables/mason/internal/config"
	"github.com/networkables/mason/internal/stackerr"
)

type (
	Event interface{}

	HistoricalEvent struct {
		E  Event
		Ts time.Time
	}

	HistoricalError struct {
		E  error
		Ts time.Time
	}
)

var _ Bus = (*memoryBus)(nil)

type Bus interface {
	AddListener() chan Event
	Publish(Event)
	Run(context.Context)
	History() []HistoricalEvent
	Errors() []HistoricalError
}

type memoryBus struct {
	inbound          chan Event
	outbound         []chan Event
	lock             sync.Mutex
	historicalEvents []HistoricalEvent
	historicalErrors []HistoricalError
	maxhistory       int
	maxerrors        int
	enableddebuglog  bool
	enableerrorlog   bool
}

func New(cfg *config.Bus) *memoryBus {
	return &memoryBus{
		inbound:          make(chan Event, *cfg.InboundSize),
		outbound:         make([]chan Event, 0),
		historicalEvents: make([]HistoricalEvent, 0, *cfg.MaxEvents),
		historicalErrors: make([]HistoricalError, 0, *cfg.MaxErrors),
		maxhistory:       *cfg.MaxEvents,
		maxerrors:        *cfg.MaxErrors,
		enableddebuglog:  *cfg.EnableDebugLog,
		enableerrorlog:   *cfg.EnableErrorLog,
	}
}

func (b *memoryBus) AddListener() chan Event {
	ch := make(chan Event)
	b.lock.Lock()
	defer b.lock.Unlock()
	b.outbound = append(b.outbound, ch)
	return ch
}

func (b *memoryBus) Publish(e Event) {
	if b.enableddebuglog {
		log.Debugf("buseevent %T : %s", e, e)
	}
	select {
	case b.inbound <- e:
	default:
		log.Warnf("dropping bus events; %T %s", e, e)
	}
}

func (b *memoryBus) recordEvent(e Event) {
	ts := time.Now()

	var se stackerr.StackErr
	err, ok := e.(error)
	if ok {
		se, ok = e.(stackerr.StackErr)
		if !ok {
			se = stackerr.New(err)
			log.Warn("non stackerr error", "error", err, "stack", "stack", se.Stack)
			// err = se
		}
		if b.enableerrorlog {
			log.Error(err)
		}
		if len(b.historicalErrors) > b.maxhistory {
			b.historicalErrors = b.historicalErrors[1:]
		}
		b.historicalErrors = append(b.historicalErrors, HistoricalError{E: err, Ts: ts})
	} else {
		// Only events are sent on the bus
		if len(b.historicalEvents) > b.maxhistory {
			// b.history = slices.Delete(b.history, 0, len(b.history)-b.maxhistory)
			b.historicalEvents = b.historicalEvents[1:]
		}
		b.historicalEvents = append(b.historicalEvents, HistoricalEvent{E: e, Ts: ts})
	}
}

func (b *memoryBus) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for _, ch := range b.outbound {
				close(ch)
			}
			return
		case e := <-b.inbound:
			b.recordEvent(e)
			b.sendEvent(e)
			// log.Debugf("sent event %T", e)
		}
	}
}

func (b *memoryBus) sendEvent(e Event) {
	// TODO: Need a watchdog on this incase a recieve blocks up
	//   - Or maybe a way to filter what is sent?
	// log.Debug("\tbus sendout start")
	for _, outch := range b.outbound {
		outch <- e
	}
	// log.Debug("\tbus sendout end")
}

func (b *memoryBus) History() []HistoricalEvent {
	return slices.Clone(b.historicalEvents)
}

func (b *memoryBus) Errors() []HistoricalError {
	return slices.Clone(b.historicalErrors)
}

func NewLogSink(ch chan Event) {
	for e := range ch {
		log.Info("logsink %T: %s", e, e)
	}
}

func (he HistoricalEvent) FmtTime() string {
	return he.Ts.Format(time.TimeOnly)
}

func (he HistoricalEvent) Type() string {
	return strings.Replace(fmt.Sprintf("%T", he.E), "mason.", "", 1)
}

func (he HistoricalError) FmtTime() string {
	return he.Ts.Format(time.TimeOnly)
}

func (he HistoricalError) Type() string {
	return fmt.Sprintf("%T", he.E)
}
