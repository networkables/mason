// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package oui

type Options struct {
	url       string
	directory string
	filename  string
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
	return &Options{
		url:      defaultUrl,
		filename: defaultFilename,
	}
}

func WithUrl(x string) Option {
	return func(o *Options) {
		o.url = x
	}
}

func WithDirectory(x string) Option {
	return func(o *Options) {
		o.directory = x
	}
}

func WithFilename(x string) Option {
	return func(o *Options) {
		o.filename = x
	}
}
