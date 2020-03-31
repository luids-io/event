// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	notifyapi "github.com/luids-io/api/event/notify"
	"github.com/luids-io/core/event"
	"github.com/luids-io/core/utils/yalogi"
)

// EventNotifyAPI is a factory
func EventNotifyAPI(notifier event.Notifier, logger yalogi.Logger) (*notifyapi.Service, error) {
	gsvc := notifyapi.NewService(notifier)
	return gsvc, nil
}
