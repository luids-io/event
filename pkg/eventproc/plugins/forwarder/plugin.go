// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

// Package forwarder implements a plugin for event forwarding.
//
// This package is a work in progress and makes no API stability promises.
package forwarder

import (
	"context"
	"errors"
	"fmt"

	"github.com/luids-io/api/event"
	"github.com/luids-io/event/pkg/eventproc"
)

// PluginClass registered.
const PluginClass = "forwarder"

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
		forwarder, ok := service.(event.Forwarder)
		if !ok {
			return nil, fmt.Errorf("service '%s' is not a forwarder instance", sname)
		}
		//return module function
		return func(e *event.Event) error {
			err := forwarder.ForwardEvent(context.Background(), *e)
			if err == nil {
				b.Logger().Debugf("forwarded event '%s' to '%s'", e.ID, sname)
			}
			return err
		}, nil
	}
}

func init() {
	eventproc.RegisterPlugin(PluginClass, Builder())
}
