// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"net"
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
			name:   "aname",
			prefix: "192.168.1.0/24",
			want: Network{
				Name:   "aname",
				Prefix: MustParsePrefix("192.168.1.0/24"),
			},
			wantErr: nil,
		},
		"EmptyName": {
			name:   "",
			prefix: "192.168.1.0/24",
			want: Network{
				Name:   "192.168.1.0/24",
				Prefix: MustParsePrefix("192.168.1.0/24"),
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
				Name:   "192.168.1.0/24",
				Prefix: MustParsePrefix("192.168.1.0/24"),
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
			want: "192.168.1.0/24 [192.168.1.0/24]",
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
		d    Device
		want bool
	}{
		"Good": {
			nw:   NewNetworkFromPrefix(netip.MustParsePrefix("192.168.1.1/24")),
			d:    Device{Addr: MustParseAddr("192.168.1.1")},
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

func TestSortNetworksByAddr(t *testing.T) {
	got := []Network{
		{Prefix: MustParsePrefix("192.168.1.3/24")},
		{Prefix: MustParsePrefix("192.168.1.2/24")},
		{Prefix: MustParsePrefix("192.168.1.1/24")},
	}
	want := []Network{
		{Prefix: MustParsePrefix("192.168.1.1/24")},
		{Prefix: MustParsePrefix("192.168.1.2/24")},
		{Prefix: MustParsePrefix("192.168.1.3/24")},
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
