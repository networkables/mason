// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/networkables/mason/internal/discovery"
	"github.com/networkables/mason/internal/model"
)

func devicesCmp(t *testing.T, want []model.Device, got []model.Device) string {
	t.Helper()
	return cmp.Diff(
		want,
		got,
		cmpopts.EquateComparable(netip.Prefix{}, netip.Addr{}),
		cmpopts.IgnoreUnexported(model.Device{}),
	)
}

func deviceCmp(t *testing.T, want model.Device, got model.Device) string {
	t.Helper()
	return cmp.Diff(
		want,
		got,
		cmpopts.EquateComparable(netip.Prefix{}, netip.Addr{}),
		cmpopts.IgnoreUnexported(model.Device{}),
	)
}

func TestSqliteStore_AddDevice(t *testing.T) {
	var err error
	ctx := context.Background()
	type test struct {
		input model.Device
		want  []model.Device
	}

	ts, err := time.Parse(time.RFC3339Nano, "2024-12-11T22:21:20.012345678-00:00")
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]test{
		"basic": {
			input: model.Device{
				Name:         "basic",
				Addr:         model.MustParseAddr("192.168.0.1"),
				DiscoveredAt: ts,
			},
			want: []model.Device{
				{
					Name:         "basic",
					Addr:         model.MustParseAddr("192.168.0.1"),
					DiscoveredAt: ts,
					Meta:         model.Meta{Tags: model.Tags{}},
				},
			},
		},
		"fullmodel": {
			input: model.Device{
				Name:         "allmodel",
				Addr:         model.MustParseAddr("1.2.3.4"),
				MAC:          model.MustParseMAC("a0:55:99:4b:1f:e2"),
				DiscoveredAt: ts,
				DiscoveredBy: discovery.ArpDiscoverySource,
				Meta: model.Meta{
					DnsName:      "allmodel.dns",
					Manufacturer: "Acme Inc",
					Tags:         model.Tags{model.RandomizedMacAddressTag},
				},
				Server: model.Server{
					Ports:    model.PortList{Ports: []int{1, 2, 3, 4}},
					LastScan: ts,
				},
				PerformancePing: model.Pinger{
					FirstSeen:  ts,
					LastSeen:   ts,
					Mean:       time.Minute,
					Maximum:    time.Hour,
					LastFailed: true,
				},
				SNMP: model.SNMP{
					Name:               "The Snmp Name",
					Description:        "A Snmp Description",
					Community:          "community string",
					Port:               161,
					LastSNMPCheck:      ts,
					HasArpTable:        true,
					LastArpTableScan:   ts,
					HasInterfaces:      true,
					LastInterfacesScan: ts,
				},
			},
			want: []model.Device{{
				Name:         "allmodel",
				Addr:         model.MustParseAddr("1.2.3.4"),
				MAC:          model.MustParseMAC("a0:55:99:4b:1f:e2"),
				DiscoveredAt: ts,
				DiscoveredBy: discovery.ArpDiscoverySource,
				Meta: model.Meta{
					DnsName:      "allmodel.dns",
					Manufacturer: "Acme Inc",
					Tags:         model.Tags{model.RandomizedMacAddressTag},
				},
				Server: model.Server{
					Ports:    model.PortList{Ports: []int{1, 2, 3, 4}},
					LastScan: ts,
				},
				PerformancePing: model.Pinger{
					FirstSeen:  ts,
					LastSeen:   ts,
					Mean:       time.Minute,
					Maximum:    time.Hour,
					LastFailed: true,
				},
				SNMP: model.SNMP{
					Name:               "The Snmp Name",
					Description:        "A Snmp Description",
					Community:          "community string",
					Port:               161,
					LastSNMPCheck:      ts,
					HasArpTable:        true,
					LastArpTableScan:   ts,
					HasInterfaces:      true,
					LastInterfacesScan: ts,
				},
			}},
		},
	}

	for name, tc := range tests {
		db := createTestDatabase(t)
		err = db.AddDevice(ctx, tc.input)
		if err != nil {
			t.Fatal(err)
		}
		got, err := db.selectDevices(ctx)
		if err != nil {
			t.Fatal(err)
		}
		diff := devicesCmp(t, tc.want, got)
		if diff != "" {
			t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
		}
	}
}

func TestSqliteStore_AddExistingDevice(t *testing.T) {
	var err error
	ctx := context.Background()
	type test struct {
		input model.Device
		want  []model.Device
		err   error
	}

	ts, err := time.Parse(time.RFC3339Nano, "2024-12-11T22:21:20.012345678-00:00")
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]test{
		"erroralreadyexists": {
			input: model.Device{
				Name:         "basic",
				Addr:         model.MustParseAddr("192.168.0.1"),
				DiscoveredAt: ts,
			},
			err:  model.ErrDeviceExists,
			want: []model.Device{},
		},
	}

	db := createTestDatabase(t)
	err = db.AddDevice(ctx,
		model.Device{
			Name:         "basic",
			Addr:         model.MustParseAddr("192.168.0.1"),
			DiscoveredAt: ts,
		})

	for name, tc := range tests {
		err = db.AddDevice(ctx, tc.input)
		if err != nil {
			if tc.err == nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(tc.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Errorf("%s error mismatch (-want +got):\n%s", name, diff)
			}
		} else {
			if tc.err != nil {
				t.Fatalf("%s did not encounter expected error: %v", name, tc.err)
			}
			got, err := db.selectDevices(ctx)
			if err != nil {
				t.Fatal(err)
			}
			diff := devicesCmp(t, tc.want, got)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		}
	}
}

