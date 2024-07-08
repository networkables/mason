// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sort"
	"time"

	"go4.org/netipx"
)

type (
	ScanAllNetworksRequest struct{}

	// NetworkFilter defines a function used to select a set of required devices.
	NetworkFilter func(Network) bool

	Network struct {
		Name     string
		Prefix   Prefix
		LastScan time.Time
		Tags     Tags
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
	rng := netipx.RangeOfPrefix(prefix)
	prng, _ := rng.Prefix()

	if name == "" || name == ns {
		name = prng.String()
	}
	return Network{
		Name:   name,
		Prefix: PrefixToModelPrefix(prng),
	}, nil
}

func NewNetworkFromPrefix(ns netip.Prefix) Network {
	n, err := New(ns.String(), ns.String())
	if err != nil {
		panic(err)
	}
	return n
}

func (n Network) String() string {
	if len(n.Tags) > 0 {
		return fmt.Sprintf("%s [%s] %s", n.Name, n.Prefix, n.Tags)
	}
	return fmt.Sprintf("%s [%s]", n.Name, n.Prefix)
}

func (n Network) Contains(d Device) bool {
	return n.Prefix.Contains(d.Addr)
}

func CompareNetwork(a Network, b Network) int {
	return ComparePrefix(a.Prefix, b.Prefix)
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
