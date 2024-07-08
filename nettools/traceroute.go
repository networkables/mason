// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

import (
	"context"
	"net/netip"
	"time"
)

func Traceroute4(ctx context.Context, target netip.Addr, opts ...Icmp4EchoOption) ([][]Icmp4EchoResponse, error) {
	return DefaultPkg.Traceroute4(ctx, target, opts...)
}

func (p *pkg) Traceroute4(ctx context.Context, target netip.Addr, opts ...Icmp4EchoOption) ([][]Icmp4EchoResponse, error) {
	traceopt := i4eApplyOptionsToDefault(opts...)
	traceopt = i4eApplyOptions(traceopt, I4EWithAllowAllErrors(true), I4EWithCount(5))
	hops := 20
	response := make([][]Icmp4EchoResponse, 0, hops)
	for i := 0; i < hops; i++ {
		hopr := make([]Icmp4EchoResponse, 0, traceopt.Count)
		for c := 0; c < traceopt.Count; c++ {
			// Can only use privileged ping for traceroute
			r, err := rawPingIcmp4(ctx, target, i+1, traceopt.ListenAddress, traceopt.ReadTimeout, traceopt.IcmpID, traceopt.IcmpSeq, traceopt.AllowAllErrors)
			hopr = append(hopr, r)
			if err == nil {
				i = hops
			}
		}
		response = append(response, hopr)
		time.Sleep(traceopt.BetweenDuration)
	}
	return response, nil
}
