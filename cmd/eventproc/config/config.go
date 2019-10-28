// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"github.com/luisguillenc/goconfig"

	cconfig "github.com/luids-io/common/config"
	iconfig "github.com/luids-io/event/internal/config"
)

// Default returns the default configuration
func Default(program string) *goconfig.Config {
	cfg, err := goconfig.New(program,
		goconfig.Section{
			Name:     "eventproc",
			Required: true,
			Short:    true,
			Data:     &iconfig.EventProcCfg{},
		},
		goconfig.Section{
			Name:     "apiservices",
			Required: false,
			Data:     &cconfig.APIServicesCfg{},
		},
		goconfig.Section{
			Name:     "stackbuild",
			Required: false,
			Short:    true,
			Data:     &iconfig.StackBuilderCfg{},
		},
		goconfig.Section{
			Name:     "grpc-notify",
			Required: true,
			Short:    true,
			Data: &cconfig.ServerCfg{
				ListenURI: "tcp://127.0.0.1:5851",
			},
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
