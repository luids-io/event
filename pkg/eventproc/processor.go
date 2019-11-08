// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

// Package eventproc includes components to implement a simple security event
// management system.
//
// This package is a work in progress and makes no API stability promises.
package eventproc

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/luisguillenc/yalogi"

	"github.com/luids-io/core/event"
)

// Processor is the main class of the package.
type Processor struct {
	opts   options
	logger yalogi.Logger
	//event channel
	events chan *Request
	// stacks
	main   *Stack
	stacks map[string]*Stack
	// hooks
	hrunner *hooksRunner
	// control
	wg     sync.WaitGroup
	closed bool
}

// Request is used to store information of the event processing
type Request struct {
	Event      event.Event
	Enqueued   time.Time
	Started    time.Time
	Finished   time.Time
	StackTrace []string
	jumps      []string
	//Peer  *peer.Peer
}

type options struct {
	logger   yalogi.Logger
	workers  int
	guidGen  GUIDGenerator
	buffSize int
}

var defaultOptions = options{
	workers:  runtime.NumCPU() * 4,
	logger:   yalogi.LogNull,
	guidGen:  defaultGUIDGen,
	buffSize: 100,
}

// Option defines Processor options
type Option func(*options)

// GUIDGenerator must returns a new unique Global ID for events
type GUIDGenerator func() string

// Workers defines the number of goroutines used to event processing
func Workers(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.workers = n
		}
	}
}

// SetLogger sets a logger for the component
func SetLogger(l yalogi.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

// SetGUIDGen sets a custom gid event generator
func SetGUIDGen(g GUIDGenerator) Option {
	return func(o *options) {
		o.guidGen = g
	}
}

var defaultGUIDGen GUIDGenerator = func() string {
	newid, err := uuid.NewV4()
	if err != nil {
		return ""
	}
	return newid.String()
}

// SetBufferSize defines the size of the event requests buffer
func SetBufferSize(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.buffSize = n
		}
	}
}

// New creates a new processor with stack as the main stack
func New(main *Stack, others []*Stack, h *Hooks, opt ...Option) *Processor {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	p := &Processor{
		opts:    opts,
		logger:  opts.logger,
		events:  make(chan *Request, opts.buffSize),
		main:    main,
		stacks:  make(map[string]*Stack, len(others)),
		hrunner: &hooksRunner{hooks: h},
	}
	p.init(others, opts.workers)
	return p
}

// Notify implements event.Notifier
func (p *Processor) Notify(ctx context.Context, e event.Event) (string, error) {
	if p.closed {
		return "", fmt.Errorf("processor not started")
	}
	e.ID = p.opts.guidGen()
	e.Received = time.Now()
	//enqueues event to process
	newreq := &Request{
		Event:      e,
		Enqueued:   time.Now(),
		StackTrace: make([]string, 0),
		jumps:      make([]string, 0),
	}
	p.logger.Debugf("notify(%s)", e.ID)
	if !p.closed {
		p.events <- newreq
	}
	return e.ID, nil
}

// Close event processor
func (p *Processor) Close() {
	if p.closed {
		return
	}
	p.logger.Debugf("closing event processor")
	p.closed = true
	close(p.events)
	p.wg.Wait()
}

func (p *Processor) init(others []*Stack, nworkers int) {
	p.logger.Debugf("starting event processor (%v workers)", nworkers)
	//add other stacks
	for _, stack := range others {
		p.logger.Debugf("adding stack '%s'", stack.name)
		p.stacks[stack.name] = stack
	}
	//create and init workers
	p.wg.Add(nworkers)
	for i := 0; i < nworkers; i++ {
		wid := i
		go func() {
			defer p.wg.Done()
			p.logger.Debugf("starting worker #%v", wid)
			p.processWorker(wid)
			p.logger.Debugf("closing worker #%v", wid)
		}()
	}
}

func (p *Processor) processWorker(workerid int) {
	for e := range p.events {
		p.hrunner.beforeProc(e)
		e.Started = time.Now()
		status, _ := p.main.process(p, e)
		e.Finished = time.Now()
		p.hrunner.afterProc(e)

		if status.Action == ActionReturn ||
			status.Action == ActionFinish ||
			status.Action == ActionNext {
			//only calls finish hooks if exits ok
			p.hrunner.finishProc(e)
		}
	}
}
