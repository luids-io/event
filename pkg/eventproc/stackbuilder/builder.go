// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

// Package stackbuilder facilitates the creation of stacks for the event
// management processor.
//
// This package is a work in progress and makes no API stability promises.
package stackbuilder

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/event/pkg/eventproc"
)

// Builder helps to create stacks using definitions structs
type Builder struct {
	opts   options
	logger yalogi.Logger

	regsvc apiservice.Discover
	stacks map[string]*eventproc.Stack

	startup  []func() error
	shutdown []func() error
}

type options struct {
	logger   yalogi.Logger
	certsDir string
	dataDir  string
	cacheDir string
}

var defaultOpts = options{
	logger: yalogi.LogNull,
}

// Option is used for builder configuration
type Option func(*options)

// SetLogger sets a logger for the component
func SetLogger(l yalogi.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

// CertsDir sets certificate dir
func CertsDir(s string) Option {
	return func(o *options) {
		o.certsDir = s
	}
}

// DataDir sets data dir
func DataDir(s string) Option {
	return func(o *options) {
		o.dataDir = s
	}
}

// CacheDir sets source dir
func CacheDir(s string) Option {
	return func(o *options) {
		o.cacheDir = s
	}
}

// New instances a new builder
func New(regsvc apiservice.Discover, opt ...Option) *Builder {
	opts := defaultOpts
	for _, o := range opt {
		o(&opts)
	}
	return &Builder{
		opts:     opts,
		logger:   opts.logger,
		regsvc:   regsvc,
		stacks:   make(map[string]*eventproc.Stack),
		startup:  make([]func() error, 0),
		shutdown: make([]func() error, 0),
	}
}

// StackNames returns the names of the stacks created by the builder
func (b *Builder) StackNames() []string {
	names := make([]string, 0)
	for k := range b.stacks {
		names = append(names, k)
	}
	return names
}

// GetStack returns the stack with the name passed, it will returns false
// if the stack has not been built
func (b *Builder) GetStack(name string) (*eventproc.Stack, bool) {
	stack, ok := b.stacks[name]
	return stack, ok
}

// Build construct a stack with the name passed and the modules defined by the
// array ModuleDef
func (b *Builder) Build(def StackDef) (*eventproc.Stack, error) {
	b.logger.Debugf("building '%s'", def.Name)
	if def.Name == "" {
		return nil, errors.New("stack name is empty")
	}
	stack, ok := b.stacks[def.Name]
	if ok {
		return nil, errors.New("stack name exists")
	}
	//check if disabled
	if def.Disabled {
		return nil, fmt.Errorf("'%s' is disabled", def.Name)
	}
	//create stack
	stack = eventproc.NewStack(def.Name)
	//create modules
	names := make(map[string]bool)
	for _, modDef := range def.Modules {
		if modDef.Name == "" {
			return nil, errors.New("module name empty")
		}
		if modDef.Disabled {
			continue
		}
		_, ok := names[modDef.Name]
		if ok {
			return nil, fmt.Errorf("module name '%s' duplicated", modDef.Name)
		}
		names[modDef.Name] = true

		module, err := b.buildModule(modDef)
		if err != nil {
			return nil, fmt.Errorf("building module '%s': %v", modDef.Name, err)
		}
		stack.Add(module)
	}
	b.stacks[def.Name] = stack
	return stack, nil
}

func (b *Builder) buildModule(def ModuleDef) (*eventproc.Module, error) {
	module := &eventproc.Module{
		Name:      def.Name,
		OnSuccess: def.OnSuccess,
		OnError:   def.OnError,
	}
	//build filters
	for _, defFilter := range def.Filters {
		filterb, ok := filterBuilders[defFilter.Class]
		if !ok {
			return nil, fmt.Errorf("filter builder for '%s' not found", defFilter.Class)
		}
		filter, err := filterb(b, defFilter)
		if err != nil {
			return nil, err
		}
		module.Filters = append(module.Filters, filter)
	}
	//build plugins
	for _, defPlugin := range def.Plugins {
		pluginb, ok := pluginBuilders[defPlugin.Class]
		if !ok {
			return nil, fmt.Errorf("plugin builder for '%s' not found", defPlugin.Class)
		}
		plugin, err := pluginb(b, defPlugin)
		if err != nil {
			return nil, err
		}
		module.Plugins = append(module.Plugins, plugin)
	}

	return module, nil
}

// Logger returns logger inside builder
func (b *Builder) Logger() yalogi.Logger {
	return b.logger
}

// Service returns apiservice with the id passed, returns false if not registered
func (b *Builder) Service(id string) (apiservice.Service, bool) {
	return b.regsvc.GetService(id)
}

// CertPath returns path for certificate
func (b Builder) CertPath(cert string) string {
	if path.IsAbs(cert) {
		return cert
	}
	output := cert
	if b.opts.certsDir != "" {
		output = b.opts.certsDir + string(os.PathSeparator) + output
	}
	return output
}

// DataPath returns path for data
func (b Builder) DataPath(data string) string {
	if path.IsAbs(data) {
		return data
	}
	output := data
	if b.opts.dataDir != "" {
		output = b.opts.dataDir + string(os.PathSeparator) + output
	}
	return output
}

// CachePath returns path for cache
func (b Builder) CachePath(data string) string {
	if path.IsAbs(data) {
		return data
	}
	output := data
	if b.opts.cacheDir != "" {
		output = b.opts.cacheDir + string(os.PathSeparator) + output
	}
	return output
}

// OnStartup registers the functions that will be executed during startup.
func (b *Builder) OnStartup(f func() error) {
	b.startup = append(b.startup, f)
}

// OnShutdown registers the functions that will be executed during shutdown.
func (b *Builder) OnShutdown(f func() error) {
	b.shutdown = append(b.shutdown, f)
}

// Start executes all registered functions.
func (b *Builder) Start() error {
	b.logger.Infof("starting stack-builder services")
	var ret error
	for _, f := range b.startup {
		err := f()
		if err != nil {
			return err
		}
	}
	return ret
}

// Shutdown executes all registered functions.
func (b *Builder) Shutdown() error {
	b.logger.Infof("shutting down stack-builder services")
	var ret error
	for _, f := range b.shutdown {
		err := f()
		if err != nil {
			ret = err
		}
	}
	return ret
}
