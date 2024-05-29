// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"net/netip"
	"time"

	"github.com/networkables/mason/internal/device"
	"github.com/networkables/mason/internal/network"
	"github.com/networkables/mason/internal/pinger"
)

type (

	// StoreAll is a complete implementation of all objects which need to be managed
	StoreAll interface {
		StoreNetwork
		StoreDevice
		StoreTimeseries
	}

	// StoreNetwork allows for the saving and fetching of network definitions.
	StoreNetwork interface {
		AddNetwork(n network.Network) error
		RemoveNetworkByName(name string) error
		UpdateNetwork(network.Network) error
		UpsertNetwork(network.Network) error
		GetNetworkByName(name string) (network.Network, error)
		GetFilteredNetworks(network.NetworkFilter) []network.Network
		ListNetworks() []network.Network
		CountNetworks() int
	}

	// StoreDevice allows for the saving and fetching of device definitions.
	StoreDevice interface {
		AddDevice(device.Device) error
		RemoveDeviceByAddr(netip.Addr) error
		UpdateDevice(device.Device) error
		GetDeviceByAddr(netip.Addr) (device.Device, error)
		GetFilteredDevices(device.DeviceFilter) []device.Device
		ListDevices() []device.Device
		CountDevices() int
	}

	// StoreTimeseries allows for the saving and fetching of timeseries data.
	StoreTimeseries interface {
		WriteTimeseriesPoint(
			context.Context,
			time.Time,
			device.Device,
			...pinger.TimeseriesPoint,
		) error
		ReadTimeseriesPoints(
			context.Context,
			device.Device,
			time.Duration,
			pinger.TimeseriesPoint,
		) ([]pinger.TimeseriesPoint, error)
	}
)
