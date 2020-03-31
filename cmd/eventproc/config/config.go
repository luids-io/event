// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
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
				StackMain: "main",
			},
		},
		goconfig.Section{
			Name:     "server-notify",
			Required: true,
			Short:    true,
			Data: &cconfig.ServerCfg{
				ListenURI: "tcp://127.0.0.1:5851",
			},
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
	return cfg
}
