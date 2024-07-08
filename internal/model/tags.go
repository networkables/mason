// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"database/sql/driver"
	"encoding/json"
)

type Tags []Tag

func (ts Tags) Value() (driver.Value, error) {
	if len(ts) == 0 {
		return "{}", nil
	}
	x, err := json.Marshal(ts)
	if err != nil {
		return nil, err
	}
	return string(x), nil
}

func (ts *Tags) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		if len(src) == 0 || src == "{}" {
			*ts = make(Tags, 0)
			return nil
		}
		return json.Unmarshal([]byte(src), ts)
	}
	return nil
}
