// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

// Package eventproc includes components to implement a simple security event
// management system.
//
// This package is a work in progress and makes no API stability promises.
package eventproc

import (
	"context"
	"errors"
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

// Option defines Processor options.
type Option func(*options)

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

// GUIDGenerator must returns a new unique Global ID for events.
type GUIDGenerator func() string

// Workers option defines the number of goroutines used to event processing.
func Workers(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.workers = n
		}
	}
}

// SetLogger option sets a logger for the component.
func SetLogger(l yalogi.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

// SetGUIDGen option sets a custom gid event generator.
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

// SetBufferSize option defines the size of the event request buffer.
func SetBufferSize(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.buffSize = n
		}
	}
}

// New creates a new processor with stack as the main stack.
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

// NotifyEvent implements event.Notifier.
func (p *Processor) NotifyEvent(ctx context.Context, e event.Event) (string, error) {
	if p.closed {
		return "", event.ErrUnavailable
	}
	// gets peer info
	peerData, peerAddr := getPeerAddr(ctx)

	// checks event
	def, err := p.validateNotify(e)
	if err != nil {
		p.logger.Warnf("eventproc: [peer=%s] notify event: %v", peerAddr, err)
		return "", event.ErrBadRequest
	}

	now := time.Now()
	// complete event data
	e.ID = p.opts.guidGen()
	e.Received = now
	e.Processors = []event.ProcessInfo{{
		Received:  now,
		Processor: event.GetDefaultSource(),
	}}
	e = def.Complete(e)

	//enqueues event to process
	return e.ID, p.queueEvent(e, peerData)
}

// ForwardEvent implements event.Forwarder.
func (p *Processor) ForwardEvent(ctx context.Context, e event.Event) error {
	if p.closed {
		return event.ErrUnavailable
	}
	// gets peer info
	peerData, peerAddr := getPeerAddr(ctx)

	// checks event
	err := p.validateForward(e)
	if err != nil {
		p.logger.Warnf("eventproc: [peer=%s] forward event: %v", peerAddr, err)
		return event.ErrBadRequest
	}

	// complete data
	procinfo := event.ProcessInfo{Received: time.Now(), Processor: event.GetDefaultSource()}
	e.Processors = append(e.Processors, procinfo)

	// enqueues event to process
	return p.queueEvent(e, peerData)
}

// Close event processor.
func (p *Processor) Close() {
	if p.closed {
		return
	}
	p.logger.Infof("closing event processor")
	p.closed = true
	close(p.events)
	p.wg.Wait()
}

// Request is used to store information of the event processing.
type Request struct {
	Event      event.Event
	Enqueued   time.Time
	Started    time.Time
	Finished   time.Time
	StackTrace []string
	Peer       *peer.Peer
	jumps      []string
}

func (p *Processor) init(nworkers int) {
	p.logger.Infof("starting event processor (%v workers)", nworkers)
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

func getPeerAddr(ctx context.Context) (p *peer.Peer, paddr string) {
	var ok bool
	p, ok = peer.FromContext(ctx)
	if ok {
		paddr = fmt.Sprintf("%v", p.Addr)
	}
	return
}

func (p *Processor) validateNotify(e event.Event) (def eventdb.EventDef, err error) {
	if e.ID != "" {
		err = errors.New("id not empty")
		return
	}
	if len(e.Processors) > 0 {
		err = errors.New("processors not empty")
		return
	}
	var ok bool
	def, ok = p.db.FindByCode(e.Code)
	if !ok {
		err = fmt.Errorf("code '%v' not found", e.Code)
		return
	}
	err = def.ValidateData(e)
	if err != nil {
		err = fmt.Errorf("data not valid: %v", err)
	}
	return
}

func (p *Processor) validateForward(e event.Event) error {
	if e.ID == "" {
		return errors.New("event id is empty")
	}
	if len(e.Processors) == 0 {
		return errors.New("event processors is empty")
	}
	//check loops
	self := event.GetDefaultSource()
	for _, s := range e.Processors {
		if self.Equals(s.Processor) {
			return errors.New("detected forward loop")
		}
	}
	return nil
}

func (p *Processor) queueEvent(e event.Event, pinfo *peer.Peer) error {
	// enqueues event to process
	newreq := &Request{Event: e, Enqueued: time.Now(), Peer: pinfo}
	if p.closed {
		return event.ErrUnavailable
	}
	p.events <- newreq
	return nil
}
