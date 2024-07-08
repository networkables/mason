// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/dustin/go-humanize"
	charts "github.com/go-echarts/go-echarts/v2/charts"
	opts "github.com/go-echarts/go-echarts/v2/opts"
	chartrender "github.com/go-echarts/go-echarts/v2/render"
	"github.com/go-echarts/go-echarts/v2/types"
	g "github.com/maragudk/gomponents"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/pinger"
)

type EChartPoint []interface{}

func (w WUI) wuiDevicePageHandler(wr http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	content := h.Main(
		h.ID("maincontent"),
		h.Class("drawer-content"),
		w.wuiDeviceMain(ctx, r),
	)
	extra := h.Script(h.Src("/static/javascript/echarts.min.js"))
	w.basePage(ctx, "devices", content, extra).Render(wr)
}

func (w WUI) wuiDeviceMain(ctx context.Context, r *http.Request) g.Node {
	// Errors need to be restricted so code does not continue to execute before getting to html
	var (
		d       model.Device
		errNode g.Node
	)
	id := r.PathValue("id")
	addr, err := w.m.StringToAddr(id)
	if err != nil {
		errNode = errAlert(err)
	}
	d, err = w.m.GetDeviceByAddr(ctx, addr)
	if err != nil {
		errNode = errAlert(err)
	}
	dur := 6 * time.Hour

	pingdata, err := w.m.ReadPerformancePings(ctx, d, dur)
	if err != nil {
		errNode = errAlert(err)
	}

	ipflow, err := w.m.FlowSummaryByIP(ctx, d.Addr)
	if err != nil {
		errNode = errAlert(err)
	}
	countryflow, err := w.m.FlowSummaryByCountry(ctx, d.Addr)
	if err != nil {
		errNode = errAlert(err)
	}
	nameflow, err := w.m.FlowSummaryByName(ctx, d.Addr)
	if err != nil {
		errNode = errAlert(err)
	}

	return grid("",
		widecard("Details", deviceToTable(d)),
		g.If(errNode != nil, widecard("Error", errNode)),
		graphcard("Ping Performance",
			lineGraph3(
				meantspoints2echartpoints(pingdata),
				maxtspoints2echartpoints(pingdata),
			),
			lineGraph4(
				losstspoints2echartpoints(pingdata),
			),
		),
		widecard("NetOrg Stats", nameflowSummIPToTable(nameflow)),
		widecard("Country Stats", countryflowSummIPToTable(countryflow)),
		widecard("IP Stats", ipflowSummIPToTable(ipflow)),
	)
}

func deviceToTable(d model.Device) g.Node {
	return h.Table(
		h.Class("table table-zebra"),
		h.TBody(
			toTHTD("Name", d.Name),
			toTHTD("DNS Name", d.Meta.DnsName),
			toTHTD("Addr", d.Addr.String()),
			toTHTD("MAC", d.MAC.String()),
			toTHTD("Manufacturer", d.Meta.Manufacturer),
			toTHTD("Discovered", d.DiscoveredAtString()+" by "+string(d.DiscoveredBy)),
			toTHTD("First Seen", d.FirstSeenString()),
			toTHTD("Last Seen", d.LastSeenString()+"("+d.LastSeenDurString(time.Since)+")"),
			toTHTD("Last Ping Mean", d.LastPingMeanString()),
			toTHTD("Last Ping Maximum", d.LastPingMaximumString()),

			toTHTD("Open Ports", fmt.Sprintf("%d", d.Server.Ports)),
			toTHTD("Last Port Scan", fmt.Sprintf("%s", model.DateTimeFmt(d.Server.LastScan))),
			toTHTD("Tags", fmt.Sprintf("%s", d.Meta.Tags)),

			toTHTD("SNMP Name", d.SNMP.Name),
			toTHTD("SNMP Description", d.SNMP.Description),
			toTHTD("SNMP Community", d.SNMP.Community),
			toTHTD("SNMP Port", strconv.Itoa(d.SNMP.Port)),
			toTHTD("SNMP LastCheck", model.DateTimeFmt(d.SNMP.LastSNMPCheck)),
			toTHTD("SNMP Has ARP Table", fmt.Sprintf("%t", d.SNMP.HasArpTable)),
			toTHTD("SNMP LastArpTableScan", model.DateTimeFmt(d.SNMP.LastArpTableScan)),
			toTHTD("SNMP Interfaces", fmt.Sprintf("%t", d.SNMP.HasInterfaces)),
			toTHTD("SNMP LastInterfacesScan", model.DateTimeFmt(d.SNMP.LastInterfacesScan)),
		),
	)
}

