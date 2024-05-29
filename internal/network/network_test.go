// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package network

import (
	"net"
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/networkables/mason/internal/device"
)

func TestNetwork_New(t *testing.T) {
	_, badPrefixErr := netip.ParsePrefix("192.168.1.1")
	tests := map[string]struct {
		name    string
		prefix  string
		want    Network
		wantErr error
	}{
		"Good": {
			name:    "aname",
			prefix:  "192.168.1.1/24",
			want:    Network{Name: "aname", Prefix: netip.MustParsePrefix("192.168.1.1/24")},
			wantErr: nil,
		},
		"EmptyName": {
			name:   "",
			prefix: "192.168.1.1/24",
			want: Network{
				Name:   "192.168.1.1/24",
				Prefix: netip.MustParsePrefix("192.168.1.1/24"),
			},
			wantErr: nil,
		},
		"BadPrefix": {
			name:    "",
			prefix:  "192.168.1.1",
			want:    Network{},
			wantErr: badPrefixErr,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tc.name, tc.prefix)
			if (tc.wantErr == nil && err != nil) ||
				tc.wantErr != nil && err != nil && tc.wantErr.Error() != err.Error() {
				t.Errorf("err want: %v, got %v", tc.wantErr, err)
			}
			diff := cmp.Diff(
				tc.want,
				got,
				cmpopts.EquateComparable(netip.Prefix{}),
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

func TestNetwork_NewNetworkFromPrefix(t *testing.T) {
	tests := map[string]struct {
		prefix  netip.Prefix
		want    Network
		wantErr error
	}{
		"Good": {
			prefix: netip.MustParsePrefix("192.168.1.1/24"),
			want: Network{
				Name:   "192.168.1.1/24",
				Prefix: netip.MustParsePrefix("192.168.1.1/24"),
			},
			wantErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewNetworkFromPrefix(tc.prefix)
			diff := cmp.Diff(
				tc.want,
				got,
				cmpopts.EquateComparable(netip.Prefix{}),
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

func TestNetwork_String(t *testing.T) {
	tests := map[string]struct {
		nw   Network
		want string
	}{
		"Good": {
			nw:   NewNetworkFromPrefix(netip.MustParsePrefix("192.168.1.1/24")),
			want: "192.168.1.1/24 > 192.168.1.1/24 tags:[]",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.nw.String()
			diff := cmp.Diff(
				tc.want,
				got,
				cmpopts.EquateComparable(netip.Prefix{}),
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

func TestNetwork_Contains(t *testing.T) {
	tests := map[string]struct {
		nw   Network
		d    device.Device
		want bool
	}{
		"Good": {
			nw:   NewNetworkFromPrefix(netip.MustParsePrefix("192.168.1.1/24")),
			d:    device.Device{Addr: netip.MustParseAddr("192.168.1.1")},
			want: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.nw.Contains(tc.d)
			diff := cmp.Diff(
				tc.want,
				got,
				cmpopts.EquateComparable(netip.Prefix{}),
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

func TestFindFirstAddr(t *testing.T) {
	tests := map[string]struct {
		prefix string
		want   netip.Addr
	}{
		"Good": {
			prefix: "192.168.1.1/24",
			want:   netip.MustParseAddr("192.168.1.0"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			px := netip.MustParsePrefix(tc.prefix)
			got := findFirstAddr(px)
			if tc.want.Compare(got) != 0 {
				t.Errorf("%s err want: %v, got: %v", name, tc.want, got)
			}
		})
	}
}

func TestFindLastAddr(t *testing.T) {
	tests := map[string]struct {
		prefix string
		want   netip.Addr
	}{
		"Good": {
			prefix: "192.168.1.1/24",
			want:   netip.MustParseAddr("192.168.1.255"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			px := netip.MustParsePrefix(tc.prefix)
			got := findLastAddr(px)
			if tc.want.Compare(got) != 0 {
				t.Errorf("%s err want: %v, got: %v", name, tc.want, got)
			}
		})
	}
}

func TestSortNetworksByAddr(t *testing.T) {
	got := []Network{
		{Prefix: netip.MustParsePrefix("192.168.1.3/24")},
		{Prefix: netip.MustParsePrefix("192.168.1.2/24")},
		{Prefix: netip.MustParsePrefix("192.168.1.1/24")},
	}
	want := []Network{
		{Prefix: netip.MustParsePrefix("192.168.1.1/24")},
		{Prefix: netip.MustParsePrefix("192.168.1.2/24")},
		{Prefix: netip.MustParsePrefix("192.168.1.3/24")},
	}
	SortNetworksByAddr(got)

	diff := cmp.Diff(
		want,
		got,
		cmpopts.EquateComparable(netip.Addr{}),
		cmpopts.IgnoreUnexported(Network{}, netip.Prefix{}),
	)
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestIsUsefulInterface(t *testing.T) {
	tests := map[string]struct {
		iface net.Interface
		want  bool
	}{
		"FailForLoopback": {
			iface: net.Interface{
				Flags: net.FlagLoopback,
			},
			want: false,
		},
		"FailForDown": {
			iface: net.Interface{
				Flags: net.Flags(0),
			},
			want: false,
		},
		"FailForNotRunning": {
			iface: net.Interface{
				Flags: net.FlagUp,
			},
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsUsefulInterface(tc.iface)
			diff := cmp.Diff(
				tc.want,
				got,
				cmpopts.EquateComparable(netip.Prefix{}),
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

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
					Prefix: netip.MustParsePrefix("192.168.1.1/24"),
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
		wantAddr netip.Addr
		wantDone bool
	}{
		"Good": {
			ni: networkIterator{
				network: Network{
					Name:   "192.168.1.1/24",
					Prefix: netip.MustParsePrefix("192.168.1.1/24"),
				},
				current: netip.MustParseAddr("192.168.1.1"),
				first:   netip.MustParseAddr("192.168.1.0"),
				last:    netip.MustParseAddr("192.168.1.255"),
				Size:    256,
			},
			wantAddr: netip.MustParseAddr("192.168.1.2"),
			wantDone: false,
		},
		"IsDone": {
			ni: networkIterator{
				network: Network{
					Name:   "192.168.1.1/24",
					Prefix: netip.MustParsePrefix("192.168.1.1/24"),
				},
				current: netip.MustParseAddr("192.168.1.255"),
				first:   netip.MustParseAddr("192.168.1.0"),
				last:    netip.MustParseAddr("192.168.1.255"),
				Size:    256,
			},
			wantAddr: netip.MustParseAddr("192.168.1.255"),
			wantDone: true,
		},
		"NotInNetwork": {
			ni: networkIterator{
				network: Network{
					Name:   "192.168.1.1/24",
					Prefix: netip.MustParsePrefix("192.168.1.1/24"),
				},
				current: netip.MustParseAddr("192.168.2.255"),
				first:   netip.MustParseAddr("192.168.1.0"),
				last:    netip.MustParseAddr("192.168.1.255"),
				Size:    256,
			},
			wantAddr: netip.MustParseAddr("192.168.1.255"),
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
					Prefix: netip.MustParsePrefix("192.168.1.1/24"),
				},
				current: netip.MustParseAddr("192.168.1.100"),
				first:   netip.MustParseAddr("192.168.1.0"),
				last:    netip.MustParseAddr("192.168.1.255"),
				Size:    256,
			},
			want: networkIterator{
				network: Network{
					Name:   "192.168.1.1/24",
					Prefix: netip.MustParsePrefix("192.168.1.1/24"),
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
		Prefix: netip.MustParsePrefix("192.168.1.1/24"),
	}
	nic := NewNetworkIteratorAsChannel(nw)
	want := nic.ni.Size - 2 // remove broadcast and zero
	got := 0
	for range nic.C {
		got++
	}
	if want != got {
		t.Errorf("size mismatch (want:%d got:%d)", want, got)
	}
}
