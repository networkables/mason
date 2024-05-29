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
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	charts "github.com/go-echarts/go-echarts/v2/charts"
	opts "github.com/go-echarts/go-echarts/v2/opts"
	chartrender "github.com/go-echarts/go-echarts/v2/render"
	"github.com/go-echarts/go-echarts/v2/types"
	g "github.com/maragudk/gomponents"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/internal/device"
	"github.com/networkables/mason/internal/pinger"
)

type EChartPoint []interface{}

func (w WUI) wuiDevicePageHandler(wr http.ResponseWriter, r *http.Request) {
	content := h.Main(
		h.ID("maincontent"),
		h.Class("drawer-content"),
		w.wuiDeviceMain(r),
	)
	extra := h.Script(h.Src("/static/javascript/echarts.min.js"))
	w.basePage("devices", content, extra).Render(wr)
}

func (w WUI) wuiDeviceMain(r *http.Request) g.Node {
	// Errors need to be restricted so code does not continue to execute before getting to html
	var (
		d       device.Device
		errNode g.Node
	)
	id := r.PathValue("id")
	addr, err := w.m.StringToAddr(id)
	if err != nil {
		errNode = errAlert(err)
	}
	d, err = w.m.GetDeviceByAddr(addr)
	if err != nil {
		errNode = errAlert(err)
	}
	dur := 6 * time.Hour
	ctx := context.Background()

	avgdata, err := w.m.ReadTimeseriesPoints(ctx, d, dur, pinger.TimeseriesAvgResponse{})
	if err != nil {
		errNode = errAlert(err)
	}
	maxdata, err := w.m.ReadTimeseriesPoints(ctx, d, dur, pinger.TimeseriesMaxResponse{})
	if err != nil {
		errNode = errAlert(err)
	}
	lossdata, err := w.m.ReadTimeseriesPoints(ctx, d, dur, pinger.TimeseriesPacketLoss{})
	if err != nil {
		errNode = errAlert(err)
	}
	return h.Div(
		h.Div(
			h.Class("grid grid-cols-12 grid-rows-[min-content] gap-y-12 p-4 lg:gap-x-12 lg:p-10"),
			h.Section(
				h.Class("card col-span-12 overflow-hidden bg-base-100 shadow-sm xl:col-span-10"),
				errNode,
				h.Div(h.Class("card-body grow-0"),
					h.H2(h.Class("card-title"),
						h.A(
							h.Class("link-hover link"),
							g.Text("Details"),
						),
					),
					h.Div(
						h.Class("overflow-x-auto"),
						deviceToTable(d),
					),
				),
			),
			h.Section(
				h.Class("card col-span-12 overflow-hidden bg-base-100 shadow-sm xl:col-span-10"),
				h.StyleEl(
					g.Text(`
						.container {margin-top:30px; display: flex;justify-content: center;align-items: center;} 
            .item {margin: auto;}
          `),
				),
				errNode,
				h.Div(
					h.Class("card-body grow-0"),
					h.H2(h.Class("card-title"),
						h.A(
							h.Class("link-hover link"),
							g.Text("Ping Performance"),
						),
					),
					lineGraph3(
						tspoints2echartpoints(avgdata),
						tspoints2echartpoints(maxdata),
					),
					lineGraph4(
						tspoints2echartpoints(lossdata),
					),
				),
			),
		),
	)
}

func deviceToTable(d device.Device) g.Node {
	return h.Table(
		h.Class("table table-zebra"),
		h.TBody(
			// h.Tr(h.Th(svgBarChart()), h.Td(g.Text(" "))),
			h.Tr(h.Th(g.Text("Name")), h.Td(g.Text(d.Name))),
			h.Tr(h.Th(g.Text("DNS Name")), h.Td(g.Text(d.Meta.DnsName))),
			h.Tr(h.Th(g.Text("Addr")), h.Td(g.Text(d.Addr.String()))),
			h.Tr(h.Th(g.Text("MAC")), h.Td(g.Text(d.MAC.String()))),
			h.Tr(h.Th(g.Text("Manufacturer")), h.Td(g.Text(d.Meta.Manufacturer))),
			h.Tr(
				h.Th(g.Text("Discovered")),
				h.Td(g.Text(d.DiscoveredAtString()+" by "+string(d.DiscoveredBy))),
			),
			h.Tr(h.Th(g.Text("First Seen")), h.Td(g.Text(d.FirstSeenString()))),
			h.Tr(
				h.Th(g.Text("Last Seen")),
				h.Td(g.Text(d.LastSeenString()+"("+d.LastSeenDurString(time.Since)+")")),
			),
			toTD("Last Ping Mean", d.LastPingMeanString()),
			toTD("Last Ping Maximum", d.LastPingMaximumString()),
			h.Tr(h.Th(g.Text("Open Ports")), h.Td(g.Text(fmt.Sprintf("%d", d.Server.Ports)))),
			h.Tr(
				h.Th(g.Text("Last Port Scan")),
				h.Td(g.Text(fmt.Sprintf("%s", device.DateTimeFmt(d.Server.LastScan)))),
			),
			h.Tr(h.Th(g.Text("Tags")), h.Td(g.Text(fmt.Sprintf("%s", d.Meta.Tags)))),

			h.Tr(h.Th(g.Text("SNMP Name")), h.Td(g.Text(d.SNMP.Name))),
			h.Tr(h.Th(g.Text("SNMP Description")), h.Td(g.Text(d.SNMP.Description))),
			h.Tr(h.Th(g.Text("SNMP Community")), h.Td(g.Text(d.SNMP.Community))),
			h.Tr(h.Th(g.Text("SNMP Port")), h.Td(g.Text(strconv.Itoa(d.SNMP.Port)))),
			h.Tr(
				h.Th(g.Text("SNMP LastCheck")),
				h.Td(g.Text(device.DateTimeFmt(d.SNMP.LastSNMPCheck))),
			),
			h.Tr(
				h.Th(g.Text("SNMP Has ARP Table")),
				h.Td(g.Text(fmt.Sprintf("%t", d.SNMP.HasArpTable))),
			),
			h.Tr(
				h.Th(g.Text("SNMP LastArpTableScan")),
				h.Td(g.Text(device.DateTimeFmt(d.SNMP.LastArpTableScan))),
			),
			h.Tr(
				h.Th(g.Text("SNMP Interfaces")),
				h.Td(g.Text(fmt.Sprintf("%t", d.SNMP.HasInterfaces))),
			),
			h.Tr(
				h.Th(g.Text("SNMP LastInterfacesScan")),
				h.Td(g.Text(device.DateTimeFmt(d.SNMP.LastInterfacesScan))),
			),
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

func tspoints2echartpoints(points []pinger.TimeseriesPoint) []EChartPoint {
	ret := make([]EChartPoint, len(points))
	for i, point := range points {
		if math.IsNaN(point.GetValue()) {
			ret[i] = EChartPoint{point.GetTime(), 0}
			continue
		}
		ret[i] = EChartPoint{point.GetTime(), point.GetValue()}
	}
	return ret
}
