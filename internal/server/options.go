// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package server provides the Mason application used for network discovery and management
package server

import (
	"github.com/networkables/mason/internal/bus"
)

type Options struct {
	scanstatus *string
	cfg        *Config
	bus        bus.Bus
	store      Storer
	nfstore    NetflowStorer
}

type Option func(*Options)

func applyOptionsToDefault(opts ...Option) *Options {
	o := defaultOptions()
	return applyOptions(o, opts...)
}

func applyOptions(base *Options, opts ...Option) *Options {
	for _, f := range opts {
		f(base)
	}
	return base
}

func defaultOptions() *Options {
	scanstat := ""
	return &Options{
		scanstatus: &scanstat,
	}
}

func WithConfig(x *Config) Option {
	return func(o *Options) {
		o.cfg = x
	}
}

func WithBus(x bus.Bus) Option {
	return func(o *Options) {
		o.bus = x
	}
}

func WithStore(x Storer) Option {
	return func(o *Options) {
		o.store = x
	}
}

func WithNetflowStorer(x NetflowStorer) Option {
	return func(o *Options) {
		o.nfstore = x
	}
}
