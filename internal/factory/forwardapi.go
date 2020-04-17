// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	forwardapi "github.com/luids-io/api/event/forward"
	"github.com/luids-io/core/event"
	"github.com/luids-io/core/utils/yalogi"
)

// EventForwardAPI is a factory
func EventForwardAPI(forwarder event.Forwarder, logger yalogi.Logger) (*forwardapi.Service, error) {
	gsvc := forwardapi.NewService(forwarder)
	return gsvc, nil
}
