// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"github.com/luids-io/api/event"
	notifyapi "github.com/luids-io/api/event/grpc/notify"
	"github.com/luids-io/core/yalogi"
)

// EventNotifyAPI is a factory
func EventNotifyAPI(notifier event.Notifier, logger yalogi.Logger) (*notifyapi.Service, error) {
	gsvc := notifyapi.NewService(notifier)
	return gsvc, nil
}
