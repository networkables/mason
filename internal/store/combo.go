// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
package store

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/netip"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	whisper "github.com/go-graphite/go-whisper"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/networkables/mason/internal/config"
	"github.com/networkables/mason/internal/device"
	"github.com/networkables/mason/internal/network"
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/internal/server"
)

var _ server.StoreAll = (*comboStore)(nil)

type comboStore struct {
	directory       string
	retentions      whisper.Retentions
	networkfilename string
	devicefilename  string
	networks        []network.Network
	devices         []device.Device
}

func NewCombo(cfg *config.Store) (*comboStore, error) {
	cs := &comboStore{
		directory:       *cfg.Directory,
		retentions:      whisper.MustParseRetentionDefs(*cfg.WSPRetention),
		networkfilename: "networks.mb",
		devicefilename:  "devices.mb",
	}

	cs.ensureDirectory(*cfg.Directory)
	err := cs.readNetworks()
	if err != nil {
		return nil, err
	}
	err = cs.readDevices()
	if err != nil {
		return nil, err
	}

	return cs, nil
}

//
// Network data
//

// AddNetwork adds a network to the store, will return error if the network already exists in the store
func (cs *comboStore) AddNetwork(newnetwork network.Network) error {
	for _, net := range cs.networks {
		if net.Prefix.Contains(newnetwork.Prefix.Addr()) {
			return network.ErrNetworkExists
		}
	}
	cs.networks = append(cs.networks, newnetwork)
	return cs.saveNetworks()
}

// RemoveNetworkByName remove the named network from the store
func (cs *comboStore) RemoveNetworkByName(name string) error {
	for idx, network := range cs.networks {
		if network.Name == name {
			cs.networks = slices.Delete(cs.networks, idx, idx+1)
			return cs.saveNetworks()
		}
	}
	return network.ErrNetworkDoesNotExist
}

// UpdateNetwork will freshen up the network using the given network
func (cs *comboStore) UpdateNetwork(newnetwork network.Network) error {
	idx := 0
	found := true
	for i, network := range cs.networks {
		if network.Prefix.String() == newnetwork.Prefix.String() {
			idx = i
			found = true
		}
	}
	if found {
		cs.networks[idx] = newnetwork
		return cs.saveNetworks()
	}
	return network.ErrNetworkExists
}

// UpsertNetwork will either add the given network and if it already exists then it will run an update
func (cs *comboStore) UpsertNetwork(net network.Network) error {
	err := cs.UpdateNetwork(net)
	if err == nil {
		return nil
	}
	if errors.Is(err, network.ErrNetworkDoesNotExist) {
		return cs.AddNetwork(net)
	}
	return err
}

// GetNetworkByName returns a network with the given name
func (cs *comboStore) GetNetworkByName(name string) (network.Network, error) {
	for _, network := range cs.networks {
		if network.Name == name {
			return network, nil
		}
	}
	return network.Network{}, network.ErrNetworkDoesNotExist
}

// GetFilteredNetworks returns the networks which match the given GetFilteredNetworks
func (cs *comboStore) GetFilteredNetworks(filter network.NetworkFilter) []network.Network {
	networks := make([]network.Network, 0)
	for _, network := range cs.networks {
		if filter(network) {
			networks = append(networks, network)
		}
	}
	return networks
}

// ListNetworks returns all stored networks
func (cs *comboStore) ListNetworks() []network.Network {
	return slices.Clone(cs.networks)
}

// CountNetworks returns the number of networks in the store
func (cs *comboStore) CountNetworks() int {
	return len(cs.networks)
}

func (cs *comboStore) saveNetworks() error {
	bytes, err := msgpack.Marshal(cs.networks)
	if err != nil {
		return err
	}
	return os.WriteFile(cs.directory+"/"+cs.networkfilename, bytes, 0644)
}

