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
	"google.golang.org/grpc/peer"

	"github.com/luids-io/api/event"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/event/pkg/eventdb"
)

// Processor is the main class of the package.
type Processor struct {
	opts   options
	logger yalogi.Logger
	//event channel
	events chan *Request
	db     eventdb.Database
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
	Peer       *peer.Peer
	jumps      []string
}

type options struct {
	logger   yalogi.Logger
	workers  int
	guidGen  GUIDGenerator
	buffSize int
	hooks    *Hooks
}

var defaultOptions = options{
	workers:  runtime.NumCPU() * 4,
	logger:   yalogi.LogNull,
	guidGen:  defaultGUIDGen,
	buffSize: 100,
	hooks:    NewHooks(),
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
func New(main *Stack, others []*Stack, db eventdb.Database, opt ...Option) *Processor {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	p := &Processor{
		opts:    opts,
		logger:  opts.logger,
		events:  make(chan *Request, opts.buffSize),
		db:      db,
		main:    main,
		stacks:  make(map[string]*Stack, len(others)),
		hrunner: &hooksRunner{hooks: opts.hooks},
	}
	//add other stacks
	for _, stack := range others {
		p.stacks[stack.name] = stack
	}
	p.init(opts.workers)
	return p
}

// NotifyEvent implements event.Notifier
func (p *Processor) NotifyEvent(ctx context.Context, e event.Event) (string, error) {
	if p.closed {
		return "", event.ErrUnavailable
	}
	var peerAddr string
	peerData, ok := peer.FromContext(ctx)
	if ok {
		peerAddr = fmt.Sprintf("%v", peerData.Addr)
	}
	def, ok := p.db.FindByCode(e.Code)
	if !ok {
		p.logger.Warnf("event code '%v' not found (%s)", e.ID, peerAddr)
		return "", event.ErrBadRequest
	}
	if err := def.ValidateData(e); err != nil {
		p.logger.Warnf("event data not valid (%s): %v", peerAddr, err)
		return "", event.ErrBadRequest
	}
	// complete data
	e.ID = p.opts.guidGen()
	e.Processors = []event.ProcessInfo{{
		Received:  time.Now(),
		Processor: event.GetDefaultSource(),
	}}
	e = def.Complete(e)

	//enqueues event to process
	newreq := &Request{
		Event:      e,
		Enqueued:   time.Now(),
		Peer:       peerData,
		StackTrace: make([]string, 0),
		jumps:      make([]string, 0),
	}
	p.logger.Debugf("notify(%s)", e.ID)
	if !p.closed {
		p.events <- newreq
	}
	return e.ID, nil
}

// ForwardEvent implements event.Forwarder
func (p *Processor) ForwardEvent(ctx context.Context, e event.Event) error {
	if p.closed {
		return event.ErrUnavailable
	}
	var peerAddr string
	peerData, ok := peer.FromContext(ctx)
	if ok {
		peerAddr = fmt.Sprintf("%v", peerData.Addr)
	}
	if e.ID == "" {
		p.logger.Warnf("event id is empty (%s)", peerAddr)
		return event.ErrBadRequest
	}
	if len(e.Processors) == 0 {
		p.logger.Warnf("event '%s' processors is empty (%s)", e.ID, peerAddr)
		return event.ErrBadRequest
	}
	//check loop
	source := event.GetDefaultSource()
	for _, s := range e.Processors {
		if source.Equals(s.Processor) {
			p.logger.Warnf("event '%s' detected forward loop (%s)", e.ID, peerAddr)
			return event.ErrInternal
		}
	}
	//add current processor
	e.Processors = append(e.Processors,
		event.ProcessInfo{
			Received:  time.Now(),
			Processor: source})
	//enqueues event to process
	newreq := &Request{
		Event:      e,
		Enqueued:   time.Now(),
		Peer:       peerData,
		StackTrace: make([]string, 0),
		jumps:      make([]string, 0),
	}
	p.logger.Debugf("notify(%s)", e.ID)
	if !p.closed {
		p.events <- newreq
	}
	return nil
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

func (p *Processor) init(nworkers int) {
	p.logger.Debugf("starting event processor (%v workers)", nworkers)
	//create and init workers
	p.wg.Add(nworkers)
	for i := 0; i < nworkers; i++ {
		wid := i
		go p.processWorker(wid)
	}
}

func (p *Processor) processWorker(workerid int) {
	defer p.wg.Done()
	p.logger.Debugf("starting worker %v", workerid)
	for e := range p.events {
		//process event
		p.hrunner.beforeProc(e)
		e.Started = time.Now()
		status, _ := p.main.process(p, e)
		e.Finished = time.Now()
		p.hrunner.afterProc(e)
		//check result action
		if status.Action == ActionReturn ||
			status.Action == ActionFinish ||
			status.Action == ActionNext {
			//only calls finish hooks if exits ok
			p.hrunner.finishProc(e)
		}
	}
	p.logger.Debugf("closing worker %v", workerid)
}
