// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"

	_ "github.com/mattn/go-sqlite3"

	"github.com/networkables/mason/internal/model"
)

func (cs *Store) UpsertAsn(ctx context.Context, asn model.Asn) error {
	return cs.upsertAsn(ctx, asn)
}

func (cs *Store) upsertAsn(ctx context.Context, asn model.Asn) error {
	_, err := cs.DB.NamedExecContext(
		ctx,
		`insert into asns (asn, country, name, iprange, created)
    values (:asn, :country, :name, :iprange, datetime('now'))
    on conflict (asn) do update set country=:country, name=:name, iprange=:iprange, created=datetime('now')`,
		asn,
	)
	return err
}

func (cs *Store) GetAsn(ctx context.Context, asn string) (am model.Asn, err error) {
	return cs.selectAsn(ctx, asn)
}

func (cs *Store) selectAsn(ctx context.Context, asn string) (am model.Asn, err error) {
	var res []model.Asn
	err = cs.DB.SelectContext(
		ctx,
		&res,
		`select asn, country, name, iprange from asns where asn = ? limit 1`,
		asn,
	)
	if err != nil {
		return am, err
	}
	if len(res) > 0 {
		return res[0], nil
	}
	return model.Asn{}, nil
}
