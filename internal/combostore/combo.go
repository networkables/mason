// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package combostore

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	whisper "github.com/go-graphite/go-whisper"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/nettools"
)

type Store struct {
	directory       string
	retentions      whisper.Retentions
	networkfilename string
	devicefilename  string
	networks        []model.Network
	devices         []model.Device
}

// var _ model.Storer = (*Store)(nil)

func New(cfg *Config) (*Store, error) {
	cs := &Store{
		directory:       cfg.Directory,
		retentions:      whisper.MustParseRetentionDefs(cfg.WSPRetention),
		networkfilename: "networks.mb",
		devicefilename:  "devices.mb",
	}

	cs.ensureDirectory(cfg.Directory)
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

func (cs *Store) Close() error {
	return nil
}

//
// Network data
//

// AddNetwork adds a network to the store, will return error if the network already exists in the store
func (cs *Store) AddNetwork(ctx context.Context, newnetwork model.Network) error {
	for _, net := range cs.networks {
		if net.Prefix.ContainsAddr(newnetwork.Prefix.Addr()) {
			return model.ErrNetworkExists
		}
	}
	cs.networks = append(cs.networks, newnetwork)
	return cs.saveNetworks()
}

// RemoveNetworkByName remove the named network from the store
func (cs *Store) RemoveNetworkByName(ctx context.Context, name string) error {
	for idx, n := range cs.networks {
		if n.Name == name {
			cs.networks = slices.Delete(cs.networks, idx, idx+1)
			return cs.saveNetworks()
		}
	}
	return model.ErrNetworkDoesNotExist
}

// UpdateNetwork will freshen up the network using the given network
func (cs *Store) UpdateNetwork(ctx context.Context, newnetwork model.Network) error {
	idx := 0
	found := true
	for i, n := range cs.networks {
		if n.Prefix.String() == newnetwork.Prefix.String() {
			idx = i
			found = true
		}
	}
	if found {
		cs.networks[idx] = newnetwork
		return cs.saveNetworks()
	}
	return model.ErrNetworkExists
}

// UpsertNetwork will either add the given network and if it already exists then it will run an update
func (cs *Store) UpsertNetwork(ctx context.Context, net model.Network) error {
	err := cs.UpdateNetwork(ctx, net)
	if err == nil {
		return nil
	}
	if errors.Is(err, model.ErrNetworkDoesNotExist) {
		return cs.AddNetwork(ctx, net)
	}
	return err
}

// GetNetworkByName returns a network with the given name
func (cs *Store) GetNetworkByName(ctx context.Context, name string) (model.Network, error) {
	for _, n := range cs.networks {
		if n.Name == name {
			return n, nil
		}
	}
	return model.Network{}, model.ErrNetworkDoesNotExist
}

// GetFilteredNetworks returns the networks which match the given GetFilteredNetworks
func (cs *Store) GetFilteredNetworks(
	ctx context.Context,
	filter model.NetworkFilter,
) []model.Network {
	networks := make([]model.Network, 0)
	for _, network := range cs.networks {
		if filter(network) {
			networks = append(networks, network)
		}
	}
	return networks
}

// ListNetworks returns all stored networks
func (cs *Store) ListNetworks(ctx context.Context) []model.Network {
	return slices.Clone(cs.networks)
}

// CountNetworks returns the number of networks in the store
func (cs *Store) CountNetworks(ctx context.Context) int {
	return len(cs.networks)
}

func (cs *Store) saveNetworks() error {
	bytes, err := msgpack.Marshal(cs.networks)
	if err != nil {
		return err
	}
	return os.WriteFile(cs.directory+"/"+cs.networkfilename, bytes, 0644)
}

func (cs *Store) readNetworks() error {
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
func (cs *Store) AddDevice(ctx context.Context, newdevice model.Device) error {
	for _, d := range cs.devices {
		if d.Addr.Compare(newdevice.Addr) == 0 {
			return model.ErrDeviceExists
		}
	}
	cs.devices = append(cs.devices, newdevice)
	return cs.saveDevices()
}

// RemoveDeviceByAddr will remove the device with the given Addr from the store
func (cs *Store) RemoveDeviceByAddr(ctx context.Context, addr model.Addr) error {
	for idx, device := range cs.devices {
		if device.Addr.Compare(addr) == 0 {
			cs.devices = slices.Delete(cs.devices, idx, idx+1)
			return cs.saveDevices()
		}
	}
	return model.ErrDeviceDoesNotExist
}

// UpdateDevice will fresnen up the device using the given device
func (cs *Store) UpdateDevice(
	ctx context.Context,
	newdevice model.Device,
) (enrich bool, err error) {
	if !newdevice.IsUpdated() {
		return enrich, nil
	}
	for idx, device := range cs.devices {
		if device.Addr.Compare(newdevice.Addr) == 0 {
			enrich = device.MAC.Compare(newdevice.MAC) != 0
			cs.devices[idx] = cs.devices[idx].Merge(newdevice)
			return enrich, cs.saveDevices()
		}
	}
	return enrich, model.ErrDeviceDoesNotExist
}

// GetDeviceByAddr returns the device with the matching Addr
func (cs *Store) GetDeviceByAddr(
	ctx context.Context,
	addr model.Addr,
) (model.Device, error) {
	for _, device := range cs.devices {
		if device.Addr.Compare(addr) == 0 {
			return device, nil
		}
	}
	return model.Device{}, model.ErrDeviceDoesNotExist
}

// GetFilteredDevices returns the devices which match the given GetFilteredDevices
func (cs *Store) GetFilteredDevices(
	ctx context.Context,
	filter model.DeviceFilter,
) []model.Device {
	devices := make([]model.Device, 0)
	for _, n := range cs.devices {
		if filter(n) {
			devices = append(devices, n)
		}
	}
	return devices
}

// ListDevices returns all the stored devices
func (cs *Store) ListDevices(ctx context.Context) []model.Device {
	return slices.Clone(cs.devices)
}

// CountDevices return the number of devices in the store
func (cs *Store) CountDevices(ctx context.Context) int {
	return len(cs.devices)
}

func (cs *Store) saveDevices() error {
	bytes, err := msgpack.Marshal(cs.devices)
	if err != nil {
		return err
	}
	return os.WriteFile(cs.directory+"/"+cs.devicefilename, bytes, 0644)
}

func (cs *Store) readDevices() error {
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
func (cs *Store) WritePerformancePing(
	ctx context.Context,
	timestamp time.Time,
	device model.Device,
	point nettools.Icmp4EchoResponseStatistics,
) error {
	avgfilename := cs.wspfilename(cs.directory, device.Addr, "pingavg")
	avgwsp, err := cs.openwsp(avgfilename)
	if err != nil {
		return err
	}
	err = avgwsp.Update(float64(point.Mean/time.Millisecond), cs.timeToWspTime(timestamp))
	if err != nil {
		return err
	}
	err = avgwsp.Close()
	if err != nil {
		return err
	}

	maxfilename := cs.wspfilename(cs.directory, device.Addr, "pingmax")
	maxwsp, err := cs.openwsp(maxfilename)
	if err != nil {
		return err
	}
	err = maxwsp.Update(float64(point.Maximum/time.Millisecond), cs.timeToWspTime(timestamp))
	if err != nil {
		return err
	}
	err = maxwsp.Close()
	if err != nil {
		return err
	}

	lossfilename := cs.wspfilename(cs.directory, device.Addr, "pingloss")
	losswsp, err := cs.openwsp(lossfilename)
	if err != nil {
		return err
	}
	err = losswsp.Update(point.PacketLoss, cs.timeToWspTime(timestamp))
	if err != nil {
		return err
	}
	err = maxwsp.Close()
	if err != nil {
		return err
	}

	return nil
}

// ReadPoints returns the points from Now() minus the duration for the given type
func (cs *Store) ReadPerformancePings(
	ctx context.Context,
	device model.Device,
	duration time.Duration,
) (points []pinger.Point, err error) {
	avgname := cs.wspfilename(cs.directory, device.Addr, "pingavg")
	avgwsp, err := cs.openwsp(avgname)
	if err != nil {
		return nil, err
	}
	defer avgwsp.Close()
	avgts, err := avgwsp.Fetch(cs.fetchlast(duration))
	if err != nil {
		return nil, err
	}
	avgpoints := avgts.Points()
	numOfPoints := len(avgpoints)
	if numOfPoints == 0 {
		return nil, errors.New("no points")
	}
	points = make([]pinger.Point, numOfPoints)
	for idx, point := range avgpoints {
		if !math.IsNaN(point.Value) {
			points[idx] = pinger.Point{
				Device:  device,
				Start:   time.Unix(int64(point.Time), 0),
				Average: time.Duration(point.Value) * time.Millisecond,
			}
		}
	}

	maxname := cs.wspfilename(cs.directory, device.Addr, "pingmax")
	maxwsp, err := cs.openwsp(maxname)
	if err != nil {
		return nil, err
	}
	defer maxwsp.Close()
	maxts, err := maxwsp.Fetch(cs.fetchlast(duration))
	if err != nil {
		return nil, err
	}
	if numOfPoints == len(avgts.Points()) {
		return nil, errors.New("max point count does not equal avg count")
	}
	for idx, point := range maxts.Points() {
		if !math.IsNaN(point.Value) {
			points[idx].Maximum = time.Duration(point.Value) * time.Millisecond
		}
	}

	lossname := cs.wspfilename(cs.directory, device.Addr, "pingloss")
	losswsp, err := cs.openwsp(lossname)
	if err != nil {
		return nil, err
	}
	defer losswsp.Close()
	lossts, err := losswsp.Fetch(cs.fetchlast(duration))
	if err != nil {
		return nil, err
	}
	if numOfPoints == len(lossts.Points()) {
		return nil, errors.New("loss point count does not equal avg count")
	}
	for idx, point := range lossts.Points() {
		if !math.IsNaN(point.Value) {
			points[idx].Loss = point.Value * 100
		}
	}

	return points, nil
}

func convertPingDuration(t time.Duration) float64 {
	return float64(t) / float64(time.Millisecond)
}

func (cs *Store) ensureDirectory(dir string) {
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

// func (cs *Store) sanitizeAddrString(addr netip.Addr) string {
// 	return strings.Replace(addr.String(), ".", "-", -1)
// }

func (cs *Store) wspfilename(dir string, addr model.Addr, name string) string {
	addrString := sanitizeAddrString(addr)
	return fmt.Sprintf("%s/%s_%s.wsp", dir, addrString, name)
}

func sanitizeAddrString(addr model.Addr) string {
	return strings.Replace(addr.String(), ".", "-", -1)
}

func (cs *Store) openwsp(filename string) (*whisper.Whisper, error) {
	wsp, err := whisper.Open(filename)
	if err == nil {
		return wsp, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return whisper.Create(filename, cs.retentions, whisper.Average, 0.5)
	}
	return nil, err
}

func (cs *Store) fetchlast(dur time.Duration) (int, int) {
	t2 := time.Now()
	t1 := t2.Add(-1 * dur)
	return cs.timeToWspTime(t1), cs.timeToWspTime(t2)
}

func (cs *Store) timeToWspTime(timestamp time.Time) int {
	return int(timestamp.Unix())
}
