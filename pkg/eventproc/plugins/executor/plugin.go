// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

// Package executor implements a plugin for exec commands.
//
// This package is a work in progress and makes no API stability promises.
package executor

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/luids-io/api/event"
	"github.com/luids-io/event/pkg/eventproc"
)

// PluginClass registered.
const PluginClass = "executor"

// Builder returns a plugin builder.
func Builder() eventproc.PluginBuilder {
	return func(b *eventproc.Builder, def *eventproc.ItemDef) (eventproc.ModulePlugin, error) {
		b.Logger().Debugf("building plugin with args: %v", def.Args)
		if len(def.Args) == 0 {
			return nil, errors.New("required arg")
		}
		//first argument is application to exec
		app := def.Args[0]
		args := make([]string, 0, len(def.Args)-1)
		for idx, arg := range def.Args {
			if idx != 0 {
				args = append(args, arg)
			}
		}
		//return module function
		return func(e *event.Event) error {
			fargs := make([]string, 0, len(args))
			for _, arg := range args {
				if strings.HasPrefix(arg, "[") && strings.HasSuffix(arg, "]") {
					field := strings.Trim(arg, "[")
					field = strings.Trim(field, "]")
					arg = getField(field, e)
				}
				fargs = append(fargs, arg)
			}
			b.Logger().Debugf("exec %v %v", app, fargs)
			cmd := exec.Command(app, fargs...)
			err := cmd.Run()
			if err != nil {
				return err
			}
			return nil
		}, nil
	}
}

func getField(field string, e *event.Event) string {
	switch field {
	case "code":
		return strconv.Itoa(int(e.Code))
	case "codename":
		return e.Codename
	case "type":
		return e.Type.String()
	case "level":
		return e.Level.String()
	case "source.hostname":
		return e.Source.Hostname
	case "source.program":
		return e.Source.Program
	}
	if strings.HasPrefix(field, "data.") {
		fields := strings.Split(field, ".")
		if len(fields) == 2 {
			v, ok := e.Get(fields[1])
			if ok {
				return fmt.Sprintf("%v", v)
			}
		}
	}
	return ""
}

func init() {
	eventproc.RegisterPlugin(PluginClass, Builder())
}
