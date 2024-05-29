// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

const (
	EnabledCfgKey         = "enabled"
	MaxWorkersCfgKey      = "maxworkers"
	CheckIntervalCfgKey   = "checkinterval"
	DefaultIntervalCfgKey = "defaultinterval"
	ServerIntervalCfgKey  = "serverinterval"
	PerAddrPauseCfgKey    = "peraddrpause"
	TimeoutCfgKey         = "timeout"
	PingCountCfgKey       = "pingcount"
	CommunityCfgKey       = "community"
	PortsCfgKey           = "ports"
	EnableDebugLogCfgKey  = "enabledebuglog"
)

func Key(keys ...string) string {
	return strings.Join(keys, ".")
}

var (
	configonce      sync.Once
	configsingleton *Config
)

type Config struct {
	Store             *Store
	Server            *Server
	Bus               *Bus
	Discovery         *Discovery
	PerformancePinger *Pinger
	Enrichment        *Enrichment
}

func GetConfig() *Config {
	configonce.Do(func() {
		configsingleton = defaultConfig()
	})
	return configsingleton
}

const (
	configName = "test"
	configType = "yaml"
)

func defaultConfig() *Config {
	c := &Config{}
	cfgStoreSetDefaults()
	cfgServerSetDefaults()
	cfgBusSetDefaults()
	cfgDiscoSetDefaults()
	cfgPerfPingSetDefaults()
	cfgEnrichSetDefaults()

	// viper.SetConfigName(configName)
	// viper.SetConfigType(configType)
	// viper.AddConfigPath(".")    // optionally look for config in the working directory
	// err := viper.ReadInConfig() // Find and read the config file
	// if err != nil {             // Handle errors reading the config file
	// 	log.Fatal("viper read", "error", err)
	// }
	// viper.SetEnvPrefix("MASON")

	err := viper.Unmarshal(c)
	if err != nil {
		log.Fatal(err)
	}

	return c
	// valof := reflect.ValueOf(c)
	// if valof.Kind() == reflect.Pointer {
	// 	valof = valof.Elem()
	// }
	// if valof.Kind() == reflect.Struct {
	// 	cfgHandleStruct(valof)
	// }
}

func (c Config) Save() error {
	fname := fmt.Sprintf("%s-%d.%s", configName, time.Now().Unix(), configType)
	return viper.WriteConfigAs(fname)
}

