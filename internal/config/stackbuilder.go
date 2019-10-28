// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/luids-io/common/util"
)

// StackBuilderCfg stores stack builder prefs
type StackBuilderCfg struct {
	CertsDir string
	DataDir  string
	CacheDir string
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *StackBuilderCfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	pflag.StringVar(&cfg.CertsDir, aprefix+"certsdir", cfg.CertsDir, "Path to certificate files.")
	pflag.StringVar(&cfg.DataDir, aprefix+"datadir", cfg.DataDir, "Path to data files.")
	pflag.StringVar(&cfg.CacheDir, aprefix+"cachedir", cfg.CacheDir, "Path to cache.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *StackBuilderCfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	util.BindViper(v, aprefix+"certsdir")
	util.BindViper(v, aprefix+"datadir")
	util.BindViper(v, aprefix+"cachedir")
}

// FromViper fill values from viper
func (cfg *StackBuilderCfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.CertsDir = v.GetString(aprefix + "certsdir")
	cfg.DataDir = v.GetString(aprefix + "datadir")
	cfg.CacheDir = v.GetString(aprefix + "cachedir")
}

// Empty returns true if configuration is empty
func (cfg StackBuilderCfg) Empty() bool {
	if cfg.CertsDir != "" {
		return false
	}
	if cfg.DataDir != "" {
		return false
	}
	if cfg.CacheDir != "" {
		return false
	}
	return true
}

// Validate checks that configuration is ok
func (cfg StackBuilderCfg) Validate() error {
	if cfg.CertsDir != "" {
		if !util.DirExists(cfg.CertsDir) {
			return fmt.Errorf("certificates dir '%v' doesn't exists", cfg.CertsDir)
		}
	}
	if cfg.DataDir != "" {
		if !util.DirExists(cfg.DataDir) {
			return fmt.Errorf("data dir '%v' doesn't exists", cfg.DataDir)
		}
	}
	if cfg.CacheDir != "" {
		if !util.DirExists(cfg.CacheDir) {
			return fmt.Errorf("cache dir '%v' doesn't exists", cfg.CacheDir)
		}
	}
	return nil
}

// Dump configuration
func (cfg StackBuilderCfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
