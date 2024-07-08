// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"context"
	"net/http"
	"strconv"

	g "github.com/maragudk/gomponents"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/internal/model"
)

func (w WUI) wuiHomePageHandler(wr http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	content := h.Main(
		h.Class("drawer-content"),
		w.dashboardContent(ctx),
	)
	w.basePage(ctx, "dashboard", content, nil).Render(wr)
}

func (w WUI) dashboardContent(ctx context.Context) g.Node {
	return grid(
		"",
		wuiStatBox("devices", strconv.Itoa(w.m.CountDevices(ctx)), ""),
		wuiStatBox(
			"networks",
			strconv.Itoa(w.m.CountNetworks(ctx)),
			"ipv4 + ipv6",
		),
		wuiStatBox(
			"ping failures",
			strconv.Itoa(len(w.m.PingFailures(ctx))),
			"",
		),
		wuiStatBox(
			"servers",
			strconv.Itoa(len(w.m.ServerDevices(ctx))),
			"devices with listening ports",
		),
		g.Group(
			g.Map(
				w.m.GetNetworkStats(ctx), func(ns model.NetworkStats) g.Node {
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
