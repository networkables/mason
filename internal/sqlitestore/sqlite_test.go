// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"os"
	"testing"
	"time"
)

var testdbdir string

func createTestDatabase(t *testing.T) *Store {
	t.Helper()
	var err error

	testdbdir, err = os.MkdirTemp("", "dbt")
	if err != nil {
		t.Fatal(err)
	}
	// t.Logf("ctd: %s", testdbdir)

	db, err := New(&Config{
		Enabled:               true,
		Directory:             testdbdir,
		Filename:              "unittest.db",
		URL:                   "",
		MaxOpenConnections:    1,
		MaxIdleConnections:    1,
		ConnectionMaxLifetime: time.Minute,
		ConnectionMaxIdle:     time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func removeTestDatabase(t *testing.T) {
	t.Helper()

	// t.Logf("rtd: %s", testdbdir)
	err := os.RemoveAll(testdbdir)
	if err != nil {
		t.Fatal(err)
	}
}
