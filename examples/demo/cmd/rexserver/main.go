package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/gophergala2016/rex/examples/demo/rexdemo"
	"github.com/gophergala2016/rex/room"
	"golang.org/x/net/context"
)

func main() {
	app := cli.NewApp()
	app.Usage = "Start a REx Demo server"
	app.Action = ServerMain
	app.Run(os.Args)
}

// ServerMain performs the main routine for the demo server.
func ServerMain(*cli.Context) {
	var fatal bool
	defer func() {
		if fatal {
			os.Exit(1)
		}
	}()
	background := context.Background()

	log.Printf("[INFO] demo server initializing")
	demo := NewDemo()
	bus := room.NewBus(background, demo)
	config := &room.ServerConfig{
		Room: rexdemo.Room,
		Bus:  bus,
	}
	server := room.NewServer(config)

	log.Printf("[INFO] starting server")
	err := server.Start()
	if err != nil {
		log.Printf("[FATAL] %v", err)
		fatal = true
		return
	}

	log.Printf("[INFO] server running at %s", server.Addr())

	log.Printf("[INFO] creating mDNS discovery server")
	zc, err := room.NewZoneConfig(server)
	if err != nil {
		log.Printf("[FATAL] failed to initialize discovery")
		fatal = true
		return
	}
	disco, err := room.DiscoveryServer(zc)
	if err != nil {
		log.Printf("[FATAL] discovery server failed to start")
		fatal = true
		return
	}
	defer disco.Close()

	err = server.Wait()
	if err != nil {
		log.Printf("[FATAL] %v", err)
		fatal = true
		return
	}
}

// DemoServer is the server side (source of truth) of the demo object.
type DemoServer rexdemo.Demo

// NewDemo wraps the result of rexdemo.NewDemo() as a DemoServer
func NewDemo() *DemoServer {
	return (*DemoServer)(rexdemo.NewDemo())
}

// HandleMessage adds to the message counter
func (d *DemoServer) HandleMessage(ctx context.Context, msg room.Msg) {
	d.Mut.Lock()
	defer d.Mut.Unlock()
	d.Counter++
	d.Last = time.Now()
	log.Printf("[DEBUG] %v session %v %q", msg.Time(), msg.Session(), msg.Text())
	log.Printf("[INFO] count: %d", d.Counter)

	js, _ := json.Marshal(d)

	go func() {
		content := room.Bytes(js)
		err := room.Broadcast(ctx, content)
		if err != nil {
			log.Printf("[ERR] %v", err)
		}
	}()
}
