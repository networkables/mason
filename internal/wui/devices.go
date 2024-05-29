// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"net/http"
	"time"

	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/internal/device"
)

func (w WUI) wuiDevicesPageHandler(wr http.ResponseWriter, r *http.Request) {
	content := h.Main(
		h.ID("maincontent"),
		h.Class("drawer-content"),
		w.wuiDevicesMain(),
	)
	w.basePage("devices", content, nil).Render(wr)
}

func (w WUI) wuiDevicesMain() g.Node {
	devs := w.m.ListDevices()
	device.SortDevicesByAddr(devs)
	return h.Div(
		hx.Get("/api/devices"),
		hx.Trigger("every 60s"),
		hx.Swap("innerHTML"),
		h.Div(
			h.Class("grid grid-cols-12 grid-rows-[min-content] gap-y-12 p-4 lg:gap-x-12 lg:p-10"),
			h.Section(
				h.Class("card col-span-12 overflow-hidden bg-base-100 shadow-sm xl:col-span-10"),
				h.Div(
					h.Class("card-body grow-0"),
					h.H2(
						h.Class("card-title"),
						h.A(
							h.Class("link-hover link"),
							g.Text("Devices as of "+time.Now().Format("15:04")),
						),
					),
					h.Div(
						h.Class("overflow-x-auto"),
						devicesToTable(devs),
					),
				),
			),
		),
	)
}

func (w WUI) wuiDevicesApiHandler(wr http.ResponseWriter, r *http.Request) {
	w.wuiDevicesMain().Render(wr)
}

func devicesToTable(devs []device.Device) g.Node {
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

func deviceToTD(d device.Device) g.Node {
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
