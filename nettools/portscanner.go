// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

import (
	"context"
	"github.com/charmbracelet/log"
	"github.com/networkables/mason/internal/workerpool"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"
)

var _ Portscanner = (*pkg)(nil)

type Portscanner interface {
	ScanTcpPorts(context.Context, netip.Addr, ...portscanRequestOptionFunc) ([]int, error)
}

func ScanTcpPorts(ctx context.Context, target netip.Addr, options ...portscanRequestOptionFunc) (ports []int, err error) {
	return DefaultPkg.ScanTcpPorts(ctx, target, options...)
}

func (p *pkg) ScanTcpPorts(ctx context.Context, target netip.Addr, options ...portscanRequestOptionFunc) (ports []int, err error) {
	opts := applyPortscanRequestOptions(options...)
	portsToCheck := make(chan int)
	var wg sync.WaitGroup
	openports := make([]int, 0)

	go func() {
		for _, port := range getPortNumbers(opts.portlist) {
			portsToCheck <- port
		}
		close(portsToCheck)
	}()
	wp := workerpool.New(
		"portscanner",
		portsToCheck,
		buildTcpPortChecker(target, opts.responseTimeout),
	)
	wg.Add(2)
	go func() {
		for port := range wp.C {
			if port != 0 {
				openports = append(openports, port)
			}
		}
		wg.Done()
	}()
	go func() {
		for err := range wp.E {
			log.Error(err) // TODO: surface these errors back up
		}
		wg.Done()
	}()

	wp.Run(ctx, opts.maxWorkers)
	wg.Wait()
	return openports, nil
}

func buildTcpPortChecker(addr netip.Addr, timeout time.Duration) func(int) (int, error) {
	return func(port int) (int, error) {
		isopen, err := isTcpPortOpen(addr, port, timeout)
		if isopen {
			return port, nil
		}
		return 0, err
	}
}

func isTcpPortOpen(addr netip.Addr, port int, timeout time.Duration) (bool, error) {
	c, err := net.DialTimeout("tcp", net.JoinHostPort(addr.String(), strconv.Itoa(port)), timeout)
	if err == nil {
		c.Close()
		return true, nil
	}
	neterr, ok := err.(net.Error)
	if ok && neterr.Timeout() {
		// Don't care about surfacing timeouts
		return false, nil
	}
	if strings.HasSuffix(err.Error(), "connection refused") {
		// Don't caare about surfacing connection refusals
		return false, nil
	}
	// log.Error("port scan dial", "error", err, "addr", addr, "port", port)

	return false, err
}

//
//
// Options available for Portscans
//

type portscanRequestOptions struct {
	responseTimeout time.Duration
	maxWorkers      int
	portlist        PortList
}

func defaultPortscanRequestOptions() *portscanRequestOptions {
	return &portscanRequestOptions{
		responseTimeout: 100 * time.Millisecond,
		maxWorkers:      1,
		portlist:        GeneralPorts,
	}
}

func WithPortscanReplyTimeout(duration time.Duration) portscanRequestOptionFunc {
	return func(o *portscanRequestOptions) {
		o.responseTimeout = duration
	}
}

func WithPortscanMaxworkers(ct int) portscanRequestOptionFunc {
	return func(o *portscanRequestOptions) {
		o.maxWorkers = ct
	}
}

func WithPortscanPortlist(list PortList) portscanRequestOptionFunc {
	return func(o *portscanRequestOptions) {
		o.portlist = list
	}
}

func WithPortscanPortlistName(name string) portscanRequestOptionFunc {
	list, err := stringToPortList(name)
	if err != nil {
		log.Fatal(err)
	}
	return func(o *portscanRequestOptions) {
		o.portlist = list
	}
}

type portscanRequestOptionFunc func(*portscanRequestOptions)

func applyPortscanRequestOptions(options ...portscanRequestOptionFunc) *portscanRequestOptions {
	opts := defaultPortscanRequestOptions()
	for _, f := range options {
		f(opts)
	}
	return opts
}

type PortList int

func (p PortList) String() string {
	switch p {
	case AllPorts:
		return "All"
	case GeneralPorts:
		return "General"
	case PriviledgedPorts:
		return "Priviledged"
	case CommonPorts:
		return "Common"
	}
	return "INVALID"
}

const (
	InvalidPortList PortList = iota
	AllPorts
	GeneralPorts
	PriviledgedPorts
	CommonPorts
)

func stringToPortList(str string) (PortList, error) {
	switch strings.ToLower(str) {
	case "all":
		return AllPorts, nil
	case "general":
		return GeneralPorts, nil
	case "priviledged":
		return PriviledgedPorts, nil
	case "common":
		return CommonPorts, nil
	}
	return InvalidPortList, ErrInvalidPortListString
}

func getPortNumbers(list PortList) []int {
	switch list {
	case AllPorts:
		return allPorts()
	case GeneralPorts:
		return generalPorts()
	case PriviledgedPorts:
		return priviledgedPorts()
	case CommonPorts:
		return commonPorts()
	}
	return []int{}
}

func allPorts() []int {
	ports := make([]int, 65536)
	for i := range ports {
		ports[i] = i + 1
	}
	return ports
}

func generalPorts() []int {
	ports := make([]int, 10000)
	for i := range ports {
		ports[i] = i + 1
	}
	return ports
}

func priviledgedPorts() []int {
	ports := make([]int, 1024)
	for i := range ports {
		ports[i] = i + 1
	}
	return ports
}

func commonPorts() []int {
	return []int{
		20,   // FTP
		21,   // FTP
		22,   // SSH
		23,   // Telnet
		25,   // SMTP
		53,   // DNS
		80,   // HTTP
		110,  // POP
		123,  // NTP
		143,  // IMAP
		145,  // ?
		161,  // SNMP
		162,  // SNMP
		443,  // HTTPS
		465,  // SMTPS
		554,  // RTSP
		585,  // IMAPS
		587,  // SMTP
		993,  // IMAP4S
		995,  // POP3S
		1433, // MSSQL
		3306, // MySQL
		3389, // RDC
		6789, // Unifi
		8080, // HTTP
		8443, // HTTPS
		8843, // HTTPS
	}
}
