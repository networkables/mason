// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

import (
	"errors"
	"net"
	"net/netip"
)

type errEmptyResponse struct{}

func (e errEmptyResponse) Error() string { return "empty response" }

type errNoResponseFromRemote struct{}

func (e errNoResponseFromRemote) Error() string { return "no response from remote" }

type errInvalidAddr struct{}

func (e errInvalidAddr) Error() string { return "invalid addr" }

type errIPv6Unsupported struct{}

func (e errIPv6Unsupported) Error() string { return "IPv6 is unsupported" }

var (
	ErrEmptyResponse        = errEmptyResponse{}
	ErrNoResponseFromRemote = errNoResponseFromRemote{}
	ErrInvalidAddr          = errInvalidAddr{}
	ErrIPv6Unsupported      = errIPv6Unsupported{}
	broadcastMAC            = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	unknownMAC              = net.HardwareAddr{0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

	// ErrNoResponse           = errors.New("no response from target")
	ErrConnectionRefused    = errors.New("connection refused")
	ErrTTLExceeded          = errors.New("ttl exceeded")
	ErrRandomizedMacAddress = errors.New("randomized mac address")

	ErrNoDnsNames = errors.New("no dns names")

	ErrInvalidPortListString = errors.New("invalid port list string")
)

type ErrNoResponseW struct {
	Target netip.Addr
	OpErr  *net.OpError
	Err    error
}

func (e ErrNoResponseW) Error() string {
	x := ""
	if e.OpErr != nil {
		x += " " + e.OpErr.Error()
	}
	return "no response from " + e.Target.String() + x
}

func newErrNoResponse(target netip.Addr, err *net.OpError) ErrNoResponseW {
	return ErrNoResponseW{
		Target: target,
		OpErr:  err,
		Err:    ErrNoResponseFromRemote,
	}
}

func (e ErrNoResponseW) Unwrap() error {
	return e.Err
}
