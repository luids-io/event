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
			Name:     "eventproc.api.notify",
			Required: false,
			Data:     &iconfig.EventNotifyAPICfg{Enable: true},
		},
		goconfig.Section{
			Name:     "eventproc.api.forward",
			Required: false,
			Data:     &iconfig.EventForwardAPICfg{Enable: false},
		},
		goconfig.Section{
			Name:     "server",
			Required: true,
			Short:    true,
			Data: &cconfig.ServerCfg{
				ListenURI: "tcp://127.0.0.1:5851",
			},
		},
		goconfig.Section{
			Name:     "luids.api",
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
		notify := cfg.Data("eventproc.api.notify").(*iconfig.EventNotifyAPICfg).Enable
		forward := cfg.Data("eventproc.api.forward").(*iconfig.EventForwardAPICfg).Enable
		if !notify && !forward {
			return errors.New("'eventproc.api.notify' or 'eventproc.api.forward' sections is required")
		}
		return nil
	})
	return cfg
}
