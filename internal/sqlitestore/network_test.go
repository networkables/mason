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

	"github.com/networkables/mason/internal/model"
)

func TestSqliteStore_UpsertNetwork(t *testing.T) {
	var err error
	ctx := context.Background()
	type test struct {
		input model.Network
		want  []model.Network
	}

	ts, err := time.Parse(time.RFC3339Nano, "2024-12-11T22:21:20.012345678-00:00")
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]test{
		"basic": {
			input: model.Network{
				Name:     "basic",
				Prefix:   model.MustParsePrefix("192.168.0.0/24"),
				LastScan: ts,
			},
			want: []model.Network{
				{
					Name:     "basic",
					Prefix:   model.MustParsePrefix("192.168.0.0/24"),
					LastScan: ts,
					Tags:     model.Tags{},
				},
			},
		},
		"update_in_place": {
			input: model.Network{
				Name:     "updated",
				Prefix:   model.MustParsePrefix("192.168.0.0/24"),
				LastScan: ts,
			},
			want: []model.Network{
				{
					Name:     "updated",
					Prefix:   model.MustParsePrefix("192.168.0.0/24"),
					LastScan: ts,
					Tags:     model.Tags{},
				},
			},
		},
	}

	db := createTestDatabase(t)
	defer func() {
		db.Close()
		removeTestDatabase(t)
	}()

	for name, tc := range tests {
		err = db.UpsertNetwork(ctx, tc.input)
		if err != nil {
			t.Fatal(err)
		}
		got, err := db.selectNetworks(ctx)
		if err != nil {
			t.Fatal(err)
		}
		diff := cmp.Diff(tc.want, got,
			cmpopts.EquateComparable(netip.Prefix{}),
		)
		if diff != "" {
			t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
		}
	}
}

func TestSqliteStore_AddNetwork(t *testing.T) {
	var err error
	ctx := context.Background()
	type test struct {
		input model.Network
		want  []model.Network
		err   error
	}

	ts, err := time.Parse(time.RFC3339Nano, "2024-12-11T22:21:20.012345678-00:00")
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]test{
		"basic": {
			input: model.Network{
				Name:     "basic",
				Prefix:   model.MustParsePrefix("192.168.0.0/24"),
				LastScan: ts,
			},
			want: []model.Network{
				{
					Name:     "basic",
					Prefix:   model.MustParsePrefix("192.168.0.0/24"),
					LastScan: ts,
					Tags:     model.Tags{},
				},
			},
		},
		"erroralreadyexists": {
			input: model.Network{
				Name:     "basic",
				Prefix:   model.MustParsePrefix("192.168.0.0/24"),
				LastScan: ts,
			},
			err:  model.ErrNetworkExists,
			want: []model.Network{},
		},
	}

	db := createTestDatabase(t)
	defer func() {
		db.Close()
		removeTestDatabase(t)
	}()

	for name, tc := range tests {
		err = db.AddNetwork(ctx, tc.input)
		if err != nil {
			if tc.err == nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(tc.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Fatalf("%s error mismatch (-want +got):\n%s", name, diff)
			}
		} else {
			if tc.err != nil {
				t.Fatalf("%s did not encounter expected error: %v", name, tc.err)
			}
			got, err := db.selectNetworks(ctx)
			if err != nil {
				t.Fatal(err)
			}
			diff := cmp.Diff(tc.want, got,
				cmpopts.EquateComparable(netip.Prefix{}),
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		}
	}
}

func TestSqliteStore_UpdateNetwork(t *testing.T) {
	var err error
	ctx := context.Background()
	type test struct {
		input model.Network
		want  []model.Network
		err   error
	}

	ts, err := time.Parse(time.RFC3339Nano, "2024-12-11T22:21:20.012345678-00:00")
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]test{
		"basic": {
			input: model.Network{
				Name:     "updated",
				Prefix:   model.MustParsePrefix("192.168.0.0/24"),
				LastScan: ts.Add(time.Hour),
			},
			want: []model.Network{
				{
					Name:     "updated",
					Prefix:   model.MustParsePrefix("192.168.0.0/24"),
					LastScan: ts.Add(time.Hour),
					Tags:     model.Tags{},
				},
			},
		},
		"errordoesnotexist": {
			input: model.Network{
				Name:     "nonet",
				Prefix:   model.MustParsePrefix("192.168.200.0/24"),
				LastScan: ts,
			},
			err:  model.ErrNetworkDoesNotExist,
			want: []model.Network{},
		},
	}

	db := createTestDatabase(t)
	defer func() {
		db.Close()
		removeTestDatabase(t)
	}()

	err = db.AddNetwork(ctx,
		model.Network{
			Name:     "initial",
			Prefix:   model.MustParsePrefix("192.168.0.0/24"),
			LastScan: ts,
		})
	if err != nil {
		t.Fatal(err)
	}

	for name, tc := range tests {
		err = db.UpdateNetwork(ctx, tc.input)
		if err != nil {
			if tc.err == nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(tc.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Fatalf("%s error mismatch (-want +got):\n%s", name, diff)
			}
			return
		} else {
			if tc.err != nil {
				t.Fatalf("%s did not encounter expected error: %v", name, tc.err)
			}
			got, err := db.selectNetworks(ctx)
			if err != nil {
				t.Fatal(err)
			}
			diff := cmp.Diff(tc.want, got,
				cmpopts.EquateComparable(netip.Prefix{}),
			)
			if diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
			}
		}
	}
}

