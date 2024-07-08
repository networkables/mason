// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"database/sql/driver"

	"go4.org/netipx"
)

type IPRange struct {
	A netipx.IPRange
}

func (na IPRange) IPRange() netipx.IPRange {
	return na.A
}

func (na IPRange) String() string {
	return na.A.String()
}

func (na IPRange) Value() (driver.Value, error) {
	x := na.A.String()
	return x, nil
}

func (na *IPRange) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		x, err := ParseIPRange(src)
		if err != nil {
			return err
		}
		na.A = x.A
	}
	return nil
}

func (na IPRange) Compare(ip2 IPRange) int {
	naaddr := na.A.Prefixes()[0].Addr()
	ip2addr := ip2.A.Prefixes()[0].Addr()
	return naaddr.Compare(ip2addr)
}

func ParseIPRange(s string) (ip IPRange, err error) {
	addr, err := netipx.ParseIPRange(s)
	if err != nil {
		return ip, err
	}
	return IPRange{A: addr}, nil
}

func MustParseIPRange(s string) IPRange {
	return IPRange{A: netipx.MustParseIPRange(s)}
}

func IPRangeToModelIPRange(a netipx.IPRange) IPRange {
	return IPRange{A: a}
}
