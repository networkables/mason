// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"

	_ "github.com/mattn/go-sqlite3"

	"github.com/networkables/mason/internal/model"
)

func (cs *Store) FlowSummaryByName(
	ctx context.Context,
	addr model.Addr,
) ([]model.FlowSummaryForAddrByName, error) {
	return cs.selectNetflowsSummaryByName(ctx, addr)
}

func (cs *Store) FlowSummaryByCountry(
	ctx context.Context,
	addr model.Addr,
) ([]model.FlowSummaryForAddrByCountry, error) {
	return cs.selectNetflowsSummaryByCountry(ctx, addr)
}

func (cs *Store) selectNetflowsSummaryByName(
	ctx context.Context,
	addr model.Addr,
) (fs []model.FlowSummaryForAddrByName, err error) {
	err = cs.DB.SelectContext(
		ctx,
		&fs,
		`select name,
            ifnull(recvbytes,0) as recvbytes,
            ifnull(xmitbytes,0) as xmitbytes
       from (
            select name,
                   sum(case when flowdirection = 0 then bytes end) as recvbytes,
                   sum(case when flowdirection = 1 then bytes end) as xmitbytes
              from (
                   select flowdirection,
                          asns.name,
                          bytes
                     from (
                           select 0 as flowdirection,
                                  srcasn as asn,
                                  bytes
                             from flows
                            where dstaddr = ?
                            --and start > datetime('now','-60 minute')
                            union
                           select 1 as flowdirection,
                                  dstasn as asn,
                                  bytes
                             from flows
                            where srcaddr = ?
                            --and start > datetime('now','-60 minute')
                           ) dat,
                           asns
                     where dat.asn = asns.asn
                   )
          group by name
          order by sum(bytes) desc
    )`,
		addr,
		addr,
	)
	return fs, err
}

func (cs *Store) selectNetflowsSummaryByCountry(
	ctx context.Context,
	addr model.Addr,
) (fs []model.FlowSummaryForAddrByCountry, err error) {
	err = cs.DB.SelectContext(
		ctx,
		&fs,
		`select country,
            name, 
            ifnull(recvbytes,0) as recvbytes,
            ifnull(xmitbytes,0) as xmitbytes
       from (
            select country,
                   name, 
                   sum(case when flowdirection = 0 then bytes end) as recvbytes,
                   sum(case when flowdirection = 1 then bytes end) as xmitbytes
              from (
                   select flowdirection,
                          asns.name as name,
                          asns.country as country,
                          bytes
                     from (
                           select 0 as flowdirection,
                                  srcasn as asn,
                                  bytes
                             from flows
                            where dstaddr = ?
                            --and start > datetime('now','-60 minute')
                            union
                           select 1 as flowdirection,
                                  dstasn as asn,
                                  bytes
                             from flows
                            where srcaddr = ?
                            --and start > datetime('now','-60 minute')
                           ) dat,
                           asns
                     where dat.asn = asns.asn
                   )
          group by country, name
          order by sum(bytes) desc
    )`,
		addr,
		addr,
	)
	return fs, err
}
