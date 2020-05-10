// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

// Package basicexpr implements a filter for event processing.
// It can be used for creating basic filters using a simple sintax.
//
// This package is a work in progress and makes no API stability promises.
package basicexpr

import (
	"errors"
	"strconv"
	"strings"

	"github.com/luids-io/api/event"
	"github.com/luids-io/event/pkg/eventproc"
	"github.com/luids-io/event/pkg/eventproc/stackbuilder"
)

// BuildClass defines default class name of component builder
const BuildClass = "basicexpr"

// Filter returns a module builder for basic expressions
func Filter() stackbuilder.FilterBuilder {
	return func(builder *stackbuilder.Builder, def *stackbuilder.ItemDef) (eventproc.ModuleFilter, error) {
		builder.Logger().Debugf("building filter with args: %v", def.Args)
		if len(def.Args) != 3 {
			return nil, errors.New("args must be 3")
		}
		field := def.Args[0]
		op := def.Args[1]
		value := def.Args[2]

		switch field {
		case "code":
			vint, err := strconv.Atoi(value)
			if err != nil {
				return nil, errors.New("invalid value")
			}
			return getCode(op, event.Code(vint))

		case "type":
			var vtype event.Type
			switch value {
			case "security":
				vtype = event.Security
			default:
				return nil, errors.New("invalid value")
			}
			return getType(op, vtype)

		case "level":
			var vlevel event.Level
			switch value {
			case "info":
				vlevel = event.Info
			case "low":
				vlevel = event.Low
			case "medium":
				vlevel = event.Medium
			case "high":
				vlevel = event.High
			case "critical":
				vlevel = event.Critical
			default:
				return nil, errors.New("invalid value")
			}
			return getLevel(op, vlevel)

		case "source.hostname":
			return getSourceHostname(op, value)

		case "source.program":
			return getSourceProgram(op, value)

		}
		if strings.HasPrefix("data.", field) {
			fields := strings.Split(field, ".")
			if len(fields) == 2 {
				return geData(fields[1], op, value)
			}
		}
		return nil, errors.New("invalid field")
	}
}

func getCode(op string, value event.Code) (eventproc.ModuleFilter, error) {
	switch op {
	case "==":
		return func(e event.Event) bool {
			if e.Code == event.Code(value) {
				return true
			}
			return false
		}, nil
	case "!=":
		return func(e event.Event) bool {
			if e.Code != event.Code(value) {
				return true
			}
			return false
		}, nil
	case "<":
		return func(e event.Event) bool {
			if e.Code < event.Code(value) {
				return true
			}
			return false
		}, nil
	case "<=":
		return func(e event.Event) bool {
			if e.Code <= event.Code(value) {
				return true
			}
			return false
		}, nil
	case ">":
		return func(e event.Event) bool {
			if e.Code > event.Code(value) {
				return true
			}
			return false
		}, nil
	case ">=":
		return func(e event.Event) bool {
			if e.Code >= event.Code(value) {
				return true
			}
			return false
		}, nil
	default:
		return nil, errors.New("invalid operator")
	}
}

func getType(op string, value event.Type) (eventproc.ModuleFilter, error) {
	switch op {
	case "==":
		return func(e event.Event) bool {
			if e.Type == value {
				return true
			}
			return false
		}, nil
	case "!=":
		return func(e event.Event) bool {
			if e.Type != value {
				return true
			}
			return false
		}, nil
	default:
		return nil, errors.New("invalid operator")
	}
}

func getLevel(op string, value event.Level) (eventproc.ModuleFilter, error) {
	switch op {
	case "==":
		return func(e event.Event) bool {
			if e.Level == value {
				return true
			}
			return false
		}, nil
	case "!=":
		return func(e event.Event) bool {
			if e.Level != value {
				return true
			}
			return false
		}, nil
	case "<":
		return func(e event.Event) bool {
			if e.Level < value {
				return true
			}
			return false
		}, nil
	case "<=":
		return func(e event.Event) bool {
			if e.Level <= value {
				return true
			}
			return false
		}, nil
	case ">":
		return func(e event.Event) bool {
			if e.Level > value {
				return true
			}
			return false
		}, nil
	case ">=":
		return func(e event.Event) bool {
			if e.Level >= value {
				return true
			}
			return false
		}, nil
	default:
		return nil, errors.New("invalid operator")
	}
}

