// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

type IpFlow struct {
	SrcAddr  Addr
	SrcPort  uint16
	SrcASN   string
	DstAddr  Addr
	DstPort  uint16
	DstASN   string
	Start    time.Time
	End      time.Time
	Bytes    int
	Packets  int
	Protocol Protocol // https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
	Flags    TcpFlags
}

func (ipf IpFlow) String() string {
	return fmt.Sprintf("src:%15s:%5d [%5s] dst:%15s:%5d [%5s] %4s %6s %3dp %s %s",
		ipf.SrcAddr, ipf.SrcPort, ipf.SrcASN,
		ipf.DstAddr, ipf.DstPort, ipf.DstASN,
		ipf.Protocol,
		humanize.Bytes(uint64(ipf.Bytes)), ipf.Packets,
		ipf.Start.Format(time.TimeOnly),
		ipf.Flags,
	)
}
