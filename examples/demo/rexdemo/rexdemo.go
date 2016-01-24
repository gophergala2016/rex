package rexdemo

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"sync"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/gophergala2016/rex/room"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/geom"
)

// Room is the room used by clients and servers for the demo.
var Room = &room.Room{
	Name:    "REx Demo",
	Service: "_rexdemo._tcp.",
}

// RemotePoint is a touch event from another client
type RemotePoint struct {
	X float64
	Y float64
}

// Pt is a simple constructor for RemotePoint.
func Pt(x, y float64) RemotePoint {
	return RemotePoint{x, y}
}

// Demo is the state of a demo a copy of the state is present in the server and
// all clients.
type Demo struct {
	Mut     *sync.Mutex `json:"-"`
	X       float64     `json:"x"`
	Y       float64     `json:"y"`
	Counter int         `json:"counter"`
	Last    time.Time   `json:"last"`
}

// NewDemo returns a new Demo object
func NewDemo() *Demo {
	// It's a little weird that a pointer is preferred over sync.Mutex.  But
	// due to how state is shored be the client and server a reference makes
	// more sense.
	return &Demo{
		Mut: new(sync.Mutex),
	}
}

// State returns a snapshot of d.
func (d *Demo) State() *Demo {
	_d := &Demo{}
	*_d = *d
	_d.Mut = nil
	return _d
}

// State is an interface satisfied by other Demo types.
type State interface {
	State() *Demo
}

// StatusPainter is an object responsible for rendering the demo status at the
// top of the client UI.
type StatusPainter struct {
	bg     image.Image
	demo   State
	ttf    *truetype.Font
	opt    *truetype.Options
	face   font.Face
	frozen *Demo
	sz     size.Event
	image  *glutil.Image
	images *glutil.Images
}

// NewStatusPainter initializes and returns a StatusPainter.
func NewStatusPainter(demo State, font *truetype.Font, bg color.Color, images *glutil.Images) *StatusPainter {
	return &StatusPainter{
		demo:   demo,
		bg:     image.NewUniform(bg),
		ttf:    font,
		images: images,
	}
}

// Release calls Release on underlying gl elements.
func (p *StatusPainter) Release() {
	p.image.Release()
}

// Draw renders the demo state to the screen
func (p *StatusPainter) Draw(sz size.Event, pad int, opt *truetype.Options) {

	if sz.WidthPx == 0 && sz.HeightPx == 0 {
		return
	}
	fsize := opt.Size
	if fsize == 0 {
		fsize = 12
	}
	pixY := int(float64(fsize)*float64(sz.PixelsPerPt)) + pad

	if p.sz != sz {
		p.sz = sz
		if p.image != nil {
			p.image.Release()
		}
		p.image = p.images.NewImage(sz.WidthPx, pixY)
		// BUG: face will not update if opt changes
		p.opt = nil

	}

	if p.opt == nil || *opt != *p.opt {
		if p.opt != nil {
			dpi := p.opt.DPI
			p.opt.DPI = opt.DPI
			if *opt == *p.opt {
				// we have already correctly computed the DPI so this opt is
				// nothing new.
				p.opt.DPI = dpi
			} else {
				p.opt = nil
			}
		}
		if p.opt == nil {
			p.opt = &truetype.Options{}
			*p.opt = *opt
			p.opt.DPI = float64(sz.PixelsPerPt) * 72
			p.face = truetype.NewFace(p.ttf, p.opt)
		}
	}

	state := p.demo.State()
	state.Mut = nil
	if p.frozen == nil || *state != *p.frozen {
		// generate a current image
		draw.Draw(p.image.RGBA, p.image.RGBA.Bounds(), p.bg, image.Pt(0, 0), draw.Over)

		originX := 2
		originY := pixY - 2

		drawer := &font.Drawer{
			Dst:  p.image.RGBA,
			Src:  image.NewUniform(color.Black),
			Face: p.face,
			Dot:  fixed.P(originX, originY),
		}

		text := fmt.Sprintf("%v Count=%d", state.Last.Format(time.RFC3339), state.Counter)
		drawer.DrawString(text)
		p.image.Upload()
	}
	p.image.Draw(
		sz,
		geom.Point{X: 0, Y: 0},
		geom.Point{X: sz.WidthPt, Y: 0},
		geom.Point{X: 0, Y: geom.Pt(float64(pixY) / float64(sz.PixelsPerPt))},
		p.image.RGBA.Bounds(),
	)

}
