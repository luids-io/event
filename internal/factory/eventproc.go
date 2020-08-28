// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"

	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/event/internal/config"
	"github.com/luids-io/event/pkg/eventdb"
	"github.com/luids-io/event/pkg/eventproc"
)

// EventProc creates an event processor
func EventProc(cfg *config.EventProcCfg, b *eventproc.Builder, db eventdb.Database, logger yalogi.Logger) (*eventproc.Processor, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("bad config: %v", err)
	}
	main, ok := b.Stack(cfg.Stack.Main)
	if !ok {
		return nil, fmt.Errorf("can't find main stack '%s'", cfg.Stack.Main)
	}
	names := b.StackNames()
	others := make([]*eventproc.Stack, 0, len(names)-1)
	for _, name := range names {
		if name != cfg.Stack.Main {
			stack, _ := b.Stack(name)
			others = append(others, stack)
		}
	}
	//creates a new processor with stacks
	processor := eventproc.New(main, others, db, eventproc.SetLogger(logger))
	return processor, nil
}
