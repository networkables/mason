// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

import (
	"context"
	"errors"
	"github.com/charmbracelet/log"
	"github.com/gosnmp/gosnmp"
	"net"
	"net/netip"
	"strings"
	"time"
)

var _ Snmper = (*pkg)(nil)

type Snmper interface {
	FetchSNMPInfo(context.Context, netip.Addr, ...snmpRequestOptionFunc) (SnmpInfo, error)
	SnmpGetSystemInfo(context.Context, netip.Addr, ...snmpRequestOptionFunc) (SnmpSystemInfo, error)
	SnmpGetInterfaces(context.Context, netip.Addr, ...snmpRequestOptionFunc) ([]netip.Prefix, error)
	SnmpGetArpTable(context.Context, netip.Addr, ...snmpRequestOptionFunc) ([]ArpEntry, error)
}

type SnmpInfo struct {
	SystemInfo SnmpSystemInfo
	Interfaces []netip.Prefix
	ArpTable   []ArpEntry
}

func FetchSNMPInfo(ctx context.Context, target netip.Addr, options ...snmpRequestOptionFunc) (SnmpInfo, error) {
	return DefaultPkg.FetchSNMPInfo(ctx, target, options...)
}

func (p pkg) FetchSNMPInfo(ctx context.Context, target netip.Addr, options ...snmpRequestOptionFunc) (info SnmpInfo, err error) {
	info.SystemInfo, err = p.SnmpGetSystemInfo(ctx, target, options...)
	if err != nil {
		return info, err
	}
	info.Interfaces, err = p.SnmpGetInterfaces(ctx, target, options...)
	if err != nil {
		return info, err
	}
	info.ArpTable, err = p.SnmpGetArpTable(ctx, target, options...)
	if err != nil {
		return info, err
	}
	return info, err
}

func SnmpGetSystemInfo(ctx context.Context, target netip.Addr, options ...snmpRequestOptionFunc) (SnmpSystemInfo, error) {
	return DefaultPkg.SnmpGetSystemInfo(ctx, target, options...)
}

func (p pkg) SnmpGetSystemInfo(ctx context.Context, target netip.Addr, options ...snmpRequestOptionFunc) (ssi SnmpSystemInfo, err error) {
	opts := applySnmpRequestOptions(options...)

	ssi.Description, err = snmpGetSingleString(target, "1.3.6.1.2.1.1.1.0", opts.community, opts.port, opts.responseTimeout)
	if err != nil {
		return ssi, err
	}
	ssi.Contact, err = snmpGetSingleString(target, "1.3.6.1.2.1.1.4.0", opts.community, opts.port, opts.responseTimeout)
	if err != nil {
		return ssi, err
	}

	ssi.Name, err = snmpGetSingleString(target, "1.3.6.1.2.1.1.5.0", opts.community, opts.port, opts.responseTimeout)
	if err != nil {
		return ssi, err
	}

	ssi.Location, err = snmpGetSingleString(target, "1.3.6.1.2.1.1.6.0", opts.community, opts.port, opts.responseTimeout)
	if err != nil {
		return ssi, err
	}

	return ssi, nil
}

type SnmpSystemInfo struct {
	Description string
	Contact     string
	Name        string
	Location    string
}

func SnmpGetInterfaces(ctx context.Context, addr netip.Addr, options ...snmpRequestOptionFunc) ([]netip.Prefix, error) {
	return DefaultPkg.SnmpGetInterfaces(ctx, addr, options...)
}

func (p pkg) SnmpGetInterfaces(ctx context.Context, addr netip.Addr, options ...snmpRequestOptionFunc) (prefixes []netip.Prefix, err error) {
	opts := applySnmpRequestOptions(options...)
	oid := "1.3.6.1.2.1.4.20.1.3"
	prefixes = make([]netip.Prefix, 0)

	client, err := snmpClient(addr, opts.community, opts.port, opts.responseTimeout)
	if err != nil {
		return prefixes, err
	}
	defer client.Conn.Close()
	err = client.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		ipstr := stripIPAddressFromSNMPOid(pdu.Name, oid)
		maskstr := pdu.Value.(string)
		prefix := ipSubnetStrToPrefix(ipstr, maskstr)
		if !prefix.Addr().IsLoopback() && !(prefix.Addr().Compare(netip.IPv4Unspecified()) == 0) {
			prefixes = append(prefixes, prefix)
		}
		return nil
	})
	err = snmpErrCheck(err)
	if err != nil {
		return prefixes, err
	}
	return prefixes, nil
}

func SnmpGetArpTable(ctx context.Context, addr netip.Addr, options ...snmpRequestOptionFunc) ([]ArpEntry, error) {
	return DefaultPkg.SnmpGetArpTable(ctx, addr, options...)
}

