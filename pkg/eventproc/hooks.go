// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package eventproc

// CbRequest defines the format of the callbacks used by the hooks
type CbRequest func(*Request)

// Hooks stores information about the hooks
type Hooks struct {
	beforeProc   []CbRequest
	afterProc    []CbRequest
	finishProc   []CbRequest
	beforeModule []CbRequest
	afterModule  []CbRequest
}

// NewHooks creates a new Hooks instance
func NewHooks() *Hooks {
	return &Hooks{
		beforeProc:   []CbRequest{},
		afterProc:    []CbRequest{},
		finishProc:   []CbRequest{},
		beforeModule: []CbRequest{},
		afterModule:  []CbRequest{},
	}
}

// BeforeProc adds a callback that will be executed before the process starts
func (h *Hooks) BeforeProc(fn CbRequest) {
	h.beforeProc = append(h.beforeProc, fn)
}

// AfterProc adds a callback that will be executed before the process end
func (h *Hooks) AfterProc(fn CbRequest) {
	h.afterProc = append(h.afterProc, fn)
}

// FinishProc adds a callback that will be executed if the process finished ok
func (h *Hooks) FinishProc(fn CbRequest) {
	h.finishProc = append(h.finishProc, fn)
}

// BeforeModule adds a callback that will be executed before a stack module starts
func (h *Hooks) BeforeModule(fn CbRequest) {
	h.beforeModule = append(h.beforeModule, fn)
}

// AfterModule adds a callback that will be executed before a stack module starts
func (h *Hooks) AfterModule(fn CbRequest) {
	h.afterModule = append(h.afterModule, fn)
}

type hooksRunner struct {
	hooks *Hooks
}

func (h *hooksRunner) beforeProc(e *Request) error {
	for _, cb := range h.hooks.beforeProc {
		cb(e)
	}
	return nil
}

func (h *hooksRunner) afterProc(e *Request) error {
	for _, cb := range h.hooks.afterProc {
		cb(e)
	}
	return nil
}

func (h *hooksRunner) finishProc(e *Request) error {
	for _, cb := range h.hooks.finishProc {
		cb(e)
	}
	return nil
}

func (h *hooksRunner) beforeModule(e *Request) error {
	for _, cb := range h.hooks.beforeModule {
		cb(e)
	}
	return nil
}

func (h *hooksRunner) afterModule(e *Request) error {
	for _, cb := range h.hooks.afterModule {
		cb(e)
	}
	return nil
}
