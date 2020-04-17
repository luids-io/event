// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

// dependency injection functions

import (
	"fmt"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"

	forwardapi "github.com/luids-io/api/event/forward"
	notifyapi "github.com/luids-io/api/event/notify"
	cconfig "github.com/luids-io/common/config"
	cfactory "github.com/luids-io/common/factory"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/core/event"
	_ "github.com/luids-io/core/event/codes"
	"github.com/luids-io/core/utils/serverd"
	"github.com/luids-io/core/utils/yalogi"
	iconfig "github.com/luids-io/event/internal/config"
	ifactory "github.com/luids-io/event/internal/factory"
	"github.com/luids-io/event/pkg/eventproc"
	"github.com/luids-io/event/pkg/eventproc/stackbuilder"
)

func createLogger(debug bool) (yalogi.Logger, error) {
	cfgLog := cfg.Data("log").(*cconfig.LoggerCfg)
	return cfactory.Logger(cfgLog, debug)
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

func createAPIServices(msrv *serverd.Manager, logger yalogi.Logger) (apiservice.Discover, error) {
	cfgServices := cfg.Data("apiservices").(*cconfig.APIServicesCfg)
	registry, err := cfactory.APIAutoloader(cfgServices, logger)
	if err != nil {
		return nil, err
	}
	msrv.Register(serverd.Service{
		Name:     "apiservices.service",
		Ping:     registry.Ping,
		Shutdown: func() { registry.CloseAll() },
	})
	return registry, nil
}

func createStacks(asvc apiservice.Discover, msrv *serverd.Manager, logger yalogi.Logger) (*stackbuilder.Builder, error) {
	cfgStacks := cfg.Data("eventproc").(*iconfig.EventProcCfg)
	builder, err := ifactory.StackBuilder(cfgStacks, asvc, logger)
	if err != nil {
		return nil, err
	}
	//create stacks
	err = ifactory.Stacks(cfgStacks, builder, logger)
	if err != nil {
		return nil, err
	}
	msrv.Register(serverd.Service{
		Name:     "stacks.service",
		Start:    builder.Start,
		Shutdown: func() { builder.Shutdown() },
	})
	return builder, nil
}

func createEventProc(stacks *stackbuilder.Builder, msrv *serverd.Manager, logger yalogi.Logger) (*eventproc.Processor, error) {
	cfgEventProc := cfg.Data("eventproc").(*iconfig.EventProcCfg)
	proc, err := ifactory.EventProc(cfgEventProc, stacks, logger)
	if err != nil {
		return nil, err
	}
	msrv.Register(serverd.Service{
		Name:     "eventproc.service",
		Shutdown: proc.Close,
	})
	return proc, nil
}

func createNotifyAPI(gsrv *grpc.Server, notifier event.Notifier, msrv *serverd.Manager, logger yalogi.Logger) error {
	gsvc, err := ifactory.EventNotifyAPI(notifier, logger)
	if err != nil {
		return err
	}
	notifyapi.RegisterServer(gsrv, gsvc)
	return nil
}

func createForwardAPI(gsrv *grpc.Server, forwarder event.Forwarder, msrv *serverd.Manager, logger yalogi.Logger) error {
	gsvc, err := ifactory.EventForwardAPI(forwarder, logger)
	if err != nil {
		return err
	}
	forwardapi.RegisterServer(gsrv, gsvc)
	return nil
}

func createNotifySrv(msrv *serverd.Manager, logger yalogi.Logger) (*grpc.Server, bool, error) {
	cfgServer := cfg.Data("server-notify").(*cconfig.ServerCfg)
	if cfgServer.Empty() {
		return nil, false, nil
	}
	glis, gsrv, err := cfactory.Server(cfgServer)
	if err == cfactory.ErrURIServerExists {
		return gsrv, true, nil
	}
	if err != nil {
		return nil, false, err
	}
	if cfgServer.Metrics {
		grpc_prometheus.Register(gsrv)
	}
	msrv.Register(serverd.Service{
		Name:     fmt.Sprintf("[%s].server", cfgServer.ListenURI),
		Start:    func() error { go gsrv.Serve(glis); return nil },
		Shutdown: gsrv.GracefulStop,
		Stop:     gsrv.Stop,
	})
	return gsrv, true, nil
}

func createForwardSrv(msrv *serverd.Manager, logger yalogi.Logger) (*grpc.Server, bool, error) {
	cfgServer := cfg.Data("server-forward").(*cconfig.ServerCfg)
	if cfgServer.Empty() {
		return nil, false, nil
	}
	glis, gsrv, err := cfactory.Server(cfgServer)
	if err == cfactory.ErrURIServerExists {
		return gsrv, true, nil
	}
	if err != nil {
		return nil, false, err
	}
	if cfgServer.Metrics {
		grpc_prometheus.Register(gsrv)
	}
	msrv.Register(serverd.Service{
		Name:     fmt.Sprintf("[%s].server", cfgServer.ListenURI),
		Start:    func() error { go gsrv.Serve(glis); return nil },
		Shutdown: gsrv.GracefulStop,
		Stop:     gsrv.Stop,
	})
	return gsrv, true, nil
}
