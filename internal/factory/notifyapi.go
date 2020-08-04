// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"errors"

	"github.com/luids-io/api/event"
	notifyapi "github.com/luids-io/api/event/grpc/notify"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/event/internal/config"
)

// EventNotifyAPI is a factory
func EventNotifyAPI(cfg *config.EventNotifyAPICfg, notifier event.Notifier, logger yalogi.Logger) (*notifyapi.Service, error) {
	if !cfg.Enable {
		return nil, errors.New("event notify service disabled")
	}
	if !cfg.Log {
		logger = yalogi.LogNull
	}
	gsvc := notifyapi.NewService(notifier, notifyapi.SetServiceLogger(logger))
	return gsvc, nil
}
