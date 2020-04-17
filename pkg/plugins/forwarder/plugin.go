// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

// Package forwarder implements a plugin for forward events.
//
// This package is a work in progress and makes no API stability promises.
package forwarder

import (
	"context"
	"errors"
	"fmt"

	"github.com/luids-io/core/event"
	"github.com/luids-io/event/pkg/eventproc"
	"github.com/luids-io/event/pkg/eventproc/stackbuilder"
)

// BuildClass defines default class name of component builder
const BuildClass = "forwarder"

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
		forwarder, ok := service.(event.Forwarder)
		if !ok {
			return nil, fmt.Errorf("service '%s' is not a forwarder instance", sname)
		}
		//return module function
		return func(e *event.Event) error {
			err := forwarder.ForwardEvent(context.Background(), *e)
			if err == nil {
				builder.Logger().Debugf("forwarded event '%s' to '%s'", e.ID, sname)
			}
			return err
		}, nil
	}
}

func init() {
	stackbuilder.RegisterPlugin(BuildClass, Plugin())
}
