// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bus

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func isMemoryBusEqual(a, b *memoryBus) bool {
	if cap(a.outbound) != cap(b.outbound) {
		return false
	}
	if cap(a.historicalEvents) != cap(b.historicalEvents) {
		return false
	}
	if cap(a.historicalErrors) != cap(b.historicalErrors) {
		return false
	}
	if a.maxhistory != b.maxhistory {
		return false
	}
	if a.maxerrors != b.maxerrors {
		return false
	}
	if a.enableddebuglog != b.enableddebuglog {
		return false
	}
	if a.enableerrorlog != b.enableerrorlog {
		return false
	}
	return true
}

func TestBus_New(t *testing.T) {
	type test struct {
		input *Config
		want  *memoryBus
	}

	tests := map[string]test{
		"standard config new": {
			input: &Config{
				MaxEvents:            1,
				MaxErrors:            1,
				InboundSize:          1,
				EnableErrorLog:       true,
				EnableDebugLog:       true,
				MinimumPriorityLevel: 0,
			},
			want: &memoryBus{
				inbound:          make(chan Event, 1),
				outbound:         make([]chan Event, 0),
				historicalEvents: make([]HistoricalEvent, 1),
				historicalErrors: make([]HistoricalError, 1),
				maxhistory:       1,
				maxerrors:        1,
				enableddebuglog:  true,
				enableerrorlog:   true,
			},
		},
		"nil config new": {
			input: nil,
			want: &memoryBus{
				inbound:          make(chan Event, 0),
				outbound:         make([]chan Event, 0),
				historicalEvents: make([]HistoricalEvent, 0),
				historicalErrors: make([]HistoricalError, 0),
				maxhistory:       0,
				maxerrors:        0,
				enableddebuglog:  false,
				enableerrorlog:   false,
			},
		},
	}

	for name, tc := range tests {
		got := New(tc.input)
		diff := cmp.Diff(
			tc.want,
			got,
			cmp.AllowUnexported(),
			cmpopts.IgnoreFields(memoryBus{}, "lock"),
			cmp.Comparer(func(a *memoryBus, b *memoryBus) bool {
				return isMemoryBusEqual(a, b)
			}),
		)
		if diff != "" {
			t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
		}
	}
}
