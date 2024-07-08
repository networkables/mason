// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

import (
	"context"
	"errors"
	"math"
	"net"
	"net/netip"
	"runtime"
	"time"

	"golang.org/x/net/bpf"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var _ Icmp4Echoer = (*pkg)(nil)

type Icmp4Echoer interface {
	Icmp4Echo(context.Context, netip.Addr, ...Icmp4EchoOption) ([]Icmp4EchoResponse, error)
}

func Icmp4Echo(ctx context.Context, target netip.Addr, opts ...Icmp4EchoOption) ([]Icmp4EchoResponse, error) {
	return DefaultPkg.Icmp4Echo(ctx, target, opts...)
}

func (p *pkg) Icmp4Echo(ctx context.Context, target netip.Addr, opts ...Icmp4EchoOption) ([]Icmp4EchoResponse, error) {
	opt := i4eApplyOptionsToDefault(opts...)
	response := make([]Icmp4EchoResponse, 0, opt.Count)
	for seqNum := opt.IcmpSeq; seqNum < opt.Count+1; seqNum++ {
		var (
			err error
			r   Icmp4EchoResponse
		)
		if opt.Privileged {
			r, err = rawPingIcmp4(ctx, target, opt.TTL, opt.ListenAddress, opt.ReadTimeout, opt.IcmpID, seqNum, opt.AllowAllErrors)
		} else {
			r, err = rawPingUdp4(ctx, target, opt.TTL, opt.ListenAddress, opt.ReadTimeout, opt.IcmpID, seqNum, opt.AllowAllErrors)
		}
		response = append(response, r)
		if err != nil {
			return response, err
		}
		time.Sleep(opt.BetweenDuration)
	}
	return response, nil
}

type Icmp4EchoOptions struct {
	Privileged      bool
	ListenAddress   netip.Addr
	ReadTimeout     time.Duration
	IcmpID          int
	IcmpSeq         int
	TTL             int
	Count           int
	BetweenDuration time.Duration
	AllowAllErrors  bool
}

type Icmp4EchoOption func(*Icmp4EchoOptions)

func i4eApplyOptionsToDefault(opts ...Icmp4EchoOption) *Icmp4EchoOptions {
	o := defaultIcmp4EchoOptions()
	return i4eApplyOptions(o, opts...)
}

func i4eApplyOptions(base *Icmp4EchoOptions, opts ...Icmp4EchoOption) *Icmp4EchoOptions {
	for _, f := range opts {
		f(base)
	}
	return base
}

func I4EWithListenAddress(target netip.Addr) Icmp4EchoOption {
	return func(o *Icmp4EchoOptions) {
		o.ListenAddress = target
	}
}

func I4EWithReadTimeout(to time.Duration) Icmp4EchoOption {
	return func(o *Icmp4EchoOptions) {
		o.ReadTimeout = to
	}
}

func I4EWithIcmpID(id int) Icmp4EchoOption {
	return func(o *Icmp4EchoOptions) {
		o.IcmpID = id
	}
}

func I4EWithIcmpSeq(seq int) Icmp4EchoOption {
	return func(o *Icmp4EchoOptions) {
		o.IcmpSeq = seq
	}
}

func I4EWithTTL(ttl int) Icmp4EchoOption {
	return func(o *Icmp4EchoOptions) {
		o.TTL = ttl
	}
}

func I4EWithCount(ct int) Icmp4EchoOption {
	return func(o *Icmp4EchoOptions) {
		o.Count = ct
	}
}

func I4EWithBetweenDuration(dur time.Duration) Icmp4EchoOption {
	return func(o *Icmp4EchoOptions) {
		o.BetweenDuration = dur
	}
}

func I4EWithAllowAllErrors(allow bool) Icmp4EchoOption {
	return func(o *Icmp4EchoOptions) {
		o.AllowAllErrors = allow
	}
}

func I4EWithPrivileged(allow bool) Icmp4EchoOption {
	return func(o *Icmp4EchoOptions) {
		o.Privileged = allow
	}
}

func defaultIcmp4EchoOptions() *Icmp4EchoOptions {
	listenAddress := netip.MustParseAddr("0.0.0.0")
	icmpID := rander.Int() & 0xFFFF
	return &Icmp4EchoOptions{
		Privileged:      true,
		ListenAddress:   listenAddress,
		ReadTimeout:     100 * time.Millisecond,
		IcmpID:          icmpID,
		IcmpSeq:         1,
		TTL:             64,
		Count:           1,
		BetweenDuration: time.Second,
	}
}

type Icmp4EchoResponse struct {
	Peer    netip.Addr
	Start   time.Time
	Elapsed time.Duration
	Err     error
}

func (r Icmp4EchoResponse) populate(addr net.Addr, start time.Time, stop time.Time, err error) Icmp4EchoResponse {
	r.Start = start
	r.Elapsed = stop.Sub(start)
	r.Err = err
	switch peer := addr.(type) {
	case *net.IPAddr:
		r.Peer, _ = netip.AddrFromSlice(peer.IP)
	case *net.UDPAddr:
		r.Peer, _ = netip.AddrFromSlice(peer.IP)
	}
	return r
}

func rawPingIcmp4(ctx context.Context, target netip.Addr, ttl int, listenAddress netip.Addr, readTimeout time.Duration, icmpID int, icmpSeq int, allowAllErrors bool) (response Icmp4EchoResponse, err error) {
	listenProto := "ip4:icmp"

	if ctx.Err() != nil {
		return response, ctx.Err()
	}
	if target.Is6() {
		return response, errors.New("ipv6 not supported")
	}
	switch runtime.GOOS {
	case "darwin", "ios":
	case "linux":
		// log.Print("you may need to adjust the net.ipv4.ping_group_range kernel state")
	default:
		// log.Print("not supported on", runtime.GOOS)
		return response, errors.New("os not supported")
	}
	// log.Info("rawPing4", "target", target, "ttl",ttl, "listenAddress", listenAddress, "readTimeout", readTimeout,"icmpID",icmpID, "icmpSeq", icmpSeq)

	ln, err := net.ListenPacket(listenProto, listenAddress.String())
	if err != nil {
		return response, err
	}
	defer ln.Close()
	pc := ipv4.NewPacketConn(ln)
	if err != nil {
		return response, err
	}
	err = pc.SetTTL(ttl)
	if err != nil {
		return response, err
	}
	err = pc.SetReadDeadline(time.Now().Add(readTimeout))
	if err != nil {
		return response, err
	}
	if !allowAllErrors {

		assembled, err := buildIcmpFilterForID(uint32(icmpID))
		if err != nil {
			return response, err
		}
		err = pc.SetBPF(assembled)
		if err != nil {
			return response, err
		}
	}

	wb := buildIcmpMessageBody(icmpID, icmpSeq)

	starttime := time.Now()
	if _, err := pc.WriteTo(wb, noControlMessage, &net.IPAddr{IP: net.IP(target.AsSlice())}); err != nil {
		return response, err
	}
	rb := make([]byte, 1500)
	n, _, peer, err := pc.ReadFrom(rb)
	endtime := time.Now()
	response = response.populate(peer, starttime, endtime, err)
	if err != nil {
		operr := err.(*net.OpError)
		response.Err = operr
		if operr.Timeout() {
			return response, ErrNoResponseFromRemote
		}
		return response, err
	}

	rm, err := icmp.ParseMessage(ProtocolICMP, rb[:n])
	if err != nil {
		return response, err
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		pkt := rm.Body.(*icmp.Echo)
		if int(icmpID) != pkt.ID {
			return response, newErrNoResponse(target, nil)
		}
		if target.Compare(response.Peer) != 0 {
			return response, newErrNoResponse(target, nil)
		}
		return response, nil
	case ipv4.ICMPTypeTimeExceeded:
		return response, ErrTTLExceeded
		// default:
		//   log.Printf("got %+v; want echo reply, perr:%+v got:%T tar:%v id:%d", rm, peer, rm, target, pktid)
		// switch pkt := rm.Body.(type) {
		// case *icmp.DstUnreach:
		// 	log.Printf("\t of type %T   %+v %s", pkt, pkt, string(pkt.Data))
		// case *icmp.TimeExceeded:
		// 	log.Printf("\t of type %T   %+v %s", pkt, pkt, string(pkt.Data))
		// default:
		// 	log.Printf("\t of type %T   %+v", pkt, pkt)
		// }
	}

	return response, newErrNoResponse(target, nil)
}

func rawPingUdp4(ctx context.Context, target netip.Addr, ttl int, listenAddress netip.Addr, readTimeout time.Duration, icmpID int, icmpSeq int, allowAllErrors bool) (response Icmp4EchoResponse, err error) {
	listenProto := "udp4"

	if ctx.Err() != nil {
		return response, ctx.Err()
	}
	if target.Is6() {
		return response, errors.New("ipv6 not supported")
	}

	ln, err := icmp.ListenPacket(listenProto, listenAddress.String())
	defer ln.Close()
	err = ln.SetReadDeadline(time.Now().Add(readTimeout))
	if err != nil {
		return response, err
	}
	pc := ln.IPv4PacketConn()
	// err = pc.SetControlMessage(ipv4.FlagTTL, true)
	// if err != nil {
	// 	return response, err
	// }
	err = pc.SetTTL(ttl)
	if err != nil {
		return response, err
	}

	wb := buildIcmpMessageBody(icmpID, icmpSeq)

	starttime := time.Now()
	if _, err := ln.WriteTo(wb, &net.UDPAddr{IP: target.AsSlice()}); err != nil {
		return response, err
	}

	rb := make([]byte, 1500)
	n, _, peer, err := pc.ReadFrom(rb)
	endtime := time.Now()
	response = response.populate(peer, starttime, endtime, err)

	if err != nil {
		operr := err.(*net.OpError)
		response.Err = operr
		if operr.Timeout() {
			return response, ErrNoResponseFromRemote
		}
		return response, err
	}
	rm, err := icmp.ParseMessage(ProtocolICMP, rb[:n])
	if err != nil {
		return response, err
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		// Linux can only match IcmpID in privileged mode
		// pkt := rm.Body.(*icmp.Echo)
		// if int(icmpID) != pkt.ID {
		// 	log.Printf("identifier mismatch want: %d, got: %d", icmpID, pkt.ID)
		// 	return response, newErrNoResponse(target, nil)
		// }
		if target.Compare(response.Peer) != 0 {
			return response, newErrNoResponse(target, nil)
		}
		return response, nil
	case ipv4.ICMPTypeTimeExceeded:
		return response, ErrTTLExceeded
		// default:
		//   log.Printf("got %+v; want echo reply, perr:%+v got:%T tar:%v id:%d", rm, peer, rm, target, pktid)
		// switch pkt := rm.Body.(type) {
		// case *icmp.DstUnreach:
		// 	log.Printf("\t of type %T   %+v %s", pkt, pkt, string(pkt.Data))
		// case *icmp.TimeExceeded:
		// 	log.Printf("\t of type %T   %+v %s", pkt, pkt, string(pkt.Data))
		// default:
		// 	log.Printf("\t of type %T   %+v", pkt, pkt)
		// }
	}

	return response, newErrNoResponse(target, nil)
}

func buildIcmpMessageBody(icmpID int, icmpSeq int) []byte {

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   icmpID,
			Seq:  icmpSeq,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return wb
}

func buildIcmpFilterForID(icmpid uint32) ([]bpf.RawInstruction, error) {
	filter := []bpf.Instruction{
		// Skip to the end of the IP header
		bpf.LoadMemShift{Off: 0},
		// Load the icmp type
		bpf.LoadIndirect{Off: 0, Size: 1},
		// continue if a reply, else jump to the end and skip this packet
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x0, SkipTrue: 1},
		bpf.RetConstant{Val: 0},
		// Load the icmp id
		bpf.LoadIndirect{Off: 4, Size: 2},
		// continue if this is the icmp id we want, else jump to the end
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: icmpid, SkipTrue: 1},
		bpf.RetConstant{Val: 0},
		// this is the packet we want, return it whole
		bpf.RetConstant{Val: 1500},
	}
	return bpf.Assemble(filter)
}

