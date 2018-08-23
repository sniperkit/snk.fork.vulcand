/*
Sniperkit-Bot
- Status: analyzed
*/

package builder

import (
	"github.com/sniperkit/snk.fork.vulcand/proxy"
	"github.com/sniperkit/snk.fork.vulcand/proxy/mux"
	"github.com/sniperkit/snk.fork.vulcand/stapler"
)

// NewProxy returns a new Proxy instance.
func NewProxy(id int, st stapler.Stapler, o proxy.Options) (proxy.Proxy, error) {
	return mux.New(id, st, o)
}
