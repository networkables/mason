// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNetworkIterator_New(t *testing.T) {
	tests := map[string]struct {
		nw   Network
		want networkIterator
	}{
		"Good": {
			nw: NewNetworkFromPrefix(netip.MustParsePrefix("192.168.1.1/24")),
			want: networkIterator{
				network: Network{
					Name:   "192.168.1.1/24",
					Prefix: MustParsePrefix("192.168.1.1/24"),
				},
				current: netip.MustParseAddr("192.168.1.1"),
				first:   netip.MustParseAddr("192.168.1.0"),
				last:    netip.MustParseAddr("192.168.1.255"),
				Size:    256,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewNetworkIterator(tc.nw)
			diff := cmp.Diff(
				tc.want,
				*got,
				cmp.AllowUnexported(networkIterator{}),
				cmpopts.IgnoreUnexported(netip.Prefix{}, netip.Addr{}),
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

func TestNetworkIterator_Next(t *testing.T) {
	tests := map[string]struct {
		ni       networkIterator
		wantAddr Addr
		wantDone bool
	}{
		"Good": {
			ni: networkIterator{
				network: Network{
					Name:   "192.168.1.1/24",
					Prefix: MustParsePrefix("192.168.1.1/24"),
				},
				current: netip.MustParseAddr("192.168.1.1"),
				first:   netip.MustParseAddr("192.168.1.0"),
				last:    netip.MustParseAddr("192.168.1.255"),
				Size:    256,
			},
			wantAddr: MustParseAddr("192.168.1.2"),
			wantDone: false,
		},
		"IsDone": {
			ni: networkIterator{
				network: Network{
					Name:   "192.168.1.1/24",
					Prefix: MustParsePrefix("192.168.1.1/24"),
				},
				current: netip.MustParseAddr("192.168.1.255"),
				first:   netip.MustParseAddr("192.168.1.0"),
				last:    netip.MustParseAddr("192.168.1.255"),
				Size:    256,
			},
			wantAddr: MustParseAddr("192.168.1.255"),
			wantDone: true,
		},
		"NotInNetwork": {
			ni: networkIterator{
				network: Network{
					Name:   "192.168.1.1/24",
					Prefix: MustParsePrefix("192.168.1.1/24"),
				},
				current: netip.MustParseAddr("192.168.2.255"),
				first:   netip.MustParseAddr("192.168.1.0"),
				last:    netip.MustParseAddr("192.168.1.255"),
				Size:    256,
			},
			wantAddr: MustParseAddr("192.168.1.255"),
			wantDone: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotAddr, gotDone := tc.ni.Next()
			diff := cmp.Diff(
				tc.wantAddr,
				gotAddr,
				// cmpopts.EquateComparable(netip.Prefix{}),
				cmp.AllowUnexported(networkIterator{}),
				cmpopts.IgnoreUnexported(netip.Prefix{}, netip.Addr{}),
			)
			if diff != "" {
				t.Errorf("%s addr mismatch (-want +got):\n%s", name, diff)
			}
			if tc.wantDone != gotDone {
				t.Errorf("%s done mismatch (want:%t got:%t)", name, tc.wantDone, gotDone)
			}
		})
	}
}

func TestNetworkIterator_Reset(t *testing.T) {
	tests := map[string]struct {
		got  networkIterator
		want networkIterator
	}{
		"Good": {
			got: networkIterator{
				network: Network{
					Name:   "192.168.1.1/24",
					Prefix: MustParsePrefix("192.168.1.1/24"),
				},
				current: netip.MustParseAddr("192.168.1.100"),
				first:   netip.MustParseAddr("192.168.1.0"),
				last:    netip.MustParseAddr("192.168.1.255"),
				Size:    256,
			},
			want: networkIterator{
				network: Network{
					Name:   "192.168.1.1/24",
					Prefix: MustParsePrefix("192.168.1.1/24"),
				},
				current: netip.MustParseAddr("192.168.1.1"),
				first:   netip.MustParseAddr("192.168.1.0"),
				last:    netip.MustParseAddr("192.168.1.255"),
				Size:    256,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.got.Reset()
			diff := cmp.Diff(
				tc.want,
				tc.got,
				cmp.AllowUnexported(networkIterator{}),
				cmpopts.IgnoreUnexported(netip.Prefix{}, netip.Addr{}),
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

func TestNetworkIteratorAsChannel(t *testing.T) {
	nw := Network{
		Prefix: MustParsePrefix("192.168.1.1/24"),
	}
	nic := NewNetworkIteratorAsChannel(nw)
	want := nic.ni.Size // remove broadcast and zero
	got := 1
	for range nic.C {
		got++
	}
	if want != got {
		t.Errorf("size mismatch (want:%d got:%d)", want, got)
	}
}
