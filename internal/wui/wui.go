// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/dustin/go-humanize"
	g "github.com/maragudk/gomponents"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/internal/config"
	"github.com/networkables/mason/internal/server"
	"github.com/networkables/mason/internal/static"
)

// WUI is responsible for the Web UI when running in server mode
type WUI struct {
	m server.MasonReaderWriter
	h *http.Server
}

func New(m server.MasonReaderWriter, cfg *config.Server) *WUI {
	w := &WUI{
		m: m,
	}
	handler := w.newHandler()
	h := &http.Server{
		Addr:    ":" + *cfg.WebPort,
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

func (w WUI) basePage(activepage string, content g.Node, extrahead g.Node) g.Node {
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
					w.sideBarContent(activepage),
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
	prefix netip.Prefix,
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

func svgHome() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path d="M11.47 3.841a.75.75 0 0 1 1.06 0l8.69 8.69a.75.75 0 1 0 1.06-1.061l-8.689-8.69a2.25 2.25 0 0 0-3.182 0l-8.69 8.69a.75.75 0 1 0 1.061 1.06l8.69-8.689Z" /><path d="m12 5.432 8.159 8.159c.03.03.06.058.091.086v6.198c0 1.035-.84 1.875-1.875 1.875H15a.75.75 0 0 1-.75-.75v-4.5a.75.75 0 0 0-.75-.75h-3a.75.75 0 0 0-.75.75V21a.75.75 0 0 1-.75.75H5.625a1.875 1.875 0 0 1-1.875-1.875v-6.198a2.29 2.29 0 0 0 .091-.086L12 5.432Z" /></svg>`,
	)
}

func svgModernHome() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path d="M19.006 3.705a.75.75 0 1 0-.512-1.41L6 6.838V3a.75.75 0 0 0-.75-.75h-1.5A.75.75 0 0 0 3 3v4.93l-1.006.365a.75.75 0 0 0 .512 1.41l16.5-6Z" /><path fill-rule="evenodd" d="M3.019 11.114 18 5.667v3.421l4.006 1.457a.75.75 0 1 1-.512 1.41l-.494-.18v8.475h.75a.75.75 0 0 1 0 1.5H2.25a.75.75 0 0 1 0-1.5H3v-9.129l.019-.007ZM18 20.25v-9.566l1.5.546v9.02H18Zm-9-6a.75.75 0 0 0-.75.75v4.5c0 .414.336.75.75.75h3a.75.75 0 0 0 .75-.75V15a.75.75 0 0 0-.75-.75H9Z" clip-rule="evenodd" /></svg>`,
	)
}

func svgMagnifyGlass() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path fill-rule="evenodd" d="M10.5 3.75a6.75 6.75 0 1 0 0 13.5 6.75 6.75 0 0 0 0-13.5ZM2.25 10.5a8.25 8.25 0 1 1 14.59 5.28l4.69 4.69a.75.75 0 1 1-1.06 1.06l-4.69-4.69A8.25 8.25 0 0 1 2.25 10.5Z" clip-rule="evenodd" /></svg>`,
	)
}

func svgBarChart() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path d="M18.375 2.25c-1.035 0-1.875.84-1.875 1.875v15.75c0 1.035.84 1.875 1.875 1.875h.75c1.035 0 1.875-.84 1.875-1.875V4.125c0-1.036-.84-1.875-1.875-1.875h-.75ZM9.75 8.625c0-1.036.84-1.875 1.875-1.875h.75c1.036 0 1.875.84 1.875 1.875v11.25c0 1.035-.84 1.875-1.875 1.875h-.75a1.875 1.875 0 0 1-1.875-1.875V8.625ZM3 13.125c0-1.036.84-1.875 1.875-1.875h.75c1.036 0 1.875.84 1.875 1.875v6.75c0 1.035-.84 1.875-1.875 1.875h-.75A1.875 1.875 0 0 1 3 19.875v-6.75Z" /></svg>`,
	)
}