type Icmp4EchoResponseStatistics struct {
	Peer         netip.Addr
	Start        time.Time
	TotalPackets int
	TotalElapsed time.Duration
	Mean         time.Duration
	Minimum      time.Duration
	Maximum      time.Duration
	StdDev       time.Duration
	SuccessCount int
	PacketLoss   float64
	Asn          string
	OrgName      string
}

func CalculateIcmp4EchoResponseStatistics(rs []Icmp4EchoResponse) (ret Icmp4EchoResponseStatistics) {
	count := len(rs)
	if count == 0 {
		return ret
	}
	ret.Peer = rs[0].Peer
	ret.TotalPackets = count
	ret.Start = rs[0].Start
	ret.Minimum = math.MaxInt64
	ret.Maximum = math.MinInt64
	for _, x := range rs {
		if x.Err != nil {
			continue
		}
		if ret.Start.After(ret.Start) {
			ret.Start = x.Start
		}
		if ret.Minimum > x.Elapsed {
			ret.Minimum = x.Elapsed
		}
		if ret.Maximum < x.Elapsed {
			ret.Maximum = x.Elapsed
		}
		ret.SuccessCount++
		ret.TotalElapsed += x.Elapsed
	}
	if ret.SuccessCount > 0 {
		ret.PacketLoss = float64(ret.TotalPackets-ret.SuccessCount) / float64(ret.SuccessCount)
		ret.Mean = ret.TotalElapsed / time.Duration(ret.SuccessCount)
		mean := float64(ret.Mean)
		variance := 0.0
		for _, x := range rs {
			if x.Err != nil {
				continue
			}
			variance += math.Pow(float64(x.Elapsed)-mean, 2)
		}
		variance /= float64(ret.SuccessCount)
		ret.StdDev = time.Duration(math.Sqrt(variance))
	} else {
		ret.Minimum = 0
		ret.Maximum = 0
	}
	return ret
}
