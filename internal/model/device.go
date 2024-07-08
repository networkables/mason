// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/networkables/mason/nettools"
)

type (
	Device struct {
		Name         string
		Addr         Addr
		MAC          MAC
		DiscoveredAt time.Time
		DiscoveredBy DiscoverySource

		Meta            Meta
		Server          Server
		PerformancePing Pinger
		SNMP            SNMP

		updated bool
	}

	// DeviceFilter defines a function used to select a set of required devices.
	DeviceFilter func(Device) bool

	Meta struct {
		DnsName      string
		Manufacturer string
		Tags         Tags
	}

	Server struct {
		Ports    PortList
		LastScan time.Time
	}

	Pinger struct {
		FirstSeen  time.Time
		LastSeen   time.Time
		Mean       time.Duration
		Maximum    time.Duration
		LastFailed bool
	}

	SNMP struct {
		Name               string
		Description        string
		Community          string
		Port               int
		LastSNMPCheck      time.Time
		HasArpTable        bool
		LastArpTableScan   time.Time
		HasInterfaces      bool
		LastInterfacesScan time.Time
	}
)

func (d Device) merge(in Device) (out Device, updated bool) {
	if in.Name != "" && d.Name != in.Name {
		d.Name = in.Name
		updated = true
	}
	if d.Addr.Compare(in.Addr) != 0 {
		d.Addr = in.Addr
		updated = true
	}
	if !in.MAC.IsEmpty() && d.MAC.String() != in.MAC.String() {
		d.MAC = in.MAC
		updated = true
	}
	if !in.DiscoveredAt.IsZero() && !d.DiscoveredAt.Equal(in.DiscoveredAt) {
		d.DiscoveredAt = in.DiscoveredAt
		updated = true
	}
	if d.DiscoveredBy.IsEmpty() && d.DiscoveredBy != in.DiscoveredBy {
		d.DiscoveredBy = in.DiscoveredBy
		updated = true
	}
	return d, updated
}

func (m Meta) merge(in Meta) (out Meta, updated bool) {
	if in.DnsName != "" && m.DnsName != in.DnsName {
		m.DnsName = in.DnsName
		updated = true
	}
	if in.Manufacturer != "" && m.Manufacturer != in.Manufacturer {
		m.Manufacturer = in.Manufacturer
		updated = true
	}
	if len(in.Tags) > 0 && !cmp.Equal(m.Tags, in.Tags) {
		m.Tags = slices.Clone(in.Tags)
		updated = true
	}
	return m, updated
}

func (s Server) merge(in Server) (out Server, updated bool) {
	if s.Ports.IsEmpty() && !cmp.Equal(s.Ports, in.Ports) {
		s.Ports = in.Ports.Clone()
		updated = true
	}
	if !in.LastScan.IsZero() && !s.LastScan.Equal(in.LastScan) {
		s.LastScan = in.LastScan
		updated = true
	}
	return s, updated
}

func (p Pinger) merge(in Pinger) (out Pinger, updated bool) {
	if !in.FirstSeen.IsZero() && !p.FirstSeen.Equal(in.FirstSeen) {
		p.FirstSeen = in.FirstSeen
		updated = true
	}
	if !in.LastSeen.IsZero() && !p.LastSeen.Equal(in.LastSeen) {
		p.LastSeen = in.LastSeen
		updated = true
	}
	if p.Mean != in.Mean {
		p.Mean = in.Mean
		updated = true
	}
	if p.Maximum != in.Maximum {
		p.Maximum = in.Maximum
		updated = true
	}
	if p.LastFailed != in.LastFailed {
		p.LastFailed = in.LastFailed
		updated = true
	}
	return p, updated
}

func (s SNMP) merge(in SNMP) (out SNMP, updated bool) {
	if in.Name != "" && s.Name != in.Name {
		s.Name = in.Name
		updated = true
	}
	if in.Description != "" && s.Description != in.Description {
		s.Description = in.Description
		updated = true
	}
	if in.Community != "" && s.Community != in.Community {
		s.Community = in.Community
		updated = true
	}
	if in.Port != 0 && s.Port != in.Port {
		s.Port = in.Port
		updated = true
	}
	if !in.LastSNMPCheck.IsZero() && !s.LastSNMPCheck.Equal(in.LastSNMPCheck) {
		s.LastSNMPCheck = in.LastSNMPCheck
		updated = true
	}
	if !s.HasArpTable && in.HasArpTable {
		s.HasArpTable = in.HasArpTable
		updated = true
	}
	if !in.LastArpTableScan.IsZero() && !s.LastArpTableScan.Equal(in.LastArpTableScan) {
		s.LastArpTableScan = in.LastArpTableScan
		updated = true
	}
	if !s.HasInterfaces && in.HasInterfaces {
		s.HasInterfaces = in.HasInterfaces
		updated = true
	}
	if !in.LastInterfacesScan.IsZero() && !s.LastInterfacesScan.Equal(in.LastInterfacesScan) {
		s.LastInterfacesScan = in.LastInterfacesScan
		updated = true
	}
	return s, updated
}

