package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gophergala2016/rex/examples/demo/rexdemo"
	"github.com/gophergala2016/rex/room"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
	"golang.org/x/net/context"
)

var (
	images   *glutil.Images
	fps      *debug.FPS
	program  gl.Program
	position gl.Attrib
	offset   gl.Uniform
	color    gl.Uniform
	buf      gl.Buffer

	green  float32
	touchX float32
	touchY float32
)

func main() {
	app.Main(func(a app.App) {
		background := context.Background()

		server := make(chan *room.ServerDisco)
		servers := make(chan *room.ServerDisco)
		go func() {
			// ignore all but the first server found for now
			defer close(server)
			var chosen *room.ServerDisco
			chosen, ok := <-servers
			if !ok {
				return
			}
			server <- chosen
		}()

		log.Printf("[INFO] Waiting for servers")
		go func() {
			err := room.LookupRoom(rexdemo.Room, servers)
			if err != nil {
				log.Printf("[ERR] Failed network lookup: %v", err)
			}
		}()

		nameos, err := os.Hostname()
		if err != nil {
			log.Printf("[ERR] Hostname: %v", err)
			nameos = "UNKNOWN HOST"
		}
		name := fmt.Sprintf("%s [%d]", nameos, os.Getpid())

		demo := NewDemo()
		client := room.NewClient(demo)

		runClient := func(ctx context.Context, client *room.Client, server *room.ServerDisco) {
			ip := server.Entry.AddrV4
			if ip == nil {
				ip = server.Entry.AddrV6
			}
			host := ip.String()
			client.Host = host
			client.Port = server.Entry.Port

			err = client.CreateSession(ctx, name)
			if err != nil {
				log.Printf("[ERR] Failed to create a session: %v", err)
			}
			defer client.Send(ctx, room.String("DISCONNECT"))

			next, err := client.Run(ctx, 0)
			if err != nil {
				log.Printf("[ERR] Event loop at index %d: %v", next, err)
				panic(err)
			}
			log.Printf("[INFO] Terminated at index %d", next)
		}

		var glctx gl.Context
		var sz size.Event
		_server := server
		for e := range a.Events() {
			select {
			case chosen, ok := <-_server:
				if !ok {
					log.Printf("[ERR] No server found")
					return
				}
				_server = nil
				b, _ := json.Marshal(chosen)
				log.Printf("[INFO] Server %s at %s: %s", chosen.Entry.Name, chosen.TCPAddr, b)
				go runClient(background, client, chosen)
			default:
			}
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					onStop(glctx)
					glctx = nil
				}
			case size.Event:
				sz = e
				touchX = float32(sz.WidthPx / 2)
				touchY = float32(sz.HeightPx / 2)
			case paint.Event:
				if glctx == nil || e.External {
					// As we are actively painting as fast as
					// we can (usually 60 FPS), skip any paint
					// events sent by the system.
					continue
				}

				onPaint(glctx, sz)
				a.Publish()
				// Drive the animation by preparing to paint the next frame
				// after this one is shown.
				a.Send(paint.Event{})
			case touch.Event:
				touchX = e.X
				touchY = e.Y
			}
		}
	})
}

// DemoClient is the client side (slave) of the corresponding DemoServer in
// rexserver.
type DemoClient rexdemo.Demo

// NewDemo wraps the result of rexdemo.NewDemo() as a DemoClient
func NewDemo() *DemoClient {
	return (*DemoClient)(rexdemo.NewDemo())
}

// HandleEvent processes events broadcast from the server.
func (c *DemoClient) HandleEvent(ctx context.Context, rc *room.Client, ev room.Event) {
	log.Printf("[INFO] HANDLING")
	// No need to lock on the client side because HandleEvent calls are serialized.

	_c := DemoClient{Mut: c.Mut}
	err := json.Unmarshal([]byte(ev.Data()), &_c)
	if err != nil {
		log.Printf("[ERR] Malformed event data from server: %q", ev.Data())
		return
	}
	*c = _c
	log.Printf("[INFO] Event: %s", ev.Data())
}

func onStart(glctx gl.Context) {
	var err error
	program, err = glutil.CreateProgram(glctx, vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	buf = glctx.CreateBuffer()
	glctx.BindBuffer(gl.ARRAY_BUFFER, buf)
	glctx.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)

	position = glctx.GetAttribLocation(program, "position")
	color = glctx.GetUniformLocation(program, "color")
	offset = glctx.GetUniformLocation(program, "offset")

	images = glutil.NewImages(glctx)
	fps = debug.NewFPS(images)
}

func onStop(glctx gl.Context) {
	glctx.DeleteProgram(program)
	glctx.DeleteBuffer(buf)
	fps.Release()
	images.Release()
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(1, 0, 0, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)

	glctx.UseProgram(program)

	green += 0.01
	if green > 1 {
		green = 0
	}
	glctx.Uniform4f(color, 0, green, 0, 1)

	glctx.Uniform2f(offset, touchX/float32(sz.WidthPx), touchY/float32(sz.HeightPx))

	glctx.BindBuffer(gl.ARRAY_BUFFER, buf)
	glctx.EnableVertexAttribArray(position)
	glctx.VertexAttribPointer(position, coordsPerVertex, gl.FLOAT, false, 0, 0)
	glctx.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	glctx.DisableVertexAttribArray(position)

	fps.Draw(sz)
}

var triangleData = f32.Bytes(binary.LittleEndian,
	0.0, 0.4, 0.0, // top left
	0.0, 0.0, 0.0, // bottom left
	0.4, 0.0, 0.0, // bottom right
)

const (
	coordsPerVertex = 3
	vertexCount     = 3
)

const vertexShader = `#version 100
uniform vec2 offset;

attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position + offset4;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`
