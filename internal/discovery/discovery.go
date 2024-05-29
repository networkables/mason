// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package discovery

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/networkables/mason/internal/config"
	"github.com/networkables/mason/internal/device"
	"github.com/networkables/mason/internal/network"
	"github.com/networkables/mason/internal/stackerr"
	"github.com/networkables/mason/nettools"
)

const (
	ArpDiscoverySource     device.DiscoverySource = "ARP"
	PingDiscoverySource    device.DiscoverySource = "PING"
	SNMPDiscoverySource    device.DiscoverySource = "SNMP"
	SNMPArpDiscoverySource device.DiscoverySource = "SNMP_ARP"
)

type (
	DiscoverDevicesFromSNMPDevice struct {
		device.Device
	}
	DiscoverNetworksFromSNMPDevice struct {
		device.Device
	}
)

type IPv6ExcludedFromDiscovery struct {
	Network network.Network
}

var ErrIPv6ExcludedFromDiscovery = IPv6ExcludedFromDiscovery{}

func (e IPv6ExcludedFromDiscovery) Error() string {
	return fmt.Sprintf("ipv6 network has been excluded from discovery %s", e.Network.String())
}

func IPv6NetworkExcluded(n network.Network) IPv6ExcludedFromDiscovery {
	return IPv6ExcludedFromDiscovery{
		Network: n,
	}
}

func (e IPv6ExcludedFromDiscovery) Is(target error) bool {
	_, ok := target.(IPv6ExcludedFromDiscovery)
	if !ok {
		return false
	}
	return true
}

type NoDeviceDiscoveredError struct {
	Addr netip.Addr
}

var ErrNoDeviceDiscovered = NoDeviceDiscoveredError{}

func (e NoDeviceDiscoveredError) Error() string {
	return fmt.Sprintf("no device discovered at %s", e.Addr.String())
}

func (e NoDeviceDiscoveredError) Is(target error) bool {
	_, ok := target.(NoDeviceDiscoveredError)
	if !ok {
		return false
	}
	return true
}

func NoDeviceDiscovered(addr netip.Addr) NoDeviceDiscoveredError {
	return NoDeviceDiscoveredError{Addr: addr}
}

type scanfunc func(netip.Addr) (device.EventDeviceDiscovered, error)

func BuildAddrScanners(cfg *config.Discovery) []scanfunc {
	ret := make([]scanfunc, 0)
	if *cfg.ArpPing.Enabled {
		ret = append(ret,
			func(addr netip.Addr) (device.EventDeviceDiscovered, error) {
				return discoverDeviceWithArp(addr, cfg.ArpPing)
			},
		)
	}
	if *cfg.IcmpPing.Enabled {
		ret = append(ret,
			func(addr netip.Addr) (device.EventDeviceDiscovered, error) {
				// TODO: Need way to pass in context
				return discoverDeviceWithICMP(context.Background(), addr, cfg.IcmpPing)
			},
		)
	}
	if *cfg.Snmp.Enabled {
		ret = append(ret,
			func(addr netip.Addr) (device.EventDeviceDiscovered, error) {
				return discoverDeviceWithSNMP(addr, cfg.Snmp)
			},
		)
	}
	return ret
}

func BuildAddrScannerFunc(funcs []scanfunc) func(netip.Addr) (device.EventDeviceDiscovered, error) {
	return func(addr netip.Addr) (device.EventDeviceDiscovered, error) {
		for _, f := range funcs {
			device, err := f(addr)
			if err == nil {
				return device, nil
			}
			// handle the errors
			if errors.Is(err, ErrNoDeviceDiscovered) {
				continue
			}
			return device, err
		}
		return device.EventDeviceDiscovered{}, NoDeviceDiscovered(addr)
	}
}

func BuildNetworkScanFunc(q chan netip.Addr, status *string) func(network.Network) (string, error) {
	return func(n network.Network) (string, error) {
		if n.Prefix.Addr().Is6() {
			return "", nil
		}

		*status = n.String()
		ni := network.NewNetworkIteratorAsChannel(n)
		for addr := range ni.C {
			q <- addr
		}
		*status = ""
		return "", nil
	}
}

