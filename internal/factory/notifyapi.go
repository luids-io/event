// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"github.com/luisguillenc/yalogi"

	notifyapi "github.com/luids-io/api/event/notify"
	"github.com/luids-io/core/event"
)

// EventNotifyAPI is a factory
func EventNotifyAPI(notifier event.Notifier, logger yalogi.Logger) (*notifyapi.Service, error) {
	gsvc := notifyapi.NewService(notifier)
	return gsvc, nil
}
