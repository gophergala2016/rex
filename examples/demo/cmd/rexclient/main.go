package main

import (
	"encoding/json"
	"log"

	"github.com/gophergala2016/rex/examples/demo/rexdemo"
	"github.com/gophergala2016/rex/room"
)

func main() {
	done := make(chan struct{})
	servers := make(chan *room.ServerDisco)
	go func() {
		defer close(done)
		for s := range servers {
			b, _ := json.Marshal(s)
			log.Printf("[INFO] found server %s at %s: %s", s.Entry.Name, s.TCPAddr, b)
		}
	}()

	err := room.LookupRoom(rexdemo.Room, servers)
	if err != nil {
		panic(err)
	}

	log.Printf("[INFO] waitisg")
	<-done
}
