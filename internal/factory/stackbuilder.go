// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"

	"github.com/luids-io/common/util"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/event/internal/config"
	"github.com/luids-io/event/pkg/eventproc"
)

// StackBuilder is a factory for stackbuilder
func StackBuilder(cfg *config.EventProcCfg, regsvc apiservice.Discover, logger yalogi.Logger) (*eventproc.Builder, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}
	b := eventproc.NewBuilder(regsvc,
		eventproc.CertsDir(cfg.CertsDir),
		eventproc.DataDir(cfg.DataDir),
		eventproc.CacheDir(cfg.CacheDir),
		eventproc.SetBuildLogger(logger))
	return b, nil
}

// Stacks creates stacks in builder
func Stacks(cfg *config.EventProcCfg, b *eventproc.Builder, logger yalogi.Logger) error {
	err := cfg.Validate()
	if err != nil {
		return fmt.Errorf("bad config: %v", err)
	}
	//get definitions
	dbfiles, err := util.GetFilesDB("json", cfg.Stack.Files, cfg.Stack.Dirs)
	if err != nil {
		return fmt.Errorf("loading dbfiles: %v", err)
	}
	defs, err := loadStackDefs(dbfiles)
	if err != nil {
		return fmt.Errorf("loading stackdefs: %v", err)
	}
	//build stacks
	for _, def := range defs {
		if def.Disabled {
			continue
		}
		_, err := b.Build(def)
		if err != nil {
			return fmt.Errorf("creating stack '%s': %v", def.Name, err)
		}
	}
	return nil
}

func loadStackDefs(dbFiles []string) ([]eventproc.StackDef, error) {
	loadedDB := make([]eventproc.StackDef, 0)
	for _, file := range dbFiles {
		entries, err := eventproc.StackDefsFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("couln't load database: %v", err)
		}
		loadedDB = append(loadedDB, entries...)
	}
	return loadedDB, nil
}
