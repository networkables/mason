// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package flagset

import (
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Key(keys ...string) string {
	return strings.Join(keys, ".")
}

func Bool(f *pflag.FlagSet, v *bool, keyMajor string, keyMinor string, def bool, desc string) {
	key := Key(keyMajor, keyMinor)
	// viper.SetDefault(key, def)
	f.BoolVar(v, key, def, desc)
	viper.BindPFlag(key, f.Lookup(key))
}

func Int(f *pflag.FlagSet, v *int, keyMajor string, keyMinor string, def int, desc string) {
	key := Key(keyMajor, keyMinor)
	// viper.SetDefault(key, def)
	f.IntVar(v, key, def, desc)
	viper.BindPFlag(key, f.Lookup(key))
}

func IntSlice(
	f *pflag.FlagSet,
	v *[]int,
	keyMajor string,
	keyMinor string,
	def []int,
	desc string,
) {
	key := Key(keyMajor, keyMinor)
	// viper.SetDefault(key, def)
	f.IntSliceVar(v, key, def, desc)
	viper.BindPFlag(key, f.Lookup(key))
}

func String(
	f *pflag.FlagSet,
	v *string,
	keyMajor string,
	keyMinor string,
	def string,
	desc string,
) {
	key := Key(keyMajor, keyMinor)
	// viper.SetDefault(key, def)
	f.StringVar(v, key, def, desc)
	viper.BindPFlag(key, f.Lookup(key))
}

func StringSlice(
	f *pflag.FlagSet,
	v *[]string,
	keyMajor string,
	keyMinor string,
	def []string,
	desc string,
) {
	key := Key(keyMajor, keyMinor)
	// viper.SetDefault(key, def)
	f.StringSliceVar(v, key, def, desc)
	viper.BindPFlag(key, f.Lookup(key))
}

func Duration(
	f *pflag.FlagSet,
	v *time.Duration,
	keyMajor string,
	keyMinor string,
	def time.Duration,
	desc string,
) {
	key := Key(keyMajor, keyMinor)
	// viper.SetDefault(key, def)
	f.DurationVar(v, key, def, desc)
	viper.BindPFlag(key, f.Lookup(key))
}
