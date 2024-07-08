// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/dustin/go-humanize"
	g "github.com/maragudk/gomponents"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/internal/enrichment"
	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/pinger"
	"github.com/networkables/mason/internal/server"
	"github.com/networkables/mason/internal/static"
	"github.com/networkables/mason/nettools"
)

type MasonReaderWriter interface {
	MasonReader
	MasonWriter
	MasonNetworker
}

type MasonReader interface {
	ListNetworks(context.Context) []model.Network
	CountNetworks(context.Context) int
	ListDevices(context.Context) []model.Device
	CountDevices(context.Context) int
	GetDeviceByAddr(context.Context, model.Addr) (model.Device, error)
	ReadPerformancePings(
		context.Context,
		model.Device,
		time.Duration,
	) ([]pinger.Point, error)
	GetConfig() *server.Config
	GetInternalsSnapshot(ctx context.Context) server.MasonInternalsView
	GetUserAgent() string
	OuiLookup(mac net.HardwareAddr) string
	GetNetworkStats(ctx context.Context) []model.NetworkStats
	PingFailures(ctx context.Context) []model.Device
	ServerDevices(ctx context.Context) []model.Device
	FlowSummaryByIP(context.Context, model.Addr) ([]model.FlowSummaryForAddrByIP, error)
	FlowSummaryByName(context.Context, model.Addr) ([]model.FlowSummaryForAddrByName, error)
	FlowSummaryByCountry(context.Context, model.Addr) ([]model.FlowSummaryForAddrByCountry, error)
	LookupIP(model.Addr) string
}

type MasonWriter interface {
	AddNetwork(context.Context, model.Network) error
	AddNetworkByName(context.Context, string, string, bool) error
}

type MasonNetworker interface {
	StringToAddr(string) (model.Addr, error)
	IcmpPingAddr(
		context.Context,
		model.Addr,
		int,
		time.Duration,
		bool,
	) (nettools.Icmp4EchoResponseStatistics, error)
	IcmpPing(
		context.Context,
		string,
		int,
		time.Duration,
		bool,
	) (nettools.Icmp4EchoResponseStatistics, error)
	ArpPing(context.Context, string, time.Duration) (model.MAC, error)
	Portscan(context.Context, string, *enrichment.PortScanConfig) ([]int, error)
	GetExternalAddr(ctx context.Context) (model.Addr, error)
	Traceroute(context.Context, string) ([]nettools.Icmp4EchoResponseStatistics, error)
	TracerouteAddr(
		context.Context,
		model.Addr,
	) ([]nettools.Icmp4EchoResponseStatistics, error)
	FetchTLSInfo(context.Context, string) (nettools.TLS, error)
	FetchSNMPInfo(context.Context, string) (nettools.SnmpInfo, error)
	FetchSNMPInfoAddr(context.Context, model.Addr) (nettools.SnmpInfo, error)
}

// WUI is responsible for the Web UI when running in server mode
type WUI struct {
	m MasonReaderWriter
	h *http.Server
}

func New(m MasonReaderWriter, listenaddress string) *WUI {
	w := &WUI{
		m: m,
	}
	handler := w.newHandler()
	h := &http.Server{
		Addr:    listenaddress,
		Handler: handler,
	}
	w.h = h

	return w
}

