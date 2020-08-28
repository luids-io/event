// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

// dependency injection functions

import (
	"fmt"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"

	"github.com/luids-io/api/event"
	forwardapi "github.com/luids-io/api/event/grpc/forward"
	notifyapi "github.com/luids-io/api/event/grpc/notify"
	cconfig "github.com/luids-io/common/config"
	cfactory "github.com/luids-io/common/factory"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/core/serverd"
	"github.com/luids-io/core/yalogi"
	iconfig "github.com/luids-io/event/internal/config"
	ifactory "github.com/luids-io/event/internal/factory"
	"github.com/luids-io/event/pkg/eventdb"
	"github.com/luids-io/event/pkg/eventproc"
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
			Name:     fmt.Sprintf("health.[%s]", cfgHealth.ListenURI),
			Start:    func() error { go health.Serve(hlis); return nil },
			Shutdown: func() { health.Close() },
		})
	}
	return nil
}

func createAPIServices(msrv *serverd.Manager, logger yalogi.Logger) (apiservice.Discover, error) {
	cfgServices := cfg.Data("ids.api").(*cconfig.APIServicesCfg)
	registry, err := cfactory.APIAutoloader(cfgServices, logger)
	if err != nil {
		return nil, err
	}
	msrv.Register(serverd.Service{
		Name:     "ids.api",
		Ping:     registry.Ping,
		Shutdown: func() { registry.CloseAll() },
	})
	return registry, nil
}

func createStacks(asvc apiservice.Discover, msrv *serverd.Manager, logger yalogi.Logger) (*eventproc.Builder, error) {
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
		Name:     "eventstacks",
		Start:    builder.Start,
		Shutdown: func() { builder.Shutdown() },
	})
	return builder, nil
}

func createEventDB(logger yalogi.Logger) (eventdb.Database, error) {
	cfgEventproc := cfg.Data("eventproc").(*iconfig.EventProcCfg)
	return ifactory.EventDB(cfgEventproc, logger)
}

func createEventProc(stacks *eventproc.Builder, db eventdb.Database, msrv *serverd.Manager, logger yalogi.Logger) (*eventproc.Processor, error) {
	cfgEventProc := cfg.Data("eventproc").(*iconfig.EventProcCfg)
	proc, err := ifactory.EventProc(cfgEventProc, stacks, db, logger)
	if err != nil {
		return nil, err
	}
	msrv.Register(serverd.Service{
		Name:     "eventproc",
		Shutdown: proc.Close,
	})
	return proc, nil
}

func createNotifyAPI(gsrv *grpc.Server, notifier event.Notifier, msrv *serverd.Manager, logger yalogi.Logger) error {
	cfgAPI := cfg.Data("service.event.notify").(*iconfig.EventNotifyAPICfg)
	if cfgAPI.Enable {
		gsvc, err := ifactory.EventNotifyAPI(cfgAPI, notifier, logger)
		if err != nil {
			return err
		}
		notifyapi.RegisterServer(gsrv, gsvc)
		msrv.Register(serverd.Service{Name: "service.event.notify"})
	}
	return nil
}

func createForwardAPI(gsrv *grpc.Server, forwarder event.Forwarder, msrv *serverd.Manager, logger yalogi.Logger) error {
	cfgAPI := cfg.Data("service.event.forward").(*iconfig.EventForwardAPICfg)
	if cfgAPI.Enable {
		gsvc, err := ifactory.EventForwardAPI(cfgAPI, forwarder, logger)
		if err != nil {
			return err
		}
		forwardapi.RegisterServer(gsrv, gsvc)
		msrv.Register(serverd.Service{Name: "service.event.forward"})
	}
	return nil
}

func createServer(msrv *serverd.Manager, logger yalogi.Logger) (*grpc.Server, error) {
	cfgServer := cfg.Data("server").(*cconfig.ServerCfg)
	glis, gsrv, err := cfactory.Server(cfgServer)
	if err == cfactory.ErrURIServerExists {
		return gsrv, nil
	}
	if err != nil {
		return nil, err
	}
	if cfgServer.Metrics {
		grpc_prometheus.Register(gsrv)
	}
	msrv.Register(serverd.Service{
		Name:     fmt.Sprintf("server.[%s]", cfgServer.ListenURI),
		Start:    func() error { go gsrv.Serve(glis); return nil },
		Shutdown: gsrv.GracefulStop,
		Stop:     gsrv.Stop,
	})
	return gsrv, nil
}
