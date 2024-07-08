// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package enrichment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/emicklei/tre"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/oui"
	"github.com/networkables/mason/nettools"
)

func PortScannerFilter(cfg *PortScanConfig) model.DeviceFilter {
	return func(d model.Device) bool {
		since := time.Since(d.Server.LastScan)
		if d.IsServer() {
			if since > cfg.ServerScanInterval {
				return true
			}
			return false
		}
		if since > cfg.DefaultScanInterval {
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
	Cfg              *Config
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
		str += "PortScan:" + e.Cfg.PortScan.PortList + " "
	}
	return str
}

func DefaultEnrichmentFields(cfg *Config) EnrichmentFields {
	return EnrichmentFields{
		PerformDNSLookup: cfg.Dns.Enabled,
		PerformOUILookup: cfg.Oui.Enabled,
		PerformPortScan:  cfg.PortScan.Enabled,
		PerformSNMPScan:  cfg.Snmp.Enabled,
		Cfg:              cfg,
	}
}

type EnrichAllDevicesEvent EnrichmentFields

type EnrichDeviceRequest struct {
	Fields EnrichmentFields
	Device model.Device
}

func (e EnrichDeviceRequest) String() string {
	return fmt.Sprintf(
		"%s %s",
		e.Device.Addr,
		e.Fields.String(),
	)
}

func (e EnrichAllDevicesEvent) String() string {
	return fmt.Sprintf("EnrichAllDevices %s", EnrichmentFields(e).String())
}

func EnrichDevice(ctx context.Context, d EnrichDeviceRequest) (model.Device, error) {
	if d.Fields.PerformDNSLookup && d.Device.Meta.DnsName == "" {
		name, err := nettools.FindHostnameOf(d.Device.Addr.Addr())
		if err != nil && !errors.Is(err, nettools.ErrNoDnsNames) {
			return d.Device, tre.New(err, "reverse lookup", "addr", d.Device.Addr)
		}
		if name != "" {
			d.Device.Meta.DnsName = name
			d.Device.SetUpdated()
		}
	}
	if d.Fields.PerformOUILookup && d.Device.Meta.Manufacturer == "" {
		if nettools.IsRandomMac(d.Device.MAC.Addr()) {
			d.Device.Meta.Tags = model.Add(model.RandomizedMacAddressTag, d.Device.Meta.Tags)
			d.Device.Meta.Manufacturer = "<randomized mac>"
			d.Device.SetUpdated()
		} else {
			manu := oui.Lookup(d.Device.MAC.Addr())
			if manu != "" {
				d.Device.Meta.Manufacturer = manu
				d.Device.Meta.Tags = model.Remove(model.RandomizedMacAddressTag, d.Device.Meta.Tags)
				d.Device.SetUpdated()
			}
		}
	}
	if d.Fields.PerformPortScan {
		openports, err := nettools.ScanTcpPorts(ctx, d.Device.Addr.Addr(),
			nettools.WithPortscanReplyTimeout(d.Fields.Cfg.PortScan.Timeout),
			nettools.WithPortscanPortlistName(d.Fields.Cfg.PortScan.PortList),
			nettools.WithPortscanMaxworkers(d.Fields.Cfg.PortScan.MaxWorkers),
		)
		if err != nil {
			return d.Device, tre.New(err, "port scan", "addr", d.Device.Addr)
		}
		d.Device.Server.Ports = model.IntSliceToPortList(openports)
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
		for _, community := range d.Fields.Cfg.Snmp.Community {
			for _, port := range d.Fields.Cfg.Snmp.Ports {
				if snmpworks {
					continue // we have a good set of credentials, skip the other tests
				}
				_, err := nettools.SnmpGetSystemInfo(ctx, d.Device.Addr.Addr(),
					nettools.WithSnmpCommunity(community),
					nettools.WithSnmpPort(port),
					nettools.WithSnmpReplyTimeout(d.Fields.Cfg.Snmp.Timeout))
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
				d.Device.Addr.Addr(),
				nettools.WithSnmpCommunity(goodcommunity),
				nettools.WithSnmpPort(goodport),
				nettools.WithSnmpReplyTimeout(d.Fields.Cfg.Snmp.Timeout))
			if err != nil {
				return d.Device, tre.New(err, "snmp check", "addr", d.Device.Addr)
			}
			d.Device.SNMP.Name = ssi.Name
			d.Device.SNMP.Description = ssi.Description
			d.Device.SetUpdated()
		}
	}
	return d.Device, nil
}
