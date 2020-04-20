// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"

	"github.com/luids-io/core/event/eventdb"
	"github.com/luids-io/core/utils/yalogi"
	"github.com/luids-io/event/internal/config"
	"github.com/luids-io/event/pkg/eventproc"
	"github.com/luids-io/event/pkg/eventproc/stackbuilder"
)

// EventProc creates an event processor
func EventProc(cfg *config.EventProcCfg, b *stackbuilder.Builder, db eventdb.Database, logger yalogi.Logger) (*eventproc.Processor, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("bad config: %v", err)
	}
	main, ok := b.GetStack(cfg.Stack.Main)
	if !ok {
		return nil, fmt.Errorf("can't find main stack '%s'", cfg.Stack.Main)
	}
	names := b.StackNames()
	others := make([]*eventproc.Stack, 0, len(names)-1)
	for _, name := range names {
		if name != cfg.Stack.Main {
			stack, _ := b.GetStack(name)
			others = append(others, stack)
		}
	}
	//creates a new processor with stacks
	processor := eventproc.New(main, others, db, eventproc.SetLogger(logger))
	return processor, nil
}
