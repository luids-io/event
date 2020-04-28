// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/luids-io/common/util"
)

// EventNotifyAPICfg stores event service preferences
type EventNotifyAPICfg struct {
	Enable bool
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *EventNotifyAPICfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	pflag.BoolVar(&cfg.Enable, aprefix+"enable", cfg.Enable, "Enable event notify api.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *EventNotifyAPICfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	util.BindViper(v, aprefix+"enable")
}

// FromViper fill values from viper
func (cfg *EventNotifyAPICfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.Enable = v.GetBool(aprefix + "enable")
}

// Empty returns true if configuration is empty
func (cfg EventNotifyAPICfg) Empty() bool {
	return false
}

// Validate checks that configuration is ok
func (cfg EventNotifyAPICfg) Validate() error {
	return nil
}

// Dump configuration
func (cfg EventNotifyAPICfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
