// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"context"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/nettools"
)

func (w WUI) wuiToolTLSHandler(wr http.ResponseWriter, r *http.Request) {
	content := h.Main(
		h.ID("maincontent"),
		h.Class("drawer-content"),
		w.wuiToolTLS(nil, nil),
	)
	w.basePage("tls", content, nil).Render(wr)
}

func (w WUI) wuiToolTLS(info *nettools.TLS, err error) g.Node {
	return grid("tlscontent",
		wuiCard("TLS",
			h.Div(
				errAlert(err),
				h.FormEl(
					hx.Post(urlApiTLS),
					hx.Target("#tlscontent"),
					hx.Swap("outerHTML"),
					h.Div(
						h.Class("form-control"),
						wuiFormInput(
							"Target",
							h.Input(
								h.Type("text"),
								h.Name(wuiToolTarget),
								h.Placeholder("https://host.name"),
								h.Class("input-bordered w-1/2"),
							),
						),
						wuiFormButton("Fetch TLS"),
					),
				),
			),
		),
		wuiTLSResultTable(info),
	)
}

func (w WUI) wuiApiToolTLSHandler(wr http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	target := r.PostFormValue(wuiToolTarget)
	info, err := w.m.FetchTLSInfo(ctx, target)
	if err != nil {
		log.Error("wuiApiToolTLSHandler", "error", err)
	}
	w.wuiToolTLS(&info, err).Render(wr)
}

func wuiTLSResultTable(info *nettools.TLS) g.Node {
	if info == nil {
		return nil
	}
	return g.Group(
		[]g.Node{
			wuiCard("Certificate Information",
				wuiTable([]string{" ", " "},
					toTD("Name", info.CommonName),
					toTD("Server Name", info.ServerName),
					toTD("Valid", fmt.Sprintf("%t", info.IsValid)),
					toTD("Days Remaining", fmt.Sprintf("%d", info.DaysTilExpire)),
					toTD("Issued By", info.IssuedBy),
					toTD("Version", info.Version),
				),
			),
			wuiCard("Certificate Chain",
				wuiTable([]string{"Name", "Valid", "Expiration", "Issued By", "Version"},
					g.Group(
						g.Map(info.Chain, func(cert nettools.CertData) g.Node {
							return h.Tr(
								h.Td(g.Text(cert.CommonName)),
								h.Td(g.Text(fmt.Sprintf("%t", cert.IsValid))),
								h.Td(g.Text(fmt.Sprintf("%d", cert.DaysTilExpire))),
								h.Td(g.Text(cert.IssuedBy)),
								h.Td(g.Text(cert.Version)),
							)
						}),
					),
				),
			),
		},
	)
}
