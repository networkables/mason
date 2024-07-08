// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

import (
	"fmt"
	"net"
	"runtime"
	"runtime/debug"
)

var _ Utiler = (*pkg)(nil)

type Utiler interface {
	IsRandomMac(net.HardwareAddr) bool
	GetUserAgent() string
}

func IsRandomMac(mac net.HardwareAddr) bool {
	return DefaultPkg.IsRandomMac(mac)
}

func (p pkg) IsRandomMac(mac net.HardwareAddr) bool {
	if len(mac) == 0 {
		return false
	}
	lsb := mac[0] & 0x0F
	switch lsb {
	case 0x2, 0x6, 0xA, 0xE:
		return true
	}
	return false
}

func GetUserAgent() string {
	return DefaultPkg.GetUserAgent()
}

func (p *pkg) GetUserAgent() string {
	if p.userAgent == "" {
		bi, ok := debug.ReadBuildInfo()
		appversion := "dev_unknown"
		if ok {
			appversion = bi.Main.Version
		}
		p.userAgent = fmt.Sprintf(
			"mason/%s (%s; %s; go:%s)",
			appversion,
			runtime.GOOS,
			runtime.GOARCH,
			runtime.Version()[2:],
		)
	}
	return p.userAgent
}

// func OuiLookup(mac net.HardwareAddr) (string, error) {
// 	return DefaultPkg.OuiLookup(mac)
// }
//
// func (p pkg) OuiLookup(mac net.HardwareAddr) (string, error) {
// 	if len(mac) == 0 {
// 		return "", nil
// 	}
// 	if p.IsRandomMac(mac) {
// 		return "", ErrRandomizedMacAddress
// 	}
//
// 	prefixString := fmt.Sprintf("%02X%02X%02X", mac[0], mac[1], mac[2])
// 	if name, ok := p.ouicache[prefixString]; ok {
// 		return name, nil
// 	}
//
// 	f, err := ouifile.Open("oui/base16_oui.txt")
// 	if err != nil {
// 		return "", err
// 	}
// 	defer f.Close()
// 	var line string
// 	scanner := bufio.NewScanner(f)
// 	for scanner.Scan() {
// 		line = scanner.Text()
// 		if strings.HasPrefix(line, prefixString) {
// 			after, found := strings.CutPrefix(line, prefixString)
// 			if !found {
// 				log.Error(
// 					"ouiLookup cutprefix not found",
// 					"line",
// 					line,
// 				) // TODO: do we need to surface this?
// 				return "", nil
// 			}
// 			p.ouicache[prefixString] = after
// 			return after, nil
// 		}
// 	}
// 	// log.Debug("ouilookup could not find mac", "prefixString", prefixString, "addr", mac)
// 	return "", nil
// }