func svgCpuChip() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path d="M16.5 7.5h-9v9h9v-9Z" /><path fill-rule="evenodd" d="M8.25 2.25A.75.75 0 0 1 9 3v.75h2.25V3a.75.75 0 0 1 1.5 0v.75H15V3a.75.75 0 0 1 1.5 0v.75h.75a3 3 0 0 1 3 3v.75H21A.75.75 0 0 1 21 9h-.75v2.25H21a.75.75 0 0 1 0 1.5h-.75V15H21a.75.75 0 0 1 0 1.5h-.75v.75a3 3 0 0 1-3 3h-.75V21a.75.75 0 0 1-1.5 0v-.75h-2.25V21a.75.75 0 0 1-1.5 0v-.75H9V21a.75.75 0 0 1-1.5 0v-.75h-.75a3 3 0 0 1-3-3v-.75H3A.75.75 0 0 1 3 15h.75v-2.25H3a.75.75 0 0 1 0-1.5h.75V9H3a.75.75 0 0 1 0-1.5h.75v-.75a3 3 0 0 1 3-3h.75V3a.75.75 0 0 1 .75-.75ZM6 6.75A.75.75 0 0 1 6.75 6h10.5a.75.75 0 0 1 .75.75v10.5a.75.75 0 0 1-.75.75H6.75a.75.75 0 0 1-.75-.75V6.75Z" clip-rule="evenodd" /></svg>`,
	)
}

func svgWifi() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path fill-rule="evenodd" d="M1.371 8.143c5.858-5.857 15.356-5.857 21.213 0a.75.75 0 0 1 0 1.061l-.53.53a.75.75 0 0 1-1.06 0c-4.98-4.979-13.053-4.979-18.032 0a.75.75 0 0 1-1.06 0l-.53-.53a.75.75 0 0 1 0-1.06Zm3.182 3.182c4.1-4.1 10.749-4.1 14.85 0a.75.75 0 0 1 0 1.061l-.53.53a.75.75 0 0 1-1.062 0 8.25 8.25 0 0 0-11.667 0 .75.75 0 0 1-1.06 0l-.53-.53a.75.75 0 0 1 0-1.06Zm3.204 3.182a6 6 0 0 1 8.486 0 .75.75 0 0 1 0 1.061l-.53.53a.75.75 0 0 1-1.061 0 3.75 3.75 0 0 0-5.304 0 .75.75 0 0 1-1.06 0l-.53-.53a.75.75 0 0 1 0-1.06Zm3.182 3.182a1.5 1.5 0 0 1 2.122 0 .75.75 0 0 1 0 1.061l-.53.53a.75.75 0 0 1-1.061 0l-.53-.53a.75.75 0 0 1 0-1.06Z" clip-rule="evenodd" /></svg>`,
	)
}

