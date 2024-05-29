// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

import (
	"context"
	"net"
	"net/netip"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/miekg/dns"
)

var _ Dnser = (*pkg)(nil)

type Dnser interface {
	FindFirstAddrOf(string) (netip.Addr, error)
	FindAddrsOf(string) ([]netip.Addr, error)
	FindHostnameOf(netip.Addr) (string, error)
	DNSCheckAllServers(context.Context, string) (map[string]map[string][]netip.Addr, error)
}

// FindFirstAddrOf will return the netip.Addr of the first A record of the target.  If no A records are returned a ErrEmptyResponse error will be returned
func FindFirstAddrOf(target string) (addr netip.Addr, err error) {
	return DefaultPkg.FindFirstAddrOf(target)
}

func (p pkg) FindFirstAddrOf(target string) (addr netip.Addr, err error) {
	records, err := DefaultPkg.FindAddrsOf(target)
	if err != nil {
		return addr, err
	}
	if len(records) > 0 {
		return records[0], nil
	}
	return addr, ErrEmptyResponse
}

// FindAddrsOf will resolve the A records into netip.Addr's for the given target.  If no A records are returned a ErrEmptyResponse error will be returned
func FindAddrsOf(target string) (addrs []netip.Addr, err error) {
	return DefaultPkg.FindAddrsOf(target)
}

func (p *pkg) FindAddrsOf(target string) (addrs []netip.Addr, err error) {
	records, err := p.findDnsRecords(dnssvrGoogleA, target, dns.TypeA)
	if err != nil {
		return addrs, err
	}
	addrs = make([]netip.Addr, len(records))
	for i, record := range records {
		arecord, ok := record.(*dns.A)
		if ok {
			addr, err := convertToAddr(arecord.A)
			if err != nil {
				return addrs, err
			}
			addrs[i] = addr
		}
	}
	return addrs, nil
}

func (p *pkg) findDnsRecords(dnsserver string, target string, recType uint16) (records []dns.RR, err error) {
	m := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			RecursionDesired: true,
		},
	}
	m.SetQuestion(dns.Fqdn(target), recType)
	response, _, err := p.dnsclient.Exchange(m, dnsserver)
	if err != nil {
		return records, err
	}
	if response == nil || len(response.Answer) == 0 {
		return records, ErrNoResponseFromRemote
	}
	return response.Answer, nil
}

func FindHostnameOf(addr netip.Addr) (string, error) {
	return DefaultPkg.FindHostnameOf(addr)
}

func (p *pkg) FindHostnameOf(addr netip.Addr) (string, error) {
	names, err := net.LookupAddr(addr.String())
	if err != nil && !strings.HasSuffix(err.Error(), "no such host") {
		return "", err
	}
	if len(names) > 0 {
		return names[0], nil
	}
	return "", ErrNoDnsNames
}

func DNSCheckAllServers(ctx context.Context, target string) (map[string]map[string][]netip.Addr, error) {
	return DefaultPkg.DNSCheckAllServers(ctx, target)
}

func (p pkg) DNSCheckAllServers(ctx context.Context, target string) (ret map[string]map[string][]netip.Addr, err error) {
	ret = make(map[string]map[string][]netip.Addr)
	for company, servers := range p.dnssvrs {
		ret[company] = make(map[string][]netip.Addr)
		for server, addy := range servers {
			recs, err := p.findDnsRecords(addy, target, dns.TypeA)
			if err != nil {
				log.Error("finderr", "error", err, "server", server, "addy", addy)
				return ret, err
			}
			numrecs := len(recs)
			ret[company][addy] = make([]netip.Addr, numrecs)
			for idx, rec := range recs {
				arec, ok := rec.(*dns.A)
				if ok {
					ip, ok := netip.AddrFromSlice(arec.A)
					if ok {
						ret[company][addy][idx] = ip
					}
				}
			}
		}
	}
	return ret, nil
}
