// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

import (
	"crypto/tls"
	"time"
  "strconv"
)

var _ TLSer = (*pkg)(nil)

type TLSer interface {
	FetchTLS(string) (TLS, error)
}

type TLS struct {
	ServerName    string
	CommonName    string
	Version       string
	IsValid       bool
	DaysTilExpire int
	IssuedBy      string
	Chain         []CertData
	TLSData       *tls.ConnectionState
}

type CertData struct {
	CommonName    string
	Version       string
	IsValid       bool
	DaysTilExpire int
	IssuedBy      string
}

func FetchTLS(url string) (TLS, error) {
	return DefaultPkg.FetchTLS(url)
}

func (p *pkg) FetchTLS(url string) (t TLS, err error) {
	resp, err := p.httpclient.Get(url)
	if err != nil {
		return t, err
	}
	defer resp.Body.Close()
	return buildTLS(resp.TLS), nil

}

func buildTLS(x *tls.ConnectionState) TLS {
	t := TLS{
		TLSData: x,
	}
	if x == nil {
		return t
	}
	t.ServerName = x.ServerName
	peercount := len(x.PeerCertificates)
	if peercount > 0 {
		t.IssuedBy = x.PeerCertificates[0].Issuer.CommonName
		t.CommonName = x.PeerCertificates[0].Subject.CommonName
		t.IsValid = x.PeerCertificates[0].BasicConstraintsValid
		t.DaysTilExpire = int(x.PeerCertificates[0].NotAfter.Sub(time.Now()).Hours() / 24)
	}
	t.Version = tls.VersionName(x.Version)
	t.Chain = make([]CertData, 0, peercount)
	for i := 0; i < peercount; i++ {
		t.Chain = append(t.Chain, CertData{
			CommonName:    x.PeerCertificates[i].Subject.CommonName,
			Version:       strconv.Itoa(x.PeerCertificates[i].Version),
			IsValid:       x.PeerCertificates[i].BasicConstraintsValid,
			DaysTilExpire: int(x.PeerCertificates[i].NotAfter.Sub(time.Now()).Hours() / 24),
			IssuedBy:      x.PeerCertificates[i].Issuer.CommonName,
		})
	}
	return t
}
