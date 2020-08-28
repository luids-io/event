// Copyright 2020 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package eventdb

import (
	"github.com/luids-io/api/event"
)

// Database defines interface for event databases
type Database interface {
	FindByCode(event.Code) (EventDef, bool)
}

type database struct {
	defs map[event.Code]EventDef
}

// New returns a in memory Database.
func New(defs []EventDef) Database {
	db := &database{defs: make(map[event.Code]EventDef)}
	for _, def := range defs {
		db.defs[def.Code] = def
	}
	return db
}

func (db *database) FindByCode(code event.Code) (EventDef, bool) {
	def, ok := db.defs[code]
	return def, ok
}
