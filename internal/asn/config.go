// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package asn

import (
	"github.com/spf13/pflag"

	"github.com/networkables/mason/internal/flagset"
)

type Config struct {
	Enabled       bool
	AsnUrl        string
	CountryUrl    string
	Directory     string
	CacheFilename string
}

const (
	defaultAsnUrl        = "https://github.com/sapics/ip-location-db/raw/main/asn/asn-ipv4.csv"
	defaultCountryUrl    = "https://github.com/sapics/ip-location-db/raw/main/geo-whois-asn-country/geo-whois-asn-country-ipv4.csv"
	defaultCacheFilename = "cache.mpz1"
)

func SetFlags(pflags *pflag.FlagSet, cfg *Config) {
	configMajorKey := "asn"

	flagset.Bool(
		pflags,
		&cfg.Enabled,
		configMajorKey,
		"enabled",
		false,
		"Enabled look ups of IP/ASN information",
	)
	flagset.String(
		pflags,
		&cfg.AsnUrl,
		configMajorKey,
		"asnurl",
		defaultAsnUrl,
		"Github url of the asn-ipv4.csv file",
	)
	flagset.String(
		pflags,
		&cfg.CountryUrl,
		configMajorKey,
		"countryurl",
		defaultCountryUrl,
		"Github url of the geo-whois-asn-country-ipv4.csv file",
	)
	flagset.String(
		pflags,
		&cfg.Directory,
		configMajorKey,
		"directory",
		"data/asn",
		"location to store the asn cache db",
	)
	flagset.String(
		pflags,
		&cfg.CacheFilename,
		configMajorKey,
		"cachefilename",
		defaultCacheFilename,
		"filename of the asn cache db",
	)
}
