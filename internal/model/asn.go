// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

type Asn struct {
	IPRange IPRange
	Asn     string
	Country string
	Name    string
}

type IpAsn struct {
	Asn
	IP      Addr
	Created time.Time
}