func (p pkg) SnmpGetArpTable(ctx context.Context, addr netip.Addr, options ...snmpRequestOptionFunc) (arps []ArpEntry, err error) {
	opts := applySnmpRequestOptions(options...)

	oid := "1.3.6.1.2.1.4.22.1.2"
	arps = make([]ArpEntry, 0)
	client, err := snmpClient(addr, opts.community, opts.port, opts.responseTimeout)
	if err != nil {
		return arps, err
	}
	defer client.Conn.Close()
	err = client.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		ipstr := strings.Join(strings.Split(pdu.Name, ".")[12:], ".")
		addr, err := netip.ParseAddr(ipstr)
		if err != nil {
			return err
		}
		mac := net.HardwareAddr(pdu.Value.([]byte))
		if mac.String() == "ff:ff:ff:ff:ff:ff" {
			return nil
		}
		arps = append(arps, ArpEntry{Addr: addr, MAC: mac})
		return nil
	})
	err = snmpErrCheck(err)
	if err != nil {
		return arps, err
	}
	return arps, nil
}

func stripIPAddressFromSNMPOid(oid, rootoid string) string {
	return strings.Replace(oid, "."+rootoid+".", "", 1)
}

func ipSubnetStrToPrefix(ip, mask string) netip.Prefix {
	// TODO: unsure if ipv6 addresses could come in this way, may need checks on data lengths
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		log.Warn("issubnetStrToPrefix is not ok", "ipstr", ip, "mask", mask, "error", err)
	}
	netmask := net.ParseIP(mask)
	masky := net.IPMask{netmask[12], netmask[13], netmask[14], netmask[15]}
	ones, _ := masky.Size()
	prefix := netip.PrefixFrom(addr, ones)
	return prefix
}

// Options available for SNMP Requests
type snmpRequestOptionFunc func(*snmpRequestOptions)

type snmpRequestOptions struct {
	community               string
	port                    int
	responseTimeout         time.Duration
}

func defaultSnmpRequestOptions() *snmpRequestOptions {
	return &snmpRequestOptions{
		community:       "public",
		port:            161,
		responseTimeout: 100 * time.Millisecond,
	}
}

func WithSnmpReplyTimeout(duration time.Duration) snmpRequestOptionFunc {
	return func(o *snmpRequestOptions) {
		o.responseTimeout = duration
	}
}

func WithSnmpPort(port int) snmpRequestOptionFunc {
	return func(o *snmpRequestOptions) {
		o.port = port
	}
}

func WithSnmpCommunity(community string) snmpRequestOptionFunc {
	return func(o *snmpRequestOptions) {
		o.community = community
	}
}


func applySnmpRequestOptions(options ...snmpRequestOptionFunc) *snmpRequestOptions {
	opts := defaultSnmpRequestOptions()
	for _, f := range options {
		f(opts)
	}
	return opts
}

func snmpClient(
	addr netip.Addr,
	community string,
	port int,
	timeout time.Duration,
) (*gosnmp.GoSNMP, error) {
	client := &gosnmp.GoSNMP{
		Target:    addr.String(),
		Port:      uint16(port),
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   timeout,
	}
	err := client.Connect()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func snmpGetSingle(
	addr netip.Addr,
	oid string,
	community string,
	port int,
	timeout time.Duration,
) (any, error) {
	client, err := snmpClient(addr, community, port, timeout)
	if err != nil {
		return "", err
	}
	defer client.Conn.Close()
	res, err := client.Get([]string{oid})
	// if err != nil {
	// 	if strings.Contains(err.Error(), "request timeout") {
	// 		return "", ErrNoResponseFromRemote
	// 	}
	// 	if strings.Contains(err.Error(), "connection refused") {
	// 		return "", ErrConnectionRefused
	// 	}
	// 	return "", err
	// }
	err = snmpErrCheck(err)
	if err != nil {
		return "", err
	}
	if res.Error != 0 {
		return "", errors.New(res.Error.String())
	}
	if len(res.Variables) > 0 {
		return res.Variables[0].Value, nil
	}

	return nil, ErrEmptyResponse
}

func snmpErrCheck(e error) error {
	if e == nil {
		return e
	}
	errstr := e.Error()
	if strings.Contains(errstr, "request timeout") {
		return ErrNoResponseFromRemote
	}
	if strings.Contains(errstr, "connection refused") {
		return ErrConnectionRefused
	}
	return e
}

func snmpGetSingleString(
	addr netip.Addr,
	oid string,
	community string,
	port int,
	timeout time.Duration,
) (string, error) {
	val, err := snmpGetSingle(addr, oid, community, port, timeout)
	if err != nil {
		return "", err
	}
	return string(val.([]byte)), nil
}