func TestSqliteStore_UpdateDevice(t *testing.T) {
	var err error
	ctx := context.Background()
	type test struct {
		input model.Device
		want  []model.Device
		err   error
	}

	ts, err := time.Parse(time.RFC3339Nano, "2024-12-11T22:21:20.012345678-00:00")
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]test{
		"basic": {
			input: model.Device{
				Name:         "updated",
				Addr:         model.MustParseAddr("192.168.0.1"),
				DiscoveredAt: ts,
			},
			want: []model.Device{
				{
					Name:         "updated",
					Addr:         model.MustParseAddr("192.168.0.1"),
					DiscoveredAt: ts,
					Meta:         model.Meta{Tags: model.Tags{}},
				},
			},
		},
		"errordoesnotexist": {
			input: model.Device{
				Name:         "nonet",
				Addr:         model.MustParseAddr("192.168.0.2"),
				DiscoveredAt: ts,
			},
			err:  model.ErrDeviceDoesNotExist,
			want: []model.Device{},
		},
	}

	db := createTestDatabase(t)

	err = db.AddDevice(ctx,
		model.Device{
			Name:         "initial",
			Addr:         model.MustParseAddr("192.168.0.1"),
			DiscoveredAt: ts,
		})
	if err != nil {
		t.Fatal(err)
	}

	for name, tc := range tests {
		tc.input.SetUpdated()
		_, err = db.UpdateDevice(ctx, tc.input)
		if err != nil {
			if tc.err == nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(tc.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Errorf("%s error mismatch (-want +got):\n%s", name, diff)
			}
		} else {
			if tc.err != nil {
				t.Fatalf("%s did not encounter expected error: %v", name, tc.err)
			}
			got, err := db.selectDevices(ctx)
			if err != nil {
				t.Fatal(err)
			}
			diff := devicesCmp(t, tc.want, got)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		}
	}
}

func TestSqliteStore_UpdateDeviceSkipNonUpdated(t *testing.T) {
	var err error
	ctx := context.Background()

	db := createTestDatabase(t)

	err = db.AddDevice(ctx,
		model.Device{
			Name: "initial",
			Addr: model.MustParseAddr("192.168.0.1"),
		})
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.UpdateDevice(ctx,
		model.Device{
			Name: "updated",
			Addr: model.MustParseAddr("192.168.0.1"),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	got, err := db.selectDevices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 device, have %d", len(got))
	}
	if got[0].Name != "initial" {
		t.Fatal("device was unexpectedly updated")
	}
}

func TestSqliteStore_GetFilteredDevices(t *testing.T) {
	ctx := context.Background()

	db := createTestDatabase(t)
	err := db.AddDevice(ctx, model.Device{
		Name: "dev0",
		Addr: model.MustParseAddr("192.168.0.0"),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AddDevice(ctx, model.Device{
		Name: "dev1",
		Addr: model.MustParseAddr("192.168.0.1"),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AddDevice(ctx, model.Device{
		Name: "dev2",
		Addr: model.MustParseAddr("192.168.0.2"),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []model.Device{{
		Name: "dev1",
		Addr: model.MustParseAddr("192.168.0.1"),
	}}

	got := db.GetFilteredDevices(ctx, func(d model.Device) bool {
		if d.Name == "dev1" {
			return true
		}
		return false
	})

	diff := devicesCmp(t, want, got)
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestSqliteStore_DeviceListAndCount(t *testing.T) {
	ctx := context.Background()

	db := createTestDatabase(t)
	err := db.AddDevice(ctx, model.Device{
		Name: "dev1",
		Addr: model.MustParseAddr("192.168.0.1"),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []model.Device{{
		Name: "dev1",
		Addr: model.MustParseAddr("192.168.0.1"),
	}}

	got := db.ListDevices(ctx)
	diff := devicesCmp(t, want, got)

	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	ct := db.CountDevices(ctx)
	if ct != 1 {
		t.Errorf("count mismatch got:%d want:1", ct)
	}
}

func TestSqliteStore_GetDeviceByAddr(t *testing.T) {
	ctx := context.Background()

	db := createTestDatabase(t)
	err := db.AddDevice(ctx, model.Device{
		Name: "basic",
		Addr: model.MustParseAddr("192.168.0.1"),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := model.Device{
		Name: "basic",
		Addr: model.MustParseAddr("192.168.0.1"),
	}

	got, err := db.GetDeviceByAddr(ctx, model.MustParseAddr("192.168.0.1"))
	if err != nil {
		t.Fatal(err)
	}
	diff := deviceCmp(t, want, got)
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	_, err = db.GetDeviceByAddr(ctx, model.MustParseAddr("192.168.100.1"))
	if err == nil {
		t.Errorf("did not get an error for get of unknown name")
	}
	wanterr := model.ErrDeviceDoesNotExist
	diff = cmp.Diff(wanterr, err, cmpopts.EquateErrors())
	if diff != "" {
		t.Fatalf("error mismatch (-want +got):\n%s", diff)
	}
}

func TestSqliteStore_RemoveDeviceByAddr(t *testing.T) {
	ctx := context.Background()

	db := createTestDatabase(t)
	err := db.AddDevice(ctx, model.Device{
		Name: "basic",
		Addr: model.MustParseAddr("192.168.0.1"),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.RemoveDeviceByAddr(ctx, model.MustParseAddr("192.168.0.1"))
	if err != nil {
		t.Fatal(err)
	}

	err = db.RemoveDeviceByAddr(ctx, model.MustParseAddr("192.168.100.1"))
	if err == nil {
		t.Errorf("did not get an error for get of unknown name")
	}
	wanterr := model.ErrDeviceDoesNotExist
	diff := cmp.Diff(wanterr, err, cmpopts.EquateErrors())
	if diff != "" {
		t.Fatalf("error mismatch (-want +got):\n%s", diff)
	}
}