func ipflowSummIPToTable(fs []model.FlowSummaryForAddrByIP) g.Node {
	return wuiTable([]string{"IP", "Country", "Org", "ASN", "In", "Out"},
		g.Group(
			g.Map(fs, func(f model.FlowSummaryForAddrByIP) g.Node {
				return h.Tr(
					h.Td(g.Text(f.Addr.String())),
					h.Td(g.Text(f.Country)),
					h.Td(g.Text(f.Name)),
					h.Td(g.Text(f.Asn)),
					h.Td(g.Text(humanize.Bytes(uint64(f.RecvBytes)))),
					h.Td(g.Text(humanize.Bytes(uint64(f.XmitBytes)))),
				)
			}),
		),
	)
}

func nameflowSummIPToTable(fs []model.FlowSummaryForAddrByName) g.Node {
	return wuiTable([]string{"Org", "In", "Out"},
		g.Group(
			g.Map(fs, func(f model.FlowSummaryForAddrByName) g.Node {
				return h.Tr(
					h.Td(g.Text(f.Name)),
					h.Td(g.Text(humanize.Bytes(uint64(f.RecvBytes)))),
					h.Td(g.Text(humanize.Bytes(uint64(f.XmitBytes)))),
				)
			}),
		),
	)
}

func countryflowSummIPToTable(fs []model.FlowSummaryForAddrByCountry) g.Node {
	return wuiTable([]string{"Country", "Org", "In", "Out"},
		g.Group(
			g.Map(fs, func(f model.FlowSummaryForAddrByCountry) g.Node {
				return h.Tr(
					h.Td(g.Text(f.Country)),
					h.Td(g.Text(f.Name)),
					h.Td(g.Text(humanize.Bytes(uint64(f.RecvBytes)))),
					h.Td(g.Text(humanize.Bytes(uint64(f.XmitBytes)))),
				)
			}),
		),
	)
}

const (
	tplName = "chart"
)

// <div class="item" id="{{ .ChartID }}" style="width:{{ .Initialization.Width }};height:{{ .Initialization.Height }};"></div>

var (
	graphTpl = `
<div class="container">
  <div class="item" id="{{ .ChartID }}" style="width:{{ .Initialization.Width }};height:{{ .Initialization.Height }};"></div>
</div>

<script type="text/javascript">
    "use strict";
    let goecharts_{{ .ChartID | safeJS }} = echarts.init(document.getElementById('{{ .ChartID | safeJS }}'), "{{ .Theme }}", { renderer: "{{  .Initialization.Renderer }}" });
    let option_{{ .ChartID | safeJS }} = {{ .JSONNotEscaped | safeJS }};
    goecharts_{{ .ChartID | safeJS }}.setOption(option_{{ .ChartID | safeJS }});

  {{- range  $listener := .EventListeners }}
    {{if .Query  }}
    goecharts_{{ $.ChartID | safeJS }}.on({{ $listener.EventName }}, {{ $listener.Query | safeJS }}, {{ injectInstance $listener.Handler "%MY_ECHARTS%"  $.ChartID | safeJS }});
    {{ else }}
    goecharts_{{ $.ChartID | safeJS }}.on({{ $listener.EventName }}, {{ injectInstance $listener.Handler "%MY_ECHARTS%"  $.ChartID | safeJS }})
    {{ end }}
  {{- end }}

    {{- range .JSFunctions.Fns }}
    {{ injectInstance . "%MY_ECHARTS%"  $.ChartID  | safeJS }}
    {{- end }}
</script>
  `
	tpl = template.
		Must(template.New(tplName).
			Funcs(template.FuncMap{
				"safeJS": func(s interface{}) template.JS {
					return template.JS(fmt.Sprint(s))
				},
				"isSet": isSet,
				"injectInstance": func(funcStr types.FuncStr, echartsInstancePlaceholder string, chartID string) string {
					instance := chartrender.EchartsInstancePrefix + chartID
					return strings.Replace(
						string(funcStr),
						echartsInstancePlaceholder,
						instance,
						-1,
					)
				},
			}).
			Parse(graphTpl),
		)
)

