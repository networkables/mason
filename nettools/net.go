// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/netip"
	"time"

	"math/rand/v2"

	"github.com/charmbracelet/log"
	"github.com/miekg/dns"
	"golang.org/x/net/ipv4"
)

var _ Nettooler = (*pkg)(nil)

type Nettooler interface {
	Arper
	Dnser
	Icmp4Echoer
	Ipifyer
	Portscanner
	Snmper
	TLSer
	Utiler
}

const (
	google        = "GOOGLE"
	cloudflare    = "CLOUDFLARE"
	controld      = "CONTROLD"
	quad9         = "QUAD9"
	opendns       = "OPENDNS"
	adguard       = "ADGUARD"
	cleanbrowsing = "CLEANBROWSING"
	alternatedns  = "ALTERNATEDNS"

	dnssvrGoogleA        = "8.8.8.8:53"
	dnssvrGoogleB        = "8.8.4.4:53"
	dnssvrCloudflareA    = "1.1.1.1:53"
	dnssvrCloudflareB    = "1.0.0.1:53"
	dnssvrControlDA      = "76.76.2.0:53"
	dnssvrControlDB      = "76.76.10.0:53"
	dnssvrQuad9A         = "9.9.9.9:53"
	dnssvrQuad9B         = "149.112.112.112:53"
	dnssvrOpenDNSA       = "208.67.222.222:53"
	dnssvrOpenDNSB       = "208.67.220.220:53"
	dnssvrAdGuardA       = "94.140.14.14:53"
	dnssvrAdGuardB       = "94.140.15.15:53"
	dnssvrCleanBrowsingA = "185.228.168.9:53"
	dnssvrCleanBrowsingB = "185.228.169.9:53"
	dnssvrAlternateDNSA  = "76.76.19.19:53"
	dnssvrAlternateDNSB  = "76.223.112.150:53"

	ProtocolICMP     = 1
	ProtocolIPv6ICMP = 58

	defaultIpifyUrl = "https://api64.ipify.org"
)

type pkg struct {
	// TODO: need better processing / usage patterns for underlying hardware & network routing
	// would like to have data at the ready when using default route gateway
	// - should be able to have a single method call given a target addr that would return
	//    everything we need to listen and/or send packets
	ifacesByMACString map[string]*net.Interface
	ifacesByName      map[string]*net.Interface
	ifacesByAddr      map[netip.Addr]*net.Interface
	ifacesByNetPrefix map[string]*net.Interface

	defaultRouteIface   *net.Interface
	defaultRouteGateway netip.Addr

	arptable map[netip.Addr]net.HardwareAddr

	dnsclient  *dns.Client
	httpclient *http.Client

	dnssvrs map[string]map[string]string

	ipifyUrl string

	userAgent string

	ouicache map[string]string
}

var DefaultPkg *pkg

func init() {
	var err error
	DefaultPkg, err = newPkg()
	if err != nil {
		log.Fatal("nettools pkg init", "error", err)
	}
}

func newPkg() (*pkg, error) {
	p := &pkg{
		arptable: make(map[netip.Addr]net.HardwareAddr),
		dnsclient: &dns.Client{
			Timeout: 5 * time.Second,
		},
		httpclient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		dnssvrs: map[string]map[string]string{
			google: {
				"A": dnssvrGoogleA,
				"B": dnssvrGoogleB,
			},
			cloudflare: {
				"A": dnssvrCloudflareA,
				"B": dnssvrCloudflareB,
			},
			controld: {
				"A": dnssvrControlDA,
				"B": dnssvrControlDB,
			},
			quad9: {
				"A": dnssvrQuad9A,
				"B": dnssvrQuad9B,
			},
			opendns: {
				"A": dnssvrOpenDNSA,
				"B": dnssvrOpenDNSB,
			},
			// adguard: {
			// 	"A": dnssvrAdGuardA,
			// 	"B": dnssvrAdGuardB,
			// },
			cleanbrowsing: {
				"A": dnssvrCleanBrowsingA,
				"B": dnssvrCleanBrowsingB,
			},
			// alternatedns: {
			// 	"A": dnssvrAlternateDNSA,
			// 	"B": dnssvrAlternateDNSB,
			// },
		},
		ipifyUrl: defaultIpifyUrl,
		ouicache: make(map[string]string),
	}
	p.GetUserAgent()
	err := p.loadInterfaces()
	if err != nil {
		return nil, err
	}
	p.loadRoutes()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func convertToAddr(ip net.IP) (addr netip.Addr, err error) {
	addr, ok := netip.AddrFromSlice(ip)
	if ok {
		return addr, nil
	}
	return addr, ErrInvalidAddr
}

func (p *pkg) loadInterfaces() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	p.ifacesByMACString = make(map[string]*net.Interface)
	p.ifacesByName = make(map[string]*net.Interface)
	p.ifacesByAddr = make(map[netip.Addr]*net.Interface)

	for _, iface := range ifaces {
		p.ifacesByMACString[iface.HardwareAddr.String()] = &iface
		p.ifacesByName[iface.Name] = &iface
		ips, err := iface.Addrs()
		if err != nil {
			return err
		}
		for _, ip := range ips {
			prefix, err := netip.ParsePrefix(ip.String())
			if err != nil {
				return err
			}
			p.ifacesByAddr[prefix.Addr()] = &iface
		}
	}

	return nil
}

func (p pkg) getDefaultInterface() *net.Interface {
	return p.defaultRouteIface
}

func (p pkg) bestInterface(target netip.Addr) (iface net.Interface, addr netip.Addr) {
	for prefixstr, ifacep := range p.ifacesByNetPrefix {
		prefix, err := netip.ParsePrefix(prefixstr)
		if err != nil {
			// This should not happen, since the table is loaded from system interface
			log.Fatal("unexpected bad prefix", "prefixstr", prefixstr)
		}
		if prefix.Contains(target) {
			return *ifacep, p.addrOfIface(*ifacep, target.Is4())
		}
	}
	return *p.getDefaultInterface(), p.addrOfIface(*p.getDefaultInterface(), target.Is4())
}

func (p pkg) addrOfIface(target net.Interface, isipv4 bool) netip.Addr {
	for addr, iface := range p.ifacesByAddr {
		if iface.Name == target.Name && isipv4 == addr.Is4() {
			return addr
		}
	}
	return netip.Addr{}
}

var noControlMessage *ipv4.ControlMessage = nil

var rander = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano()/2)))
