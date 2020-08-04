// Copyright 2020 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package eventdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/luids-io/api/event"
)

// EventDef defines events
type EventDef struct {
	Code        event.Code `json:"code"`
	Type        event.Type `json:"type"`
	Codename    string     `json:"codename"`
	Tags        []string   `json:"tags,omitempty"`
	Description string     `json:"description,omitempty"`
	Fields      []FieldDef `json:"fields,omitempty"`
	RaisedBy    []string   `json:"raised_by,omitempty"`
}

// FieldDef stores field definitions
type FieldDef struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// Complete event information from definition
func (def *EventDef) Complete(e event.Event) event.Event {
	ret := e
	ret.Type = def.Type
	ret.Codename = def.Codename
	ret.Description = def.getDesc(e)
	if len(def.Tags) > 0 {
		ret.Tags = make([]string, len(def.Tags), len(def.Tags))
		copy(ret.Tags, def.Tags)
	}
	return ret
}

func (def *EventDef) getDesc(e event.Event) string {
	if reBetweenBrackets.MatchString(def.Description) {
		return reBetweenBrackets.ReplaceAllStringFunc(def.Description, func(s string) string {
			element := strings.Trim(s, "[")
			element = strings.Trim(element, "]")
			if strings.HasPrefix(element, "data.") {
				field := strings.TrimPrefix(element, "data.")
				value, ok := e.Data[field]
				if !ok {
					return ""
				}
				return fmt.Sprintf("%v", value)
			}
			return s
		})
	}
	return def.Description
}

var reBetweenBrackets = regexp.MustCompile(`\[([^\[\]]*)\]`)

// ValidateData if event data is ok
func (def *EventDef) ValidateData(e event.Event) error {
	metas := make(map[string]FieldDef, len(def.Fields))
	for _, f := range def.Fields {
		metas[f.Name] = f
	}
	for field := range e.Data {
		_, ok := metas[field]
		if !ok {
			return fmt.Errorf("data field '%s' undefined", field)
		}
	}
	for field, md := range metas {
		value, ok := e.Data[field]
		if !ok {
			if md.Required {
				return fmt.Errorf("data field '%s' is required", field)
			}
			continue
		}
		switch md.Type {
		case "string":
			if _, ok := value.(string); !ok {
				return fmt.Errorf("data field '%s' is not a valid string", field)
			}
		case "int":
			if _, ok := value.(int); !ok {
				return fmt.Errorf("data field '%s' is not a valid int", field)
			}
		case "float":
			if _, ok := value.(float64); !ok {
				return fmt.Errorf("data field '%s' is not a valid float", field)
			}
		}
	}
	return nil
}

// DefsFromFile returns event definitions from files
func DefsFromFile(path string) ([]EventDef, error) {
	var stacks []EventDef
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
