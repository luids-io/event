// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

// Package archiver implements a plugin for archive events.
//
// This package is a work in progress and makes no API stability promises.
package archiver

import (
	"context"
	"errors"
	"fmt"

	"github.com/luids-io/api/event"
	"github.com/luids-io/event/pkg/eventproc"
	"github.com/luids-io/event/pkg/eventproc/stackbuilder"
)

// BuildClass defines default class name of component builder
const BuildClass = "archiver"

// Plugin returns a plugin that archive events
func Plugin() stackbuilder.PluginBuilder {
	return func(builder *stackbuilder.Builder, def *stackbuilder.ItemDef) (eventproc.ModulePlugin, error) {
		builder.Logger().Debugf("building plugin with args: %v", def.Args)
		if len(def.Args) != 1 {
			return nil, errors.New("required arg")
		}
		//first argument is output filename
		sname := def.Args[0]
		service, ok := builder.Service(sname)
		if !ok {
			return nil, fmt.Errorf("service '%s' doesn't exist", sname)
		}
		archive, ok := service.(event.Archiver)
		if !ok {
			return nil, fmt.Errorf("service '%s' is not an archiver instance", sname)
		}
		//return module function
		return func(e *event.Event) error {
			sid, err := archive.SaveEvent(context.Background(), *e)
			if err == nil {
				builder.Logger().Debugf("saved event: %s", sid)
			}
			return err
		}, nil
	}
}

func init() {
	stackbuilder.RegisterPlugin(BuildClass, Plugin())
}
