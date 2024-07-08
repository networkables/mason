// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"database/sql/driver"
)

type DiscoverySource string

func (ds DiscoverySource) String() string {
	return string(ds)
}

func (ds DiscoverySource) IsEmpty() bool {
	return string(ds) == ""
}

func (ds DiscoverySource) Value() (driver.Value, error) {
	x := string(ds)
	return x, nil
}

func (ds *DiscoverySource) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		if src == "" {
			return nil
		}
		x := DiscoverySource(src)
		*ds = x
	}
	return nil
}
