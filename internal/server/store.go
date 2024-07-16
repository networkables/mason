// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"time"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/nettools"
)

type (

	// Storeer is a complete implementation of all objects which need to be managed
	Storer interface {
		NetworkStorer
		DeviceStorer
		PerformancePingStorer
		Close() error
	}

	// NetworkStorer allows for the saving and fetching of network definitions.
	NetworkStorer interface {
		AddNetwork(context.Context, model.Network) error
		RemoveNetworkByName(context.Context, string) error
		UpdateNetwork(context.Context, model.Network) error
		GetNetworkByName(context.Context, string) (model.Network, error)
		GetFilteredNetworks(context.Context, model.NetworkFilter) []model.Network
		ListNetworks(context.Context) []model.Network
		CountNetworks(context.Context) int
	}

	// DeviceStorer allows for the saving and fetching of device definitions.
	DeviceStorer interface {
		AddDevice(context.Context, model.Device) error
		RemoveDeviceByAddr(context.Context, model.Addr) error
		UpdateDevice(context.Context, model.Device) (bool, error)
		GetDeviceByAddr(context.Context, model.Addr) (model.Device, error)
		GetFilteredDevices(context.Context, model.DeviceFilter) []model.Device
		ListDevices(context.Context) []model.Device
		CountDevices(context.Context) int
	}

	// PerformancePingStorer allows for the saving and fetching of timeseries data.
	PerformancePingStorer interface {
		WritePerformancePing(
			context.Context,
			time.Time,
			model.Device,
			nettools.Icmp4EchoResponseStatistics,
		) error
		ReadPerformancePings(
			context.Context,
			model.Device,
			time.Duration,
		) ([]pinger.Point, error)
	}

	NetflowStorer interface {
		AsnStorer
		AddNetflows(context.Context, []model.IpFlow) error
		GetNetflows(context.Context, model.Addr) ([]model.IpFlow, error)
		FlowSummaryByIP(context.Context, model.Addr) ([]model.FlowSummaryForAddrByIP, error)
		FlowSummaryByName(context.Context, model.Addr) ([]model.FlowSummaryForAddrByName, error)
		FlowSummaryByCountry(
			context.Context,
			model.Addr,
		) ([]model.FlowSummaryForAddrByCountry, error)
	}

	AsnStorer interface {
		StartAsnLoad() func(*error)
		UpsertAsn(context.Context, model.Asn) error
		GetAsn(context.Context, string) (model.Asn, error)
	}
)
