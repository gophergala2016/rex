package rexdemo

import "github.com/gophergala2016/rex/room"

// Room is the room used by clients and servers for the demo.
var Room = &room.Room{
	Name:    "REx Demo",
	Service: "_rexdemo_._tcp",
}
