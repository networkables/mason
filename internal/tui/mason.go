// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tui

import (
	"context"
	"net"
	"time"

	"github.com/networkables/mason/internal/enrichment"
	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/internal/server"
	"github.com/networkables/mason/nettools"
)

type MasonReaderWriter interface {
	MasonReader
	MasonWriter
	MasonNetworker
}

type MasonReader interface {
	ListNetworks(context.Context) []model.Network
	CountNetworks(context.Context) int
	ListDevices(context.Context) []model.Device
	CountDevices(context.Context) int
	GetDeviceByAddr(context.Context, model.Addr) (model.Device, error)
	ReadPerformancePings(
		context.Context,
		model.Device,
		time.Duration,
	) ([]pinger.Point, error)
	// GetConfig() *Config
	GetInternalsSnapshot(ctx context.Context) server.MasonInternalsView
	GetUserAgent() string
	OuiLookup(mac net.HardwareAddr) string
	GetNetworkStats(ctx context.Context) []model.NetworkStats
	PingFailures(ctx context.Context) []model.Device
	ServerDevices(ctx context.Context) []model.Device
	FlowSummaryByIP(context.Context, model.Addr) ([]model.FlowSummaryForAddrByIP, error)
	FlowSummaryByName(context.Context, model.Addr) ([]model.FlowSummaryForAddrByName, error)
	FlowSummaryByCountry(context.Context, model.Addr) ([]model.FlowSummaryForAddrByCountry, error)
	LookupIP(model.Addr) string
}

type MasonWriter interface {
	AddNetwork(context.Context, model.Network) error
	AddNetworkByName(context.Context, string, string, bool) error
}

type MasonNetworker interface {
	StringToAddr(string) (model.Addr, error)
	IcmpPingAddr(
		context.Context,
		model.Addr,
		int,
		time.Duration,
		bool,
	) (nettools.Icmp4EchoResponseStatistics, error)
	IcmpPing(
		context.Context,
		string,
		int,
		time.Duration,
		bool,
	) (nettools.Icmp4EchoResponseStatistics, error)
	ArpPing(context.Context, string, time.Duration) (model.MAC, error)
	Portscan(context.Context, string, *enrichment.PortScanConfig) ([]int, error)
	GetExternalAddr(ctx context.Context) (model.Addr, error)
	Traceroute(context.Context, string) ([]nettools.Icmp4EchoResponseStatistics, error)
	TracerouteAddr(
		context.Context,
		model.Addr,
	) ([]nettools.Icmp4EchoResponseStatistics, error)
	FetchTLSInfo(context.Context, string) (nettools.TLS, error)
	FetchSNMPInfo(context.Context, string) (nettools.SnmpInfo, error)
	FetchSNMPInfoAddr(context.Context, model.Addr) (nettools.SnmpInfo, error)
}
