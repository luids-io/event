// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/luids-io/core/utils/serverd"
	"github.com/luids-io/event/cmd/eventproc/config"
)

//Variables for version output
var (
	Program  = "eventproc"
	Build    = "unknown"
	Version  = "unknown"
	Revision = "unknown"
)

var (
	cfg = config.Default(Program)
	//behaviour
	configFile = ""
	version    = false
	help       = false
	debug      = false
	dryRun     = false
)

func init() {
	//config mapped params
	cfg.PFlags()
	//behaviour params
	pflag.StringVar(&configFile, "config", configFile, "Use explicit config file.")
	pflag.BoolVar(&version, "version", version, "Show version.")
	pflag.BoolVarP(&help, "help", "h", help, "Show this help.")
	pflag.BoolVar(&debug, "debug", debug, "Enable debug.")
	pflag.BoolVar(&dryRun, "dry-run", dryRun, "Checks configuration but not start service.")
	pflag.Parse()
}

func main() {
	if version {
		fmt.Printf("version: %s\nrevision: %s\nbuild: %s\n", Version, Revision, Build)
		os.Exit(0)
	}
	if help {
		pflag.Usage()
		os.Exit(0)
	}

	// load configuration
	err := cfg.LoadIfFile(configFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	//creates logger
	logger, err := createLogger(debug)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// echo version and config
	logger.Infof("%s (version: %s build: %s)", Program, Version, Build)
	if debug {
		logger.Debugf("configuration dump:\n%v", cfg.Dump())
	}

	// creates main server manager
	msrv := serverd.New(serverd.SetLogger(logger))

	// create registry api services
	apisvc, err := createAPIServices(msrv, logger)
	if err != nil {
		logger.Fatalf("couldn't create registry: %v", err)
	}

	// create event database
	db, err := createEventDB(logger)
	if err != nil {
		logger.Fatalf("couldn't create event database: %v", err)
	}

	// create stacks
	stacks, err := createStacks(apisvc, msrv, logger)
	if err != nil {
		logger.Fatalf("couldn't create stacks: %v", err)
	}

	// create event processor
	eproc, err := createEventProc(stacks, db, msrv, logger)
	if err != nil {
		logger.Fatalf("couldn't create eventproc: %v", err)
	}

	if dryRun {
		fmt.Println("configuration seems ok")
		os.Exit(0)
	}

	// creates server
	gsrv, err := createServer(msrv, logger)
	if err != nil {
		logger.Fatalf("couldn't create event server: %v", err)
	}
	if notifyAPIEnabled() {
		// creates and register notify api
		err = createNotifyAPI(gsrv, eproc, msrv, logger)
		if err != nil {
			logger.Fatalf("couldn't create notify api: %v", err)
		}
	}
	if forwardAPIEnabled() {
		// creates and register forward api
		err = createForwardAPI(gsrv, eproc, msrv, logger)
		if err != nil {
			logger.Fatalf("couldn't create forward api: %v", err)
		}
	}

	// creates health server
	err = createHealthSrv(msrv, logger)
	if err != nil {
		logger.Fatalf("couldn't create health server: %v", err)
	}

	//run server
	err = msrv.Run()
	if err != nil {
		logger.Errorf("running server: %v", err)
	}
	logger.Infof("%s finished", Program)
}
