// Copyright 2020 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package eventdb_test

import (
	"strings"
	"testing"

	"github.com/luids-io/api/event"
	"github.com/luids-io/event/pkg/eventdb"
)

func TestEventDef(t *testing.T) {
	var err error
	var def eventdb.EventDef
	var e event.Event

	// test event complete and basic error
	def = eventdb.EventDef{
		Code:     1234,
		Codename: "prueba1",
		Type:     event.Security,
	}
	e = event.New(1234, event.Low)
	err = def.ValidateData(e)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	e = def.Complete(e)
	if e.Type != event.Security {
		t.Error("type missmatch")
	}
	if e.Codename != "prueba1" {
		t.Error("codename misstmatch")
	}
	if e.Level != event.Low {
		t.Error("level missmatch")
	}

	// test undefined
	e = event.New(1234, event.Info)
	e.Set("message", "la cagaste burlancaster")
	err = def.ValidateData(e)
	if err == nil {
		t.Error("expected error")
	}
	if !strings.Contains(err.Error(), "undefined") {
		t.Errorf("unexpected error: %v", err)
	}

	// test description complete with field
	def = eventdb.EventDef{
		Code:        1234,
		Codename:    "prueba1",
		Type:        event.Security,
		Description: "este es el mensaje: [data.message]",
		Fields: []eventdb.FieldDef{
			{Name: "message", Type: "string", Required: true},
		},
	}
	e = event.New(1234, event.Info)
	e.Set("message", "la cagaste burlancaster")
	err = def.ValidateData(e)
	if err != nil {
		t.Errorf("unexecter error: %v", err)
	}
	e = def.Complete(e)
	if e.Description != "este es el mensaje: la cagaste burlancaster" {
		t.Errorf("unexpected desc: %s", e.Description)
	}

	// test with integer value
	def = eventdb.EventDef{
		Code:        1234,
		Codename:    "prueba1",
		Type:        event.Security,
		Description: "este es el mensaje: [data.message]",
		Fields: []eventdb.FieldDef{
			{Name: "message", Type: "string", Required: true},
			{Name: "score", Type: "int", Required: true},
		},
	}
	e = event.New(1234, event.Info)
	e.Set("message", "la cagaste burlancaster")
	err = def.ValidateData(e)
	if err == nil {
		t.Error("execter error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("unexpected error: %v", err)
	}
	// test bad int
	e.Set("score", "malvalor")
	err = def.ValidateData(e)
	if err == nil {
		t.Error("expected error")
	}
	if !strings.Contains(err.Error(), "valid int") {
		t.Errorf("unexpected error: %v", err)
	}
	// test int value
	e.Set("score", 100)
	err = def.ValidateData(e)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// test with integer value
	def = eventdb.EventDef{
		Code:        1234,
		Codename:    "prueba1",
		Type:        event.Security,
		Description: "[data.score] [data.prob]",
		Fields: []eventdb.FieldDef{
			{Name: "message", Type: "string", Required: true},
			{Name: "score", Type: "int", Required: true},
			{Name: "prob", Type: "float"},
		},
	}
	e = event.New(1234, event.Info)
	e.Set("message", "la cagaste burlancaster")
	e.Set("score", 100)
	err = def.ValidateData(e)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	e = def.Complete(e)
	if e.Description != "100 " {
		t.Errorf("unexpected desc: %s", e.Description)
	}
	e.Set("prob", 0.2)
	err = def.ValidateData(e)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	e = def.Complete(e)
	if e.Description != "100 0.2" {
		t.Errorf("unexpected desc: %s", e.Description)
	}
}
