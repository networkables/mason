// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/nettools"
)

// WritePoint stores a given point for a device
func (cs *Store) WritePerformancePing(
	ctx context.Context,
	timestamp time.Time,
	device model.Device,
	point nettools.Icmp4EchoResponseStatistics,
) (err error) {
	err = cs.insertPerformancePing(ctx, timestamp, device.Addr, point)
	if err != nil {
		return err
	}

	return nil
}

// ReadPoints returns the points from Now() minus the duration for the given type
func (cs *Store) ReadPerformancePings(
	ctx context.Context,
	device model.Device,
	duration time.Duration,
) (points []pinger.Point, err error) {
	points, err = cs.selectPerformancePings(ctx, device.Addr, time.Now().Add(-1*duration))
	if err != nil {
		return points, err
	}
	return points, nil
}

func (cs *Store) selectPerformancePings(
	ctx context.Context,
	addr model.Addr,
	from time.Time,
) (points []pinger.Point, err error) {
	err = cs.DB.SelectContext(
		ctx,
		&points,
		`select 
      start, minimum, average, maximum, loss
    from performancepings
    where addr = ? and start > ?`,
		addr,
		from,
	)
	return points, err
}

func (cs *Store) insertPerformancePing(
	ctx context.Context,
	ts time.Time,
	addr model.Addr,
	p nettools.Icmp4EchoResponseStatistics,
) (err error) {
	_, err = cs.DB.ExecContext(
		ctx,
		`insert into performancepings (start, addr, minimum, average, maximum, loss)
    values (?, ?, ?, ?, ?, ?)`,
		ts, addr, p.Minimum, p.Mean, p.Maximum, p.PacketLoss,
	)
	return err
}
