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

func (w WUI) wuiToolTracerouteHandler(wr http.ResponseWriter, r *http.Request) {
	content := h.Main(
		h.ID("maincontent"),
		h.Class("drawer-content"),
		w.wuiToolTraceroute(nil, nil),
	)
	w.basePage("traceroute", content, nil).Render(wr)
}

func (w WUI) wuiToolTraceroute(tr []nettools.Icmp4EchoResponseStatistics, err error) g.Node {
	return grid("traceroutecontent",
		wuiCard("Traceroute",
			h.Div(
				errAlert(err),
				h.FormEl(
					hx.Post(urlApiTraceroute),
					hx.Target("#traceroutecontent"),
					hx.Swap("outerHTML"),
					h.Div(
						h.Class("form-control"),
						wuiFormInput(
							"Target",
							h.Input(
								h.Type("text"),
								h.Name(wuiToolTarget),
								h.Placeholder("192.168.1.1 or host.name"),
								h.Class("input-bordered w-1/2"),
							),
						),
						wuiFormButton("Traceroute"),
					),
				),
			),
		),
		wuiTracerouteResultTable(tr),
	)
}

func (w WUI) wuiApiToolTracerouteHandler(wr http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	target := r.PostFormValue(wuiToolTarget)
	tr, err := w.m.Traceroute(ctx, target)
	if err != nil {
		log.Error("wuiApiToolTracerouteHandler", "error", err)
	}
	w.wuiToolTraceroute(tr, err).Render(wr)
}

func wuiTracerouteResultTable(tr []nettools.Icmp4EchoResponseStatistics) g.Node {
	if len(tr) == 0 {
		return nil
	}
	x := 0
	return wuiCard("Traceroute Results",
		wuiTable([]string{"Hop", "Peer", "PacketLoss", "Mean", "Max"},
			g.Group(
				g.Map(tr, func(hop nettools.Icmp4EchoResponseStatistics) g.Node {
					x += 1
					return h.Tr(
						h.Td(g.Text(fmt.Sprintf("%d", x))),
						h.Td(g.Text(hop.Peer.String())),
						h.Td(g.Text(fmt.Sprintf("%3.0f %%", hop.PacketLoss))),
						h.Td(g.Text(fmtDur(hop.Mean))),
						h.Td(g.Text(fmtDur(hop.Maximum))),
					)
				}),
			),
		),
	)
}