// func lineGraph2(avg []EChartPoint, maxi []EChartPoint, loss []EChartPoint) g.Node {
// 	line := charts.NewLine()
// 	line.Initialization.Width = "800px"
// 	//line.Theme = "wonderland"
//
// 	// preformat data
// 	avgdata := make([]opts.LineData, len(avg))
// 	for i, point := range avg {
// 		avgdata[i] = opts.LineData{Value: point}
// 	}
// 	maxdata := make([]opts.LineData, len(maxi))
// 	for i, point := range maxi {
// 		maxdata[i] = opts.LineData{Value: point}
// 	}
// 	lossdata := make([]opts.LineData, len(loss))
// 	for i, point := range loss {
// 		lossdata[i] = opts.LineData{Value: point}
// 	}
//
// 	line.AddSeries("Average Response", avgdata, charts.WithLabelOpts(
// 		opts.Label{Show: opts.Bool(true), Position: "bottom"},
// 	))
// 	line.AddSeries("Max Response", maxdata, charts.WithLabelOpts(
// 		opts.Label{Show: opts.Bool(true), Position: "bottom"},
// 	))
// 	line.AddSeries("Packet Loss", lossdata, charts.WithLabelOpts(
// 		opts.Label{Show: opts.Bool(true), Position: "bottom"},
// 	))
// 	line.SetGlobalOptions(
// 		charts.WithTooltipOpts(opts.Tooltip{
// 			Trigger: "axis",
// 			AxisPointer: &opts.AxisPointer{
// 				Type: "cross",
// 			},
// 		}),
// 		charts.WithXAxisOpts(opts.XAxis{
// 			Name:         "Time",
// 			NameLocation: "middle",
// 			Type:         "time",
// 		}),
// 		charts.WithYAxisOpts(opts.YAxis{
// 			Name:         "duration (ms)",
// 			NameLocation: "end",
// 			Type:         "value",
// 			AxisLabel: &opts.AxisLabel{
// 				Formatter: "{value} ms",
// 			},
// 		}),
// 	)
// 	line.ExtendYAxis(opts.YAxis{
// 		Name:         "packet loss (%)",
// 		NameLocation: "end",
// 		Type:         "value",
// 		Show:         opts.Bool(true),
// 		Min:          0,
// 		Max:          100,
// 		AxisLabel: &opts.AxisLabel{
// 			Formatter: "{value} %",
// 		},
// 		/*
// 			AxisLine: &opts.AxisLine{
// 				Show: opts.Bool(false),
// 				LineStyle: &opts.LineStyle{
// 					Opacity: 0,
// 				},
// 			},
// 		*/
// 	})
// 	line.SetSeriesOptions(
// 		charts.WithLineChartOpts(opts.LineChart{
// 			Smooth: opts.Bool(true),
// 		}),
// 		charts.WithLabelOpts(opts.Label{
// 			Show:      opts.Bool(false),
// 			Formatter: "{a}",
// 		}),
// 	)
// 	line.Renderer = newSnippetRenderer(line, line.Validate)
// 	htmlsnippet := renderToString(line)
//
// 	return g.Raw(htmlsnippet)
// }

func lineGraph3(avg []EChartPoint, maxi []EChartPoint) g.Node {
	line := charts.NewLine()
	line.Initialization.Width = "800px"
	//line.Theme = "wonderland"

	// preformat data
	avgdata := make([]opts.LineData, len(avg))
	for i, point := range avg {
		avgdata[i] = opts.LineData{Value: point}
	}
	maxdata := make([]opts.LineData, len(maxi))
	for i, point := range maxi {
		maxdata[i] = opts.LineData{Value: point}
	}

	line.AddSeries("Average Response", avgdata, charts.WithLabelOpts(
		opts.Label{Show: opts.Bool(true), Position: "bottom"},
	))
	line.AddSeries("Max Response", maxdata, charts.WithLabelOpts(
		opts.Label{Show: opts.Bool(true), Position: "bottom"},
	))
	line.SetGlobalOptions(
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			AxisPointer: &opts.AxisPointer{
				Type: "cross",
			},
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name:         "Time",
			NameLocation: "middle",
			Type:         "time",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name:         "duration (ms)",
			NameLocation: "end",
			Type:         "value",
			AxisLabel: &opts.AxisLabel{
				Formatter: "{value} ms",
			},
		}),
	)
	line.SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
		charts.WithLabelOpts(opts.Label{
			Show:      opts.Bool(false),
			Formatter: "{a}",
		}),
	)
	line.Renderer = newSnippetRenderer(line, line.Validate)
	htmlsnippet := renderToString(line)

	return g.Raw(htmlsnippet)
}

func lineGraph4(loss []EChartPoint) g.Node {
	line := charts.NewLine()
	line.Initialization.Width = "800px"
	//line.Theme = "wonderland"

	lossdata := make([]opts.LineData, len(loss))
	for i, point := range loss {
		lossdata[i] = opts.LineData{Value: point}
	}

	line.AddSeries("Packet Loss", lossdata, charts.WithLabelOpts(
		opts.Label{Show: opts.Bool(true), Position: "bottom"},
	))
	line.SetGlobalOptions(
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			AxisPointer: &opts.AxisPointer{
				Type: "cross",
			},
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name:         "Time",
			NameLocation: "middle",
			Type:         "time",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name:         "packet loss (%)",
			NameLocation: "end",
			Type:         "value",
			AxisLabel: &opts.AxisLabel{
				Formatter: "{value} %",
			},
		}),
	)
	line.SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
		charts.WithLabelOpts(opts.Label{
			Show:      opts.Bool(false),
			Formatter: "{a}",
		}),
	)
	line.Renderer = newSnippetRenderer(line, line.Validate)
	htmlsnippet := renderToString(line)

	return g.Raw(htmlsnippet)
}

