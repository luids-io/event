// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"errors"

	cconfig "github.com/luids-io/common/config"
	"github.com/luids-io/core/utils/goconfig"
	iconfig "github.com/luids-io/event/internal/config"
)

// Default returns the default configuration
func Default(program string) *goconfig.Config {
	cfg, err := goconfig.New(program,
		goconfig.Section{
			Name:     "eventproc",
			Required: true,
			Data: &iconfig.EventProcCfg{
				Stack: iconfig.StackCfg{Main: "main"},
			},
		},
		goconfig.Section{
			Name:     "server-notify",
			Required: false,
			Short:    false,
			Data:     &cconfig.ServerCfg{},
		},
		goconfig.Section{
			Name:     "server-forward",
			Required: false,
			Short:    false,
			Data:     &cconfig.ServerCfg{},
		},
		goconfig.Section{
			Name:     "apiservices",
			Required: false,
			Data:     &cconfig.APIServicesCfg{},
		},
		goconfig.Section{
			Name:     "log",
			Required: true,
			Data: &cconfig.LoggerCfg{
				Level: "info",
			},
		},
		goconfig.Section{
			Name:     "health",
			Required: false,
			Data:     &cconfig.HealthCfg{},
		},
	)
	if err != nil {
		panic(err)
	}
	// add aditional validators
	cfg.AddValidator(func(cfg *goconfig.Config) error {
		noNotifyServer := cfg.Data("server-notify").Empty()
		noForwardServer := cfg.Data("server-forward").Empty()
		if noNotifyServer && noForwardServer {
			return errors.New("'server-notify' or 'server-forward' sections required")
		}
		return nil
	})
	return cfg
}
