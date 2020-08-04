// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"errors"

	"github.com/luids-io/api/event"
	forwardapi "github.com/luids-io/api/event/grpc/forward"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/event/internal/config"
)

// EventForwardAPI is a factory
func EventForwardAPI(cfg *config.EventForwardAPICfg, forwarder event.Forwarder, logger yalogi.Logger) (*forwardapi.Service, error) {
	if !cfg.Enable {
		return nil, errors.New("event forward service disabled")
	}
	if !cfg.Log {
		logger = yalogi.LogNull
	}
	gsvc := forwardapi.NewService(forwarder, forwardapi.SetServiceLogger(logger))
	return gsvc, nil
}
