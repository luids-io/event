// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package eventproc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/luids-io/core/event"
)

// Module defines the information that will be stacked for the processing
type Module struct {
	// Name of the module, it must be unique in the stack
	Name string
	// Filters that will be applied before the plugins are executed. If one of
	// them returns false, then will not be executed and the module result
	// will be Next. If all of them returns true, then all plugin will be
	// executed and the OnSuccess action will be returned (if no errors).
	Filters []ModuleFilter
	// Plugins will be executed if all filters returns true (or if Filters is
	// empty). If there is an error in any of the plugins, the OnError action
	// will be returned.
	Plugins []ModulePlugin
	// OnSucess will be returned to the processor if all the filters apply and
	// the plugins execution don't returns errors.
	OnSuccess StackAction
	// OnError will be returned to the processor if there is an error in
	// plugin execution.
	OnError StackAction
}

// ModuleFilter is a signature for functions that filters events
type ModuleFilter func(e event.Event) (result bool)

// ModulePlugin is a signature for functions that process events
type ModulePlugin func(e *event.Event) error

// StackAction defines the actions returned by the modules to define the
// processing flow
type StackAction struct {
	Action Action
	Label  string
}

// Action defines the behaviours
type Action uint8

// Several actions in rulechain
const (
	ActionNext Action = iota
	ActionStop
	ActionFinish
	ActionJump
	ActionReturn
)

func (a StackAction) String() string {
	switch a.Action {
	case ActionNext:
		return "next"
	case ActionStop:
		return "stop"
	case ActionFinish:
		return "finish"
	case ActionJump:
		return fmt.Sprintf("jump %s", a.Label)
	case ActionReturn:
		return "return"
	}
	return fmt.Sprintf("unkown(%d)", a.Action)
}

// MarshalJSON implements interface
func (a StackAction) MarshalJSON() ([]byte, error) {
	s := ""
	switch a.Action {
	case ActionNext:
		s = "next"
	case ActionStop:
		s = "stop"
	case ActionFinish:
		s = "finish"
	case ActionJump:
		s = fmt.Sprintf("jump %s", a.Label)
	case ActionReturn:
		s = "return"
	default:
		return nil, fmt.Errorf("invalid value '%v' for action", s)
	}
	return json.Marshal(s)
}

// UnmarshalJSON implements interface
func (a *StackAction) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "next":
		a.Action = ActionNext
		return nil
	case "stop":
		a.Action = ActionStop
		return nil
	case "finish":
		a.Action = ActionFinish
		return nil
	case "return":
		a.Action = ActionReturn
		return nil
	}
	if strings.HasPrefix(s, "jump ") {
		st := strings.Split(s, " ")
		if len(st) == 2 {
			a.Action = ActionJump
			a.Label = st[1]
			return nil
		}
	}
	return fmt.Errorf("cannot unmarshal action '%s'", s)
}

// Stack is the struct used by the processor and contains the the modules that
// will be executed
type Stack struct {
	name    string
	modules []*Module
}

// NewStack returns a new Stack
func NewStack(name string) *Stack {
	c := &Stack{
		name:    name,
		modules: make([]*Module, 0),
	}
	return c
}

// Name returns the name of the stack
func (c Stack) Name() string {
	return c.name
}

// Add appends a module to the stack
func (c *Stack) Add(m *Module) {
	c.modules = append(c.modules, m)
}

func (c *Stack) process(p *Processor, e *Request) (status StackAction, last int) {
	for idx, r := range c.modules {
		e.StackTrace = append(e.StackTrace, fmt.Sprintf("%s.%s", c.name, r.Name))
		p.hrunner.beforeModule(e)

		last = idx
		status = StackAction{Action: ActionNext}
		apply := true
		//check filters
		if len(r.Filters) > 0 {
			for _, filter := range r.Filters {
				apply = filter(e.Event)
				if !apply {
					break //stop filtering
				}
			}
		}
		if apply {
			status = r.OnSuccess
			//exec plugins
			if len(r.Plugins) > 0 {
				for idx, plugin := range r.Plugins {
					err := plugin(&e.Event)
					if err != nil {
						p.logger.Warnf("plugin execution trace %v idx %v: %v", e.StackTrace, idx, err)
						status = r.OnError
						break //stop exec
					}
				}
			}
		}
		p.hrunner.afterModule(e)

	LOOPJUMP:
		for status.Action == ActionJump {
			if status.Label == c.name {
				p.logger.Errorf("loop autoreference in stack '%s': trace %v", c.name, e.StackTrace)
				status = StackAction{Action: ActionStop}
				break LOOPJUMP
			}
			for _, prev := range e.jumps {
				if status.Label == prev {
					p.logger.Errorf("loop find in stack '%s': trace %v", c.name, e.StackTrace)
					status = StackAction{Action: ActionStop}
					break LOOPJUMP
				}
			}
			jmpstack, ok := p.stacks[status.Label]
			if !ok {
				p.logger.Errorf("can't find stack '%s': trace %v", status.Label, e.StackTrace)
				status = StackAction{Action: ActionStop}
				break LOOPJUMP
			}
			e.jumps = append(e.jumps, c.name)
			status, _ = jmpstack.process(p, e)
			e.jumps = e.jumps[:len(e.jumps)-1]
		}

		if status.Action != ActionNext {
			return
		}
	}
	return
}
