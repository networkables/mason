// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"net/http"

	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/internal/network"
)

func (w WUI) wuiNetworksPageHandler(wr http.ResponseWriter, r *http.Request) {
	content := h.Main(
		h.ID("maincontent"),
		h.Class("drawer-content"),
		w.wuiNetworksMain(nil),
	)
	w.basePage("networks", content, nil).Render(wr)
}

const (
	wuiNetworksFormName    = "netname"
	wuiNetworksFormPrefix  = "netprefix"
	wuiNetworksFormScanNow = "scannow"
)

func (w *WUI) wuiNetworksApiCreate(wr http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue(wuiNetworksFormName)
	prefix := r.PostFormValue(wuiNetworksFormPrefix)
	scannowstr := r.PostFormValue(wuiNetworksFormScanNow)

	scannow := false
	if scannowstr == "on" {
		scannow = true
	}
	err := w.m.AddNetworkByName(name, prefix, scannow)

	w.wuiNetworksMain(err).Render(wr)
}

func (w WUI) wuiNetworksMain(err error) g.Node {
	var errNode g.Node
	if err != nil {
		errNode = errAlert(err)
	}
	nets := w.m.ListNetworks()
	network.SortNetworksByAddr(nets)
	return grid("networkscontent",
		wuiCard("Networks",
			networksToTable(nets),
		),
		wuiCard("Add Network",
			h.Div(
				errNode,
				h.FormEl(
					hx.Post("/api/networks"),
					hx.Target("#networkscontent"),
					hx.Swap("outerHTML"),
					h.Div(
						h.Class("form-control"),
						h.Label(
							h.Class("label"),
							h.Span(h.Class("label-text"), g.Text("Name")),
							h.Input(
								h.Type("text"),
								h.Name(wuiNetworksFormName),
								h.Placeholder("Custom Name"),
								h.Class("input input-bordered w-1/2"),
							),
						),
						h.Label(
							h.Class("label"),
							h.Span(h.Class("label-text"), g.Text("Prefix")),
							h.Input(
								h.Type("text"),
								h.Name(wuiNetworksFormPrefix),
								h.Placeholder("192.168.1.1/24"),
								h.Class("input input-bordered w-1/2"),
							),
						),
						h.Label(
							h.Class("label cursor-pointer"),
							h.Span(h.Class("label-text"), g.Text("Scan network immediately")),
							h.Input(
								h.Type("checkbox"),
								h.Name(wuiNetworksFormScanNow),
								g.Attr("checked", "checked"),
								h.Class("checkbox checkbox-primary"),
							),
						),
					),
					h.Div(
						h.Class("flex gap-4 py-4"),
						h.Button(h.Class("btn btn-primary grow"), g.Text("Add Network")),
					),
				),
			),
		),
	)
}

func networksToTable(nets []network.Network) g.Node {
	return wuiTable(
		[]string{"Name", "Prefix"},
		g.Group(
			g.Map(
				nets,
				func(n network.Network) g.Node {
					return networkToTD(n)
				}),
		),
	)
}

func networkToTD(n network.Network) g.Node {
	return h.Tr(
		h.Td(g.Text(n.Name)),
		h.Td(g.Text(n.Prefix.String())),
	)
}