var (
	ErrDeviceExists       = errors.New("device exists")
	ErrDeviceDoesNotExist = errors.New("device does not exists")

	EmptyDevice = Device{}
)

func (d Device) IsUpdated() bool {
	return d.updated
}

func (d *Device) SetUpdated() {
	d.updated = true
}

func (d *Device) ClearUpdated() {
	d.updated = false
}

func (d Device) String() string {
	return d.toString(time.Since)
}

func (d Device) toString(f func(time.Time) time.Duration) string {
	return fmt.Sprintf(
		"%5s %-15s [%17s <%s>] fs:%s ls:%s tags:%s",
		d.Name,
		d.Addr,
		d.MAC,
		d.Meta.Manufacturer,
		d.FirstSeenString(),
		d.LastSeenDurString(f),
		d.Meta.Tags,
	)
}

func DateTimeFmt(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format(time.DateTime)
}

func (d Device) DiscoveredAtString() string {
	return d.DiscoveredAt.Format(time.DateOnly)
}

func (d Device) FirstSeenString() string {
	return DateTimeFmt(d.PerformancePing.FirstSeen)
}

func (d Device) LastSeenString() string {
	return DateTimeFmt(d.PerformancePing.LastSeen)
}

func (d Device) LastSeenDurString(f func(time.Time) time.Duration) string {
	lsdur := f(d.PerformancePing.LastSeen)
	dls := "never"
	if !d.PerformancePing.LastSeen.IsZero() {
		dls = lsdur.Round(time.Minute).String()
	}
	return dls
}

func (d Device) LastPingMeanString() string {
	if d.PerformancePing.FirstSeen.IsZero() {
		return ""
	}
	if d.PerformancePing.LastFailed {
		return "failure"
	}
	return d.PerformancePing.Mean.Round(50 * time.Microsecond).String()
}

func (d Device) LastPingMaximumString() string {
	if d.PerformancePing.FirstSeen.IsZero() {
		return ""
	}
	if d.PerformancePing.LastFailed {
		return "failure"
	}
	return d.PerformancePing.Maximum.Round(50 * time.Microsecond).String()
}

func (d *Device) UpdateFromPingStats(stats nettools.Icmp4EchoResponseStatistics, ts time.Time) {
	if stats.SuccessCount > 0 {
		d.updated = true
		d.PerformancePing.LastFailed = false
		d.PerformancePing.LastSeen = ts
		d.PerformancePing.Mean = stats.Mean
		d.PerformancePing.Maximum = stats.Maximum
		if d.PerformancePing.FirstSeen.IsZero() {
			d.PerformancePing.FirstSeen = ts
		}
	} else {
		d.PerformancePing.LastFailed = true
	}
}

func (d Device) Merge(in Device) Device {
	var baseUpdated, metaUpdated, serverUpdated, pingerUpdated, snmpUpdated bool
	d, baseUpdated = d.merge(in)
	d.Meta, metaUpdated = d.Meta.merge(in.Meta)
	d.Server, serverUpdated = d.Server.merge(in.Server)
	d.PerformancePing, pingerUpdated = d.PerformancePing.merge(in.PerformancePing)
	d.SNMP, snmpUpdated = d.SNMP.merge(in.SNMP)
	d.updated = baseUpdated || metaUpdated || serverUpdated || pingerUpdated || snmpUpdated

	if d.Name == "" || (d.IsNameAddr() && d.Meta.DnsName != "") {
		d.Name = d.Addr.String()
		if d.Meta.DnsName != "" {
			d.Name = d.Meta.DnsName
		}
	}
	return d
}

func (d Device) IsNameAddr() bool {
	return d.Name == d.Addr.String()
}

func (d Device) IsServer() bool {
	if d.Server.Ports.Len() > 0 {
		return true
	}
	return false
}

func SortDevicesByAddr(devs []Device) {
	sort.SliceStable(devs, func(i, j int) bool {
		return devs[i].Addr.Compare(devs[j].Addr) == -1
	})
}
