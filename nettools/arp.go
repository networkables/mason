// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

import (
	"context"
	"net"
	"net/netip"
	"time"

	"github.com/mdlayher/arp"
)

var _ Arper = (*pkg)(nil)

type Arper interface {
	FindHardwareAddrOf(context.Context, netip.Addr, ...arpRequestOptionFunc) (ArpEntry, error)
	FindUsingIfNameHardwareAddrOf(context.Context, string, netip.Addr, ...arpRequestOptionFunc) (ArpEntry, error)
}

func FindHardwareAddrOf(ctx context.Context, target netip.Addr, options ...arpRequestOptionFunc) (entry ArpEntry, err error) {
	return DefaultPkg.FindHardwareAddrOf(ctx, target, options...)
}

func FindUsingIfNameHardwareAddrOf(ctx context.Context, ifname string, target netip.Addr, options ...arpRequestOptionFunc) (entry ArpEntry, err error) {
	return DefaultPkg.FindUsingIfNameHardwareAddrOf(ctx, ifname, target, options...)
}

type ArpEntry struct {
	Addr netip.Addr
	MAC  net.HardwareAddr
}

func (p *pkg) FindHardwareAddrOf(ctx context.Context, target netip.Addr, options ...arpRequestOptionFunc) (entry ArpEntry, err error) {
	iface, _ := p.bestInterface(target)
	return p.FindUsingIfNameHardwareAddrOf(ctx, iface.Name, target, options...)
}

func (p *pkg) FindUsingIfNameHardwareAddrOf(ctx context.Context, ifname string, target netip.Addr, options ...arpRequestOptionFunc) (entry ArpEntry, err error) {
	opts := applyArpRequestOptions(options...)
	entry.Addr = target

	if !opts.skipcache {
		if mac, ok := p.arptable[target]; ok {
			entry.MAC = mac
			return entry, nil
		}
	}

	// Ensure valid network interface
	ifi, err := net.InterfaceByName(ifname)
	if err != nil {
		return entry, err
	}

	// Set up ARP client with socket
	c, err := arp.Dial(ifi)
	if err != nil {
		return entry, err
	}
	defer c.Close()

	// Set request deadline from flag
	if err := c.SetDeadline(time.Now().Add(opts.responseTimeout)); err != nil {
		return entry, err
	}

	// Request hardware address for IP address
	mac, err := c.Resolve(target)
	if err != nil {
		if err, ok := err.(*net.OpError); ok {
			if err.Timeout() {
				return entry, ErrNoResponseFromRemote
			}
		}
		return entry, err
	}
	entry.MAC = mac

	// TODO: multi writer enabled
	p.arptable[target] = mac

	return entry, nil
}

//
// Options available for ARP Requests
//

type arpRequestOptions struct {
	skipcache       bool
	responseTimeout time.Duration
}

func defaultArpRequestOptions() *arpRequestOptions {
	return &arpRequestOptions{
		skipcache:       false,
		responseTimeout: 100 * time.Millisecond,
	}
}

func WithArpReplyTimeout(duration time.Duration) arpRequestOptionFunc {
	return func(o *arpRequestOptions) {
		o.responseTimeout = duration
	}
}

func WithArpNoCache() arpRequestOptionFunc {
	return func(o *arpRequestOptions) {
		o.skipcache = true
	}
}

type arpRequestOptionFunc func(*arpRequestOptions)

func applyArpRequestOptions(options ...arpRequestOptionFunc) *arpRequestOptions {
	opts := defaultArpRequestOptions()
	for _, f := range options {
		f(opts)
	}
	return opts
}
