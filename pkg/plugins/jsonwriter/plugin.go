// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

// Package jsonwriter implements a plugin for event processing.
// It can be used to write events into a json file.
//
// This package is a work in progress and makes no API stability promises.
package jsonwriter

import (
	"errors"

	"github.com/luids-io/core/event"
	"github.com/luids-io/event/pkg/eventproc"
	"github.com/luids-io/event/pkg/eventproc/stackbuilder"
)

// BuildClass defines default class name of component builder
const BuildClass = "jsonwriter"

const dataBuffSize = 100

// Plugin returns a plugin that writes events into a file using json format
func Plugin() stackbuilder.PluginBuilder {
	return func(builder *stackbuilder.Builder, def *stackbuilder.ItemDef) (eventproc.ModulePlugin, error) {
		builder.Logger().Debugf("building plugin with args: %v", def.Args)
		if len(def.Args) != 1 {
			return nil, errors.New("required arg")
		}
		//first argument is output filename
		fpath := builder.DataPath(def.Args[0])
		file := getJSONFile(fpath)

		builder.OnStartup(func() error {
			return file.open(dataBuffSize)
		})
		builder.OnShutdown(func() error {
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
	stackbuilder.RegisterPlugin(BuildClass, Plugin())
}
