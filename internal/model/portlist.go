// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"database/sql/driver"
	"slices"
	"strconv"
	"strings"
)

type PortList struct {
	Ports []int
}

func (pl PortList) IsEmpty() bool {
	if pl.Ports == nil {
		return true
	}
	return len(pl.Ports) == 0
}

const portSeperator = " "

func (pl PortList) String() string {
	str := ""
	for _, port := range pl.Ports {
		if str != "" {
			str += portSeperator
		}
		str += strconv.Itoa(port)
	}
	return str
}

func (pl PortList) Value() (driver.Value, error) {
	x := pl.String()
	return x, nil
}

func (pl *PortList) Scan(src interface{}) (err error) {
	switch src := src.(type) {
	case string:
		if src == "" {
			return nil
		}
		p1, err := ParsePortList(src)
		if err != nil {
			return err
		}
		pl.Ports = p1.Ports
	}
	return nil
}

func (pl PortList) Clone() PortList {
	c := slices.Clone(pl.Ports)
	return PortList{Ports: c}
}

func (pl PortList) Len() int {
	return len(pl.Ports)
}

func ParsePortList(s string) (pl PortList, err error) {
	// s = strings.TrimSpace(s)
	intstrs := strings.Split(s, portSeperator)
	pl.Ports = make([]int, len(intstrs))
	for idx, intstr := range intstrs {
		pl.Ports[idx], err = strconv.Atoi(intstr)
		if err != nil {
			return pl, err
		}
	}
	return pl, nil
}

func MustParsePortList(s string) PortList {
	pl, err := ParsePortList(s)
	if err != nil {
		panic(err)
	}
	return pl
}

func IntSliceToPortList(s []int) PortList {
	return PortList{Ports: slices.Clone(s)}
}