//
// func cfgHandleStruct(val reflect.Value) {
// 	tp := val.Type()
// 	for i := 0; i < val.NumField(); i++ {
// 		valf := val.Field(i)
// 		fieldf := tp.Field(i)
//
// 		if valf.Kind() == reflect.Struct {
// 			cfgHandleStruct(valf)
// 			continue
// 		}
// 		setFieldDefault(valf, fieldf)
// 		// cfgHandleField(i, valf, fieldf)
// 	}
// }
//
// func cfgHandleField(idx int, valf reflect.Value, sf reflect.StructField) {
// 	log.Printf(
// 		"field%d  kind:%s  type:%s  name:%s tag:%s\n",
// 		idx,
// 		valf.Kind(),
// 		valf.Type(),
// 		sf.Name,
// 		sf.Tag.Get("default"),
// 	)
// }
//
// const (
// 	cfgDefaultValueTag = "default"
// )
//
// func setFieldDefault(valf reflect.Value, sf reflect.StructField) {
// 	if valf.CanSet() {
// 		defaultValueString := sf.Tag.Get(cfgDefaultValueTag)
// 		switch valf.Kind() {
// 		case reflect.String:
// 			valf.SetString(defaultValueString)
// 		case reflect.Int:
// 			if defaultValueString == "" {
// 				valf.SetInt(0)
// 				return
// 			}
// 			iv, err := strconv.Atoi(defaultValueString)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			valf.SetInt(int64(iv))
// 		case reflect.Bool:
// 			if defaultValueString == "" {
// 				valf.SetBool(false)
// 				return
// 			}
// 			b, err := strconv.ParseBool(defaultValueString)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			valf.SetBool(b)
// 		default:
// 			// non std lib types
// 			switch valf.Type().String() {
// 			case "time.Duration":
// 				if defaultValueString == "" {
// 					var e time.Duration
// 					valf.Set(reflect.ValueOf(e))
// 					return
// 				}
// 				dur, err := time.ParseDuration(defaultValueString)
// 				if err != nil {
// 					log.Fatal(err)
// 				}
// 				valf.Set(reflect.ValueOf(dur))
// 			default:
// 				log.Warn(
// 					"cfg setdefault unhandled field type",
// 					"fieldname",
// 					sf.Name,
// 					"type",
// 					valf.Type(),
// 					"kind",
// 					valf.Kind(),
// 				)
// 			}
// 		}
// 	} else {
// 		log.Warn("cfg cannot set default value", "fieldname", sf.Name)
// 	}
// }
//
// func (c Config) String() string {
// 	s := ""
// 	for _, t := range c.ToTuple() {
// 		s += fmt.Sprintf("%-40s : %s\n", t.Name, t.Value)
// 	}
// 	return s
// }
//
// type ConfigTuple struct {
// 	Name  string
// 	Value string
// }
//
// func (c Config) ToTuple() []ConfigTuple {
// 	tuples := make([]ConfigTuple, 0)
//
// 	valof := reflect.ValueOf(c)
// 	if valof.Kind() == reflect.Pointer {
// 		valof = valof.Elem()
// 	}
// 	if valof.Kind() == reflect.Struct {
// 		tp := valof.Type()
// 		t := cfgTupleStruct("root", valof, tp)
// 		tuples = append(tuples, t...)
// 	}
//
// 	return tuples
// }
//
// func cfgTupleStruct(prefix string, val reflect.Value, tp reflect.Type) []ConfigTuple {
// 	tuples := make([]ConfigTuple, 0, val.NumField())
// 	for i := 0; i < val.NumField(); i++ {
// 		valf := val.Field(i)
// 		fieldf := tp.Field(i)
//
// 		if valf.Kind() == reflect.Struct {
// 			tuples = append(tuples, cfgTupleStruct(prefix+"."+fieldf.Name, valf, fieldf.Type)...)
// 			continue
// 		}
// 		tuple := ConfigTuple{
// 			Name:  prefix + "." + fieldf.Name,
// 			Value: cfgStringValue(valf),
// 		}
// 		tuples = append(tuples, tuple)
// 	}
// 	return tuples
// }
//
// func cfgStringValue(val reflect.Value) string {
// 	i := val.Interface()
// 	switch i.(type) {
// 	case string:
// 		return i.(string)
// 	case bool:
// 		return fmt.Sprintf("%t", i.(bool))
// 	case int:
// 		return strconv.Itoa(i.(int))
// 	case time.Duration:
// 		return (i.(time.Duration)).String()
// 	default:
// 		log.Warnf("cfgStringValue did not handle type: %T", i)
// 	}
// 	return ""
// }
//
// /*
// func (c *Config) SetupFlags(flags *flag.FlagSet) {
// }
//
// func cfgSetFlagStruct(flags *flag.FlagSet, prefix string, val reflect.Value, tp reflect.Type) {
// 	for i := 0; i < val.NumField(); i++ {
// 		valf := val.Field(i)
// 		fieldf := tp.Field(i)
//
// 		if valf.Kind() == reflect.Struct {
// 			cfgSetFlagStruct(flags, prefix+"."+fieldf.Name, valf, fieldf.Type)
// 			continue
// 		}
// 		cfgSetFlag(flags, prefix+"."+fieldf.Name, valf, fieldf.Tag.Get("flagdesc"))
// 	}
// }
//
// func cfgSetFlag(
// 	flags *flag.FlagSet,
// 	name string,
// 	val reflect.Value,
// 	flagdesc string,
// ) {
// 	i := val.Interface()
// 	switch i.(type) {
// 	case string:
// 	case bool:
// 		flags.BoolVar((val.Addr().Pointer()).(*bool), name, i.(bool), flagdesc)
// 	default:
// 		log.Warnf("cfgSetFlag did not handle type: %T", i)
// 	}
// }
// */