func (w *WUI) Start() error {
	log.Info("starting http server", "addr", w.h.Addr)
	err := w.h.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (w *WUI) Shutdown(ctx context.Context) error {
	if w.h != nil {
		if err := w.h.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (w *WUI) newHandler() http.Handler {
	mux := http.NewServeMux()
	w.addRoutes(mux)
	var handler http.Handler = mux
	// middleware
	// handler = someMiddleware(handler)
	return handler
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, static.StaticFiles, "/static/images/favicon.ico")
}

func (w WUI) basePage(
	ctx context.Context,
	activepage string,
	content g.Node,
	extrahead g.Node,
) g.Node {
	return h.Doctype(
		h.HTML(
			h.DataAttr("theme", "light"),
			h.Lang("en"),
			h.Head(
				h.Meta(h.Charset("utf-8")),
				h.TitleEl(g.Text("Mason")),
				h.Meta(
					h.Name("viewport"),
					h.Content("width=device-width, initial-scale=1"),
				),
				h.Link(
					h.Rel("stylesheet"),
					h.Href("/static/css/daisyui-4.11.1.css"),
				),
				h.Script(h.Src("/static/javascript/tailwindcss-3.4.3.js")),
				h.Script(h.Src("/static/javascript/htmx.js")),
				h.Script(h.Src("/static/javascript/theme-change.js")),
				extrahead,
			),
			h.Body(
				h.Class("drawer min-h-screen bg-base-200 lg:drawer-open"),
				h.Input(
					h.ID("my-drawer"),
					h.Type("checkbox"),
					h.Class("drawer-toggle"),
				),
				content,
				h.Aside(
					h.ID("sidebar"),
					h.Class("drawer-side z-10"),
					w.sideBarContent(ctx, activepage),
				),
			),
		),
	)
}

func isSelectd(name string, selected string) g.Node {
	if strings.ToUpper(name) == strings.ToUpper(selected) {
		return h.Class("active")
	}
	return nil
}

func fmtDur(d time.Duration) string {
	return d.Round(50 * time.Microsecond).String()
}

func netStatBox2(
	netname string,
	prefix netip.Prefix,
	usedip uint64,
	totalip float64,
	avgping time.Duration,
	maxping time.Duration,
) g.Node {
	pct := fmt.Sprintf("%3.0f", (float64(usedip)/totalip)*100)
	nm := g.Text(netname)
	if len(netname) > 25 {
		nm = h.Span(h.Class("text-2xl"), g.Text(netname))
	}
	// if len(netname) > 26 {
	// 	nm = h.Span(h.Class("text-sm"), g.Text(netname))
	// }
	// if len(netname) > 35 {
	// 	nm = h.Span(h.Class("text-xs"), g.Text(netname))
	// }
	return h.Section(
		h.Class(
			// "stats stats-horizontal col-span-12 w-full",
			"stats stats-horizontal col-span-12 xl:col-span-6",
		),
		h.Div(
			h.Class("stat"),
			h.Div(
				h.Class("stat-title"),
				g.Text("network"),
			),
			h.Div(
				h.Class("stat-value"),
				nm,
			),
			h.Div(
				h.Class("stat-desc"),
				h.Span(
					h.Class("text-secondary"),
					g.Text(prefix.String()),
				),
				h.Span(
					g.Text(pct+"% used"),
				),
			),
		),
		h.Div(
			h.Class("stat"),
			h.Div(
				h.Class("stat-title"),
				g.Text("avg ping"),
			),
			h.Div(
				h.Class("stat-value"),
				g.Text(fmtDur(avgping)),
			),
			h.Div(
				h.Class("stat-desc"),
				g.Text("max "+fmtDur(maxping)),
			),
		),
	)
}

func netStatBox(
	netname string,
	prefix model.Prefix,
	usedip uint64,
	totalip float64,
	avgping time.Duration,
	maxping time.Duration,
) g.Node {
	pct := fmt.Sprintf("%3.0f", (float64(usedip)/totalip)*100)
	aspace := fmt.Sprintf("%d of %s", usedip, humanize.SIWithDigits(totalip, 0, ""))
	nm := g.Text(netname)
	if len(netname) > 25 {
		nm = h.Span(h.Class("text-2xl"), g.Text(netname))
	}
	// if len(netname) > 15 {
	// 	nm = h.Span(h.Class("text-base"), g.Text(netname))
	// }
	// if len(netname) > 26 {
	// 	nm = h.Span(h.Class("text-xs"), g.Text(netname))
	// }
	return h.Section(
		h.Class(
			// "stats stats-horizontal col-span-12 w-full",
			"stats stats-horizontal col-span-12 xl:col-span-6",
		),
		h.Div(
			h.Class("stat"),
			h.Div(
				h.Class("stat-title"),
				g.Text("network"),
			),
			h.Div(
				h.Class("stat-value"),
				nm,
			),
			h.Div(
				h.Class("stat-desc text-secondary"),
				g.Text(prefix.String()),
			),
		),
		h.Div(
			h.Class("stat"),
			h.Div(
				h.Class("stat-title"),
				g.Text("avg ping"),
			),
			h.Div(
				h.Class("stat-value"),
				g.Text(fmtDur(avgping)),
			),
			h.Div(
				h.Class("stat-desc"),
				g.Text("max "+fmtDur(maxping)),
			),
		),
		h.Div(
			h.Class("stat"),
			h.Div(
				h.Class("stat-title"),
				g.Text("ip usage"),
			),
			h.Div(
				h.Class("stat-value"),
				g.Text(pct+" %"),
			),
			h.Div(
				h.Class("stat-desc"),
				g.Text(aspace),
			),
		),
	)
}

func wuiStatBox(title string, value string, desc string) g.Node {
	var descNode g.Node
	if desc != "" {
		descNode = h.Div(
			h.Class("stat-desc"),
			g.Text(desc),
		)
	}
	return h.Section(
		h.Class(
			// "stats stats-vertical col-span-12 w-1/2 shadow-sm xl:stats-horizontal",
			"stats col-span-3 xl:col-span-2",
		),
		h.Div(
			h.Class("stat"),
			h.Div(
				h.Class("stat-title"),
				g.Text(title),
			),
			h.Div(
				h.Class("stat-value"),
				g.Text(value),
			),
			descNode,
		),
	)
}

func widecard(title string, body g.Node) g.Node {
	return h.Section(
		h.Class("card bg-base-100 col-span-12 shadow-sm"),
		h.Div(
			h.Class("card-body"),
			h.H2(
				h.Class("card-title"),
				g.Text(title),
			),
			body,
			// h.P(g.Text(body)),
		),
	)
}

func graphcard(title string, graphs ...g.Node) g.Node {
	return h.Section(
		h.Class("card bg-base-100 col-span-12 shadow-sm"),
		h.StyleEl(
			g.Text(`
						.container {margin-top:30px; display: flex;justify-content: center;align-items: center;} 
            .item {margin: auto;}
          `),
		),
		h.Div(
			h.Class("card-body"),
			h.H2(
				h.Class("card-title"),
				g.Text(title),
			),
			g.Group(graphs),
		),
	)
}

func card(title string, body string) g.Node {
	return h.Section(
		h.Class("card bg-base-100 col-span-3"),
		h.Div(
			h.Class("card-body"),
			h.H2(
				h.Class("card-title"),
				g.Text(title),
			),
			h.P(g.Text(body)),
		),
	)
}

func grid(id string, cards ...g.Node) g.Node {
	return h.Div(
		g.If(id != "", h.ID(id)),
		h.Div(
			h.Class(
				// "grid grid-cols-12 grid-rows-[min-content] gap-y-12 p-4 lg:gap-x-12 lg:p-10",
				"grid grid-cols-12 grid-rows-[min-content] gap-4 p-4",
			),
			g.Group(cards),
		),
	)
}

func wuiCard(title string, content g.Node) g.Node {
	return h.Section(
		h.Class("card col-span-12 overflow-hidden bg-base-100 shadow-sm xl:col-span-10"),
		h.Div(
			h.Class("card-body grow-0"),
			h.H2(
				h.Class("card-title"),
				h.A(
					h.Class("link-hover link"),
					g.Text(title),
				),
			),
			h.Div(
				h.Class("overflow-x-auto"),
				content,
			),
		),
	)
}

func errAlert(err error) g.Node {
	if err == nil {
		return nil
	}
	return h.Div(
		h.Class("alert alert-error"),
		g.Raw(
			`<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-5 w-5" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`,
		),
		h.Span(g.Text(err.Error())),
	)
}

func wuiTable(names []string, rows ...g.Node) g.Node {
	return h.Table(
		h.Class("table table-zebra"),
		h.THead(
			h.Tr(
				g.Group(
					g.Map(names, func(n string) g.Node {
						return h.Th(g.Text(n))
					}),
				),
			),
		),
		h.TBody(
			rows...,
		),
	)
}

func wuiFormInput(name string, elem g.Node) g.Node {
	return h.Label(
		h.Class("label"),
		h.Span(h.Class("label-text"), g.Text(name)),
		elem,
	)
}

func wuiFormButton(text string) g.Node {
	return h.Div(
		h.Class("flex gap-4 py-4"),
		h.Button(h.Class("btn btn-primary grow"), g.Text(text)),
	)
}
