// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package pinger

import (
	"context"
	"errors"
	"time"

	"github.com/networkables/mason/internal/config"
	"github.com/networkables/mason/internal/device"
	"github.com/networkables/mason/internal/stackerr"
	"github.com/networkables/mason/nettools"
)

type (
	PerfPingDevicesEvent struct{}

	TsPointBase struct {
		Time  time.Time
		Value float64
	}

	TimeseriesPoint interface {
		GetName() string
		GetValue() float64
		GetTime() time.Time
		Hydrate(int, float64) TimeseriesPoint
	}

	PerformancePingResponseEvent struct {
		Device   device.Device
		Stats    nettools.Icmp4EchoResponseStatistics
		Start    time.Time
		Duration time.Duration
	}

	TimeseriesMaxResponse TsPointBase
	TimeseriesAvgResponse TsPointBase
	TimeseriesPacketLoss  TsPointBase
)

func (tmr TimeseriesMaxResponse) GetName() string    { return "pingmax" }
func (tmr TimeseriesMaxResponse) GetTime() time.Time { return tmr.Time }
func (tmr TimeseriesMaxResponse) GetValue() float64  { return tmr.Value }
func (tmr TimeseriesMaxResponse) Hydrate(t int, v float64) TimeseriesPoint {
	tmr.Time = time.Unix(int64(t), 0)
	tmr.Value = float64(v) / float64(time.Millisecond)
	return tmr
}

func (tar TimeseriesAvgResponse) GetName() string    { return "pingavg" }
func (tar TimeseriesAvgResponse) GetTime() time.Time { return tar.Time }
func (tar TimeseriesAvgResponse) GetValue() float64  { return tar.Value }
func (tar TimeseriesAvgResponse) Hydrate(t int, v float64) TimeseriesPoint {
	tar.Time = time.Unix(int64(t), 0)
	tar.Value = float64(v) / float64(time.Millisecond)
	return tar
}

func (tpl TimeseriesPacketLoss) GetName() string    { return "pingloss" }
func (tpl TimeseriesPacketLoss) GetTime() time.Time { return tpl.Time }
func (tpl TimeseriesPacketLoss) GetValue() float64  { return tpl.Value }
func (tpl TimeseriesPacketLoss) Hydrate(t int, v float64) TimeseriesPoint {
	tpl.Time = time.Unix(int64(t), 0)
	tpl.Value = v * 100.0
	return tpl
}

var (
	_ TimeseriesPoint = (*TimeseriesMaxResponse)(nil)
	_ TimeseriesPoint = (*TimeseriesAvgResponse)(nil)
	_ TimeseriesPoint = (*TimeseriesPacketLoss)(nil)
)

func (pre PerformancePingResponseEvent) Points() []TimeseriesPoint {
	return []TimeseriesPoint{
		TimeseriesMaxResponse{pre.Start, float64(pre.Stats.Maximum.Nanoseconds())},
		TimeseriesAvgResponse{pre.Start, float64(pre.Stats.Mean.Nanoseconds())},
		TimeseriesPacketLoss{pre.Start, pre.Stats.PacketLoss},
	}
}

func BuildPingDevice(
	ctx context.Context,
	cfg *config.Pinger,
) func(device.Device) (PerformancePingResponseEvent, error) {
	return func(device device.Device) (pre PerformancePingResponseEvent, err error) {
		responses, err := nettools.Icmp4Echo(
			ctx,
			device.Addr,
			nettools.I4EWithCount(*cfg.PingCount),
			nettools.I4EWithReadTimeout(*cfg.Timeout),
		)
		if err != nil && !errors.Is(err, nettools.ErrNoResponseFromRemote) {
			return pre, stackerr.New(err)
		}
		stats := nettools.CalculateIcmp4EchoResponseStatistics(responses)
		device.UpdateFromPingStats(stats, stats.Start)
		pre = PerformancePingResponseEvent{
			Start:    stats.Start,
			Device:   device,
			Duration: stats.TotalElapsed,
			Stats:    stats,
		}

		return pre, nil
	}
}

func PerformancePingerFilter(cfg *config.Pinger) device.DeviceFilter {
	return func(d device.Device) bool {
		if d.PerformancePing.LastSeen.IsZero() {
			return true
		}
		since := time.Since(d.PerformancePing.LastSeen)
		if d.IsServer() {
			if since > *cfg.ServerInterval {
				return true
			}
			return false
		}
		if since > *cfg.DefaultInterval {
			return true
		}
		return false
	}
}
