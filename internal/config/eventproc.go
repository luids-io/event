// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/luids-io/common/util"
)

// EventProcCfg defines the configuration of a processor
type EventProcCfg struct {
	DatabaseDirs  []string
	DatabaseFiles []string
	Workers       int
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *EventProcCfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	if short {
		pflag.StringSliceVarP(&cfg.DatabaseDirs, aprefix+"dirs", "S", cfg.DatabaseDirs, "Config stack dirs.")
		pflag.StringSliceVarP(&cfg.DatabaseFiles, aprefix+"files", "s", cfg.DatabaseFiles, "Config stack files.")
	} else {
		pflag.StringSliceVar(&cfg.DatabaseDirs, aprefix+"dirs", cfg.DatabaseDirs, "Config stack dirs.")
		pflag.StringSliceVar(&cfg.DatabaseFiles, aprefix+"files", cfg.DatabaseFiles, "Config stack files.")
	}
	pflag.IntVar(&cfg.Workers, aprefix+"workers", cfg.Workers, "Number of workers.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *EventProcCfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	util.BindViper(v, aprefix+"dirs")
	util.BindViper(v, aprefix+"files")
	util.BindViper(v, aprefix+"workers")
}

// FromViper fill values from viper
func (cfg *EventProcCfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.DatabaseDirs = v.GetStringSlice(aprefix + "dirs")
	cfg.DatabaseFiles = v.GetStringSlice(aprefix + "files")
	cfg.Workers = v.GetInt(aprefix + "workers")
}

// Empty returns true if configuration is empty
func (cfg EventProcCfg) Empty() bool {
	if len(cfg.DatabaseFiles) > 0 {
		return false
	}
	if len(cfg.DatabaseDirs) > 0 {
		return false
	}
	if cfg.Workers > 0 {
		return false
	}
	return true
}

// Validate checks that configuration is ok
func (cfg EventProcCfg) Validate() error {
	for _, file := range cfg.DatabaseFiles {
		if !util.FileExists(file) {
			return fmt.Errorf("stack config file '%v' doesn't exists", file)
		}
	}
	for _, dir := range cfg.DatabaseDirs {
		if !util.DirExists(dir) {
			return fmt.Errorf("stack config dir '%v' doesn't exists", dir)
		}
	}
	if cfg.Workers < 0 {
		return errors.New("invalid workers value")
	}
	return nil
}

// Dump configuration
func (cfg EventProcCfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
