// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

//go:build windows

package combostore

import (
	"context"
	"errors"
	"time"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/nettools"
)

type Store struct{}

// var _ model.Storer = (*Store)(nil)

var unsupported = errors.New("platform is not supported")

func New(cfg *Config) (*Store, error) {
	return nil, unsupported
}

func (cs *Store) Close() error {
	return unsupported
}

//
// Network data
//

// AddNetwork adds a network to the store, will return error if the network already exists in the store
func (cs *Store) AddNetwork(ctx context.Context, newnetwork model.Network) error {
	return unsupported
}

// RemoveNetworkByName remove the named network from the store
func (cs *Store) RemoveNetworkByName(ctx context.Context, name string) error {
	return unsupported
}

// UpdateNetwork will freshen up the network using the given network
func (cs *Store) UpdateNetwork(ctx context.Context, newnetwork model.Network) error {
	return unsupported
}

// UpsertNetwork will either add the given network and if it already exists then it will run an update
func (cs *Store) UpsertNetwork(ctx context.Context, net model.Network) error {
	return unsupported
}

// GetNetworkByName returns a network with the given name
func (cs *Store) GetNetworkByName(ctx context.Context, name string) (model.Network, error) {
	return model.Network{}, unsupported
}

// GetFilteredNetworks returns the networks which match the given GetFilteredNetworks
func (cs *Store) GetFilteredNetworks(
	ctx context.Context,
	filter model.NetworkFilter,
) []model.Network {
	return nil
}

// ListNetworks returns all stored networks
func (cs *Store) ListNetworks(ctx context.Context) []model.Network {
	return nil
}

// CountNetworks returns the number of networks in the store
func (cs *Store) CountNetworks(ctx context.Context) int {
	return 0
}

//
// Device data
//

// AddDevice adds a device to the store, will return error if the device already exists
func (cs *Store) AddDevice(ctx context.Context, newdevice model.Device) error {
	return unsupported
}

// RemoveDeviceByAddr will remove the device with the given Addr from the store
func (cs *Store) RemoveDeviceByAddr(ctx context.Context, addr model.Addr) error {
	return unsupported
}

// UpdateDevice will fresnen up the device using the given device
func (cs *Store) UpdateDevice(
	ctx context.Context,
	newdevice model.Device,
) (enrich bool, err error) {
	return false, unsupported
}

// GetDeviceByAddr returns the device with the matching Addr
func (cs *Store) GetDeviceByAddr(
	ctx context.Context,
	addr model.Addr,
) (model.Device, error) {
	return model.Device{}, unsupported
}

// GetFilteredDevices returns the devices which match the given GetFilteredDevices
func (cs *Store) GetFilteredDevices(
	ctx context.Context,
	filter model.DeviceFilter,
) []model.Device {
	return nil
}

// ListDevices returns all the stored devices
func (cs *Store) ListDevices(ctx context.Context) []model.Device {
	return nil
}

// CountDevices return the number of devices in the store
func (cs *Store) CountDevices(ctx context.Context) int {
	return 0
}

//
// Timeseries data
//

// WritePoint stores a given point for a device
func (cs *Store) WritePerformancePing(
	ctx context.Context,
	timestamp time.Time,
	device model.Device,
	point nettools.Icmp4EchoResponseStatistics,
) error {
	return unsupported
}

// ReadPoints returns the points from Now() minus the duration for the given type
func (cs *Store) ReadPerformancePings(
	ctx context.Context,
	device model.Device,
	duration time.Duration,
) (points []pinger.Point, err error) {
	return nil, unsupported
}
