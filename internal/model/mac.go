// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"bytes"
	"database/sql/driver"
	"net"
)

type MAC struct {
	M net.HardwareAddr
}

func (m MAC) IsEmpty() bool {
	return len(m.M) == 0
}

func (m MAC) String() string {
	return m.M.String()
}

func (m MAC) Addr() net.HardwareAddr { return m.M }

func (m MAC) Value() (driver.Value, error) {
	x := m.M.String()
	return x, nil
}

func (m *MAC) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		if src == "" {
			return nil
		}
		x, err := net.ParseMAC(src)
		if err != nil {
			return err
		}
		m.M = x
	}
	return nil
}

func (m MAC) Compare(x MAC) int {
	return bytes.Compare([]byte(m.M), []byte(x.M))
}

func ParseMAC(s string) (m MAC, err error) {
	m1, err := net.ParseMAC(s)
	if err != nil {
		return m, err
	}
	return MAC{M: m1}, nil
}

func MustParseMAC(s string) MAC {
	mac, err := ParseMAC(s)
	if err != nil {
		panic(err)
	}
	return mac
}

func HardwareAddrToMAC(a net.HardwareAddr) MAC {
	return MAC{M: a}
}
