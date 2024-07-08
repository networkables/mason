// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"
	"slices"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/networkables/mason/internal/model"
)

// AddDevice adds a device to the store, will return error if the device already exists
func (cs *Store) AddDevice(ctx context.Context, newdevice model.Device) error {
	for _, d := range cs.devices {
		if d.Addr.Compare(newdevice.Addr) == 0 {
			return model.ErrDeviceExists
		}
	}
	cs.devices = append(cs.devices, newdevice)
	return cs.saveDevices(ctx)
}

// RemoveDeviceByAddr will remove the device with the given Addr from the store
func (cs *Store) RemoveDeviceByAddr(ctx context.Context, addr model.Addr) error {
	for idx, device := range cs.devices {
		if device.Addr.Compare(addr) == 0 {
			cs.devices = slices.Delete(cs.devices, idx, idx+1)
			return cs.saveDevices(ctx)
		}
	}
	return model.ErrDeviceDoesNotExist
}

// UpdateDevice will fresnen up the device using the given device
func (cs *Store) UpdateDevice(
	ctx context.Context,
	newdevice model.Device,
) (enrich bool, err error) {
	// if !newdevice.IsUpdated() {
	// 	return enrich, nil
	// }
	for idx, device := range cs.devices {
		if device.Addr.Compare(newdevice.Addr) == 0 {
			enrich = !newdevice.MAC.IsEmpty() && device.MAC.Compare(newdevice.MAC) != 0
			cs.devices[idx] = cs.devices[idx].Merge(newdevice)
			return enrich, cs.saveDevices(ctx)
		}
	}
	return enrich, model.ErrDeviceDoesNotExist
}

// GetDeviceByAddr returns the device with the matching Addr
func (cs *Store) GetDeviceByAddr(
	ctx context.Context,
	addr model.Addr,
) (model.Device, error) {
	for _, d := range cs.devices {
		if d.Addr.Compare(addr) == 0 {
			return d, nil
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

func (cs *Store) saveDevices(ctx context.Context) (err error) {
	for _, device := range cs.devices {
		err = cs.upsertDevice(ctx, device)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs *Store) readDevicesInitial(ctx context.Context) (err error) {
	err = cs.readDevices(ctx)
	if err != nil && strings.EqualFold(err.Error(), "no such table: devices") {
		return nil
	}
	return err
}

func (cs *Store) readDevices(ctx context.Context) (err error) {
	cs.devices, err = cs.selectDevices(ctx)
	return err
}

func (cs *Store) selectDevices(ctx context.Context) (devices []model.Device, err error) {
	err = cs.DB.SelectContext(
		ctx,
		&devices,
		`select 
      name, addr, mac, discoveredat, discoveredby,
      metadnsname as "meta.dnsname", metamanufacturer as "meta.manufacturer", metatags as "meta.tags",
      serverports as "server.ports", serverlastscan as "server.lastscan",
      perfpingfirstseen as "performanceping.firstseen", perfpinglastseen as "performanceping.lastseen", perfpingmeanping as "performanceping.mean", perfpingmaxping as "performanceping.maximum", perfpinglastfailed as "performanceping.lastfailed",
      snmpname as "snmp.name", snmpdescription as "snmp.description", snmpcommunity as "snmp.community", snmpport as "snmp.port", snmplastcheck as "snmp.lastsnmpcheck", snmphasarptable as "snmp.hasarptable", snmplastarptablescan as "snmp.lastarptablescan", snmphasinterfaces as "snmp.hasinterfaces", snmplastinterfacesscan as "snmp.lastinterfacesscan"
    from devices`,
	)
	return devices, err
}

func (cs *Store) upsertDevice(ctx context.Context, d model.Device) error {
	_, err := cs.DB.NamedExecContext(
		ctx,
		`insert into devices (
      name, addr, mac, discoveredat, discoveredby,
      metadnsname, metamanufacturer, metatags,
      serverports, serverlastscan,
      perfpingfirstseen, perfpinglastseen, perfpingmeanping, perfpingmaxping, perfpinglastfailed,
      snmpname, snmpdescription, snmpcommunity, snmpport, snmplastcheck, snmphasarptable, snmplastarptablescan, snmphasinterfaces, snmplastinterfacesscan
    )
    values (
      :name, :addr, :mac, :discoveredat, :discoveredby,
      :meta.dnsname, :meta.manufacturer, :meta.tags,
      :server.ports, :server.lastscan,
      :performanceping.firstseen, :performanceping.lastseen, :performanceping.mean, :performanceping.maximum, :performanceping.lastfailed,
      :snmp.name, :snmp.description, :snmp.community, :snmp.port, :snmp.lastsnmpcheck, :snmp.hasarptable, :snmp.lastarptablescan, :snmp.hasinterfaces, :snmp.lastinterfacesscan
    )
    on conflict (addr) do update set 
      name=:name, addr=:addr, mac=:mac, discoveredat=:discoveredat, discoveredby=:discoveredby,
      metadnsname=:meta.dnsname, metamanufacturer=:meta.manufacturer, metatags=:meta.tags,
      serverports=:server.ports, serverlastscan=:server.lastscan,
      perfpingfirstseen=:performanceping.firstseen, perfpinglastseen=:performanceping.lastseen, perfpingmeanping=:performanceping.mean, perfpingmaxping=:performanceping.maximum, perfpinglastfailed=:performanceping.lastfailed,
      snmpname=:snmp.name, snmpdescription=:snmp.description, snmpcommunity=:snmp.community, snmpport=:snmp.port, snmplastcheck=:snmp.lastsnmpcheck, 
      snmphasarptable=:snmp.hasarptable, snmplastarptablescan=:snmp.lastarptablescan, 
      snmphasinterfaces=:snmp.hasinterfaces, snmplastinterfacesscan=:snmp.lastinterfacesscan
    `,
		d,
	)
	return err
}
