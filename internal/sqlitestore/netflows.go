// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"

	_ "github.com/mattn/go-sqlite3"

	"github.com/networkables/mason/internal/model"
)

// AddNetflows adds a network to the store, will return error if the network already exists in the store
func (cs *Store) AddNetflows(ctx context.Context, flows []model.IpFlow) (err error) {
	for _, flow := range flows {
		err = cs.insertNetflow(ctx, flow)
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

func (cs *Store) insertNetflow(ctx context.Context, n model.IpFlow) error {
	_, err := cs.DB.NamedExecContext(
		ctx,
		`insert into flows (start, end, srcaddr, srcport, srcasn, dstaddr, dstport, dstasn, protocol, bytes, packets)
    values (:start, :end, :srcaddr, :srcport, :srcasn, :dstaddr, :dstport, :dstasn, :protocol, :bytes, :packets)
		`,
		n,
	)
	return err
}

func (cs *Store) selectNetflow(
	ctx context.Context,
	addr model.Addr,
) ([]model.IpFlow, error) {
	var fs []model.IpFlow
	err := cs.DB.SelectContext(
		ctx,
		&fs,
		`select start, end, srcaddr, srcport, srcasn, dstaddr, dstport, dstasn, protocol, bytes, packets from flows where srcaddr = ? or dstaddr = ?`,
		addr,
		addr,
	)
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
	err = cs.DB.SelectContext(
		ctx,
		&fs,
		`select country,
            name,
            asn,
            addr,
            ifnull(recvbytes,0) as recvbytes,
            ifnull(xmitbytes,0) as xmitbytes
       from (
            select asns.country,
                   asns.name,
                   asns.asn,
                   dat.addr,
                   sum(case when flowdirection = 0 then bytes end) as recvbytes,
                   sum(case when flowdirection = 1 then bytes end) as xmitbytes
              from (
                   select 0 as flowdirection,
                          srcasn as asn,
                          srcaddr as addr,
                          bytes
                     from flows
                    where dstaddr = ?
                    --and start > datetime('now','-60 minute')
                    union
                   select 1 as flowdirection,
                          dstasn as asn,
                          dstaddr as addr,
                          bytes
                     from flows
                    where srcaddr = ?
                    --and start > datetime('now','-60 minute')
                   ) dat, 
                   asns
              where dat.asn = asns.asn
             group by asns.country, asns.name, asns.asn, dat.addr
             order by sum(bytes) desc
            )`,
		addr,
		addr,
	)
	return fs, err
}
