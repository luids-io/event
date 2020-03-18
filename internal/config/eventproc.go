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
	StackDirs  []string
	StackFiles []string
	StackMain  string
	Workers    int
	CertsDir   string
	DataDir    string
	CacheDir   string
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *EventProcCfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	if short {
		pflag.StringSliceVarP(&cfg.StackDirs, aprefix+"dirs", "S", cfg.StackDirs, "Config stack dirs.")
		pflag.StringSliceVarP(&cfg.StackFiles, aprefix+"files", "s", cfg.StackFiles, "Config stack files.")
	} else {
		pflag.StringSliceVar(&cfg.StackDirs, aprefix+"dirs", cfg.StackDirs, "Config stack dirs.")
		pflag.StringSliceVar(&cfg.StackFiles, aprefix+"files", cfg.StackFiles, "Config stack files.")
	}
	pflag.StringVar(&cfg.StackMain, aprefix+"main", cfg.StackMain, "Stack main name.")
	pflag.IntVar(&cfg.Workers, aprefix+"workers", cfg.Workers, "Number of workers.")
	pflag.StringVar(&cfg.CertsDir, aprefix+"certsdir", cfg.CertsDir, "Path to certificate files.")
	pflag.StringVar(&cfg.DataDir, aprefix+"datadir", cfg.DataDir, "Path to data files.")
	pflag.StringVar(&cfg.CacheDir, aprefix+"cachedir", cfg.CacheDir, "Path to cache.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *EventProcCfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	util.BindViper(v, aprefix+"dirs")
	util.BindViper(v, aprefix+"files")
	util.BindViper(v, aprefix+"main")
	util.BindViper(v, aprefix+"workers")
	util.BindViper(v, aprefix+"certsdir")
	util.BindViper(v, aprefix+"datadir")
	util.BindViper(v, aprefix+"cachedir")
}

// FromViper fill values from viper
func (cfg *EventProcCfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.StackDirs = v.GetStringSlice(aprefix + "dirs")
	cfg.StackFiles = v.GetStringSlice(aprefix + "files")
	cfg.StackMain = v.GetString(aprefix + "main")
	cfg.Workers = v.GetInt(aprefix + "workers")
	cfg.CertsDir = v.GetString(aprefix + "certsdir")
	cfg.DataDir = v.GetString(aprefix + "datadir")
	cfg.CacheDir = v.GetString(aprefix + "cachedir")
}

// Empty returns true if configuration is empty
func (cfg EventProcCfg) Empty() bool {
	if len(cfg.StackFiles) > 0 {
		return false
	}
	if len(cfg.StackDirs) > 0 {
		return false
	}
	if cfg.StackMain != "" {
		return false
	}
	if cfg.Workers > 0 {
		return false
	}
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
func (cfg EventProcCfg) Validate() error {
	for _, file := range cfg.StackFiles {
		if !util.FileExists(file) {
			return fmt.Errorf("stack config file '%v' doesn't exists", file)
		}
	}
	for _, dir := range cfg.StackDirs {
		if !util.DirExists(dir) {
			return fmt.Errorf("stack config dir '%v' doesn't exists", dir)
		}
	}
	if cfg.StackMain == "" {
		return errors.New("stack main label can't be empty")
	}
	if cfg.Workers < 0 {
		return errors.New("invalid workers value")
	}
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
func (cfg EventProcCfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
