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
)

func (w WUI) sidebarApiHandler(wr http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	selected := r.URL.Query().Get("selected")
	if selected == "" {
		selected = "dashboard"
	}
	w.sideBarContent(ctx, selected).Render(wr)
}

func sideBarLink(name string, selected string, url string, icon func() g.Node) g.Node {
	// iconSvg := g.Node(nil)
	// if icon != nil {
	// 	iconSvg = icon()
	// }
	// return h.Li(
	// 	h.A(
	// 		isSelectd(name, selected),
	// 		iconSvg,
	// 		h.Href(url),
	// 		g.Text(name),
	// 	),
	// )
	return baseSideBarLink(name, g.Text(name), selected, url, icon)
}

func sideBarLinkDevices(count int, selected string) g.Node {
	name := "Devices"
	return baseSideBarLink(
		name,
		g.Group(
			[]g.Node{
				g.Text(name),
				h.Span(
					h.Class("badge badge-info badge-sm"),
					g.Text(strconv.Itoa(count)),
				),
			},
		),
		selected,
		urlDevices,
		svgCpuChip,
	)
}

func baseSideBarLink(
	name string,
	namenode g.Node,
	selected string,
	url string,
	icon func() g.Node,
) g.Node {
	iconSvg := g.Node(nil)
	if icon != nil {
		iconSvg = icon()
	}
	return h.Li(
		h.A(
			isSelectd(name, selected),
			iconSvg,
			h.Href(url),
			namenode,
		),
	)
}

func sideBarSubsection(name string, icon func() g.Node, links ...g.Node) g.Node {
	iconSvg := g.Node(nil)
	if icon != nil {
		iconSvg = icon()
	}
	return h.Li(
		h.Details(
			g.Attr("open"),
			h.Summary(
				iconSvg,
				g.Text(name),
			),
			h.Ul(
				links...,
			),
		),
	)
}

func (w WUI) sideBarContent(ctx context.Context, selected string) g.Node {
	return h.Div(
		h.Label(h.For("my-drawer"), h.Class("drawer-overlay")),
		h.Nav(
			// h.Class("flex min-h-screen w-72 flex-col gap-2 overflow-y-auto bg-base-100 px-6 py-10"),
			h.Class("flex min-h-screen flex-col gap-2 overflow-y-auto bg-base-100"),
			h.Div(
				h.Class("mx-4 my-4 flex items-center gap-2 font-black"),
				g.Text("Mason"),
			),
			h.Ul(
				h.Class("menu"),
				sideBarLink("Dashboard", selected, urlRoot, svgModernHome),
				sideBarLinkDevices(len(w.m.ListDevices(ctx)), selected),
				sideBarLink("Networks", selected, urlNetworks, svgWifi),
				sideBarSubsection(
					"Tools", svgWrenchScrewdriver,
					// sideBarLink("Investigator", selected, urlInvestigator, svgFingerPrint),
					sideBarLink("Ping", selected, urlPing, svgCursorArrowRipple),
					sideBarLink("Traceroute", selected, urlTraceroute, svgArrowTrendingUp),
					sideBarLink("TLS", selected, urlTLS, svgLockClosed),
				),
				sideBarSubsection(
					"System", svgAdjustmentVertical,
					sideBarLink("Config", selected, urlConfig, svgCog),
					sideBarLink("Internals", selected, urlInternals, svgEye),
				),
			),
		),
	)
}
