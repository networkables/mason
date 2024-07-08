// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"database/sql/driver"
	"net/netip"

	"go4.org/netipx"
)

type Prefix struct {
	P netip.Prefix
}

func (np Prefix) String() string {
	return np.P.String()
}

func (np Prefix) Value() (driver.Value, error) {
	x := np.P.String()
	return x, nil
}

func (np *Prefix) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		x, err := netip.ParsePrefix(src)
		if err != nil {
			return err
		}
		np.P = x
	}
	return nil
}

func PrefixToModelPrefix(p netip.Prefix) Prefix {
	return Prefix{P: p}
}

func MustParsePrefix(s string) Prefix {
	return Prefix{P: netip.MustParsePrefix(s)}
}

func (np Prefix) Contains(addr Addr) bool {
	return np.ContainsAddr(addr.Addr())
}

func (np Prefix) ContainsAddr(addr netip.Addr) bool {
	return np.P.Contains(addr)
}

func (np Prefix) Addr() netip.Addr {
	return np.P.Addr()
}

func (np Prefix) Is6() bool {
	return np.P.Addr().Is6()
}

func (np Prefix) Bits() int {
	return np.P.Bits()
}

func ComparePrefix(a Prefix, b Prefix) int {
	return netipx.ComparePrefix(a.P, b.P)
}
