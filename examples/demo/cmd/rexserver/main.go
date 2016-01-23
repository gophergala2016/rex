package main

import (
	"log"
	"os"

	"github.com/coreos/etcd/Godeps/_workspace/src/github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = "Start a REx Demo server"
	app.Action = ServerMain
	app.Run(os.Args)
}

// ServerMain performs the main routine for the demo server.
func ServerMain(*cli.Context) {
	log.Printf("[INFO] demo server initializing")
	log.Printf("[INFO] TODO")
}