func SnmpArpTableRescanFilter(cfg *config.DiscoverySNMPConfig) device.DeviceFilter {
	return func(d device.Device) bool {
		if d.SNMP.LastArpTableScan.IsZero() {
			return true
		}
		if !d.SNMP.HasArpTable {
			return false
		}
		since := time.Since(d.SNMP.LastArpTableScan)
		if since > *cfg.ArpTableRescanInterval {
			return true
		}
		return false
	}
}

func SnmpInterfaceRescanFilter(cfg *config.DiscoverySNMPConfig) device.DeviceFilter {
	return func(d device.Device) bool {
		if d.SNMP.LastInterfacesScan.IsZero() {
			return true
		}
		if !d.SNMP.HasInterfaces {
			return false
		}
		since := time.Since(d.SNMP.LastInterfacesScan)
		if since > *cfg.InterfaceRescanInterval {
			return true
		}
		return false
	}
}

func NetworkRescanFilter(cfg *config.Discovery) network.NetworkFilter {
	return func(network network.Network) bool {
		if network.LastScan.IsZero() {
			return true
		}
		since := time.Since(network.LastScan)
		if since > *cfg.NetworkScanInterval {
			return true
		}
		return false
	}
}

func discoverDeviceWithArp(
	addr netip.Addr,
	cfg *config.DiscoveryArpConfig,
) (device.EventDeviceDiscovered, error) {
	entry, err := nettools.FindHardwareAddrOf(
		context.TODO(),
		addr,
		nettools.WithArpReplyTimeout(*cfg.Timeout),
	)
	if err != nil {
		if errors.Is(err, nettools.ErrNoResponseFromRemote) {
			return device.EmptyDiscoveredDevice, NoDeviceDiscovered(addr)
		}
		return device.EmptyDiscoveredDevice, stackerr.New(err)
	}
	if err == nil {
		return device.EventDeviceDiscovered{
			Addr:         addr,
			MAC:          entry.MAC,
			DiscoveredBy: ArpDiscoverySource,
			DiscoveredAt: time.Now(),
		}, nil
	}
	return device.EmptyDiscoveredDevice, NoDeviceDiscovered(addr)
}

func discoverDeviceWithICMP(
	ctx context.Context,
	addr netip.Addr,
	cfg *config.DiscoveryICMPConfig,
) (event device.EventDeviceDiscovered, err error) {
	responses, err := nettools.Icmp4Echo(
		ctx,
		addr,
		nettools.I4EWithCount(*cfg.PingCount),
		nettools.I4EWithReadTimeout(*cfg.Timeout),
	)
	if err != nil {
		if errors.Is(err, nettools.ErrNoResponseFromRemote) {
			return event, NoDeviceDiscovered(addr)
		}
		return event, stackerr.New(err)
	}
	stats := nettools.CalculateIcmp4EchoResponseStatistics(responses)
	if stats.SuccessCount > 0 {
		ts := time.Now()
		edd := device.Device{
			Addr:         addr,
			DiscoveredBy: PingDiscoverySource,
			DiscoveredAt: ts,
		}
		edd.UpdateFromPingStats(stats, ts)
		return device.EventDeviceDiscovered(edd), nil
	}

	return event, NoDeviceDiscovered(addr)
}

func discoverDeviceWithSNMP(
	addr netip.Addr,
	cfg *config.DiscoverySNMPConfig,
) (device.EventDeviceDiscovered, error) {
	ctx := context.TODO()
	for _, port := range cfg.Ports {
		for _, community := range cfg.Community {
			ssi, err := nettools.SnmpGetSystemInfo(ctx, addr,
				nettools.WithSnmpCommunity(community),
				nettools.WithSnmpPort(port),
				nettools.WithSnmpReplyTimeout(*cfg.Timeout))
			if err != nil {
				if errors.Is(err, nettools.ErrConnectionRefused) ||
					errors.Is(err, nettools.ErrNoResponseFromRemote) {
					continue
				}
				return device.EventDeviceDiscovered{}, stackerr.New(err)
			}
			if ssi.Description != "" {
				return device.EventDeviceDiscovered{
					Addr:         addr,
					DiscoveredBy: SNMPDiscoverySource,
					DiscoveredAt: time.Now(),
					SNMP: device.SNMP{
						Name:        ssi.Name,
						Description: ssi.Description,
						Community:   community,
						Port:        port,
					},
				}, nil
			}

		}
	}
	return device.EventDeviceDiscovered{}, NoDeviceDiscovered(addr)
}
