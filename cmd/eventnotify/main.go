// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/spf13/pflag"

	"github.com/luids-io/core/event"
	"github.com/luids-io/event/cmd/eventnotify/config"
)

//Variables for version output
var (
	Program  = "eventnotify"
	Build    = "unknown"
	Version  = "unknown"
	Revision = "unknown"
)

var (
	cfg = config.Default(Program)
	//behaviour
	configFile = ""
	version    = false
	debug      = false
	help       = false
	//input
	inStdin = false
	inFile  = ""
)

func init() {
	//config mapped params
	cfg.PFlags()
	//behaviour params
	pflag.StringVar(&configFile, "config", configFile, "Use explicit config file.")
	pflag.BoolVar(&version, "version", version, "Show version.")
	pflag.BoolVarP(&help, "help", "h", help, "Show this help.")
	pflag.BoolVar(&debug, "debug", debug, "Enable debug.")
	//input params
	pflag.BoolVar(&inStdin, "stdin", inStdin, "From stdin.")
	pflag.StringVarP(&inFile, "file", "f", inFile, "File for input.")
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
	// check args
	if len(pflag.Args()) == 0 && !inStdin && inFile == "" {
		fmt.Fprintln(os.Stderr, "required event data")
		os.Exit(1)
	}
	// load configuration
	err := cfg.LoadIfFile(configFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// creates logger
	logger, err := createLogger(debug)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// create grpc client
	client, err := createClient(logger)
	if err != nil {
		logger.Fatalf("couldn't create client: %v", err)
	}
	defer client.Close()

	//read events from stdin or file
	reader := os.Stdin
	if inFile != "" {
		file, err := os.Open(inFile)
		if err != nil {
			logger.Fatalf("opening file: %v", err)
		}
		defer file.Close()
		reader = file
	}
	var events []event.Event
	byteValue, err := ioutil.ReadAll(reader)
	if err != nil {
		logger.Fatalf("reading event data: %v", err)
	}
	err = json.Unmarshal(byteValue, &events)
	if err != nil {
		logger.Fatalf("unmarshalling events: %v", err)
	}
	//get default source
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	defaultSource := event.Source{
		Hostname: hostname,
		Program:  filepath.Base(os.Args[0]),
		Instance: strconv.Itoa(os.Getpid()),
	}
	// notify events
	for _, e := range events {
		if e.ID == "" {
			nid, err := uuid.NewV4()
			if err != nil {
				logger.Fatalf("generating uuid: %v", err)
			}
			e.ID = nid.String()
		}
		if e.Source.Hostname == "" || e.Source.Program == "" {
			e.Source = defaultSource
		}
		if e.Timestamp.IsZero() {
			e.Timestamp = time.Now()
		}
		reqid, err := client.Notify(context.Background(), e)
		if err != nil {
			logger.Fatalf("notify event: %v", err)
		}
		fmt.Println("event notified with request id:", reqid)
	}
}