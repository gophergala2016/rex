package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	_color "image/color"
	"log"
	"net"
	"os"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/gophergala2016/rex/examples/demo/rexdemo"
	"github.com/gophergala2016/rex/examples/exutil/exfont"
	"github.com/gophergala2016/rex/room"
	"golang.org/x/image/font"
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

const (
	bgRed   = 224
	bgGreen = 235
	bgBlue  = 245
)

var (
	remotePt chan RemotePoint
	messages chan room.Content
	demo     *DemoClient
	images   *glutil.Images
	fps      *debug.FPS
	program  gl.Program
	position gl.Attrib
	offset   gl.Uniform
	color    gl.Uniform
	buf      gl.Buffer

	statusFont    *truetype.Font
	statusFace    font.Face
	statusFaceOpt = &truetype.Options{}
	statusPainter *rexdemo.StatusPainter
	statusBG      = _color.RGBA{R: bgRed, G: bgGreen, B: bgBlue, A: 255}

	green          float32
	touchThreshold = 100 * time.Millisecond
	touched        bool
	touchTime      time.Time
	touchX         float32
	touchY         float32
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

		demo = NewDemo()
		client := room.NewClient(demo)
		messages = make(chan room.Content, 1)
		remotePt = make(chan RemotePoint, 1)

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

			clientShutdown := make(chan struct{})
			mdone := make(chan struct{})
			defer func() {
				close(clientShutdown)
			}()
			go func() {
				defer close(mdone)
				for {
					select {
					case mc := <-messages:
						err := client.Send(ctx, mc)
						if err != nil {
							log.Printf("[ERR] Sending message: %v", err)
						}
					case <-clientShutdown:
						return
					}
				}
			}()

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
			case pt := <-remotePt:
				touchX = float32(pt.X * float64(sz.WidthPx))
				touchY = float32(pt.Y * float64(sz.HeightPx))
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
				if e.Type == touch.TypeEnd {
					if touched && time.Since(touchTime) > touchThreshold {
						touched = false
					}
					if !touched {
						touched = true
						touchTime = time.Now()
						_x := float64(touchX) / float64(sz.WidthPx)
						_y := float64(touchY) / float64(sz.HeightPx)
						pt := fmt.Sprintf("%g,%g", _x, _y)
						select {
						case messages <- room.String(pt):
							log.Printf("[INFO] Touch event sent")
						default:
						}
					}
				}
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
	c.Mut.Lock()
	defer c.Mut.Unlock()

	_c := DemoClient{Mut: c.Mut}
	err := json.Unmarshal([]byte(ev.Data()), &_c)
	if err != nil {
		log.Printf("[ERR] Malformed event data from server: %q", ev.Data())
		return
	}
	*c = _c

	// try to update the local touch data... don't try too hard
	// TODO: make this more resilient.
	pt := RemotePoint{X: c.X, Y: c.Y}
	select {
	case remotePt <- pt:
	default:
	}

	log.Printf("[INFO] Event: %s", ev.Data())
}

// State returns the current demo state of the demo.
func (c *DemoClient) State() *rexdemo.Demo {
	c.Mut.Lock()
	defer c.Mut.Unlock()

	d := &rexdemo.Demo{}
	*d = (rexdemo.Demo)(*c)

	return d
}

func onStart(glctx gl.Context) {
	var err error
	program, err = glutil.CreateProgram(glctx, vertexShader, fragmentShader)
	if err != nil {
		log.Printf("[ERR] Failed creating GL program: %v", err)
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

	statusFont, statusFace, err = exfont.LoadAsset("Tuffy.ttf", statusFaceOpt)
	if err != nil {
		log.Printf("[ERR] Failed to load status font: %v", err)
	}
	statusPainter = rexdemo.NewStatusPainter(demo, statusFont, statusBG, images)

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("[ERR] Failed to retreived interfaces")
	} else {
		log.Printf("[DEBUG] %d network interfaces", len(ifaces))
		for _, iface := range ifaces {
			log.Printf("[DEBUG] IFACE %d %s", iface.Index, iface.Name)
		}
	}
}

func onStop(glctx gl.Context) {
	glctx.DeleteProgram(program)
	glctx.DeleteBuffer(buf)
	fps.Release()
	if statusPainter != nil {
		statusPainter.Release()
	}
	images.Release()
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(float32(bgRed)/255, float32(bgGreen)/255, float32(bgBlue)/255, 1)
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

	statusPainter.Draw(sz, actionBarPad, statusFaceOpt)
	fps.Draw(sz)
}

func paintState(glctx gl.Context, sz size.Event) {
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

// RemotePoint is a touch event from another client
type RemotePoint struct {
	X float64
	Y float64
}
