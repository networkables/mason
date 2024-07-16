// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"github.com/networkables/mason/internal/model"
)

// AddNetflows adds a network to the store, will return error if the network already exists in the store
func (cs *Store) AddNetflows(ctx context.Context, flows []model.IpFlow) (err error) {
	conn, err := cs.Pool.Take(ctx)
	if err != nil {
		return err
	}
	fn := sqlitex.Transaction(conn)
	defer func() {
		fn(&err)
		cs.Pool.Put(conn)
	}()
	for _, flow := range flows {
		err = insertNetflow(conn, flow)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs *Store) GetNetflows(
	ctx context.Context,
	addr model.Addr,
) (flows []model.IpFlow, err error) {
	return cs.selectNetflow(ctx, addr)
}

func insertNetflow(conn *sqlite.Conn, n model.IpFlow) error {
	stmt, err := conn.Prepare(
		`INSERT INTO flows (start, end, srcaddr, srcport, srcasn, dstaddr, dstport, dstasn, protocol, bytes, packets)
    VALUES (:start, :end, :srcaddr, :srcport, :srcasn, :dstaddr, :dstport, :dstasn, :protocol, :bytes, :packets)
		`,
	)
	if err != nil {
		return err
	}
	stmt.SetText(":start", n.Start.Format(time.RFC3339Nano))
	stmt.SetText(":end", n.End.Format(time.RFC3339Nano))
	stmt.SetText(":srcaddr", n.SrcAddr.String())
	stmt.SetInt64(":srcport", int64(n.SrcPort))
	stmt.SetText(":srcasn", n.SrcASN)
	stmt.SetText(":dstaddr", n.DstAddr.String())
	stmt.SetInt64(":dstport", int64(n.DstPort))
	stmt.SetText(":dstasn", n.DstASN)
	stmt.SetText(":protocol", n.Protocol.String())
	stmt.SetInt64(":bytes", int64(n.Bytes))
	stmt.SetInt64(":packets", int64(n.Packets))
	_, err = stmt.Step()
	return err
}

func (cs *Store) selectNetflow(
	ctx context.Context,
	addr model.Addr,
) (fs []model.IpFlow, err error) {
	stmt, err := cs.DB.Prepare(
		`SELECT start, end, srcaddr, srcport, srcasn, dstaddr, dstport, dstasn, protocol, bytes, packets
     FROM flows 
    WHERE srcaddr = :srcaddr OR dstaddr = :dstaddr`,
	)
	var hasRow bool
	for {
		fmt.Println("device")
		hasRow, err = stmt.Step()
		if err != nil {
			return fs, err
		}
		if !hasRow {
			break
		}
		flow := model.IpFlow{
			SrcPort: uint16(stmt.GetInt64("srcport")),
			SrcASN:  stmt.GetText("srcasn"),
			DstPort: uint16(stmt.GetInt64("dstport")),
			DstASN:  stmt.GetText("dstasn"),
			Bytes:   int(stmt.GetInt64("bytes")),
			Packets: int(stmt.GetInt64("packets")),
		}
		flow.Start, err = time.Parse(time.RFC3339Nano, stmt.GetText("start"))
		if err != nil {
			return fs, err
		}
		flow.End, err = time.Parse(time.RFC3339Nano, stmt.GetText("end"))
		if err != nil {
			return fs, err
		}
		err = flow.SrcAddr.Scan(stmt.GetText("srcaddr"))
		if err != nil {
			return fs, err
		}
		err = flow.DstAddr.Scan(stmt.GetText("dstaddr"))
		if err != nil {
			return fs, err
		}
		p, err := strconv.Atoi(stmt.GetText("protocol"))
		if err != nil {
			return fs, err
		}
		flow.Protocol = model.Protocol(p)

		fs = append(fs, flow)
	}
	return fs, err
}

func (cs *Store) FlowSummaryByIP(
	ctx context.Context,
	addr model.Addr,
) ([]model.FlowSummaryForAddrByIP, error) {
	return cs.selectNetflowsSummaryByIP(ctx, addr)
}

func (cs *Store) selectNetflowsSummaryByIP(
	ctx context.Context,
	addr model.Addr,
) (fs []model.FlowSummaryForAddrByIP, err error) {
	stmt, err := cs.DB.Prepare(
		`SELECT country,
            name,
            asn,
            addr,
            ifnull(recvbytes,0) AS recvbytes,
            ifnull(xmitbytes,0) AS xmitbytes
       FROM (
            SELECT asns.country,
                   asns.name,
                   asns.asn,
                   dat.addr,
                   SUM(CASE WHEN flowdirection = 0 THEN bytes END) AS recvbytes,
                   SUM(CASE WHEN flowdirection = 1 THEN bytes END) AS xmitbytes
              FROM (
                   SELECT 0 AS flowdirection,
                          srcasn AS asn,
                          srcaddr AS addr,
                          bytes
                     FROM flows
                    WHERE dstaddr = :addr
                    --and start > datetime('now','-60 minute')
                    UNION
                   SELECT 1 AS flowdirection,
                          dstasn AS asn,
                          dstaddr AS addr,
                          bytes
                     FROM flows
                    WHERE srcaddr = :addr
                    --and start > datetime('now','-60 minute')
                   ) dat, 
                   asns
              WHERE dat.asn = asns.asn
             GROUP BY asns.country, asns.name, asns.asn, dat.addr
             ORDER BY SUM(bytes) DESC
            )`)
	if err != nil {
		return fs, err
	}
	stmt.SetText(":addr", addr.String())
	var hasRow bool
	for {
		hasRow, err = stmt.Step()
		if err != nil {
			return fs, err
		}
		if !hasRow {
			break
		}
		f := model.FlowSummaryForAddrByIP{
			Country:   stmt.GetText("country"),
			Name:      stmt.GetText("name"),
			Asn:       stmt.GetText("asn"),
			RecvBytes: int(stmt.GetInt64("recvbytes")),
			XmitBytes: int(stmt.GetInt64("xmitbytes")),
		}
		err = f.Addr.Scan(stmt.GetText("addr"))
		if err != nil {
			return fs, err
		}

		fs = append(fs, f)
	}
	return fs, err
}
