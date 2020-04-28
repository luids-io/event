// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/luids-io/common/util"
)

// EventForwardAPICfg stores event service preferences
type EventForwardAPICfg struct {
	Enable bool
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *EventForwardAPICfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	pflag.BoolVar(&cfg.Enable, aprefix+"enable", cfg.Enable, "Enable event forward api.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *EventForwardAPICfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	util.BindViper(v, aprefix+"enable")
}

// FromViper fill values from viper
func (cfg *EventForwardAPICfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.Enable = v.GetBool(aprefix + "enable")
}

// Empty returns true if configuration is empty
func (cfg EventForwardAPICfg) Empty() bool {
	return false
}

// Validate checks that configuration is ok
func (cfg EventForwardAPICfg) Validate() error {
	return nil
}

// Dump configuration
func (cfg EventForwardAPICfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
