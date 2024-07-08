// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import "strings"

type TcpFlags byte

func (tf TcpFlags) String() string {
	// return fmt.Sprintf("%08b", tf)
	str := ""

	if tf == 0 {
		return str
	}

	if tf&0b00000001 > 0 {
		str += ",FIN"
	}
	if tf&0b00000010 > 0 {
		str += ",SYN"
	}
	if tf&0b00000100 > 0 {
		str += ",RST"
	}
	if tf&0b00001000 > 0 {
		str += ",PSH"
	}
	if tf&0b00010000 > 0 {
		str += ",ACK"
	}
	if tf&0b00100000 > 0 {
		str += ",URG"
	}

	str = strings.TrimPrefix(str, ",")
	return str
}
