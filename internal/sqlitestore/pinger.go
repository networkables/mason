// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"
	"time"

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
	stmt, err := cs.DB.Prepare(
		`select 
      start, minimum, average, maximum, loss
    from performancepings
    where addr = :addr and start > :start`)
	if err != nil {
		return points, err
	}
	stmt.SetText(":addr", addr.String())
	stmt.SetText(":start", from.Format(time.RFC3339Nano))

	var hasRow bool
	for {
		hasRow, err = stmt.Step()
		if err != nil {
			return points, err
		}
		if !hasRow {
			break
		}

		p := pinger.Point{
			Minimum: time.Duration(stmt.GetInt64("minimum")),
			Average: time.Duration(stmt.GetInt64("average")),
			Maximum: time.Duration(stmt.GetInt64("maximum")),
			Loss:    stmt.GetFloat("loss"),
		}
		p.Start, err = time.Parse(time.RFC3339Nano, stmt.GetText("start"))

		points = append(points, p)
	}

	return points, err
}

func (cs *Store) insertPerformancePing(
	ctx context.Context,
	ts time.Time,
	addr model.Addr,
	p nettools.Icmp4EchoResponseStatistics,
) (err error) {
	stmt, err := cs.DB.Prepare(
		`insert into performancepings (start, addr, minimum, average, maximum, loss)
    values (:start, :addr, :minimum, :average, :maximum, :loss)`)
	if err != nil {
		return err
	}
	stmt.SetText(":start", ts.Format(time.RFC3339Nano))
	stmt.SetText(":addr", addr.String())
	stmt.SetInt64(":minimum", p.Minimum.Nanoseconds())
	stmt.SetInt64(":average", p.Mean.Nanoseconds())
	stmt.SetInt64(":maximum", p.Maximum.Nanoseconds())
	stmt.SetFloat(":loss", p.PacketLoss)

	_, err = stmt.Step()

	return err
}
