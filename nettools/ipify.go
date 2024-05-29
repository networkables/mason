// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

import (
	"context"
	"io"
	"net/http"
	"net/netip"
)

var _ Ipifyer = (*pkg)(nil)

type Ipifyer interface {
	GetExternalAddr(context.Context) (netip.Addr, error)
}

func GetExternalAddr(ctx context.Context) (netip.Addr, error) {
	return DefaultPkg.GetExternalAddr(ctx)
}

func (p *pkg) GetExternalAddr(ctx context.Context) (netip.Addr, error) {
	str, err := getExternalIPString(p.ipifyUrl, p.GetUserAgent())
	if err != nil {
		return netip.Addr{}, err
	}
	addr, err := netip.ParseAddr(str)
	if err != nil {
		return addr, err
	}
	return addr, nil
}

func getExternalIPString(url string, useragent string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", useragent)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
