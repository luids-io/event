// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"

	"github.com/luids-io/api/event/eventdb"
	"github.com/luids-io/common/util"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/event/internal/config"
)

// EventDB is a factory for event database
func EventDB(cfg *config.EventProcCfg, logger yalogi.Logger) (eventdb.Database, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}
	//get definitions
	dbfiles, err := util.GetFilesDB("json", cfg.DB.Files, cfg.DB.Dirs)
	if err != nil {
		return nil, fmt.Errorf("loading dbfiles: %v", err)
	}
	defs, err := loadEventDefs(dbfiles)
	if err != nil {
		return nil, fmt.Errorf("loading event definitions: %v", err)
	}
	return eventdb.New(defs), nil
}

func loadEventDefs(dbFiles []string) ([]eventdb.EventDef, error) {
	loadedDB := make([]eventdb.EventDef, 0)
	for _, file := range dbFiles {
		entries, err := eventdb.DefsFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("couln't load database: %v", err)
		}
		loadedDB = append(loadedDB, entries...)
	}
	return loadedDB, nil
}
