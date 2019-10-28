// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package stackbuilder

import "github.com/luids-io/event/pkg/eventproc"

var filterBuilders map[string]FilterBuilder
var pluginBuilders map[string]PluginBuilder

// FilterBuilder defines the signature for the constuctors of the filters
type FilterBuilder func(builder *Builder, def *ItemDef) (eventproc.ModuleFilter, error)

// PluginBuilder defines the signature for the constuctors of the plugins
type PluginBuilder func(builder *Builder, def *ItemDef) (eventproc.ModulePlugin, error)

// RegisterFilter register a filter for the class name passed
func RegisterFilter(class string, f FilterBuilder) {
	filterBuilders[class] = f
}

// RegisterPlugin register a plugin for the class name passed
func RegisterPlugin(class string, f PluginBuilder) {
	pluginBuilders[class] = f
}

func init() {
	filterBuilders = make(map[string]FilterBuilder)
	pluginBuilders = make(map[string]PluginBuilder)
}
