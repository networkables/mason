// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type Protocol byte

func (p Protocol) String() string {
	switch p {
	case 0:
		return "HOPOPT"
	case 1:
		return "ICMP"
	case 2:
		return "IGMP"
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	default:
		return "Unknown: [" + string(p) + "]"
	}
}
