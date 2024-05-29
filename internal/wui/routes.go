// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"net/http"

	"github.com/networkables/mason/internal/static"
)

func (w WUI) addRoutes(mux *http.ServeMux) {
	// mime.AddExtensionType(".js", "application/javascript")
	w.addApiRoutes(mux)
	w.addPageRoutes(mux)
	mux.Handle("/static/", http.FileServerFS(static.StaticFiles))
	mux.HandleFunc("/favicon.ico", faviconHandler)
}

const (
	urlConfig          = "/config"
	urlInternals       = "/internals"
	urlNetworks        = "/networks"
	urlDevices         = "/devices"
	urlDevice          = "/device"
	urlRoot            = "/"
	urlApiNetworks     = "/api/networks"
	urlApiDevices      = "/api/devices"
	urlApiPing         = "/api/ping"
	urlApiTraceroute   = "/api/traceroute"
	urlApiTLS          = "/api/tls"
	urlApiInvestigator = "/api/investigator"
	urlInvestigator    = "/investigator"
	urlPing            = "/ping"
	urlTraceroute      = "/traceroute"
	urlTLS             = "/tls"
)

func (w WUI) addPageRoutes(mux *http.ServeMux) {
	mux.HandleFunc(urlInvestigator, w.wuiToolInvestigatorHandler)
	mux.HandleFunc(urlPing, w.wuiToolPingHandler)
	mux.HandleFunc(urlTraceroute, w.wuiToolTracerouteHandler)
	mux.HandleFunc(urlTLS, w.wuiToolTLSHandler)

	mux.HandleFunc(urlConfig, w.wuiConfigPageHandler)
	mux.HandleFunc(urlInternals, w.wuiInternalsPageHandler)
	mux.HandleFunc(urlNetworks, w.wuiNetworksPageHandler)
	mux.HandleFunc(urlDevices, w.wuiDevicesPageHandler)
	mux.HandleFunc(urlDevice+"/{id}", w.wuiDevicePageHandler)
	mux.HandleFunc(urlRoot, w.wuiHomePageHandler)
}

func (w WUI) addApiRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST "+urlApiNetworks, w.wuiNetworksApiCreate)
	mux.HandleFunc(urlApiDevices, w.wuiDevicesApiHandler)
	mux.HandleFunc(urlApiPing, w.wuiApiToolPingHandler)
	mux.HandleFunc(urlApiTraceroute, w.wuiApiToolTracerouteHandler)
	mux.HandleFunc(urlApiTLS, w.wuiApiToolTLSHandler)
	mux.HandleFunc(urlApiInvestigator, w.wuiApiToolInvestigatorHandler)
}