func getSourceHostname(op string, value string) (eventproc.ModuleFilter, error) {
	switch op {
	case "==":
		return func(e event.Event) bool {
			if e.Source.Hostname == value {
				return true
			}
			return false
		}, nil
	case "!=":
		return func(e event.Event) bool {
			if e.Source.Hostname != value {
				return true
			}
			return false
		}, nil
	default:
		return nil, errors.New("invalid operator")
	}
}

func getSourceProgram(op string, value string) (eventproc.ModuleFilter, error) {
	switch op {
	case "==":
		return func(e event.Event) bool {
			if e.Source.Program == value {
				return true
			}
			return false
		}, nil
	case "!=":
		return func(e event.Event) bool {
			if e.Source.Program != value {
				return true
			}
			return false
		}, nil
	default:
		return nil, errors.New("invalid operator")
	}
}

func geData(field, op, value string) (eventproc.ModuleFilter, error) {
	switch op {
	case "isset":
		return func(e event.Event) bool {
			_, ok := e.Get(field)
			return ok
		}, nil
	case "==":
		return func(e event.Event) bool {
			datav, ok := e.Get(field)
			if !ok {
				return false
			}
			datas, ok := datav.(string)
			if !ok {
				return false
			}
			if datas == value {
				return true
			}
			return false
		}, nil
	case "!=":
		return func(e event.Event) bool {
			datav, ok := e.Get(field)
			if !ok {
				return true
			}
			datas, ok := datav.(string)
			if !ok {
				return true
			}
			if datas != value {
				return true
			}
			return false
		}, nil
	case "eq":
		vint, err := strconv.Atoi(value)
		if err != nil {
			return nil, errors.New("invalid value")
		}
		return func(e event.Event) bool {
			datav, ok := e.Get(field)
			if !ok {
				return false
			}
			datai, ok := datav.(int)
			if !ok {
				return false
			}
			if datai == vint {
				return true
			}
			return false
		}, nil
	case "ne":
		vint, err := strconv.Atoi(value)
		if err != nil {
			return nil, errors.New("invalid value")
		}
		return func(e event.Event) bool {
			datav, ok := e.Get(field)
			if !ok {
				return true
			}
			datai, ok := datav.(int)
			if !ok {
				return true
			}
			if datai != vint {
				return true
			}
			return false
		}, nil
	case "lt":
		vint, err := strconv.Atoi(value)
		if err != nil {
			return nil, errors.New("invalid value")
		}
		return func(e event.Event) bool {
			datav, ok := e.Get(field)
			if !ok {
				return false
			}
			datai, ok := datav.(int)
			if !ok {
				return false
			}
			if datai < vint {
				return true
			}
			return false
		}, nil
	case "le":
		vint, err := strconv.Atoi(value)
		if err != nil {
			return nil, errors.New("invalid value")
		}
		return func(e event.Event) bool {
			datav, ok := e.Get(field)
			if !ok {
				return false
			}
			datai, ok := datav.(int)
			if !ok {
				return false
			}
			if datai <= vint {
				return true
			}
			return false
		}, nil
	case "gt":
		vint, err := strconv.Atoi(value)
		if err != nil {
			return nil, errors.New("invalid value")
		}
		return func(e event.Event) bool {
			datav, ok := e.Get(field)
			if !ok {
				return false
			}
			datai, ok := datav.(int)
			if !ok {
				return false
			}
			if datai > vint {
				return true
			}
			return false
		}, nil
	case "ge":
		vint, err := strconv.Atoi(value)
		if err != nil {
			return nil, errors.New("invalid value")
		}
		return func(e event.Event) bool {
			datav, ok := e.Get(field)
			if !ok {
				return false
			}
			datai, ok := datav.(int)
			if !ok {
				return false
			}
			if datai >= vint {
				return true
			}
			return false
		}, nil

	}
	return nil, errors.New("invalid field")
}

func init() {
	stackbuilder.RegisterFilter(BuildClass, Filter())
}
