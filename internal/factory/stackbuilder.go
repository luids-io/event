// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"github.com/luisguillenc/yalogi"

	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/event/internal/config"
	"github.com/luids-io/event/pkg/eventproc/stackbuilder"
)

// StackBuilder is a factory for stackbuilder
func StackBuilder(cfg *config.StackBuilderCfg, regsvc apiservice.Discover, logger yalogi.Logger) (*stackbuilder.Builder, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}
	b := stackbuilder.New(regsvc,
		stackbuilder.CertsDir(cfg.CertsDir),
		stackbuilder.DataDir(cfg.DataDir),
		stackbuilder.CacheDir(cfg.CacheDir),
		stackbuilder.SetLogger(logger))
	return b, nil
}
