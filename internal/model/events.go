// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
)

type (
	EventDeviceDiscovered Device
	EventDeviceAdded      Device
	EventDeviceUpdated    Device
)

var EmptyDiscoveredDevice EventDeviceDiscovered

func (dd EventDeviceDiscovered) String() string {
	return fmt.Sprintf("%s %s", dd.DiscoveredBy, dd.Addr)
}

func (nd EventDeviceAdded) String() string { return nd.Addr.String() }

func (ude EventDeviceUpdated) String() string {
	return fmt.Sprintf("%s [%s %s]", ude.Name, ude.Addr, ude.MAC)
}
