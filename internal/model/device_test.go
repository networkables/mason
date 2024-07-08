// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/networkables/mason/nettools"
)

func TestDevice_IsUpdated(t *testing.T) {
	type test struct {
		input Device
		want  bool
	}

	tests := []test{
		{input: Device{}, want: false},
		{input: Device{updated: true}, want: true},
	}

	for _, tc := range tests {
		got := tc.input.IsUpdated()
		if got != tc.want {
			t.Errorf("got: %t, want: %t", got, tc.want)
		}
	}
}

func TestDevice_SetUpdated(t *testing.T) {
	d := Device{}
	d.SetUpdated()
	want := true
	got := d.updated

	if got != want {
		t.Errorf("got: %t, want: %t", got, want)
	}
}

func TestDevice_toString(t *testing.T) {
	mac, _ := net.ParseMAC("00:00:5e:00:53:01")
	firstseen, _ := time.Parse(time.DateOnly, "2011-01-01")
	lastseen, _ := time.Parse(time.DateOnly, "2022-02-02")

	f := func(ts time.Time) time.Duration {
		return lastseen.Add(time.Hour).Sub(ts)
	}

	tests := map[string]struct {
		input Device
		want  string
	}{
		"empty": {
			input: Device{},
			want:  "      invalid IP      [                  <>] fs:never ls:never tags:[]",
		},
		"filled": {
			input: Device{
				Name:            "name",
				Addr:            MustParseAddr("192.168.1.1"),
				MAC:             HardwareAddrToMAC(mac),
				Meta:            Meta{Manufacturer: "A Company"},
				PerformancePing: Pinger{FirstSeen: firstseen, LastSeen: lastseen},
			},
			want: " name 192.168.1.1     [00:00:5e:00:53:01 <A Company>] fs:2011-01-01 00:00:00 ls:1h0m0s tags:[]",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.input.toString(f)
			diff := cmp.Diff(
				tc.want,
				got,
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

func TestDevice_DiscoveredAtString(t *testing.T) {
	firstseen, _ := time.Parse(time.DateOnly, "2011-01-01")
	d := Device{DiscoveredAt: firstseen}
	want := "2011-01-01"
	got := d.DiscoveredAtString()

	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

func TestDevice_LastSeenString(t *testing.T) {
	lastseen, _ := time.Parse(time.DateOnly, "2011-01-01")
	d := Device{PerformancePing: Pinger{LastSeen: lastseen}}
	want := "2011-01-01 00:00:00"
	got := d.LastSeenString()

	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

func TestDevice_MeanString(t *testing.T) {
	firstseen, _ := time.Parse(time.DateOnly, "2011-01-01")
	tests := map[string]struct {
		input Device
		want  string
	}{
		"Empty":     {input: Device{}, want: ""},
		"Zero Time": {input: Device{PerformancePing: Pinger{FirstSeen: firstseen}}, want: "0s"},
		"Millisecond": {
			input: Device{PerformancePing: Pinger{FirstSeen: firstseen, Mean: time.Millisecond}},
			want:  "1ms",
		},
		"10 microseconds": {
			input: Device{
				PerformancePing: Pinger{FirstSeen: firstseen, Mean: 10 * time.Microsecond},
			},
			want: "0s",
		},
		"150 microseconds": {
			input: Device{
				PerformancePing: Pinger{FirstSeen: firstseen, Mean: 150 * time.Microsecond},
			},
			want: "150µs",
		},
		"failure": {
			input: Device{PerformancePing: Pinger{FirstSeen: firstseen, LastFailed: true}},
			want:  "failure",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.input.LastPingMeanString()
			diff := cmp.Diff(
				tc.want,
				got,
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

func TestDevice_MaximumString(t *testing.T) {
	firstseen, _ := time.Parse(time.DateOnly, "2011-01-01")
	tests := map[string]struct {
		input Device
		want  string
	}{
		"Empty":     {input: Device{}, want: ""},
		"Zero Time": {input: Device{PerformancePing: Pinger{FirstSeen: firstseen}}, want: "0s"},
		"Millisecond": {
			input: Device{PerformancePing: Pinger{FirstSeen: firstseen, Maximum: time.Millisecond}},
			want:  "1ms",
		},
		"10 microseconds": {
			input: Device{
				PerformancePing: Pinger{FirstSeen: firstseen, Maximum: 10 * time.Microsecond},
			},
			want: "0s",
		},
		"150 microseconds": {
			input: Device{
				PerformancePing: Pinger{FirstSeen: firstseen, Maximum: 150 * time.Microsecond},
			},
			want: "150µs",
		},
		"failure": {
			input: Device{PerformancePing: Pinger{FirstSeen: firstseen, LastFailed: true}},
			want:  "failure",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.input.LastPingMaximumString()
			diff := cmp.Diff(
				tc.want,
				got,
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

func TestDevice_IsServer(t *testing.T) {
	tests := map[string]struct {
		input Device
		want  bool
	}{
		"NotServer": {input: Device{}, want: false},
		"ServerOpenPorts": {
			input: Device{Server: Server{Ports: IntSliceToPortList([]int{1, 2, 3})}},
			want:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.input.IsServer()
			diff := cmp.Diff(
				tc.want,
				got,
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

func TestSortDevicesByAddr(t *testing.T) {
	got := []Device{
		{Addr: MustParseAddr("192.168.1.3")},
		{Addr: MustParseAddr("192.168.1.2")},
		{Addr: MustParseAddr("192.168.1.1")},
	}
	want := []Device{
		{Addr: MustParseAddr("192.168.1.1")},
		{Addr: MustParseAddr("192.168.1.2")},
		{Addr: MustParseAddr("192.168.1.3")},
	}
	SortDevicesByAddr(got)

	diff := cmp.Diff(
		want,
		got,
		cmpopts.EquateComparable(netip.Addr{}),
		cmpopts.IgnoreUnexported(Device{}),
	)
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestDevice_UpdateFromPingStats(t *testing.T) {
	ts, _ := time.Parse(time.DateOnly, "2011-01-01")
	tests := map[string]struct {
		got  Device
		pre  nettools.Icmp4EchoResponseStatistics
		want Device
	}{
		"GoodPing": {
			got: Device{},
			want: Device{PerformancePing: Pinger{
				FirstSeen: ts, LastSeen: ts, Mean: time.Microsecond, Maximum: time.Millisecond,
			}},
			pre: nettools.Icmp4EchoResponseStatistics{
				SuccessCount: 1,
				Mean:         time.Microsecond,
				Maximum:      time.Millisecond,
			},
		},
		"PingFailed": {got: Device{}, want: Device{PerformancePing: Pinger{LastFailed: true}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.got.UpdateFromPingStats(tc.pre, ts)
			diff := cmp.Diff(
				tc.want,
				tc.got,
				cmpopts.EquateComparable(netip.Addr{}),
				cmpopts.IgnoreUnexported(Device{}),
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

func TestDevice_MetaMerge(t *testing.T) {
	tests := map[string]struct {
		starting    Meta
		in          Meta
		want        Meta
		wantUpdated bool
	}{
		"Empty": {
			starting:    Meta{},
			in:          Meta{},
			want:        Meta{},
			wantUpdated: false,
		},
		"Updated": {
			starting: Meta{},
			in: Meta{
				DnsName:      "dns1",
				Manufacturer: "company1",
				Tags:         Tags{Tag{Val: "tag1"}},
			},
			want: Meta{
				DnsName:      "dns1",
				Manufacturer: "company1",
				Tags:         Tags{Tag{Val: "tag1"}},
			},
			wantUpdated: true,
		},
		"NoUpdate": {
			starting: Meta{
				DnsName:      "dns1",
				Manufacturer: "company1",
				Tags:         Tags{Tag{Val: "tag1"}},
			},
			in: Meta{
				DnsName:      "dns1",
				Manufacturer: "company1",
				Tags:         Tags{Tag{Val: "tag1"}},
			},
			want: Meta{
				DnsName:      "dns1",
				Manufacturer: "company1",
				Tags:         Tags{Tag{Val: "tag1"}},
			},
			wantUpdated: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotUpdated := tc.starting.merge(tc.in)
			diff := cmp.Diff(
				tc.want,
				got,
				cmpopts.EquateComparable(netip.Addr{}),
				cmpopts.IgnoreUnexported(Device{}),
			)
			if diff != "" {
				t.Errorf("%s device mismatch (-want +got):\n%s", name, diff)
			}
			if tc.wantUpdated != gotUpdated {
				t.Errorf("%s updated mismatch (want:%t got:%t)", name, tc.wantUpdated, gotUpdated)
			}
		})
	}
}

func TestDevice_ServerMerge(t *testing.T) {
	ts, _ := time.Parse(time.DateOnly, "2011-01-01")
	tests := map[string]struct {
		starting    Server
		in          Server
		want        Server
		wantUpdated bool
	}{
		"Empty": {
			starting:    Server{},
			in:          Server{},
			want:        Server{},
			wantUpdated: false,
		},
		"Updated": {
			starting: Server{},
			in: Server{
				Ports:    IntSliceToPortList([]int{1, 2, 3}),
				LastScan: ts,
			},
			want: Server{
				Ports:    IntSliceToPortList([]int{1, 2, 3}),
				LastScan: ts,
			},
			wantUpdated: true,
		},
		"NoUpdate": {
			starting: Server{
				Ports:    IntSliceToPortList([]int{1, 2, 3}),
				LastScan: ts,
			},
			in: Server{
				Ports:    IntSliceToPortList([]int{1, 2, 3}),
				LastScan: ts,
			},
			want: Server{
				Ports:    IntSliceToPortList([]int{1, 2, 3}),
				LastScan: ts,
			},
			wantUpdated: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotUpdated := tc.starting.merge(tc.in)
			diff := cmp.Diff(
				tc.want,
				got,
				cmpopts.EquateComparable(netip.Addr{}),
				cmpopts.IgnoreUnexported(Device{}),
			)
			if diff != "" {
				t.Errorf("%s device mismatch (-want +got):\n%s", name, diff)
			}
			if tc.wantUpdated != gotUpdated {
				t.Errorf("%s updated mismatch (want:%t got:%t)", name, tc.wantUpdated, gotUpdated)
			}
		})
	}
}

func TestDevice_PingerMerge(t *testing.T) {
	ts1, _ := time.Parse(time.DateOnly, "2011-01-01")
	ts2, _ := time.Parse(time.DateOnly, "2022-02-02")
	d1 := time.Microsecond
	d2 := time.Millisecond
	tests := map[string]struct {
		starting    Pinger
		in          Pinger
		want        Pinger
		wantUpdated bool
	}{
		"Empty": {
			starting:    Pinger{},
			in:          Pinger{},
			want:        Pinger{},
			wantUpdated: false,
		},
		"Updated": {
			starting: Pinger{},
			in: Pinger{
				FirstSeen:  ts1,
				LastSeen:   ts2,
				Mean:       d1,
				Maximum:    d2,
				LastFailed: true,
			},
			want: Pinger{
				FirstSeen:  ts1,
				LastSeen:   ts2,
				Mean:       d1,
				Maximum:    d2,
				LastFailed: true,
			},
			wantUpdated: true,
		},
		"NoUpdate": {
			starting: Pinger{
				FirstSeen:  ts1,
				LastSeen:   ts2,
				Mean:       d1,
				Maximum:    d2,
				LastFailed: true,
			},
			in: Pinger{
				FirstSeen:  ts1,
				LastSeen:   ts2,
				Mean:       d1,
				Maximum:    d2,
				LastFailed: true,
			},
			want: Pinger{
				FirstSeen:  ts1,
				LastSeen:   ts2,
				Mean:       d1,
				Maximum:    d2,
				LastFailed: true,
			},
			wantUpdated: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotUpdated := tc.starting.merge(tc.in)
			diff := cmp.Diff(
				tc.want,
				got,
				cmpopts.EquateComparable(netip.Addr{}),
				cmpopts.IgnoreUnexported(Device{}),
			)
			if diff != "" {
				t.Errorf("%s device mismatch (-want +got):\n%s", name, diff)
			}
			if tc.wantUpdated != gotUpdated {
				t.Errorf("%s updated mismatch (want:%t got:%t)", name, tc.wantUpdated, gotUpdated)
			}
		})
	}
}

func TestDevice_SNMPMerge(t *testing.T) {
	ts1, _ := time.Parse(time.DateOnly, "2011-01-01")
	ts2, _ := time.Parse(time.DateOnly, "2022-02-02")
	ts3, _ := time.Parse(time.DateOnly, "2033-03-03")
	tests := map[string]struct {
		starting    SNMP
		in          SNMP
		want        SNMP
		wantUpdated bool
	}{
		"Empty": {},
		"Updated": {
			starting: SNMP{},
			in: SNMP{
				Name:               "aname",
				Description:        "adesc",
				Community:          "acommunity",
				Port:               161,
				LastSNMPCheck:      ts1,
				HasArpTable:        true,
				LastArpTableScan:   ts2,
				HasInterfaces:      true,
				LastInterfacesScan: ts3,
			},
			want: SNMP{
				Name:               "aname",
				Description:        "adesc",
				Community:          "acommunity",
				Port:               161,
				LastSNMPCheck:      ts1,
				HasArpTable:        true,
				LastArpTableScan:   ts2,
				HasInterfaces:      true,
				LastInterfacesScan: ts3,
			},
			wantUpdated: true,
		},
		"NoUpdate": {
			starting: SNMP{
				Name:               "aname",
				Description:        "adesc",
				Community:          "acommunity",
				Port:               161,
				LastSNMPCheck:      ts1,
				HasArpTable:        true,
				LastArpTableScan:   ts2,
				HasInterfaces:      true,
				LastInterfacesScan: ts3,
			},
			in: SNMP{
				Name:               "aname",
				Description:        "adesc",
				Community:          "acommunity",
				Port:               161,
				LastSNMPCheck:      ts1,
				HasArpTable:        true,
				LastArpTableScan:   ts2,
				HasInterfaces:      true,
				LastInterfacesScan: ts3,
			},
			want: SNMP{
				Name:               "aname",
				Description:        "adesc",
				Community:          "acommunity",
				Port:               161,
				LastSNMPCheck:      ts1,
				HasArpTable:        true,
				LastArpTableScan:   ts2,
				HasInterfaces:      true,
				LastInterfacesScan: ts3,
			},
			wantUpdated: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotUpdated := tc.starting.merge(tc.in)
			diff := cmp.Diff(
				tc.want,
				got,
				cmpopts.EquateComparable(netip.Addr{}),
				cmpopts.IgnoreUnexported(Device{}),
			)
			if diff != "" {
				t.Errorf("%s device mismatch (-want +got):\n%s", name, diff)
			}
			if tc.wantUpdated != gotUpdated {
				t.Errorf("%s updated mismatch (want:%t got:%t)", name, tc.wantUpdated, gotUpdated)
			}
		})
	}
}

func TestDevice_BaseMerge(t *testing.T) {
	addr := MustParseAddr("192.168.1.1")
	hwaddr, _ := net.ParseMAC("00:00:5e:00:53:01")
	mac := HardwareAddrToMAC(hwaddr)
	ts, _ := time.Parse(time.DateOnly, "2011-01-01")
	src := DiscoverySource("TEST")
	tests := map[string]struct {
		starting    Device
		in          Device
		want        Device
		wantUpdated bool
	}{
		"Empty": {
			starting:    Device{},
			in:          Device{},
			want:        Device{},
			wantUpdated: false,
		},
		"Updated": {
			starting: Device{},
			in: Device{
				Name:         "aname",
				Addr:         addr,
				MAC:          mac,
				DiscoveredAt: ts,
				DiscoveredBy: src,
			},
			want: Device{
				Name:         "aname",
				Addr:         addr,
				MAC:          mac,
				DiscoveredAt: ts,
				DiscoveredBy: src,
			},
			wantUpdated: true,
		},
		"NoUpdate": {
			starting: Device{
				Name:         "aname",
				Addr:         addr,
				MAC:          mac,
				DiscoveredAt: ts,
				DiscoveredBy: src,
			},
			in: Device{
				Name:         "aname",
				Addr:         addr,
				MAC:          mac,
				DiscoveredAt: ts,
				DiscoveredBy: src,
			},
			want: Device{
				Name:         "aname",
				Addr:         addr,
				MAC:          mac,
				DiscoveredAt: ts,
				DiscoveredBy: src,
			},
			wantUpdated: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotUpdated := tc.starting.merge(tc.in)
			diff := cmp.Diff(
				tc.want,
				got,
				cmpopts.EquateComparable(netip.Addr{}),
				cmpopts.IgnoreUnexported(Device{}),
			)
			if diff != "" {
				t.Errorf("%s device mismatch (-want +got):\n%s", name, diff)
			}
			if tc.wantUpdated != gotUpdated {
				t.Errorf("%s updated mismatch (want:%t got:%t)", name, tc.wantUpdated, gotUpdated)
			}
		})
	}
}

func TestDevice_Merge(t *testing.T) {
	addr := MustParseAddr("192.168.1.1")
	hwaddr, _ := net.ParseMAC("00:00:5e:00:53:01")
	mac := HardwareAddrToMAC(hwaddr)
	ts, _ := time.Parse(time.DateOnly, "2011-01-01")
	src := DiscoverySource("TEST")
	d := time.Minute
	tests := map[string]struct {
		starting    Device
		in          Device
		want        Device
		wantUpdated bool
	}{
		"Empty": {
			starting: Device{},
			in:       Device{},
			want: Device{
				Name: "invalid IP",
			},
			wantUpdated: false,
		},
		"Updated": {
			starting: Device{},
			in: Device{
				Name:         "aname",
				Addr:         addr,
				MAC:          mac,
				DiscoveredAt: ts,
				DiscoveredBy: src,
				Meta: Meta{
					DnsName:      "dns1",
					Manufacturer: "company1",
					Tags:         Tags{Tag{Val: "tag1"}},
				},
				Server: Server{
					Ports:    IntSliceToPortList([]int{1, 2, 3}),
					LastScan: ts,
				},
				PerformancePing: Pinger{
					FirstSeen:  ts,
					LastSeen:   ts,
					Mean:       d,
					Maximum:    d,
					LastFailed: true,
				},
				SNMP: SNMP{
					Name:               "aname",
					Description:        "adesc",
					Community:          "acommunity",
					Port:               161,
					LastSNMPCheck:      ts,
					HasArpTable:        true,
					LastArpTableScan:   ts,
					HasInterfaces:      true,
					LastInterfacesScan: ts,
				},
			},
			want: Device{
				Name:         "aname",
				Addr:         addr,
				MAC:          mac,
				DiscoveredAt: ts,
				DiscoveredBy: src,
				Meta: Meta{
					DnsName:      "dns1",
					Manufacturer: "company1",
					Tags:         Tags{Tag{Val: "tag1"}},
				},
				Server: Server{
					Ports:    IntSliceToPortList([]int{1, 2, 3}),
					LastScan: ts,
				},
				PerformancePing: Pinger{
					FirstSeen:  ts,
					LastSeen:   ts,
					Mean:       d,
					Maximum:    d,
					LastFailed: true,
				},
				SNMP: SNMP{
					Name:               "aname",
					Description:        "adesc",
					Community:          "acommunity",
					Port:               161,
					LastSNMPCheck:      ts,
					HasArpTable:        true,
					LastArpTableScan:   ts,
					HasInterfaces:      true,
					LastInterfacesScan: ts,
				},
			},
			wantUpdated: true,
		},
		"NoUpdate": {
			starting: Device{
				Name:         "aname",
				Addr:         addr,
				MAC:          mac,
				DiscoveredAt: ts,
				DiscoveredBy: src,
				Meta: Meta{
					DnsName:      "dns1",
					Manufacturer: "company1",
					Tags:         Tags{Tag{Val: "tag1"}},
				},
				Server: Server{
					Ports:    IntSliceToPortList([]int{1, 2, 3}),
					LastScan: ts,
				},
				PerformancePing: Pinger{
					FirstSeen:  ts,
					LastSeen:   ts,
					Mean:       d,
					Maximum:    d,
					LastFailed: true,
				},
				SNMP: SNMP{
					Name:               "aname",
					Description:        "adesc",
					Community:          "acommunity",
					Port:               161,
					LastSNMPCheck:      ts,
					HasArpTable:        true,
					LastArpTableScan:   ts,
					HasInterfaces:      true,
					LastInterfacesScan: ts,
				},
			},
			in: Device{
				Name:         "aname",
				Addr:         addr,
				MAC:          mac,
				DiscoveredAt: ts,
				DiscoveredBy: src,
				Meta: Meta{
					DnsName:      "dns1",
					Manufacturer: "company1",
					Tags:         Tags{Tag{Val: "tag1"}},
				},
				Server: Server{
					Ports:    IntSliceToPortList([]int{1, 2, 3}),
					LastScan: ts,
				},
				PerformancePing: Pinger{
					FirstSeen:  ts,
					LastSeen:   ts,
					Mean:       d,
					Maximum:    d,
					LastFailed: true,
				},
				SNMP: SNMP{
					Name:               "aname",
					Description:        "adesc",
					Community:          "acommunity",
					Port:               161,
					LastSNMPCheck:      ts,
					HasArpTable:        true,
					LastArpTableScan:   ts,
					HasInterfaces:      true,
					LastInterfacesScan: ts,
				},
			},
			want: Device{
				Name:         "aname",
				Addr:         addr,
				MAC:          mac,
				DiscoveredAt: ts,
				DiscoveredBy: src,
				Meta: Meta{
					DnsName:      "dns1",
					Manufacturer: "company1",
					Tags:         Tags{Tag{Val: "tag1"}},
				},
				Server: Server{
					Ports:    IntSliceToPortList([]int{1, 2, 3}),
					LastScan: ts,
				},
				PerformancePing: Pinger{
					FirstSeen:  ts,
					LastSeen:   ts,
					Mean:       d,
					Maximum:    d,
					LastFailed: true,
				},
				SNMP: SNMP{
					Name:               "aname",
					Description:        "adesc",
					Community:          "acommunity",
					Port:               161,
					LastSNMPCheck:      ts,
					HasArpTable:        true,
					LastArpTableScan:   ts,
					HasInterfaces:      true,
					LastInterfacesScan: ts,
				},
			},
			wantUpdated: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.starting.Merge(tc.in)
			diff := cmp.Diff(
				tc.want,
				got,
				cmpopts.EquateComparable(netip.Addr{}),
				cmpopts.IgnoreUnexported(Device{}),
			)
			if diff != "" {
				t.Errorf("%s device mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}
