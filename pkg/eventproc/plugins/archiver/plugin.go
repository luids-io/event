// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

// Package archiver implements a plugin for event archiving.
//
// This package is a work in progress and makes no API stability promises.
package archiver

import (
	"context"
	"errors"
	"fmt"

	"github.com/luids-io/api/event"
	"github.com/luids-io/event/pkg/eventproc"
)

// PluginClass registered.
const PluginClass = "archiver"

// Builder returns a plugin builder.
func Builder() eventproc.PluginBuilder {
	return func(b *eventproc.Builder, def *eventproc.ItemDef) (eventproc.ModulePlugin, error) {
		b.Logger().Debugf("building plugin with args: %v", def.Args)
		if len(def.Args) != 1 {
			return nil, errors.New("required arg")
		}
		//first argument is output filename
		sname := def.Args[0]
		service, ok := b.Service(sname)
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
				b.Logger().Debugf("saved event: %s", sid)
			}
			return err
		}, nil
	}
}

func init() {
	eventproc.RegisterPlugin(PluginClass, Builder())
}
