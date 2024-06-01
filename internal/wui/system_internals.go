// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"

	"github.com/dustin/go-humanize"
	g "github.com/maragudk/gomponents"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/internal/bus"
	"github.com/networkables/mason/internal/server"
	"github.com/networkables/mason/internal/stackerr"
)

func (w WUI) wuiInternalsPageHandler(wr http.ResponseWriter, r *http.Request) {
	content := h.Main(
		h.ID("maincontent"),
		h.Class("drawer-content"),
		w.wuiInternalsMain(),
	)
	w.basePage("internals", content, nil).Render(wr)
}

func (w WUI) wuiInternalsMain() g.Node {
	internals := w.m.GetInternalsSnapshot()
	return grid("",
		wuiCard("Mason", masonInternalsToTable(internals)),
		wuiCard("Errors", wuiErrorsToTable(internals.Errors)),
		wuiCard("Events", wuiEventsToTable(internals.Events)),
		wuiCard("Go", goInternalsToTable(internals)),
	)
}

func masonInternalsToTable(iv server.MasonInternalsView) g.Node {
	return wuiTable([]string{"Name", "Value"},
		toTD("Networks", fmt.Sprint(iv.NetworkStoreCount)),
		toTD("Devices", fmt.Sprint(iv.DeviceStoreCount)),
		toTD(
			"NetworkScan Workers",
			fmt.Sprintf("%d / %d", iv.NetworkScanActive, iv.NetworkScanMaxWorkers),
		),
		toTD(
			"Discovery Workers",
			fmt.Sprintf("%d / %d", iv.AddressScanActive, iv.DiscoveryMaxWorkers),
		),
		toTD(
			"Enrichment Workers",
			fmt.Sprintf("%d / %d", iv.DeviceEnrichActive, iv.EnrichmentMaxWorkers),
		),
		toTD(
			"PerformancePinger Workers",
			fmt.Sprintf("%d / %d", iv.PerfPingActive, iv.PingerMaxWorkers),
		),
		toTD("PortScan MaxWorkers", fmt.Sprint(iv.PortScanMaxWorkers)),
		toTD("Current Network Scan", fmt.Sprint(iv.CurrentNetworkScan)),
	)
}

func goInternalsToTable(iv server.MasonInternalsView) g.Node {
	return wuiTable([]string{"Name", "Value"},
		toTD("Go Routines", fmt.Sprint(iv.NumberOfGoProcs)),
		toTD("Mem Alloc", humanize.Bytes(iv.Memstats.Alloc)),
		toTD("Mem CumlativeAlloc", humanize.Bytes(iv.Memstats.TotalAlloc)),
		toTD("Mem HeapAlloc", humanize.Bytes(iv.Memstats.HeapAlloc)),
		toTD("Mem HeapSys", humanize.Bytes(iv.Memstats.HeapSys)),

		toTD("Path", fmt.Sprintf("%+v", iv.Buildinfo.Path)),
		toTD(
			"Main",
			fmt.Sprintf(
				"%s (%s) %s",
				iv.Buildinfo.Main.Path,
				iv.Buildinfo.Main.Version,
				iv.Buildinfo.Main.Sum,
			),
		),
		h.Tr(h.Td(g.Text("Deps")), godepsToNode(iv.Buildinfo.Deps)),
		toTD("Settings", fmt.Sprintf("%+v", iv.Buildinfo.Settings)),
		toTD("GoVersion", fmt.Sprintf("%+v", iv.Buildinfo.GoVersion)),
	)
}

func godepsToNode(deps []*debug.Module) g.Node {
	num := len(deps)
	if num == 0 {
		return h.Td(g.Text("none"))
	}
	nodes := make([]g.Node, num)
	for i, dep := range deps {
		nodes[i] = h.Div(
			h.Span(g.Text(dep.Path)),
			h.Span(g.Text(dep.Version)),
		)
	}
	return h.Td(g.Group(nodes))
}

func toTD(name string, value string) g.Node {
	return h.Tr(
		h.Td(g.Text(name)),
		h.Td(g.Text(value)),
	)
}

func toTHTD(name string, value string) g.Node {
	return h.Tr(
		h.Th(g.Text(name)),
		h.Td(g.Text(value)),
	)
}

func wuiEventsToTable(events []bus.HistoricalEvent) g.Node {
	return wuiTable([]string{"Time", "Name", "Value"},
		g.Group(
			g.Map(events, func(he bus.HistoricalEvent) g.Node {
				valstr := ""
				estr, ok := he.E.(fmt.Stringer)
				if ok {
					valstr = estr.String()
				} else {
					valstr = fmt.Sprintf("%v", he.E)
				}
				return h.Tr(
					h.Td(g.Text(he.FmtTime())),
					h.Td(g.Text(he.Type())),
					h.Td(g.Text(valstr)),
				)
			}),
		),
	)
}

func wuiErrorsToTable(errors []bus.HistoricalError) g.Node {
	return wuiTable(
		[]string{"Time", "Type", "Error", "Stack"},
		g.Group(
			g.Map(errors, func(he bus.HistoricalError) g.Node {
				// stack := ""
				tp := he.Type()
				se, ok := he.E.(stackerr.StackErr)
				if ok {
					tp = fmt.Sprintf("%T", se.Unwrap())
					// stack = stackerr.Stack()
				}
				return h.Tr(
					h.Td(g.Text(he.FmtTime())),
					h.Td(g.Text(tp)),
					h.Td(g.Text(he.E.Error())),
					h.Td(
						g.If(
							ok,
							g.Group(g.Map(se.Frames, func(frame stackerr.Frame) g.Node {
								return wuiStackFrame(frame)
							})),
						),
					),
				)
			}),
		),
	)
}

func wuiStackFrame(frame stackerr.Frame) g.Node {
	return h.Div(
		h.Div(
			h.Span(
				h.Class("text-xs text-green-400"),
				g.Text(frame.File+":"+strconv.Itoa(frame.Line)),
			),
			// h.Span(
			// 	h.Class("text-xs"),
			// 	g.Text(" ("+frame.PCStr+")"),
			// ),
		),
		// h.Div(
		// 	h.Class("pl-4"),
		// 	h.Span(
		// 		h.Class("text-xs text-blue-400"),
		// 		g.Text(frame.Function),
		// 	),
		// 	h.Span(
		// 		h.Class("text-xs"),
		// 		g.Text(" "+frame.Source),
		// 	),
		// ),
	)
}
