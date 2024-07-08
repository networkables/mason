// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package netflows

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	ipfixVersion uint16 = 10
)

var ntpStart = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

type IpfixHeader struct {
	DataSize            int
	ExportedAt          time.Time
	SequenceNumber      int
	ObservationDomainID int
}

func handlePacket(dat []byte) (flows []RawFlow, err error) {
	idx := 0

	hdrsize := 16
	if len(dat) < hdrsize {
		return flows, errors.New("data is smaller than ipfix header")
	}

	hdr, err := parseHeader(dat[idx:hdrsize])
	if err != nil {
		return flows, err
	}
	idx += hdrsize

	flows, err = parseSets(hdr.ObservationDomainID, dat[idx:hdr.DataSize])
	if err != nil {
		return flows, err
	}

	return flows, nil
}

func parseHeader(dat []byte) (hdr IpfixHeader, err error) {
	ver := binary.BigEndian.Uint16(dat[0:2])
	if ver != ipfixVersion {
		return hdr, fmt.Errorf("wrong ipfix header version: %d", ver)
	}

	size := binary.BigEndian.Uint16(dat[2:4]) - 4
	hdr.DataSize = int(size)

	ts := binary.BigEndian.Uint32(dat[4:8])
	hdr.ExportedAt = time.Unix(int64(ts), 0)

	seq := binary.BigEndian.Uint32(dat[8:12])
	hdr.SequenceNumber = int(seq)

	oid := binary.BigEndian.Uint32(dat[12:16])
	hdr.ObservationDomainID = int(oid)

	// fmt.Printf("hdr: %+v\n", hdr)
	return hdr, nil
}

var ErrSetHeaderTooSmall = errors.New("set header too small")

func parseSets(obsid int, dat []byte) (flows []RawFlow, err error) {
	size := len(dat)
	idx := 0
	hdrsize := 4

	for idx < size {
		hdr, err := parseSetHeader(dat[idx : idx+hdrsize])
		if err != nil {
			if err == ErrSetHeaderTooSmall {
				return flows, nil
			}
			return flows, err
		}
		// fmt.Printf("sethdr: %+v\n", hdr)
		idx += hdrsize
		flowz, err := parseSet(hdr, dat[idx:idx+hdr.DataLength])
		if err != nil {
			return flows, err
		}
		flows = append(flows, flowz...)
		idx += hdr.DataLength
	}

	return flows, nil
}

type SetHeader struct {
	ID         int
	DataLength int
}

func parseSetHeader(dat []byte) (hdr SetHeader, err error) {
	if len(dat) < 4 {
		return hdr, ErrSetHeaderTooSmall
	}

	id := binary.BigEndian.Uint16(dat[0:2])
	hdr.ID = int(id)
	length := binary.BigEndian.Uint16(dat[2:4]) - 4
	hdr.DataLength = int(length)

	return hdr, nil
}

const (
	TemplateSetDef       = 2
	OptionTemplateSetDef = 3
)

var (
	v2templates  = map[int]TemplateDef{}
	templateLock sync.Mutex
)

func parseSet(hdr SetHeader, dat []byte) (flows []RawFlow, err error) {
	switch {
	case hdr.ID == TemplateSetDef:
		t, err := parseTemplateDef(dat[0:hdr.DataLength])
		if err != nil {
			return flows, err
		}
		// TODO: need to incorporate ObservationDomainID, what about a bit shift of obsid up above the template id range?
		templateLock.Lock()
		v2templates[t.ID] = t
		templateLock.Unlock()
	case hdr.ID == OptionTemplateSetDef:
	case hdr.ID > 255:
		templateLock.Lock()
		t, ok := v2templates[hdr.ID]
		templateLock.Unlock()
		if ok {
			flows = parseFlow(t, dat)
			// ff := rawsToIpFlows(rf)
			// for i, f := range ff {
			// 	fmt.Printf("  ff%02d: %+v\n", i, f)
			// }
			// } else {
			// 	fmt.Printf("warn: unknown template for data parse: %d\n", hdr.ID)
		}
	default:
		return flows, fmt.Errorf("unknown set header id: %d", hdr.ID)
	}
	return flows, nil
}

func nanosecondsFromFlowTime(dat []byte) time.Time {
	u64 := binary.BigEndian.Uint64(dat)
	frac := time.Duration(u64&0xffffffff) * time.Second
	nsec := frac >> 32
	if uint32(frac) >= 0x80000000 {
		nsec++
	}
	return ntpStart.Add((time.Duration(u64>>32) * time.Second) + nsec)
}

type TemplateDef struct {
	ID     int
	Fields []FieldDef
}

type FieldDef struct {
	ID               uint16
	DataLength       int
	EnterpriseNumber uint32
}

func parseTemplateDef(dat []byte) (t TemplateDef, err error) {
	idx := 0
	id := binary.BigEndian.Uint16(dat[0:2])
	count := binary.BigEndian.Uint16(dat[2:4])
	idx += 4
	t.ID = int(id)
	t.Fields = make([]FieldDef, count)

	for i := 0; i < int(count); i++ {
		var fd FieldDef
		fd.ID = binary.BigEndian.Uint16(dat[idx : idx+2])
		idx += 2
		length := binary.BigEndian.Uint16(dat[idx : idx+2])
		fd.DataLength = int(length)
		idx += 2
		if fd.ID>>15 > 0 {
			fd.EnterpriseNumber = binary.BigEndian.Uint32(dat[idx : idx+4])
			idx += 4
		}
		t.Fields[i] = fd
	}
	return t, nil
}

type RawField struct {
	ID   uint16
	Data []byte
}

type RawFlow struct {
	Fields []RawField
}

func parseFlow(t TemplateDef, dat []byte) (flows []RawFlow) {
	idx := 0
	length := len(dat)

	for idx < length {
		var flow RawFlow
		flow.Fields = make([]RawField, len(t.Fields))
		for i, field := range t.Fields {
			flow.Fields[i].ID = field.ID
			flow.Fields[i].Data = make([]byte, field.DataLength)
			copy(flow.Fields[i].Data, dat[idx:idx+field.DataLength])
			idx += field.DataLength
		}
		flows = append(flows, flow)
	}
	return flows
}

func dumpByteSlice(b []byte) (str string) {
	var a [16]byte
	n := (len(b) + 15) &^ 15
	for i := 0; i < n; i++ {
		if i%16 == 0 {
			str += fmt.Sprintf("%4d", i)
		}
		if i%8 == 0 {
			str += fmt.Sprint(" ")
		}
		if i < len(b) {
			str += fmt.Sprintf(" %02X", b[i])
		} else {
			str += fmt.Sprint("   ")
		}
		if i >= len(b) {
			a[i%16] = ' '
		} else if b[i] < 32 || b[i] > 126 {
			a[i%16] = '.'
		} else {
			a[i%16] = b[i]
		}
		if i%16 == 15 {
			str += fmt.Sprintf("  %s\n", string(a[:]))
		}
	}
	return str
}
