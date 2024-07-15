// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

//go:build windows

package nettools

import "net"

func (p *pkg) loadRoutes() error {
	p.ifacesByNetPrefix = make(map[string]*net.Interface)
	return nil
}
