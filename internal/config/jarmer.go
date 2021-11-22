package config

import (
	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"reflect"
)

type JarmConfig struct {
	Fingerprints map[string]string `fig:"fingerprints,required"`
	DefaultPort  int               `fig:"default_port,required"`
}

func (c *config) Jarm() *JarmConfig {
	return c.jarm.Do(func() interface{} {
		var config JarmConfig

		err := figure.
			Out(&config).
			With(figure.BaseHooks, StringMapStringHook).
			From(kv.MustGetStringMap(c.getter, "jarm")).
			Please()
		if err != nil {
			panic(err)
		}

		return &config
	}).(*JarmConfig)
}

var StringMapStringHook = figure.Hooks{
	"map[string]string": func(value interface{}) (reflect.Value, error) {
		result := cast.ToStringMapString(value)
		return reflect.ValueOf(result), nil
	},
}
