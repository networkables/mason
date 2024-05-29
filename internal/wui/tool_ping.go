// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/nettools"
)

func (w WUI) wuiToolPingHandler(wr http.ResponseWriter, r *http.Request) {
	content := h.Main(
		h.ID("maincontent"),
		h.Class("drawer-content"),
		w.wuiToolPing(nil, nil, nil),
	)
	w.basePage("ping", content, nil).Render(wr)
}

func wuiPingResultTable(
	stats *nettools.Icmp4EchoResponseStatistics,
	mac net.HardwareAddr,
	manu string,
) g.Node {
	var macstr string
	if stats == nil {
		return nil
	}
	if mac != nil {
		macstr = mac.String()
	}
	return wuiCard("Ping Results for "+stats.Peer.String(),
		wuiTable([]string{" ", " "},
			toTD("Peer", stats.Peer.String()),
			toTD("Pings", fmt.Sprintf("%d", stats.TotalPackets)),
			toTD("Packet Loss", fmt.Sprintf("%3.0f %%", stats.PacketLoss)),
			toTD("Minimum", fmtDur(stats.Minimum)),
			toTD("Mean", fmtDur(stats.Mean)),
			toTD("Maximum", fmtDur(stats.Maximum)),
			toTD("StdDev", stats.StdDev.String()),
			toTD("Elapsed", fmtDur(stats.TotalElapsed)),
			g.If(
				mac != nil,
				toTD("MAC", macstr),
			),
			g.If(
				manu == "",
				toTD("Manufacturer", manu),
			),
		),
	)
}

func wuiIcmpStatsTable(stats *nettools.Icmp4EchoResponseStatistics) g.Node {
	if stats == nil {
		return nil
	}
	return wuiTable([]string{" ", " "},
		toTD("Peer", stats.Peer.String()),
		toTD("Pings", fmt.Sprintf("%d", stats.TotalPackets)),
		toTD("Packet Loss", fmt.Sprintf("%3.0f %%", stats.PacketLoss)),
		toTD("Minimum", stats.Minimum.String()),
		toTD("Mean", stats.Mean.String()),
		toTD("Maximum", stats.Maximum.String()),
		toTD("StdDev", stats.StdDev.String()),
		toTD("Elapsed", stats.TotalElapsed.String()),
	)
}

func (w WUI) wuiToolPing(
	icmpstats *nettools.Icmp4EchoResponseStatistics,
	mac net.HardwareAddr,
	err error,
) g.Node {
	var manu string
	if mac != nil {
		manu, _ = w.m.OuiLookup(mac)
	}
	var inValue g.Node
	if icmpstats != nil {
		inValue = h.Value(icmpstats.Peer.String())
	}

	return grid("pingcontent",
		wuiCard("Ping",
			h.Div(
				errAlert(err),
				h.FormEl(
					hx.Post(urlApiPing),
					hx.Target("#pingcontent"),
					hx.Swap("outerHTML"),
					h.Div(
						h.Class("form-control"),
						wuiFormInput(
							"Target",
							h.Input(
								h.Type("text"),
								h.Name(wuiToolTarget),
								inValue,
								h.Placeholder("192.168.1.1 or host.name"),
								h.Class("input-bordered w-1/2"),
							),
						),
						wuiFormButton("Ping"),
					),
				),
			),
		),
		wuiPingResultTable(icmpstats, mac, manu),
	)
}

func (w WUI) wuiApiToolPingHandler(wr http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	target := r.PostFormValue(wuiToolTarget)
	count := 10
	timeout := 100 * time.Millisecond
	stats, err := w.m.IcmpPing(ctx, target, count, timeout)
	if err != nil {
		log.Error("wuiApiToolPingHandler", "error", err)
	}
	mac, _ := w.m.ArpPing(ctx, target, timeout)
	w.wuiToolPing(&stats, mac, err).Render(wr)
}
