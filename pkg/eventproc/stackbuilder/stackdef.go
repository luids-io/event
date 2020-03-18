// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package stackbuilder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/luids-io/event/pkg/eventproc"
)

// StackDef defines stack configuration
type StackDef struct {
	Name     string      `json:"name"`
	Disabled bool        `json:"disabled"`
	Modules  []ModuleDef `json:"modules,omitempty"`
}

// ModuleDef defines configuration of the modules in the stack
type ModuleDef struct {
	Name      string                `json:"name"`
	Filters   []*ItemDef            `json:"filters,omitempty"`
	Plugins   []*ItemDef            `json:"plugins,omitempty"`
	OnSuccess eventproc.StackAction `json:"onsuccess"`
	OnError   eventproc.StackAction `json:"onerror"`
	Disabled  bool                  `json:"disabled"`
}

// ItemDef defines a generic configuration item for filters and plugins
type ItemDef struct {
	Class string                 `json:"class"`
	Args  []string               `json:"args,omitempty"`
	Opts  map[string]interface{} `json:"opts,omitempty"`
}

// DefsFromFile returns all stack definitions in a file in json format
func DefsFromFile(path string) ([]StackDef, error) {
	var stacks []StackDef
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, fmt.Errorf("opening file '%s': %v", path, err)
	}
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading file '%s': %v", path, err)
	}
	err = json.Unmarshal(byteValue, &stacks)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling stacks from json file '%s': %v", path, err)
	}
	return stacks, nil
}
