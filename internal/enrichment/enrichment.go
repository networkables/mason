// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package enrichment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/networkables/mason/internal/config"
	"github.com/networkables/mason/internal/device"
	"github.com/networkables/mason/internal/stackerr"
	"github.com/networkables/mason/internal/tags"
	"github.com/networkables/mason/nettools"
)

func PortScannerFilter(cfg *config.PortScannerConfig) device.DeviceFilter {
	return func(d device.Device) bool {
		since := time.Since(d.Server.LastScan)
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

// TODO: This should probably go away and just use the EnrichmentConfig
type EnrichmentFields struct {
	PerformDNSLookup bool
	PerformOUILookup bool
	PerformPortScan  bool
	PerformSNMPScan  bool
	Cfg              *config.Enrichment
}

func (e EnrichmentFields) String() string {
	str := ""
	if e.PerformDNSLookup {
		str += "DNS "
	}
	if e.PerformOUILookup {
		str += "OUI "
	}
	if e.PerformSNMPScan {
		str += "SNMP "
	}
	if e.PerformPortScan {
		str += "PortScan:" + *e.Cfg.PortScanner.PortList + " "
	}
	return str
}

func DefaultEnrichmentFields(cfg *config.Enrichment) EnrichmentFields {
	return EnrichmentFields{
		PerformDNSLookup: *cfg.DnsLookup.Enabled,
		PerformOUILookup: *cfg.OuiLookup.Enabled,
		PerformPortScan:  *cfg.PortScanner.Enabled,
		PerformSNMPScan:  *cfg.SNMP.Enabled,
		Cfg:              cfg,
	}
}

type EnrichAllDevicesEvent EnrichmentFields

type EnrichDeviceRequest struct {
	Fields EnrichmentFields
	Device device.Device
}

func (e EnrichDeviceRequest) String() string {
	return fmt.Sprintf(
		"%s %s",
		e.Device.Addr,
		e.Fields.String(),
	)
}

func (e EnrichAllDevicesEvent) String() string {
	return fmt.Sprintf("ScanAllDevices %s", EnrichmentFields(e).String())
}

func EnrichDevice(d EnrichDeviceRequest) (device.Device, error) {
	ctx := context.TODO()
	if d.Fields.PerformDNSLookup && d.Device.Meta.DnsName == "" {
		name, err := nettools.FindHostnameOf(d.Device.Addr)
		if err != nil && !errors.Is(err, nettools.ErrNoDnsNames) {
			// log.Error("dnsReverseLookup problem", "error", err)
			return d.Device, stackerr.New(err)
		}
		if name != "" {
			d.Device.Meta.DnsName = name
			d.Device.SetUpdated()
		}
	}
	if d.Fields.PerformOUILookup && d.Device.Meta.Manufacturer == "" {
		manu, err := nettools.OuiLookup(d.Device.MAC)
		switch {
		case err == nil:
			d.Device.Meta.Manufacturer = manu
			d.Device.Meta.Tags = tags.Remove(tags.RandomizedMacAddressTag, d.Device.Meta.Tags)
			d.Device.SetUpdated()
		case errors.Is(err, nettools.ErrRandomizedMacAddress):
			d.Device.Meta.Tags = tags.Add(tags.RandomizedMacAddressTag, d.Device.Meta.Tags)
			d.Device.Meta.Manufacturer = "<randomized mac>"
			d.Device.SetUpdated()
		default:
			return d.Device, stackerr.New(err)
		}
	}
	if d.Fields.PerformPortScan {
		openports, err := nettools.ScanTcpPorts(ctx, d.Device.Addr,
			nettools.WithPortscanReplyTimeout(*d.Fields.Cfg.PortScanner.PortTimeout),
			nettools.WithPortscanPortlistName(*d.Fields.Cfg.PortScanner.PortList),
			nettools.WithPortscanMaxworkers(*d.Fields.Cfg.PortScanner.MaxWorkers),
		)
		if err != nil {
			return d.Device, stackerr.New(err)
		}
		d.Device.Server.Ports = openports
		d.Device.Server.LastScan = time.Now()
		d.Device.SetUpdated()
	}
	if d.Fields.PerformSNMPScan {
		var (
			snmpworks     bool
			goodcommunity string
			goodport      int
		)
		d.Device.SNMP.LastSNMPCheck = time.Now()
		d.Device.SetUpdated()
		for _, community := range d.Fields.Cfg.SNMP.Community {
			for _, port := range d.Fields.Cfg.SNMP.Ports {
				if snmpworks {
					continue // we have a good set of credentials, skip the other tests
				}
				_, err := nettools.SnmpGetSystemInfo(ctx, d.Device.Addr,
					nettools.WithSnmpCommunity(community),
					nettools.WithSnmpPort(port),
					nettools.WithSnmpReplyTimeout(*d.Fields.Cfg.SNMP.Timeout))
				if err != nil {
					if errors.Is(err, nettools.ErrConnectionRefused) ||
						errors.Is(err, nettools.ErrNoResponseFromRemote) {
						continue
					}
					continue
				}
				// good hit
				snmpworks = true
				goodcommunity = community
				goodport = port
			}
		}
		if snmpworks {
			d.Device.SNMP.Community = goodcommunity
			d.Device.SNMP.Port = goodport
			ssi, err := nettools.SnmpGetSystemInfo(
				ctx,
				d.Device.Addr,
				nettools.WithSnmpCommunity(goodcommunity),
				nettools.WithSnmpPort(goodport),
				nettools.WithSnmpReplyTimeout(*d.Fields.Cfg.SNMP.Timeout))
			if err != nil {
				return d.Device, stackerr.New(err)
			}
			d.Device.SNMP.Name = ssi.Name
			d.Device.SNMP.Description = ssi.Description
			d.Device.SetUpdated()
		}
	}
	return d.Device, nil
}