// func lineGraph(title string, seriesname string, points []EChartPoint) g.Node {
// 	line := charts.NewLine()
// 	line.Initialization.Width = "800px"
// 	//line.Theme = "wonderland"
//
// 	// preformat data
// 	data := make([]opts.LineData, len(points))
// 	for i, point := range points {
// 		data[i] = opts.LineData{Value: point}
// 	}
//
// 	yes := true
// 	no := false
// 	line.AddSeries(seriesname, data)
// 	line.SetGlobalOptions(
// 		charts.WithTooltipOpts(opts.Tooltip{
// 			Trigger: "axis",
// 			AxisPointer: &opts.AxisPointer{
// 				Type: "cross",
// 			},
// 		}),
// 		charts.WithTitleOpts(opts.Title{
// 			Title: title,
// 		}),
// 		charts.WithXAxisOpts(opts.XAxis{
// 			Name:         "Time",
// 			NameLocation: "middle",
// 			Type:         "time",
// 		}),
// 		/*
// 			charts.WithYAxisOpts(opts.YAxis{
// 				// Name:         "milliseconds (ms)",
// 				NameLocation: "left",
// 				Type:         "value",
// 				AxisLabel: &opts.AxisLabel{
// 					Formatter: "{value} ms",
// 				},
// 			}),
// 		*/
// 	)
// 	line.SetSeriesOptions(
// 		charts.WithLineChartOpts(opts.LineChart{
// 			Smooth: types.Bool(&yes),
// 		}),
// 		charts.WithLabelOpts(opts.Label{
// 			Show:      types.Bool(&no),
// 			Formatter: "{a}",
// 		}),
// 	)
// 	line.Renderer = newSnippetRenderer(line, line.Validate)
// 	htmlsnippet := renderToString(line)
//
// 	return g.Raw(htmlsnippet)
// }

func isSet(name string, data interface{}) bool {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return false
	}

	return v.FieldByName(name).IsValid()
}

type snippetRenderer struct {
	c      interface{}
	before []func()
}

func newSnippetRenderer(c interface{}, before ...func()) chartrender.Renderer {
	return &snippetRenderer{c: c, before: before}
}

func (r *snippetRenderer) Render(w io.Writer) error {
	for _, fn := range r.before {
		fn()
	}

	err := tpl.ExecuteTemplate(w, tplName, r.c)
	return err
}

func (r *snippetRenderer) RenderContent() []byte {
	return []byte{}
}

func renderToHtml(c interface{}) template.HTML {
	var buf bytes.Buffer
	r := c.(chartrender.Renderer)
	err := r.Render(&buf)
	if err != nil {
		log.Error(
			"failed to render chart",
			"error",
			err,
		) // TODO: surface this error (UI or Bus or Both)
		return ""
	}
	// fmt.Println(buf.String())

	return template.HTML(buf.String())
}

func renderToString(c interface{}) string {
	var buf bytes.Buffer
	r := c.(chartrender.Renderer)
	err := r.Render(&buf)
	if err != nil {
		log.Error(
			"failed to render chart",
			"error",
			err,
		) // TODO: surface this error (UI or Bus or Both)
		return ""
	}
	// fmt.Println(buf.String())

	return buf.String()
}

// func tspoints2echartpoints(points []pinger.TimeseriesPoint) []EChartPoint {
// 	ret := make([]EChartPoint, len(points))
// 	for i, point := range points {
// 		if math.IsNaN(point.GetValue()) {
// 			ret[i] = EChartPoint{point.GetTime(), 0}
// 			continue
// 		}
// 		ret[i] = EChartPoint{point.GetTime(), point.GetValue()}
// 	}
// 	return ret
// }

func meantspoints2echartpoints(points []pinger.Point) []EChartPoint {
	ret := make([]EChartPoint, len(points))
	for i, point := range points {
		ret[i] = EChartPoint{point.Start, float64(point.Average) / float64(time.Millisecond)}
	}
	return ret
}

func maxtspoints2echartpoints(points []pinger.Point) []EChartPoint {
	ret := make([]EChartPoint, len(points))
	for i, point := range points {
		ret[i] = EChartPoint{point.Start, float64(point.Maximum) / float64(time.Millisecond)}
	}
	return ret
}

func losstspoints2echartpoints(points []pinger.Point) []EChartPoint {
	ret := make([]EChartPoint, len(points))
	for i, point := range points {
		ret[i] = EChartPoint{point.Start, point.Loss * 100.0}
	}
	return ret
}
