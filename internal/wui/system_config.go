// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wui

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	g "github.com/maragudk/gomponents"
	h "github.com/maragudk/gomponents/html"

	"github.com/networkables/mason/internal/server"
)

func (w WUI) wuiConfigPageHandler(wr http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	content := h.Main(
		h.ID("maincontent"),
		h.Class("drawer-content"),
		w.wuiConfigMain(),
	)
	w.basePage(ctx, "config", content, nil).Render(wr)
}

func (w WUI) wuiConfigMain() g.Node {
	return grid("",
		wuiCard("Theme",
			h.Div(
				h.Button(
					h.DataAttr("act-class", "shadow-outline"),
					h.DataAttr("set-theme", ""),
					h.Class("btn"),
					g.Text("Default"),
				),
				h.Button(
					h.DataAttr("act-class", "shadow-outline"),
					h.DataAttr("set-theme", "dark"),
					h.Class("btn"),
					g.Text("Dark"),
				),
				h.Button(
					h.DataAttr("act-class", "shadow-outline"),
					h.DataAttr("set-theme", "light"),
					h.Class("btn"),
					g.Text("Light"),
				),
			),
		),
		wuiCard("Config",
			configToTable(w.m.GetConfig()),
		),
	)
}

func configToTable(cfg *server.Config) g.Node {
	val := reflect.ValueOf(cfg)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	rows := cfgProcLevel("", val)
	return h.Table(
		h.Class("table table-zebra"),
		h.THead(
			h.Tr(
				h.Th(g.Text("Name")),
				h.Th(g.Text("Value")),
			),
		),
		h.TBody(
			rows...,
		),
	)
}

func cfgProcLevel(prefix string, val reflect.Value) []g.Node {
	tp := val.Type()
	nodes := make([]g.Node, 0)

	for i := 0; i < val.NumField(); i++ {
		field := tp.Field(i)
		fieldType := field.Type
		if field.Type.Kind() == reflect.Pointer {
			fieldType = field.Type.Elem()
		}

		fieldval := val.Field(i)
		if fieldval.Kind() == reflect.Pointer {
			fieldval = fieldval.Elem()
		}

		fieldname := field.Name
		if prefix != "" {
			fieldname = prefix + "." + fieldname
		}
		switch fieldType.Kind() {
		case reflect.Struct:
			nodes = append(nodes, cfgProcLevel(fieldname, fieldval)...)
		default:
			nodes = append(nodes, cfgValToNode(fieldname, fieldval))
		}
	}
	return nodes
}

func cfgValToNode(name string, val reflect.Value) g.Node {
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Int:
		return h.Tr(
			h.Td(g.Text(name)),
			h.Td(g.Text(strconv.Itoa(int(val.Int())))),
		)

	case reflect.Bool:
		return h.Tr(
			h.Td(g.Text(name)),
			h.Td(g.Text(strconv.FormatBool(val.Bool()))),
		)

	case reflect.Slice:
		switch val.Type().Elem().Kind() {
		case reflect.Int:
			return h.Tr(
				h.Td(g.Text(name)),
				h.Td(g.Text(fmt.Sprintf("%d", val.Interface().([]int)))),
			)
		}
	}
	return h.Tr(
		h.Td(g.Text(name)),
		h.Td(g.Text(fmt.Sprintf("%s", val.Interface()))),
	)
}
