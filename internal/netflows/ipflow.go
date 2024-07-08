// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package netflows

import (
	"encoding/binary"
	"net/netip"

	"github.com/networkables/mason/internal/model"
)

func addrFromSlice(s []byte) model.Addr {
	x, _ := netip.AddrFromSlice(s)
	return model.AddrToModelAddr(x)
}

func rawToIpFlow(dat RawFlow) (f model.IpFlow) {
	for _, field := range dat.Fields {
		switch field.ID {
		case IPFIX_FIELD_sourceIPv4Address:
			f.SrcAddr = addrFromSlice(field.Data)
		case IPFIX_FIELD_destinationIPv4Address:
			f.DstAddr = addrFromSlice(field.Data)
		case IPFIX_FIELD_sourceTransportPort:
			f.SrcPort = binary.BigEndian.Uint16(field.Data)
		case IPFIX_FIELD_destinationTransportPort:
			f.DstPort = binary.BigEndian.Uint16(field.Data)
		case IPFIX_FIELD_flowStartNanoseconds:
			f.Start = nanosecondsFromFlowTime(field.Data)
		case IPFIX_FIELD_flowEndNanoseconds:
			f.End = nanosecondsFromFlowTime(field.Data)
		case IPFIX_FIELD_octetDeltaCount:
			f.Bytes = int(binary.BigEndian.Uint32(field.Data))
		case IPFIX_FIELD_packetDeltaCount:
			f.Packets = int(binary.BigEndian.Uint32(field.Data))
		case IPFIX_FIELD_protocolIdentifier:
			f.Protocol = model.Protocol(field.Data[0])
		case IPFIX_FIELD_tcpControlBits:
			f.Flags = model.TcpFlags(field.Data[0])
		}
	}

	return f
}

// func addAsns(in IpFlow) IpFlow {
// 	in.SrcASN, _ = asndb.FindAsn(in.SrcAddr)
// 	in.DstASN, _ = asndb.FindAsn(in.DstAddr)
// 	return in
// }

func rawsToIpFlows(flows []RawFlow) (f []model.IpFlow) {
	f = make([]model.IpFlow, len(flows))
	for i, flow := range flows {
		f[i] = rawToIpFlow(flow)
	}
	return f
}
