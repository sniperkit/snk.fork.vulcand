/*
Sniperkit-Bot
- Status: analyzed
*/

package main

import (
	"os"

	"github.com/sniperkit/snk.fork.vulcand/plugin/registry"
	"github.com/sniperkit/snk.fork.vulcand/vctl/command"
)

var vulcanUrl string

func main() {
	cmd := command.NewCommand(registry.GetRegistry())
	err := cmd.Run(os.Args)
	if err != nil {
		cmd.PrintError(err)
	}
}
