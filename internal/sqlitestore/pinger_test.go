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
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/nettools"
)

func TestSqliteStore_WritePerformancePing(t *testing.T) {
	var err error
	ctx := context.Background()

	ts, err := time.Parse(time.RFC3339Nano, "2024-12-11T22:21:20.012345678-00:00")
	if err != nil {
		t.Fatal(err)
	}
	dev := model.Device{Addr: model.MustParseAddr("192.168.86.1")}
	point := pinger.Point{
		Start:   ts,
		Average: time.Second,
		Minimum: time.Minute,
		Maximum: time.Hour,
		Loss:    0.5,
	}

	db := createTestDatabase(t)
	defer func() {
		db.Close()
		removeTestDatabase(t)
	}()
	err = db.WritePerformancePing(
		ctx,
		ts,
		dev,
		nettools.Icmp4EchoResponseStatistics{
			Start:      ts,
			Mean:       time.Second,
			Minimum:    time.Minute,
			Maximum:    time.Hour,
			PacketLoss: 0.5,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	points, err := db.ReadPerformancePings(ctx, dev, 2*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	diff := cmp.Diff(
		points,
		[]pinger.Point{point},
		cmpopts.EquateComparable(netip.Addr{}),
		cmpopts.IgnoreUnexported(model.Device{}),
	)
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}