func svgCog() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path d="M17.004 10.407c.138.435-.216.842-.672.842h-3.465a.75.75 0 0 1-.65-.375l-1.732-3c-.229-.396-.053-.907.393-1.004a5.252 5.252 0 0 1 6.126 3.537ZM8.12 8.464c.307-.338.838-.235 1.066.16l1.732 3a.75.75 0 0 1 0 .75l-1.732 3c-.229.397-.76.5-1.067.161A5.23 5.23 0 0 1 6.75 12a5.23 5.23 0 0 1 1.37-3.536ZM10.878 17.13c-.447-.098-.623-.608-.394-1.004l1.733-3.002a.75.75 0 0 1 .65-.375h3.465c.457 0 .81.407.672.842a5.252 5.252 0 0 1-6.126 3.539Z" /><path fill-rule="evenodd" d="M21 12.75a.75.75 0 1 0 0-1.5h-.783a8.22 8.22 0 0 0-.237-1.357l.734-.267a.75.75 0 1 0-.513-1.41l-.735.268a8.24 8.24 0 0 0-.689-1.192l.6-.503a.75.75 0 1 0-.964-1.149l-.6.504a8.3 8.3 0 0 0-1.054-.885l.391-.678a.75.75 0 1 0-1.299-.75l-.39.676a8.188 8.188 0 0 0-1.295-.47l.136-.77a.75.75 0 0 0-1.477-.26l-.136.77a8.36 8.36 0 0 0-1.377 0l-.136-.77a.75.75 0 1 0-1.477.26l.136.77c-.448.121-.88.28-1.294.47l-.39-.676a.75.75 0 0 0-1.3.75l.392.678a8.29 8.29 0 0 0-1.054.885l-.6-.504a.75.75 0 1 0-.965 1.149l.6.503a8.243 8.243 0 0 0-.689 1.192L3.8 8.216a.75.75 0 1 0-.513 1.41l.735.267a8.222 8.222 0 0 0-.238 1.356h-.783a.75.75 0 0 0 0 1.5h.783c.042.464.122.917.238 1.356l-.735.268a.75.75 0 0 0 .513 1.41l.735-.268c.197.417.428.816.69 1.191l-.6.504a.75.75 0 0 0 .963 1.15l.601-.505c.326.323.679.62 1.054.885l-.392.68a.75.75 0 0 0 1.3.75l.39-.679c.414.192.847.35 1.294.471l-.136.77a.75.75 0 0 0 1.477.261l.137-.772a8.332 8.332 0 0 0 1.376 0l.136.772a.75.75 0 1 0 1.477-.26l-.136-.771a8.19 8.19 0 0 0 1.294-.47l.391.677a.75.75 0 0 0 1.3-.75l-.393-.679a8.29 8.29 0 0 0 1.054-.885l.601.504a.75.75 0 0 0 .964-1.15l-.6-.503c.261-.375.492-.774.69-1.191l.735.267a.75.75 0 1 0 .512-1.41l-.734-.267c.115-.439.195-.892.237-1.356h.784Zm-2.657-3.06a6.744 6.744 0 0 0-1.19-2.053 6.784 6.784 0 0 0-1.82-1.51A6.705 6.705 0 0 0 12 5.25a6.8 6.8 0 0 0-1.225.11 6.7 6.7 0 0 0-2.15.793 6.784 6.784 0 0 0-2.952 3.489.76.76 0 0 1-.036.098A6.74 6.74 0 0 0 5.251 12a6.74 6.74 0 0 0 3.366 5.842l.009.005a6.704 6.704 0 0 0 2.18.798l.022.003a6.792 6.792 0 0 0 2.368-.004 6.704 6.704 0 0 0 2.205-.811 6.785 6.785 0 0 0 1.762-1.484l.009-.01.009-.01a6.743 6.743 0 0 0 1.18-2.066c.253-.707.39-1.469.39-2.263a6.74 6.74 0 0 0-.408-2.309Z" clip-rule="evenodd" /></svg>`,
	)
}

func svgEye() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z" /><path fill-rule="evenodd" d="M1.323 11.447C2.811 6.976 7.028 3.75 12.001 3.75c4.97 0 9.185 3.223 10.675 7.69.12.362.12.752 0 1.113-1.487 4.471-5.705 7.697-10.677 7.697-4.97 0-9.186-3.223-10.675-7.69a1.762 1.762 0 0 1 0-1.113ZM17.25 12a5.25 5.25 0 1 1-10.5 0 5.25 5.25 0 0 1 10.5 0Z" clip-rule="evenodd" /></svg>`,
	)
}