func TestSqliteStore_GetFilteredNetworks(t *testing.T) {
	ctx := context.Background()

	db := createTestDatabase(t)
	defer func() {
		db.Close()
		removeTestDatabase(t)
	}()
	err := db.AddNetwork(ctx, model.Network{
		Name:   "net0",
		Prefix: model.MustParsePrefix("192.168.0.0/24"),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AddNetwork(ctx, model.Network{
		Name:   "net1",
		Prefix: model.MustParsePrefix("192.168.1.0/24"),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AddNetwork(ctx, model.Network{
		Name:   "net2",
		Prefix: model.MustParsePrefix("192.168.2.0/24"),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []model.Network{{
		Name:   "net1",
		Prefix: model.MustParsePrefix("192.168.1.0/24"),
	}}

	got := db.GetFilteredNetworks(ctx, func(n model.Network) bool {
		if n.Name == "net1" {
			return true
		}
		return false
	})

	diff := cmp.Diff(want, got,
		cmpopts.EquateComparable(netip.Prefix{}),
	)
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestSqliteStore_NetworkListAndCount(t *testing.T) {
	ctx := context.Background()

	db := createTestDatabase(t)
	defer func() {
		db.Close()
		removeTestDatabase(t)
	}()
	err := db.AddNetwork(ctx, model.Network{
		Name:   "net1",
		Prefix: model.MustParsePrefix("192.168.0.0/24"),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []model.Network{{
		Name:   "net1",
		Prefix: model.MustParsePrefix("192.168.0.0/24"),
	}}

	got := db.ListNetworks(ctx)
	diff := cmp.Diff(want, got,
		cmpopts.EquateComparable(netip.Prefix{}),
	)
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	ct := db.CountNetworks(ctx)
	if ct != 1 {
		t.Errorf("count mismatch got:%d want:1", ct)
	}
}

func TestSqliteStore_GetNetworkByName(t *testing.T) {
	ctx := context.Background()

	db := createTestDatabase(t)
	defer func() {
		db.Close()
		removeTestDatabase(t)
	}()
	err := db.AddNetwork(ctx, model.Network{
		Name:   "basic",
		Prefix: model.MustParsePrefix("192.168.0.0/24"),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := model.Network{
		Name:   "basic",
		Prefix: model.MustParsePrefix("192.168.0.0/24"),
	}

	got, err := db.GetNetworkByName(ctx, "basic")
	if err != nil {
		t.Fatal(err)
	}
	diff := cmp.Diff(want, got,
		cmpopts.EquateComparable(netip.Prefix{}),
	)
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	_, err = db.GetNetworkByName(ctx, "notfound")
	if err == nil {
		t.Errorf("did not get an error for get of unknown name")
	}
	wanterr := model.ErrNetworkDoesNotExist
	diff = cmp.Diff(wanterr, err, cmpopts.EquateErrors())
	if diff != "" {
		t.Fatalf("error mismatch (-want +got):\n%s", diff)
	}
}

func TestSqliteStore_RemoveNetworkByName(t *testing.T) {
	ctx := context.Background()

	db := createTestDatabase(t)
	defer func() {
		db.Close()
		removeTestDatabase(t)
	}()
	err := db.AddNetwork(ctx, model.Network{
		Name:   "basic",
		Prefix: model.MustParsePrefix("192.168.0.0/24"),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.RemoveNetworkByName(ctx, "basic")
	if err != nil {
		t.Fatal(err)
	}

	err = db.RemoveNetworkByName(ctx, "notfound")
	if err == nil {
		t.Errorf("did not get an error for get of unknown name")
	}
	wanterr := model.ErrNetworkDoesNotExist
	diff := cmp.Diff(wanterr, err, cmpopts.EquateErrors())
	if diff != "" {
		t.Fatalf("error mismatch (-want +got):\n%s", diff)
	}
}
