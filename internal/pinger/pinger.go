// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package pinger

import (
	"context"
	"errors"
	"time"

	"github.com/emicklei/tre"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/nettools"
)

type (
	PerfPingDevicesEvent struct{}

	TsPointBase struct {
		Time  time.Time
		Value float64
	}

	Point struct {
		Device  model.Device
		Start   time.Time
		Minimum time.Duration
		Average time.Duration
		Maximum time.Duration
		Loss    float64
	}

	PerformancePingResponseEvent struct {
		Device   model.Device
		Stats    nettools.Icmp4EchoResponseStatistics
		Start    time.Time
		Duration time.Duration
	}
)

func BuildPingDevice(
	cfg *Config,
) func(context.Context, model.Device) (PerformancePingResponseEvent, error) {
	return func(ctx context.Context, d model.Device) (pre PerformancePingResponseEvent, err error) {
		responses, err := nettools.Icmp4Echo(
			ctx,
			d.Addr.Addr(),
			nettools.I4EWithCount(cfg.PingCount),
			nettools.I4EWithReadTimeout(cfg.Timeout),
			nettools.I4EWithPrivileged(cfg.Privileged),
		)
		if err != nil && !errors.Is(err, nettools.ErrNoResponseFromRemote) {
			return pre, tre.New(err, "icmp4 echo")
		}
		stats := nettools.CalculateIcmp4EchoResponseStatistics(responses)
		d.UpdateFromPingStats(stats, stats.Start)
		pre = PerformancePingResponseEvent{
			Start:    stats.Start,
			Device:   d,
			Duration: stats.TotalElapsed,
			Stats:    stats,
		}

		return pre, nil
	}
}

func PerformancePingerFilter(cfg *Config) model.DeviceFilter {
	return func(d model.Device) bool {
		if d.PerformancePing.LastSeen.IsZero() {
			return true
		}
		since := time.Since(d.PerformancePing.LastSeen)
		if d.IsServer() {
			if since > cfg.ServerInterval {
				return true
			}
			return false
		}
		if since > cfg.DefaultInterval {
			return true
		}
		return false
	}
}
