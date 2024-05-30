// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"net/http"
	"strconv"

	g "github.com/maragudk/gomponents"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/internal/network"
)

func (w WUI) wuiHomePageHandler(wr http.ResponseWriter, r *http.Request) {
	content := h.Main(
		h.Class("drawer-content"),
		w.dashboardContent(),
	)
	w.basePage("dashboard", content, nil).Render(wr)
}

func (w WUI) dashboardContent() g.Node {
	return grid(
		"",
		wuiStatBox("devices", strconv.Itoa(w.m.CountDevices()), ""),
		wuiStatBox(
			"networks",
			strconv.Itoa(w.m.CountNetworks()),
			"ipv4 + ipv6",
		),
		wuiStatBox(
			"ping failures",
			strconv.Itoa(len(w.m.PingFailures())),
			"",
		),
		wuiStatBox(
			"servers",
			strconv.Itoa(len(w.m.ServerDevices())),
			"devices with listening ports",
		),
		g.Group(
			g.Map(
				w.m.GetNetworkStats(), func(ns network.NetworkStats) g.Node {
					return netStatBox(
						ns.Name,
						ns.Prefix,
						ns.IPUsed,
						ns.IPTotal,
						ns.AvgPing,
						ns.MaxPing,
					)
				},
			),
		),
	)
}
