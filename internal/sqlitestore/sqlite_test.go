// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlitestore

import (
	"os"
	"testing"
	"time"
)

func createTestDatabase(t *testing.T) *Store {
	t.Helper()

	os.Remove("unittest.db")
	os.Remove("unittest.db-shm")
	os.Remove("unittest.db-wal")
	db, err := New(&Config{
		Enabled:               true,
		Directory:             "",
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

	os.Remove("unittest.db")
	os.Remove("unittest.db-shm")
	os.Remove("unittest.db-wal")
}
