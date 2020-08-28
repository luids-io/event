// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package eventproc

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/core/yalogi"
)

// FilterBuilder defines the signature for the constuctors of the filters.
type FilterBuilder func(*Builder, *ItemDef) (ModuleFilter, error)

// PluginBuilder defines the signature for the constuctors of the plugins.
type PluginBuilder func(*Builder, *ItemDef) (ModulePlugin, error)

// Builder helps to create stacks using definitions structs.
type Builder struct {
	opts   buildOpts
	logger yalogi.Logger

	regsvc apiservice.Discover
	stacks map[string]*Stack

	startup  []func() error
	shutdown []func() error
}

// BuilderOption is used for builder configuration.
type BuilderOption func(*buildOpts)

type buildOpts struct {
	logger   yalogi.Logger
	certsDir string
	dataDir  string
	cacheDir string
}

var defaultBuildOpts = buildOpts{logger: yalogi.LogNull}

// SetBuildLogger sets a logger for the component.
func SetBuildLogger(l yalogi.Logger) BuilderOption {
	return func(o *buildOpts) {
		o.logger = l
	}
}

// CertsDir sets certificate dir.
func CertsDir(s string) BuilderOption {
	return func(o *buildOpts) {
		o.certsDir = s
	}
}

// DataDir sets data dir.
func DataDir(s string) BuilderOption {
	return func(o *buildOpts) {
		o.dataDir = s
	}
}

// CacheDir sets source dir.
func CacheDir(s string) BuilderOption {
	return func(o *buildOpts) {
		o.cacheDir = s
	}
}

// NewBuilder instances a new builder.
func NewBuilder(regsvc apiservice.Discover, opt ...BuilderOption) *Builder {
	opts := defaultBuildOpts
	for _, o := range opt {
		o(&opts)
	}
	return &Builder{
		opts:   opts,
		logger: opts.logger,
		regsvc: regsvc,
		stacks: make(map[string]*Stack),
	}
}

// StackNames returns the names of the stacks created by the builder.
func (b *Builder) StackNames() []string {
	names := make([]string, 0, len(b.stacks))
	for k := range b.stacks {
		names = append(names, k)
	}
	return names
}

// Stack returns the stack with the name passed, it will returns false
// if the stack has not been built.
func (b *Builder) Stack(name string) (*Stack, bool) {
	stack, ok := b.stacks[name]
	return stack, ok
}

// Build construct a stack with the name passed and the modules defined by the
// array ModuleDef
func (b *Builder) Build(def StackDef) (*Stack, error) {
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
	stack = NewStack(def.Name)
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

func (b *Builder) buildModule(def ModuleDef) (*Module, error) {
	module := &Module{
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

// Logger returns logger inside builder.
func (b *Builder) Logger() yalogi.Logger {
	return b.logger
}

// Service returns apiservice with the id passed, returns false if not registered.
func (b *Builder) Service(id string) (apiservice.Service, bool) {
	return b.regsvc.GetService(id)
}

// CertPath returns path for certificate.
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

// DataPath returns path for data.
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

// CachePath returns path for cache.
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

// RegisterFilter register a filter for the class name passed.
func RegisterFilter(class string, f FilterBuilder) {
	filterBuilders[class] = f
}

// RegisterPlugin register a plugin for the class name passed.
func RegisterPlugin(class string, f PluginBuilder) {
	pluginBuilders[class] = f
}

var filterBuilders map[string]FilterBuilder
var pluginBuilders map[string]PluginBuilder

func init() {
	filterBuilders = make(map[string]FilterBuilder)
	pluginBuilders = make(map[string]PluginBuilder)
}
