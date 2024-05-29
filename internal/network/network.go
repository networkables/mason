// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package network

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sort"
	"time"

	"github.com/networkables/mason/internal/device"
	"github.com/networkables/mason/internal/tags"
)

type (
	ScanAllNetworksRequest struct{}

	// NetworkFilter defines a function used to select a set of required devices.
	NetworkFilter func(Network) bool

	Network struct {
		Name     string
		Prefix   netip.Prefix
		LastScan time.Time
		Tags     []tags.Tag
	}
)

type ScanNetworkRequest Network

type DiscoveredNetwork Network

// NetworkAddEvent
type NetworkAddedEvent Network

var (
	ErrNetworkExists       = errors.New("network exists")
	ErrNetworkDoesNotExist = errors.New("network does not exists")
)

func New(name string, ns string) (Network, error) {
	prefix, err := netip.ParsePrefix(ns)
	if err != nil {
		return Network{}, err
	}
	if name == "" {
		name = ns
	}
	return Network{
		Name:   name,
		Prefix: prefix,
	}, nil
}

func findFirstAddr(prefix netip.Prefix) netip.Addr {
	var prev netip.Addr
	addr := prefix.Addr()
	for {
		prev = addr.Prev()
		if prefix.Contains(prev) {
			addr = prev
			continue
		}
		return addr
	}
}

func findLastAddr(prefix netip.Prefix) netip.Addr {
	var next netip.Addr
	addr := prefix.Addr()
	for {
		next = addr.Next()
		if prefix.Contains(next) {
			addr = next
			continue
		}
		return addr
	}
}

func NewNetworkFromPrefix(ns netip.Prefix) Network {
	return Network{
		Name:   ns.String(),
		Prefix: ns,
	}
}

func (n Network) String() string {
	return fmt.Sprintf("%s > %s tags:%s", n.Name, n.Prefix, n.Tags)
}

func (n Network) Contains(d device.Device) bool {
	return n.Prefix.Contains(d.Addr)
}

// networkIterator
type networkIterator struct {
	network Network
	current netip.Addr
	first   netip.Addr
	last    netip.Addr
	Size    int
}

func NewNetworkIterator(n Network) *networkIterator {
	first := findFirstAddr(n.Prefix)
	last := findLastAddr(n.Prefix)
	curr := first.Next()
	size := 1
	x := first
	for {
		if x.Compare(last) == 0 {
			break
		}
		x = x.Next()
		size++
	}
	return &networkIterator{
		network: n,
		current: curr,
		first:   first,
		last:    last,
		Size:    size,
	}
}

func (ni *networkIterator) Next() (ip netip.Addr, done bool) {
	ip = ni.current
	ni.current = ni.current.Next()
	if ip.Compare(ni.last) == 0 {
		return netip.Addr{}, true
	}
	if !ni.network.Prefix.Contains(ip) {
		return netip.Addr{}, true
	}
	return ip, false
}

func (ni *networkIterator) Reset() {
	ni.current = ni.network.Prefix.Addr()
}

// NetworkIteratorAsChannel
type networkIteratorAsChannel struct {
	ni *networkIterator
	C  chan netip.Addr
}

func NewNetworkIteratorAsChannel(n Network) *networkIteratorAsChannel {
	ch := make(chan netip.Addr)
	ni := NewNetworkIterator(n)

	go func() {
		for {
			addr, done := ni.Next()
			if done {
				close(ch)
				return
			}
			ch <- addr
		}
	}()

	return &networkIteratorAsChannel{
		ni: ni,
		C:  ch,
	}
}

func IsUsefulInterface(iface net.Interface) bool {
	if iface.Flags&net.FlagLoopback != 0 {
		return false
	}
	if iface.Flags&net.FlagUp == 0 {
		return false
	}
	if iface.Flags&net.FlagRunning == 0 {
		return false
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return false
	}
	if len(addrs) == 0 {
		return false
	}
	return true
}

func SortNetworksByAddr(nets []Network) {
	sort.SliceStable(nets, func(i, j int) bool {
		return nets[i].Prefix.Addr().Compare(nets[j].Prefix.Addr()) == -1
	})
}

type NetworkStats struct {
	Network
	IPUsed  uint64
	IPTotal float64
	AvgPing time.Duration
	MaxPing time.Duration
}
