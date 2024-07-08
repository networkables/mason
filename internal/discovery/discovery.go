// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package discovery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/emicklei/tre"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/nettools"
)

const (
	ArpDiscoverySource     model.DiscoverySource = "ARP"
	PingDiscoverySource    model.DiscoverySource = "PING"
	SNMPDiscoverySource    model.DiscoverySource = "SNMP"
	SNMPArpDiscoverySource model.DiscoverySource = "SNMP_ARP"
)

type (
	DiscoverDevicesFromSNMPDevice struct {
		model.Device
	}
	DiscoverNetworksFromSNMPDevice struct {
		model.Device
	}
)

type IPv6ExcludedFromDiscovery struct {
	Network model.Network
}

var ErrIPv6ExcludedFromDiscovery = IPv6ExcludedFromDiscovery{}

func (e IPv6ExcludedFromDiscovery) Error() string {
	return fmt.Sprintf("ipv6 network has been excluded from discovery %s", e.Network.String())
}

func IPv6NetworkExcluded(n model.Network) IPv6ExcludedFromDiscovery {
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
	Addr model.Addr
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

func NoDeviceDiscovered(addr model.Addr) NoDeviceDiscoveredError {
	return NoDeviceDiscoveredError{Addr: addr}
}

type scanfunc func(context.Context, model.Addr) (model.EventDeviceDiscovered, error)

func BuildAddrScanners(cfg *Config) []scanfunc {
	ret := make([]scanfunc, 0)
	if cfg.Arp.Enabled {
		ret = append(ret,
			func(ctx context.Context, addr model.Addr) (model.EventDeviceDiscovered, error) {
				return discoverDeviceWithArp(ctx, addr, cfg.Arp)
			},
		)
	}
	if cfg.Icmp.Enabled {
		ret = append(ret,
			func(ctx context.Context, addr model.Addr) (model.EventDeviceDiscovered, error) {
				return discoverDeviceWithICMP(ctx, addr, cfg.Icmp)
			},
		)
	}
	if cfg.Snmp.Enabled {
		ret = append(ret,
			func(ctx context.Context, addr model.Addr) (model.EventDeviceDiscovered, error) {
				return discoverDeviceWithSNMP(ctx, addr, cfg.Snmp)
			},
		)
	}
	return ret
}

func BuildAddrScannerFunc(
	funcs []scanfunc,
) func(context.Context, model.Addr) (model.EventDeviceDiscovered, error) {
	return func(ctx context.Context, addr model.Addr) (model.EventDeviceDiscovered, error) {
		for _, f := range funcs {
			device, err := f(ctx, addr)
			if err == nil {
				return device, nil
			}
			// handle the errors
			if errors.Is(err, ErrNoDeviceDiscovered) {
				continue
			}
			return device, err
		}
		return model.EventDeviceDiscovered{}, NoDeviceDiscovered(addr)
	}
}

func BuildNetworkScanFunc(
	q chan model.Addr,
	status *string,
) func(context.Context, model.Network) (string, error) {
	return func(ctx context.Context, n model.Network) (string, error) {
		if n.Prefix.Is6() {
			return "", nil
		}

		*status = n.String()
		ni := model.NewNetworkIteratorAsChannel(n)
		for addr := range ni.C {
			if ctx.Err() != nil {
				return "", nil
			}
			select {
			case <-ctx.Done():
				break
			case q <- addr:
			}
		}
		*status = ""
		return "", nil
	}
}

func SnmpArpTableRescanFilter(cfg *SNMPConfig) model.DeviceFilter {
	return func(d model.Device) bool {
		if d.SNMP.LastArpTableScan.IsZero() {
			return true
		}
		if !d.SNMP.HasArpTable {
			return false
		}
		since := time.Since(d.SNMP.LastArpTableScan)
		if since > cfg.ArpTableRescanInterval {
			return true
		}
		return false
	}
}

func SnmpInterfaceRescanFilter(cfg *SNMPConfig) model.DeviceFilter {
	return func(d model.Device) bool {
		if d.SNMP.LastInterfacesScan.IsZero() {
			return true
		}
		if !d.SNMP.HasInterfaces {
			return false
		}
		since := time.Since(d.SNMP.LastInterfacesScan)
		if since > cfg.InterfaceRescanInterval {
			return true
		}
		return false
	}
}

func NetworkRescanFilter(cfg *Config) model.NetworkFilter {
	return func(network model.Network) bool {
		if network.LastScan.IsZero() {
			return true
		}
		since := time.Since(network.LastScan)
		if since > cfg.NetworkScanInterval {
			return true
		}
		return false
	}
}

func discoverDeviceWithArp(
	ctx context.Context,
	addr model.Addr,
	cfg *ArpConfig,
) (model.EventDeviceDiscovered, error) {
	entry, err := nettools.FindHardwareAddrOf(
		ctx,
		addr.Addr(),
		nettools.WithArpReplyTimeout(cfg.Timeout),
	)
	if err != nil {
		if errors.Is(err, nettools.ErrNoResponseFromRemote) {
			return model.EmptyDiscoveredDevice, NoDeviceDiscovered(addr)
		}
		return model.EmptyDiscoveredDevice, tre.New(
			err,
			"find hardware addr of",
			"addr",
			addr.Addr(),
		)
	}
	if err == nil {
		return model.EventDeviceDiscovered{
			Addr:         addr,
			MAC:          model.HardwareAddrToMAC(entry.MAC),
			DiscoveredBy: ArpDiscoverySource,
			DiscoveredAt: time.Now(),
		}, nil
	}
	return model.EmptyDiscoveredDevice, NoDeviceDiscovered(addr)
}

func discoverDeviceWithICMP(
	ctx context.Context,
	addr model.Addr,
	cfg *ICMPConfig,
) (event model.EventDeviceDiscovered, err error) {
	responses, err := nettools.Icmp4Echo(
		ctx,
		addr.Addr(),
		nettools.I4EWithCount(cfg.PingCount),
		nettools.I4EWithReadTimeout(cfg.Timeout),
		nettools.I4EWithPrivileged(cfg.Privileged),
		nettools.I4EWithBetweenDuration(cfg.SleepBetween),
	)
	if err != nil {
		if errors.Is(err, nettools.ErrNoResponseFromRemote) {
			return event, NoDeviceDiscovered(addr)
		}
		return event, tre.New(err, "icmp4 echo")
	}
	stats := nettools.CalculateIcmp4EchoResponseStatistics(responses)
	if stats.SuccessCount > 0 {
		ts := time.Now()
		edd := model.Device{
			Addr:         addr,
			DiscoveredBy: PingDiscoverySource,
			DiscoveredAt: ts,
		}
		edd.UpdateFromPingStats(stats, ts)
		return model.EventDeviceDiscovered(edd), nil
	}

	return event, NoDeviceDiscovered(addr)
}

func discoverDeviceWithSNMP(
	ctx context.Context,
	addr model.Addr,
	cfg *SNMPConfig,
) (model.EventDeviceDiscovered, error) {
	for _, port := range cfg.Ports {
		for _, community := range cfg.Community {
			ssi, err := nettools.SnmpGetSystemInfo(ctx, addr.Addr(),
				nettools.WithSnmpCommunity(community),
				nettools.WithSnmpPort(port),
				nettools.WithSnmpReplyTimeout(cfg.Timeout))
			if err != nil {
				if errors.Is(err, nettools.ErrConnectionRefused) ||
					errors.Is(err, nettools.ErrNoResponseFromRemote) {
					continue
				}
				return model.EventDeviceDiscovered{}, tre.New(err, "snmp check")
			}
			if ssi.Description != "" {
				return model.EventDeviceDiscovered{
					Addr:         addr,
					DiscoveredBy: SNMPDiscoverySource,
					DiscoveredAt: time.Now(),
					SNMP: model.SNMP{
						Name:        ssi.Name,
						Description: ssi.Description,
						Community:   community,
						Port:        port,
					},
				}, nil
			}

		}
	}
	return model.EventDeviceDiscovered{}, NoDeviceDiscovered(addr)
}
