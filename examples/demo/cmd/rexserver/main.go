package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	_color "image/color"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/cli"
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

var (
	remotePt chan rexdemo.RemotePoint
	demo     *DemoServer
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
	statusBG      = image.NewUniform(_color.White)

	green  float32
	touchX float32
	touchY float32
)

func main() {
	app := cli.NewApp()
	app.Usage = "Start a REx Demo server"
	app.Action = AppMain
	app.Run(os.Args)
}

// AppMain runs the app package main loop.
func AppMain(*cli.Context) {
	app.Main(ServerMain)
}

// ServerMain performs the main routine for the demo server.
func ServerMain(a app.App) {
	background := context.Background()
	remotePt = make(chan rexdemo.RemotePoint, 1)
	demo = NewDemo()

	go RunDiscovery(background, demo)

	var glctx gl.Context
	var sz size.Event
	for e := range a.Events() {
		select {
		case pt := <-remotePt:
			touchX = float32(pt.X * float64(sz.WidthPx))
			touchY = float32(pt.Y * float64(sz.HeightPx))
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
	statusPainter = rexdemo.NewStatusPainter(demo, statusFont, _color.White, images)
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

	statusPainter.Draw(sz, actionBarPad, statusFaceOpt)
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

// RunDiscovery runs the discover server
func RunDiscovery(background context.Context, demo *DemoServer) {
	var bestAddr string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("[ERR] Failed to retreived interfaces")
	} else {
		log.Printf("[INFO] %d network interface addresses", len(addrs))
		for _, addr := range addrs {
			str := addr.String()
			if strings.HasPrefix(str, "10.") || strings.HasPrefix(str, "192.") {
				// looks like a local address. we will try to bind to it.
				bestAddr = addr.String()
				bestAddr = strings.SplitN(bestAddr, "/", 2)[0]
				bestAddr += ":0"
			}
			log.Printf("[DEBUG] IFACE %s", addr)
		}
	}
	if bestAddr == "" {
		log.Printf("[WARN] Unable to locate a good address for binding")
	}

	log.Printf("[INFO] demo server initializing")
	bus := room.NewBus(background, demo)
	config := &room.ServerConfig{
		Room: rexdemo.Room,
		Bus:  bus,
		Addr: bestAddr,
	}
	server := room.NewServer(config)

	log.Printf("[INFO] starting server")
	err = server.Start()
	if err != nil {
		log.Printf("[FATAL] %v", err)
		return
	}

	log.Printf("[INFO] server running at %s", server.Addr())

	log.Printf("[INFO] creating mDNS discovery server")
	zc, err := room.NewZoneConfig(server)
	if err != nil {
		log.Printf("[FATAL] Failed to initialize discovery")
		return
	}
	disco, err := room.DiscoveryServer(zc)
	if err != nil {
		log.Printf("[FATAL] Discovery server failed to start: %v", err)
		return
	}
	defer disco.Close()

	err = server.Wait()
	if err != nil {
		log.Printf("[FATAL] %v", err)
		return
	}
}

// DemoServer is the server side (source of truth) of the demo object.
type DemoServer rexdemo.Demo

// NewDemo wraps the result of rexdemo.NewDemo() as a DemoServer
func NewDemo() *DemoServer {
	return (*DemoServer)(rexdemo.NewDemo())
}

// State implements rexdemo.State
func (d *DemoServer) State() *rexdemo.Demo {
	return (*rexdemo.Demo)(d).State()
}

// HandleMessage adds to the message counter
func (d *DemoServer) HandleMessage(ctx context.Context, msg room.Msg) {
	var okpt bool
	var x, y float64
	data := msg.Text()
	_, err := fmt.Sscanf(data, "%g,%g", &x, &y)
	if err == nil {
		log.Printf("[INFO] Got a point [%0.03g,%0.03g]", x, y)
		okpt = true
	}

	d.Mut.Lock()
	defer d.Mut.Unlock()
	d.Counter++
	d.Last = time.Now()
	if okpt {
		d.X = x
		d.Y = y
		pt := rexdemo.Pt(x, y)
		// TODO: more resilient transfer of state.
		select {
		case remotePt <- pt:
			log.Printf("[INFO] Sent point [%0.03g,%0.03g]", x, y)
		default:
		}
	}
	log.Printf("[DEBUG] %v session %v %q", msg.Time(), msg.Session(), data)
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
