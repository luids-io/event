// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/luisguillenc/serverd"
	"github.com/luisguillenc/yalogi"

	cconfig "github.com/luids-io/common/config"
	cfactory "github.com/luids-io/common/factory"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/core/event"
	_ "github.com/luids-io/core/event/codes"
	"github.com/luids-io/core/event/services/notify"
	iconfig "github.com/luids-io/event/internal/config"
	ifactory "github.com/luids-io/event/internal/factory"
	"github.com/luids-io/event/pkg/eventproc"
	"github.com/luids-io/event/pkg/eventproc/stackbuilder"

	// api services
	_ "github.com/luids-io/core/event/services/archive"

	// event plugins
	_ "github.com/luids-io/event/pkg/filters/basicexpr"
	_ "github.com/luids-io/event/pkg/plugins/executor"
	_ "github.com/luids-io/event/pkg/plugins/jsonwriter"
	_ "github.com/luids-io/event/pkg/plugins/archiver"
)

func createLogger(debug bool) (yalogi.Logger, error) {
	cfgLog := cfg.Data("log").(*cconfig.LoggerCfg)
	return cfactory.Logger(cfgLog, debug)
}

func createAPIServices(srv *serverd.Manager, logger yalogi.Logger) (*apiservice.Registry, error) {
	cfgServices := cfg.Data("apiservices").(*cconfig.APIServicesCfg)
	registry, err := cfactory.APIServices(cfgServices, logger)
	if err != nil {
		return nil, err
	}
	srv.Register(serverd.Service{
		Name:     "apiservices.service",
		Ping:     registry.Ping,
		Shutdown: func() { registry.CloseAll() },
	})
	return registry, nil
}

// create stack builder
func createStackBuilder(srv *serverd.Manager, regsvc *apiservice.Registry, logger yalogi.Logger) (*stackbuilder.Builder, error) {
	cfgStackBuilder := cfg.Data("stackbuild").(*iconfig.StackBuilderCfg)
	builder, err := ifactory.StackBuilder(cfgStackBuilder, regsvc, logger)
	if err != nil {
		return nil, err
	}
	srv.Register(serverd.Service{
		Name:     "stack-builder.service",
		Start:    builder.Start,
		Shutdown: func() { builder.Shutdown() },
	})
	return builder, nil
}

// create event processor
func createEventProc(srv *serverd.Manager, builder *stackbuilder.Builder, logger yalogi.Logger) (*eventproc.Processor, error) {
	cfgEventProc := cfg.Data("eventproc").(*iconfig.EventProcCfg)
	proc, err := ifactory.EventProc(cfgEventProc, builder, logger)
	if err != nil {
		return nil, err
	}
	srv.Register(serverd.Service{
		Name:     "eventproc.service",
		Shutdown: proc.Close,
	})
	return proc, nil
}

// create notify server
func createNotifySrv(srv *serverd.Manager, notifier event.Notifier, logger yalogi.Logger) error {
	//create server
	cfgServer := cfg.Data("grpc-notify").(*cconfig.ServerCfg)
	glis, gsrv, err := cfactory.Server(cfgServer)
	if err != nil {
		return err
	}
	// create service
	service := notify.NewService(notifier)
	notify.RegisterServer(gsrv, service)
	if cfgServer.Metrics {
		grpc_prometheus.Register(gsrv)
	}
	srv.Register(serverd.Service{
		Name:     "grpc-notify.server",
		Start:    func() error { go gsrv.Serve(glis); return nil },
		Shutdown: gsrv.GracefulStop,
		Stop:     gsrv.Stop,
	})
	return nil
}

func createHealthSrv(srv *serverd.Manager, logger yalogi.Logger) error {
	cfgHealth := cfg.Data("health").(*cconfig.HealthCfg)
	if !cfgHealth.Empty() {
		hlis, health, err := cfactory.Health(cfgHealth, srv, logger)
		if err != nil {
			logger.Fatalf("creating health server: %v", err)
		}
		srv.Register(serverd.Service{
			Name:     "health.server",
			Start:    func() error { go health.Serve(hlis); return nil },
			Shutdown: func() { health.Close() },
		})
	}
	return nil
}
