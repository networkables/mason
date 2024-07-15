// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

//go:build linux || freebsd || openbsd || darwin

package nettools

import (
	"net"
	"net/netip"

	"github.com/vishvananda/netlink"
)

func (p *pkg) loadRoutes() error {
	p.ifacesByNetPrefix = make(map[string]*net.Interface)
	// using 0 for family, unsure what the range is, doesn't seem to change with values provided
	routes, err := netlink.RouteList(nil, 0)
	if err != nil {
		return err
	}
	for _, route := range routes {
		iface, err := net.InterfaceByIndex(route.LinkIndex)
		if err != nil {
			return err
		}
		if route.Dst != nil {
			p.ifacesByNetPrefix[route.Dst.String()] = iface
			continue
		}

		// This handles the default route
		addr, ok := netip.AddrFromSlice(route.Gw)
		if ok {
			p.defaultRouteIface = iface
			p.defaultRouteGateway = addr
		}
	}
	return nil
}