func svgLockClosed() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path fill-rule="evenodd" d="M12 1.5a5.25 5.25 0 0 0-5.25 5.25v3a3 3 0 0 0-3 3v6.75a3 3 0 0 0 3 3h10.5a3 3 0 0 0 3-3v-6.75a3 3 0 0 0-3-3v-3c0-2.9-2.35-5.25-5.25-5.25Zm3.75 8.25v-3a3.75 3.75 0 1 0-7.5 0v3h7.5Z" clip-rule="evenodd" /></svg>`,
	)
}

func svgCursorArrowRipple() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path fill-rule="evenodd" d="M17.303 5.197A7.5 7.5 0 0 0 6.697 15.803a.75.75 0 0 1-1.061 1.061A9 9 0 1 1 21 10.5a.75.75 0 0 1-1.5 0c0-1.92-.732-3.839-2.197-5.303Zm-2.121 2.121a4.5 4.5 0 0 0-6.364 6.364.75.75 0 1 1-1.06 1.06A6 6 0 1 1 18 10.5a.75.75 0 0 1-1.5 0c0-1.153-.44-2.303-1.318-3.182Zm-3.634 1.314a.75.75 0 0 1 .82.311l5.228 7.917a.75.75 0 0 1-.777 1.148l-2.097-.43 1.045 3.9a.75.75 0 0 1-1.45.388l-1.044-3.899-1.601 1.42a.75.75 0 0 1-1.247-.606l.569-9.47a.75.75 0 0 1 .554-.68Z" clip-rule="evenodd" /></svg>`,
	)
}

func svgArrowTrendingUp() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path fill-rule="evenodd" d="M15.22 6.268a.75.75 0 0 1 .968-.431l5.942 2.28a.75.75 0 0 1 .431.97l-2.28 5.94a.75.75 0 1 1-1.4-.537l1.63-4.251-1.086.484a11.2 11.2 0 0 0-5.45 5.173.75.75 0 0 1-1.199.19L9 12.312l-6.22 6.22a.75.75 0 0 1-1.06-1.061l6.75-6.75a.75.75 0 0 1 1.06 0l3.606 3.606a12.695 12.695 0 0 1 5.68-4.974l1.086-.483-4.251-1.632a.75.75 0 0 1-.432-.97Z" clip-rule="evenodd" /></svg>`,
	)
}

func svgFingerPrint() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path fill-rule="evenodd" d="M12 3.75a6.715 6.715 0 0 0-3.722 1.118.75.75 0 1 1-.828-1.25 8.25 8.25 0 0 1 12.8 6.883c0 3.014-.574 5.897-1.62 8.543a.75.75 0 0 1-1.395-.551A21.69 21.69 0 0 0 18.75 10.5 6.75 6.75 0 0 0 12 3.75ZM6.157 5.739a.75.75 0 0 1 .21 1.04A6.715 6.715 0 0 0 5.25 10.5c0 1.613-.463 3.12-1.265 4.393a.75.75 0 0 1-1.27-.8A6.715 6.715 0 0 0 3.75 10.5c0-1.68.503-3.246 1.367-4.55a.75.75 0 0 1 1.04-.211ZM12 7.5a3 3 0 0 0-3 3c0 3.1-1.176 5.927-3.105 8.056a.75.75 0 1 1-1.112-1.008A10.459 10.459 0 0 0 7.5 10.5a4.5 4.5 0 1 1 9 0c0 .547-.022 1.09-.067 1.626a.75.75 0 0 1-1.495-.123c.041-.495.062-.996.062-1.503a3 3 0 0 0-3-3Zm0 2.25a.75.75 0 0 1 .75.75c0 3.908-1.424 7.485-3.781 10.238a.75.75 0 0 1-1.14-.975A14.19 14.19 0 0 0 11.25 10.5a.75.75 0 0 1 .75-.75Zm3.239 5.183a.75.75 0 0 1 .515.927 19.417 19.417 0 0 1-2.585 5.544.75.75 0 0 1-1.243-.84 17.915 17.915 0 0 0 2.386-5.116.75.75 0 0 1 .927-.515Z" clip-rule="evenodd" /></svg>`,
	)
}

