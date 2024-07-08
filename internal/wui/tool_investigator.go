// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"context"
	"net/http"

	g "github.com/maragudk/gomponents"
	h "github.com/maragudk/gomponents/html"
)

func (w WUI) wuiToolInvestigatorHandler(wr http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	content := h.Main(
		h.ID("maincontent"),
		h.Class("drawer-content"),
		w.wuiToolInvestigator(),
	)
	w.basePage(ctx, "investigator", content, nil).Render(wr)
}

func (w WUI) wuiToolInvestigator() g.Node {
	return grid("investigatorcontent",
		wuiCard("Investigator",
			h.Div(),
		),
	)
}

func (w WUI) wuiApiToolInvestigatorHandler(wr http.ResponseWriter, r *http.Request) {
	// ctx := context.TODO()
	// target := r.PostFormValue(wuiToolTarget)
	// count := 10
	// timeout := 100 * time.Millisecond
	// stats, err := w.m.IcmpPing(ctx, target, count, timeout)
	// if err != nil {
	// 	log.Error("wuiApiToolPingHandler", "error", err)
	// }
	// mac, err := w.m.ArpPing(ctx, target, timeout)
	// w.wuiToolPing(&stats, mac, err).Render(wr)
}
