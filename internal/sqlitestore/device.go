// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"
	"slices"
	"strings"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

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
	conn, err := cs.Pool.Get(ctx)
	if err != nil {
		return err
	}
	fn := sqlitex.Transaction(conn)
	defer func() {
		fn(&err)
		cs.Pool.Put(conn)
	}()
	for _, device := range cs.devices {
		err = upsertDevice(conn, device)
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
	stmt, err := cs.DB.Prepare(
		`SELECT 
      name, addr, mac, discoveredat, discoveredby,
      metadnsname AS "meta.dnsname", metamanufacturer AS "meta.manufacturer", metatags AS "meta.tags",
      serverports AS "server.ports", serverlastscan AS "server.lastscan",
      perfpingfirstseen AS "performanceping.firstseen", perfpinglastseen AS "performanceping.lastseen", perfpingmeanping AS "performanceping.mean", perfpingmaxping AS "performanceping.maximum", perfpinglastfailed AS "performanceping.lastfailed",
      snmpname AS "snmp.name", snmpdescription AS "snmp.description", snmpcommunity AS "snmp.community", snmpport AS "snmp.port", snmplastcheck AS "snmp.lastsnmpcheck", snmphasarptable AS "snmp.hasarptable", snmplastarptablescan AS "snmp.lastarptablescan", snmphasinterfaces AS "snmp.hasinterfaces", snmplastinterfacesscan AS "snmp.lastinterfacesscan"
    FROM devices`,
	)
	if err != nil {
		return devices, err
	}

	var hasRow bool
	for {
		hasRow, err = stmt.Step()
		if err != nil {
			return devices, err
		}
		if !hasRow {
			break
		}
		device := model.Device{
			Name: stmt.GetText("name"),
			Meta: model.Meta{
				DnsName:      stmt.GetText("meta.dnsname"),
				Manufacturer: stmt.GetText("meta.manufacturer"),
			},
			PerformancePing: model.Pinger{
				LastFailed: stmt.GetBool("performanceping.lastfailed"),
				Mean:       time.Duration(stmt.GetInt64("performanceping.mean")),
				Maximum:    time.Duration(stmt.GetInt64("performanceping.maximum")),
			},
			SNMP: model.SNMP{
				Name:          stmt.GetText("snmp.name"),
				Description:   stmt.GetText("snmp.description"),
				Community:     stmt.GetText("snmp.community"),
				Port:          int(stmt.GetInt64("snmp.port")),
				HasArpTable:   stmt.GetBool("snmp.hasarptable"),
				HasInterfaces: stmt.GetBool("snmp.hasinterfaces"),
			},
		}
		err = device.Addr.Scan(stmt.GetText("addr"))
		if err != nil {
			return devices, err
		}
		err = device.MAC.Scan(stmt.GetText("mac"))
		if err != nil {
			return devices, err
		}
		device.DiscoveredAt, err = time.Parse(time.RFC3339Nano, stmt.GetText("discoveredat"))
		if err != nil {
			return devices, err
		}
		err = device.DiscoveredBy.Scan(stmt.GetText("discoveredby"))
		if err != nil {
			return devices, err
		}
		err = device.Meta.Tags.Scan(stmt.GetText("meta.tags"))
		if err != nil {
			return devices, err
		}

		err = device.Server.Ports.Scan(stmt.GetText("server.ports"))
		if err != nil {
			return devices, err
		}
		device.Server.LastScan, err = time.Parse(time.RFC3339Nano, stmt.GetText("server.lastscan"))
		if err != nil {
			return devices, err
		}

		device.PerformancePing.FirstSeen, err = time.Parse(
			time.RFC3339Nano,
			stmt.GetText("performanceping.firstseen"),
		)
		if err != nil {
			return devices, err
		}
		device.PerformancePing.LastSeen, err = time.Parse(
			time.RFC3339Nano,
			stmt.GetText("performanceping.lastseen"),
		)
		if err != nil {
			return devices, err
		}
		device.SNMP.LastSNMPCheck, err = time.Parse(
			time.RFC3339Nano,
			stmt.GetText("snmp.lastsnmpcheck"),
		)
		if err != nil {
			return devices, err
		}
		device.SNMP.LastArpTableScan, err = time.Parse(
			time.RFC3339Nano,
			stmt.GetText("snmp.lastarptablescan"),
		)
		if err != nil {
			return devices, err
		}
		device.SNMP.LastInterfacesScan, err = time.Parse(
			time.RFC3339Nano,
			stmt.GetText("snmp.lastinterfacesscan"),
		)
		if err != nil {
			return devices, err
		}

		devices = append(devices, device)
	}

	return devices, err
}

