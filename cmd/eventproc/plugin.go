// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

// imported plugins

import (
	// api services
	_ "github.com/luids-io/api/event/grpc/archive"
	_ "github.com/luids-io/api/event/grpc/forward"

	// event plugins
	_ "github.com/luids-io/event/pkg/filters/basicexpr"
	_ "github.com/luids-io/event/pkg/plugins/archiver"
	_ "github.com/luids-io/event/pkg/plugins/executor"
	_ "github.com/luids-io/event/pkg/plugins/forwarder"
	_ "github.com/luids-io/event/pkg/plugins/jsonwriter"
)
