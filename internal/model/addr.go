// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"database/sql/driver"
	"net/netip"
)

type Addr struct {
	A netip.Addr
}

func (na Addr) Addr() netip.Addr {
	return na.A
}

func (na Addr) String() string {
	return na.A.String()
}

func (na Addr) Value() (driver.Value, error) {
	x := na.A.String()
	return x, nil
}

func (na *Addr) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		x, err := netip.ParseAddr(src)
		if err != nil {
			return err
		}
		na.A = x
	}
	return nil
}

func (na Addr) Compare(ip2 Addr) int {
	return na.A.Compare(ip2.A)
}

func ParseAddr(s string) (ip Addr, err error) {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return ip, err
	}
	return Addr{A: addr}, nil
}

func MustParseAddr(s string) Addr {
	return Addr{A: netip.MustParseAddr(s)}
}

func AddrToModelAddr(a netip.Addr) Addr {
	return Addr{A: a}
}
