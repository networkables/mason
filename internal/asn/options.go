// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package asn

type Options struct {
	asnurl        string
	countryurl    string
	directory     string
	cachefilename string
	store         asnstorer
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
		asnurl:        defaultAsnUrl,
		countryurl:    defaultCountryUrl,
		cachefilename: defaultCacheFilename,
	}
}

func WithAsnUrl(x string) Option {
	return func(o *Options) {
		o.asnurl = x
	}
}

func WithCountryUrl(x string) Option {
	return func(o *Options) {
		o.countryurl = x
	}
}

func WithDirectory(x string) Option {
	return func(o *Options) {
		o.directory = x
	}
}

func WithCacheFilename(x string) Option {
	return func(o *Options) {
		o.cachefilename = x
	}
}

func WithStorer(x asnstorer) Option {
	return func(o *Options) {
		o.store = x
	}
}
