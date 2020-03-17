// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"errors"
	"fmt"

	"github.com/luisguillenc/yalogi"

	"github.com/luids-io/common/util"
	"github.com/luids-io/event/internal/config"
	"github.com/luids-io/event/pkg/eventproc"
	"github.com/luids-io/event/pkg/eventproc/stackbuilder"
)

// EventProc creates an event processor
func EventProc(cfg *config.EventProcCfg, builder *stackbuilder.Builder, logger yalogi.Logger) (*eventproc.Processor, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid eventproc config: %v", err)
	}
	//get definitions
	dbfiles, err := util.GetFilesDB("json", cfg.DatabaseFiles, cfg.DatabaseDirs)
	if err != nil {
		return nil, fmt.Errorf("loading dbfiles: %v", err)
	}
	defs, err := loadStackDefs(dbfiles)
	if err != nil {
		return nil, fmt.Errorf("loading stackdefs: %v", err)
	}
	//build stacks
	var main *eventproc.Stack
	others := make([]*eventproc.Stack, 0, len(defs))
	for _, def := range defs {
		if def.Name == "" {
			return nil, errors.New("stack name required")
		}
		stack, err := builder.Build(def.Name, def.Modules)
		if err != nil {
			return nil, fmt.Errorf("creating stack %s: %v", def.Name, err)
		}
		if def.Name == "main" {
			main = stack
		} else {
			others = append(others, stack)
		}
	}
	//creates a new processor with stacks
	processor := eventproc.New(main, others, eventproc.SetLogger(logger))
	return processor, nil
}

func loadStackDefs(dbFiles []string) ([]stackbuilder.StackDef, error) {
	loadedDB := make([]stackbuilder.StackDef, 0)
	for _, file := range dbFiles {
		entries, err := stackbuilder.DefsFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("couln't load database: %v", err)
		}
		loadedDB = append(loadedDB, entries...)
	}
	return loadedDB, nil
}
