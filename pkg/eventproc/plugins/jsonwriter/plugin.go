// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

// Package jsonwriter implements a plugin for event archiving.
//
// This package is a work in progress and makes no API stability promises.
package jsonwriter

import (
	"errors"

	"github.com/luids-io/api/event"
	"github.com/luids-io/event/pkg/eventproc"
)

// PluginCass registered.
const PluginCass = "jsonwriter"

// Builder returns a plugin builder.
func Builder() eventproc.PluginBuilder {
	return func(b *eventproc.Builder, def *eventproc.ItemDef) (eventproc.ModulePlugin, error) {
		b.Logger().Debugf("building plugin with args: %v", def.Args)
		if len(def.Args) != 1 {
			return nil, errors.New("required arg")
		}
		//first argument is output filename
		fpath := b.DataPath(def.Args[0])
		file := getJSONFile(fpath)

		b.OnStartup(func() error {
			return file.open(dataBuffSize)
		})
		b.OnShutdown(func() error {
			file.close()
			return nil
		})
		//return module function
		return func(e *event.Event) error {
			file.write(e)
			return nil
		}, nil
	}
}

func init() {
	eventproc.RegisterPlugin(PluginCass, Builder())
}