func (cs *comboStore) readNetworks() error {
	bytes, err := os.ReadFile(cs.directory + "/" + cs.networkfilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	err = msgpack.Unmarshal(bytes, &cs.networks)
	return err
}

//
// Device data
//

// AddDevice adds a device to the store, will return error if the device already exists
func (cs *comboStore) AddDevice(newdevice device.Device) error {
	for _, d := range cs.devices {
		if d.Addr.Compare(newdevice.Addr) == 0 {
			return device.ErrDeviceExists
		}
	}
	cs.devices = append(cs.devices, newdevice)
	return cs.saveDevices()
}

// RemoveDeviceByAddr will remove the device with the given Addr from the store
func (cs *comboStore) RemoveDeviceByAddr(addr netip.Addr) error {
	for idx, device := range cs.devices {
		if device.Addr.Compare(addr) == 0 {
			cs.devices = slices.Delete(cs.devices, idx, idx+1)
			return cs.saveDevices()
		}
	}
	return device.ErrDeviceDoesNotExist
}

// UpdateDevice will fresnen up the device using the given device
func (cs *comboStore) UpdateDevice(newdevice device.Device) error {
	if !newdevice.IsUpdated() {
		return nil
	}
	for idx, device := range cs.devices {
		if device.Addr.Compare(newdevice.Addr) == 0 {
			cs.devices[idx] = cs.devices[idx].Merge(newdevice)
			return cs.saveDevices()
		}
	}
	return device.ErrDeviceDoesNotExist
}

// GetDeviceByAddr returns the device with the matching Addr
func (cs *comboStore) GetDeviceByAddr(addr netip.Addr) (device.Device, error) {
	for _, device := range cs.devices {
		if device.Addr.Compare(addr) == 0 {
			return device, nil
		}
	}
	return device.Device{}, device.ErrDeviceDoesNotExist
}

// GetFilteredDevices returns the devices which match the given GetFilteredDevices
func (cs *comboStore) GetFilteredDevices(filter device.DeviceFilter) []device.Device {
	devices := make([]device.Device, 0)
	for _, n := range cs.devices {
		if filter(n) {
			devices = append(devices, n)
		}
	}
	return devices
}

// ListDevices returns all the stored devices
func (cs *comboStore) ListDevices() []device.Device {
	return slices.Clone(cs.devices)
}

// CountDevices return the number of devices in the store
func (cs *comboStore) CountDevices() int {
	return len(cs.devices)
}

func (cs *comboStore) saveDevices() error {
	bytes, err := msgpack.Marshal(cs.devices)
	if err != nil {
		return err
	}
	return os.WriteFile(cs.directory+"/"+cs.devicefilename, bytes, 0644)
}

func (cs *comboStore) readDevices() error {
	bytes, err := os.ReadFile(cs.directory + "/" + cs.devicefilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	err = msgpack.Unmarshal(bytes, &cs.devices)
	return err
}

//
// Timeseries data
//

// WritePoint stores a given point for a device
func (cs *comboStore) WriteTimeseriesPoint(
	ctx context.Context,
	timestamp time.Time,
	device device.Device,
	points ...pinger.TimeseriesPoint,
) error {
	for _, point := range points {

		filename := cs.wspfilename(cs.directory, device.Addr, point.GetName())
		wsp, err := cs.openwsp(filename)
		if err != nil {
			return err
		}

		err = wsp.Update(point.GetValue(), cs.timeToWspTime(timestamp))
		if err != nil {
			return err
		}

		err = wsp.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadPoints returns the points from Now() minus the duration for the given type
func (cs *comboStore) ReadTimeseriesPoints(
	ctx context.Context,
	device device.Device,
	duration time.Duration,
	hydrator pinger.TimeseriesPoint,
) ([]pinger.TimeseriesPoint, error) {
	fname := cs.wspfilename(cs.directory, device.Addr, hydrator.GetName())
	wsp, err := cs.openwsp(fname)
	if err != nil {
		return nil, err
	}
	defer wsp.Close()
	ts, err := wsp.Fetch(cs.fetchlast(duration))
	if err != nil {
		return nil, err
	}
	points := ts.Points()
	numOfPoints := len(points)
	if numOfPoints == 0 {
		return nil, errors.New("no points")
	}
	ret := make([]pinger.TimeseriesPoint, 0)
	for _, point := range points {
		if !math.IsNaN(point.Value) {
			ret = append(ret, hydrator.Hydrate(point.Time, point.Value))
		}
	}
	return ret, nil
}

func (cs *comboStore) ensureDirectory(dir string) {
	stat, err := os.Stat(dir)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	if stat.IsDir() {
		return
	}
	log.Fatal("not a directory", "dir", dir)
}

// func (cs *comboStore) sanitizeAddrString(addr netip.Addr) string {
// 	return strings.Replace(addr.String(), ".", "-", -1)
// }

func (cs *comboStore) wspfilename(dir string, addr netip.Addr, name string) string {
	addrString := sanitizeAddrString(addr)
	return fmt.Sprintf("%s/%s_%s.wsp", dir, addrString, name)
}

func sanitizeAddrString(addr netip.Addr) string {
	return strings.Replace(addr.String(), ".", "-", -1)
}

func (cs *comboStore) openwsp(filename string) (*whisper.Whisper, error) {
	wsp, err := whisper.Open(filename)
	if err == nil {
		return wsp, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return whisper.Create(filename, cs.retentions, whisper.Average, 0.5)
	}
	return nil, err
}

func (cs *comboStore) fetchlast(dur time.Duration) (int, int) {
	t2 := time.Now()
	t1 := t2.Add(-1 * dur)
	return cs.timeToWspTime(t1), cs.timeToWspTime(t2)
}

func (cs *comboStore) timeToWspTime(timestamp time.Time) int {
	return int(timestamp.Unix())
}
