// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type FlowSummaryForAddrByIP struct {
	Country   string
	Name      string
	Asn       string
	Addr      Addr
	RecvBytes int
	XmitBytes int
}

type FlowSummaryForAddrByName struct {
	Name      string
	RecvBytes int
	XmitBytes int
}

type FlowSummaryForAddrByCountry struct {
	Country   string
	Name      string
	RecvBytes int
	XmitBytes int
}
