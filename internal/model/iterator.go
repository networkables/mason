// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"net/netip"

	"go4.org/netipx"
)

// networkIterator
type networkIterator struct {
	network Network
	current netip.Addr
	first   netip.Addr
	last    netip.Addr
	Size    int
}

func NewNetworkIterator(n Network) *networkIterator {
	rng := netipx.RangeOfPrefix(n.Prefix.P)

	first := rng.From()
	last := rng.To()
	curr := first
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

func (ni *networkIterator) Next() (addr Addr, done bool) {
	ip := ni.current
	ni.current = ni.current.Next()
	if ip.Compare(ni.last) == 0 {
		return AddrToModelAddr(ip), true
	}
	if !ni.network.Prefix.ContainsAddr(ip) {
		return AddrToModelAddr(ip), true
	}
	return AddrToModelAddr(ip), false
}

func (ni *networkIterator) Reset() {
	ni.current = ni.network.Prefix.Addr()
}

// NetworkIteratorAsChannel
type networkIteratorAsChannel struct {
	ni *networkIterator
	C  chan Addr
}

func NewNetworkIteratorAsChannel(n Network) *networkIteratorAsChannel {
	ch := make(chan Addr)
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