func svgWrenchScrewdriver() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path fill-rule="evenodd" d="M12 6.75a5.25 5.25 0 0 1 6.775-5.025.75.75 0 0 1 .313 1.248l-3.32 3.319c.063.475.276.934.641 1.299.365.365.824.578 1.3.64l3.318-3.319a.75.75 0 0 1 1.248.313 5.25 5.25 0 0 1-5.472 6.756c-1.018-.086-1.87.1-2.309.634L7.344 21.3A3.298 3.298 0 1 1 2.7 16.657l8.684-7.151c.533-.44.72-1.291.634-2.309A5.342 5.342 0 0 1 12 6.75ZM4.117 19.125a.75.75 0 0 1 .75-.75h.008a.75.75 0 0 1 .75.75v.008a.75.75 0 0 1-.75.75h-.008a.75.75 0 0 1-.75-.75v-.008Z" clip-rule="evenodd" /><path d="m10.076 8.64-2.201-2.2V4.874a.75.75 0 0 0-.364-.643l-3.75-2.25a.75.75 0 0 0-.916.113l-.75.75a.75.75 0 0 0-.113.916l2.25 3.75a.75.75 0 0 0 .643.364h1.564l2.062 2.062 1.575-1.297Z" /><path fill-rule="evenodd" d="m12.556 17.329 4.183 4.182a3.375 3.375 0 0 0 4.773-4.773l-3.306-3.305a6.803 6.803 0 0 1-1.53.043c-.394-.034-.682-.006-.867.042a.589.589 0 0 0-.167.063l-3.086 3.748Zm3.414-1.36a.75.75 0 0 1 1.06 0l1.875 1.876a.75.75 0 1 1-1.06 1.06L15.97 17.03a.75.75 0 0 1 0-1.06Z" clip-rule="evenodd" /></svg>`,
	)
}

func svgAdjustmentHorizontal() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path d="M18.75 12.75h1.5a.75.75 0 0 0 0-1.5h-1.5a.75.75 0 0 0 0 1.5ZM12 6a.75.75 0 0 1 .75-.75h7.5a.75.75 0 0 1 0 1.5h-7.5A.75.75 0 0 1 12 6ZM12 18a.75.75 0 0 1 .75-.75h7.5a.75.75 0 0 1 0 1.5h-7.5A.75.75 0 0 1 12 18ZM3.75 6.75h1.5a.75.75 0 1 0 0-1.5h-1.5a.75.75 0 0 0 0 1.5ZM5.25 18.75h-1.5a.75.75 0 0 1 0-1.5h1.5a.75.75 0 0 1 0 1.5ZM3 12a.75.75 0 0 1 .75-.75h7.5a.75.75 0 0 1 0 1.5h-7.5A.75.75 0 0 1 3 12ZM9 3.75a2.25 2.25 0 1 0 0 4.5 2.25 2.25 0 0 0 0-4.5ZM12.75 12a2.25 2.25 0 1 1 4.5 0 2.25 2.25 0 0 1-4.5 0ZM9 15.75a2.25 2.25 0 1 0 0 4.5 2.25 2.25 0 0 0 0-4.5Z" /></svg>`,
	)
}

func svgAdjustmentVertical() g.Node {
	return g.Raw(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path d="M6 12a.75.75 0 0 1-.75-.75v-7.5a.75.75 0 1 1 1.5 0v7.5A.75.75 0 0 1 6 12ZM18 12a.75.75 0 0 1-.75-.75v-7.5a.75.75 0 0 1 1.5 0v7.5A.75.75 0 0 1 18 12ZM6.75 20.25v-1.5a.75.75 0 0 0-1.5 0v1.5a.75.75 0 0 0 1.5 0ZM18.75 18.75v1.5a.75.75 0 0 1-1.5 0v-1.5a.75.75 0 0 1 1.5 0ZM12.75 5.25v-1.5a.75.75 0 0 0-1.5 0v1.5a.75.75 0 0 0 1.5 0ZM12 21a.75.75 0 0 1-.75-.75v-7.5a.75.75 0 0 1 1.5 0v7.5A.75.75 0 0 1 12 21ZM3.75 15a2.25 2.25 0 1 0 4.5 0 2.25 2.25 0 0 0-4.5 0ZM12 11.25a2.25 2.25 0 1 1 0-4.5 2.25 2.25 0 0 1 0 4.5ZM15.75 15a2.25 2.25 0 1 0 4.5 0 2.25 2.25 0 0 0-4.5 0Z" /></svg>`,
	)
}
