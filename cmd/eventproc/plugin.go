// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

// imported plugins

import (
	// api services
	_ "github.com/luids-io/api/event/grpc/archive"
	_ "github.com/luids-io/api/event/grpc/forward"

	// event plugins
	_ "github.com/luids-io/event/pkg/eventproc/filters/basicexpr"
	_ "github.com/luids-io/event/pkg/eventproc/plugins/archiver"
	_ "github.com/luids-io/event/pkg/eventproc/plugins/executor"
	_ "github.com/luids-io/event/pkg/eventproc/plugins/forwarder"
	_ "github.com/luids-io/event/pkg/eventproc/plugins/jsonwriter"
)
