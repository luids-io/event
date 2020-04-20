// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/luids-io/common/util"
)

// StackCfg defines the configuration of stack
type StackCfg struct {
	Dirs  []string
	Files []string
	Main  string
}

// EventDBCfg defines the configuration of the event database
type EventDBCfg struct {
	Dirs  []string
	Files []string
}

// EventProcCfg defines the configuration of a processor
type EventProcCfg struct {
	Stack    StackCfg
	DB       EventDBCfg
	Workers  int
	CertsDir string
	DataDir  string
	CacheDir string
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *EventProcCfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	if short {
		pflag.StringSliceVarP(&cfg.Stack.Dirs, aprefix+"stack.dirs", "S", cfg.Stack.Dirs, "Config stack dirs.")
		pflag.StringSliceVarP(&cfg.Stack.Files, aprefix+"stack.files", "s", cfg.Stack.Files, "Config stack files.")
	} else {
		pflag.StringSliceVar(&cfg.Stack.Dirs, aprefix+"stack.dirs", cfg.Stack.Dirs, "Config stack dirs.")
		pflag.StringSliceVar(&cfg.Stack.Files, aprefix+"stack.files", cfg.Stack.Files, "Config stack files.")
	}
	pflag.StringVar(&cfg.Stack.Main, aprefix+"stack.main", cfg.Stack.Main, "Stack main name.")
	pflag.StringSliceVar(&cfg.DB.Dirs, aprefix+"db.dirs", cfg.DB.Dirs, "Config event database dirs.")
	pflag.StringSliceVar(&cfg.DB.Files, aprefix+"db.files", cfg.DB.Files, "Config event database files.")
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
	util.BindViper(v, aprefix+"stack.dirs")
	util.BindViper(v, aprefix+"stack.files")
	util.BindViper(v, aprefix+"stack.main")
	util.BindViper(v, aprefix+"db.dirs")
	util.BindViper(v, aprefix+"db.files")
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
	cfg.Stack.Dirs = v.GetStringSlice(aprefix + "stack.dirs")
	cfg.Stack.Files = v.GetStringSlice(aprefix + "stack.files")
	cfg.Stack.Main = v.GetString(aprefix + "stack.main")
	cfg.DB.Dirs = v.GetStringSlice(aprefix + "db.dirs")
	cfg.DB.Files = v.GetStringSlice(aprefix + "db.files")
	cfg.Workers = v.GetInt(aprefix + "workers")
	cfg.CertsDir = v.GetString(aprefix + "certsdir")
	cfg.DataDir = v.GetString(aprefix + "datadir")
	cfg.CacheDir = v.GetString(aprefix + "cachedir")
}

// Empty returns true if configuration is empty
func (cfg EventProcCfg) Empty() bool {
	if len(cfg.Stack.Files) > 0 {
		return false
	}
	if len(cfg.Stack.Dirs) > 0 {
		return false
	}
	if cfg.Stack.Main != "" {
		return false
	}
	if len(cfg.DB.Files) > 0 {
		return false
	}
	if len(cfg.DB.Dirs) > 0 {
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
	for _, file := range cfg.Stack.Files {
		if !util.FileExists(file) {
			return fmt.Errorf("stack config file '%v' doesn't exists", file)
		}
	}
	for _, dir := range cfg.Stack.Dirs {
		if !util.DirExists(dir) {
			return fmt.Errorf("stack config dir '%v' doesn't exists", dir)
		}
	}
	if cfg.Stack.Main == "" {
		return errors.New("stack main label can't be empty")
	}
	for _, file := range cfg.DB.Files {
		if !util.FileExists(file) {
			return fmt.Errorf("event database file '%v' doesn't exists", file)
		}
	}
	for _, dir := range cfg.DB.Dirs {
		if !util.DirExists(dir) {
			return fmt.Errorf("event database dir '%v' doesn't exists", dir)
		}
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
