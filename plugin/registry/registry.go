/*
Sniperkit-Bot
- Status: analyzed
*/

// This file will be generated to include all customer specific middlewares
package registry

import (
	"github.com/sniperkit/snk.fork.vulcand/plugin"
	"github.com/sniperkit/snk.fork.vulcand/plugin/cbreaker"
	"github.com/sniperkit/snk.fork.vulcand/plugin/connlimit"
	"github.com/sniperkit/snk.fork.vulcand/plugin/ratelimit"
	"github.com/sniperkit/snk.fork.vulcand/plugin/rewrite"
	"github.com/sniperkit/snk.fork.vulcand/plugin/trace"
)

func GetRegistry() *plugin.Registry {
	r := plugin.NewRegistry()

	specs := []*plugin.MiddlewareSpec{
		ratelimit.GetSpec(),
		connlimit.GetSpec(),
		rewrite.GetSpec(),
		cbreaker.GetSpec(),
		trace.GetSpec(),
	}

	for _, spec := range specs {
		if err := r.AddSpec(spec); err != nil {
			panic(err)
		}
	}

	return r
}
