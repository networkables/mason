// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"

	"zombiezen.com/go/sqlite/sqlitex"

	"github.com/networkables/mason/internal/model"
)

func (cs *Store) StartAsnLoad() func(*error) {
	return sqlitex.Transaction(cs.DB)
}

func (cs *Store) UpsertAsn(ctx context.Context, asn model.Asn) error {
	return cs.upsertAsn(ctx, asn)
}

func (cs *Store) upsertAsn(ctx context.Context, asn model.Asn) error {
	stmt, err := cs.DB.Prepare(
		`INSERT INTO asns (asn, country, name, iprange, created)
    VALUES (:asn, :country, :name, :iprange, datetime('now'))
    ON CONFLICT (asn) DO UPDATE SET country=:country, name=:name, iprange=:iprange, created=datetime('now')`,
	)
	if err != nil {
		return err
	}
	stmt.SetText(":asn", asn.Asn)
	stmt.SetText(":country", asn.Country)
	stmt.SetText(":name", asn.Name)
	stmt.SetText(":iprange", asn.IPRange.String())
	_, err = stmt.Step()
	return err
}

func (cs *Store) GetAsn(ctx context.Context, asn string) (am model.Asn, err error) {
	return cs.selectAsn(ctx, asn)
}

func (cs *Store) selectAsn(ctx context.Context, asn string) (am model.Asn, err error) {
	stmt, err := cs.DB.Prepare(
		"SELECT asn, country, name, iprange FROM asns WHERE asn = :asn LIMIT 1",
	)
	if err != nil {
		return am, err
	}
	stmt.SetText(":asn", asn)
	hasRow, err := stmt.Step()
	if err != nil {
		return am, err
	}
	if !hasRow {
		return am, err
	}
	am.Asn = stmt.GetText("asn")
	am.Country = stmt.GetText("country")
	am.Name = stmt.GetText("name")
	err = am.IPRange.Scan(stmt.GetText("iprange"))
	if err != nil {
		return am, err
	}
	return am, nil
}
