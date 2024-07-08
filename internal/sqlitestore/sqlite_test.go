// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"context"
	"testing"
	"time"
)

func createTestDatabase(t *testing.T) *Store {
	t.Helper()

	db, err := New(&Config{
		Enabled:               true,
		Directory:             "",
		Filename:              ":memory:",
		URL:                   "",
		MaxOpenConnections:    1,
		MaxIdleConnections:    1,
		ConnectionMaxLifetime: time.Minute,
		ConnectionMaxIdle:     time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.MigrateUp(context.Background()); err != nil {
		t.Fatal(err)
	}

	return db
}
