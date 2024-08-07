// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"context"
	"net/http"
	"time"

	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/internal/model"
)

func (w WUI) wuiDevicesPageHandler(wr http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	content := h.Main(
		h.ID("maincontent"),
		h.Class("drawer-content"),
		w.wuiDevicesMain(ctx),
	)
	w.basePage(ctx, "devices", content, nil).Render(wr)
}

func (w WUI) wuiDevicesMain(ctx context.Context) g.Node {
	devs := w.m.ListDevices(ctx)
	model.SortDevicesByAddr(devs)
	return h.Div(
		hx.Get("/api/devices"),
		hx.Trigger("every 60s"),
		hx.Swap("innerHTML"),
		grid("",
			wuiCard(
				"Devices as of "+time.Now().Format("15:04"),
				devicesToTable(devs),
			),
		),
	)
}

func (w WUI) wuiDevicesApiHandler(wr http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	w.wuiDevicesMain(ctx).Render(wr)
}

func devicesToTable(devs []model.Device) g.Node {
	rows := make([]g.Node, 0, len(devs))
	for _, dev := range devs {
		rows = append(rows, deviceToTD(dev))
	}
	return h.Table(
		h.Class("table table-zebra"),
		h.THead(
			h.Tr(
				h.Th(g.Text("")),
				h.Th(g.Text("Name")),
				h.Th(g.Text("IP")),
				h.Th(g.Text("Last Seen")),
				h.Th(g.Text("Ping")),
			),
		),
		h.TBody(
			rows...,
		),
	)
}

func deviceToTD(d model.Device) g.Node {
	url := "/device/" + d.Addr.String()
	detailsBtn := h.A(h.Href(url), svgMagnifyGlass())
	// graphBtn := h.A(h.Href(url), svgBarChart())
	return h.Tr(
		h.Td(
			detailsBtn,
			// graphBtn,
		),
		h.Td(g.Text(d.Name)),
		h.Td(g.Text(d.Addr.String())),
		h.Td(g.Text(d.LastSeenDurString(time.Since))),
		h.Td(g.Text(d.LastPingMeanString())),
	)
}