func upsertDevice(conn *sqlite.Conn, d model.Device) error {
	stmt, err := conn.Prepare(
		`INSERT INTO devices (
      name, addr, mac, discoveredat, discoveredby,
      metadnsname, metamanufacturer, metatags,
      serverports, serverlastscan,
      perfpingfirstseen, perfpinglastseen, perfpingmeanping, perfpingmaxping, perfpinglastfailed,
      snmpname, snmpdescription, snmpcommunity, snmpport, snmplastcheck, snmphasarptable, snmplastarptablescan, snmphasinterfaces, snmplastinterfacesscan
    )
    VALUES (
      :name, :addr, :mac, :discoveredat, :discoveredby,
      :metadnsname, :metamanufacturer, :metatags,
      :serverports, :serverlastscan,
      :performancepingfirstseen, :performancepinglastseen, :performancepingmean, :performancepingmaximum, :performancepinglastfailed,
      :snmpname, :snmpdescription, :snmpcommunity, :snmpport, :snmplastsnmpcheck, :snmphasarptable, :snmplastarptablescan, :snmphasinterfaces, :snmplastinterfacesscan
    )
    ON CONFLICT (addr) DO UPDATE SET 
      name=:name, addr=:addr, mac=:mac, discoveredat=:discoveredat, discoveredby=:discoveredby,
      metadnsname=:metadnsname, metamanufacturer=:metamanufacturer, metatags=:metatags,
      serverports=:serverports, serverlastscan=:serverlastscan,
      perfpingfirstseen=:performancepingfirstseen, perfpinglastseen=:performancepinglastseen, perfpingmeanping=:performancepingmean, perfpingmaxping=:performancepingmaximum, perfpinglastfailed=:performancepinglastfailed,
      snmpname=:snmpname, snmpdescription=:snmpdescription, snmpcommunity=:snmpcommunity, snmpport=:snmpport, snmplastcheck=:snmplastsnmpcheck, 
      snmphasarptable=:snmphasarptable, snmplastarptablescan=:snmplastarptablescan, 
      snmphasinterfaces=:snmphasinterfaces, snmplastinterfacesscan=:snmplastinterfacesscan
    `)
	if err != nil {
		return err
	}
	stmt.SetText(":name", d.Name)
	stmt.SetText(":addr", d.Addr.String())
	stmt.SetText(":mac", d.MAC.String())
	stmt.SetText(":discoveredat", d.DiscoveredAt.Format(time.RFC3339Nano))
	stmt.SetText(":discoveredby", d.DiscoveredBy.String())
	stmt.SetText(":metadnsname", d.Meta.DnsName)
	stmt.SetText(":metamanufacturer", d.Meta.Manufacturer)
	stmt.SetText(":metatags", d.Meta.Tags.String())
	stmt.SetText(":serverports", d.Server.Ports.String())
	stmt.SetText(":serverlastscan", d.Server.LastScan.Format(time.RFC3339Nano))
	stmt.SetText(":performancepingfirstseen", d.PerformancePing.FirstSeen.Format(time.RFC3339Nano))
	stmt.SetText(":performancepinglastseen", d.PerformancePing.LastSeen.Format(time.RFC3339Nano))
	stmt.SetInt64(":performancepingmean", d.PerformancePing.Mean.Nanoseconds())
	stmt.SetInt64(":performancepingmaximum", d.PerformancePing.Maximum.Nanoseconds())
	stmt.SetBool(":performancepinglastfailed", d.PerformancePing.LastFailed)
	stmt.SetText(":snmpname", d.SNMP.Name)
	stmt.SetText(":snmpdescription", d.SNMP.Description)
	stmt.SetText(":snmpcommunity", d.SNMP.Community)
	stmt.SetInt64(":snmpport", int64(d.SNMP.Port))
	stmt.SetText(":snmplastsnmpcheck", d.SNMP.LastSNMPCheck.Format(time.RFC3339Nano))
	stmt.SetBool(":snmphasarptable", d.SNMP.HasArpTable)
	stmt.SetText(":snmplastarptablescan", d.SNMP.LastArpTableScan.Format(time.RFC3339Nano))
	stmt.SetBool(":snmphasinterfaces", d.SNMP.HasInterfaces)
	stmt.SetText(":snmplastinterfacesscan", d.SNMP.LastInterfacesScan.Format(time.RFC3339Nano))

	_, err = stmt.Step()
	return err
}
