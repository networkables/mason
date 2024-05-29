// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package device

import (
	"net"
	"net/netip"
	"testing"
)

func TestEventDeviceDiscoveredString(t *testing.T) {
	mac, _ := net.ParseMAC("00:00:5e:00:53:01")
	d := EventDeviceDiscovered{
		Name:         "aname",
		MAC:          mac,
		Addr:         netip.MustParseAddr("192.168.1.1"),
		DiscoveredBy: DiscoverySource("TEST"),
	}
	want := "TEST 192.168.1.1"
	got := d.String()

	if got != want {
		t.Errorf("expected: %s, got: %s", want, got)
	}
}

func TestEventDeviceAddedString(t *testing.T) {
	mac, _ := net.ParseMAC("00:00:5e:00:53:01")
	d := EventDeviceAdded{
		Name:         "aname",
		MAC:          mac,
		Addr:         netip.MustParseAddr("192.168.1.1"),
		DiscoveredBy: DiscoverySource("TEST"),
	}
	want := "192.168.1.1"
	got := d.String()

	if got != want {
		t.Errorf("expected: %s, got: %s", want, got)
	}
}

func TestEventDeviceUpdatedString(t *testing.T) {
	mac, _ := net.ParseMAC("00:00:5e:00:53:01")
	d := EventDeviceUpdated{
		Name:         "aname",
		MAC:          mac,
		Addr:         netip.MustParseAddr("192.168.1.1"),
		DiscoveredBy: DiscoverySource("TEST"),
	}
	want := "aname [192.168.1.1 00:00:5e:00:53:01]"
	got := d.String()

	if got != want {
		t.Errorf("expected: %s, got: %s", want, got)
	}
}
