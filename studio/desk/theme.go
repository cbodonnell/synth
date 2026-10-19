package desk

import (
	"hash/fnv"
	"image"
	"image/color"
	"math"

	"github.com/cbodonnell/ebui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

var (
	ground    = color.RGBA{0x1b, 0x1d, 0x21, 0xff}
	panel     = color.RGBA{0x23, 0x26, 0x2b, 0xff}
	laneEven  = color.RGBA{0x20, 0x23, 0x28, 0xff}
	laneOdd   = color.RGBA{0x24, 0x27, 0x2c, 0xff}
	selected  = color.RGBA{0x2c, 0x33, 0x3d, 0xff}
	beatLine  = color.RGBA{0x29, 0x2d, 0x33, 0xff}
	barLine   = color.RGBA{0x36, 0x3b, 0x42, 0xff}
	edge      = color.RGBA{0x12, 0x13, 0x16, 0xff}
	ink       = color.RGBA{0xd6, 0xd8, 0xdb, 0xff}
	dim       = color.RGBA{0x84, 0x8b, 0x94, 0xff}
	faint     = color.RGBA{0x55, 0x5b, 0x63, 0xff}
	button    = color.RGBA{0x2f, 0x34, 0x3b, 0xff}
	hovered   = color.RGBA{0x3a, 0x40, 0x48, 0xff}
	pressed   = color.RGBA{0x4a, 0x51, 0x5b, 0xff}
	lit       = color.RGBA{0x3d, 0x6b, 0x8f, 0xff}
	playhead  = color.RGBA{0xff, 0xa0, 0x40, 0xff}
	loopColor = color.RGBA{0x3c, 0xc0, 0xb0, 0xff}
	soloColor = color.RGBA{0xe0, 0xc0, 0x40, 0xff}
	muteColor = color.RGBA{0xd0, 0x60, 0x50, 0xff}
	meterLow  = color.RGBA{0x50, 0xc0, 0x70, 0xff}
	meterHigh = color.RGBA{0xe0, 0x50, 0x40, 0xff}
	failColor = color.RGBA{0xe0, 0x70, 0x60, 0xff}
)

// Roles a button takes, for the painter.
const (
	roleOn = "on" // a toggle that's on
)

// fonts are the faces at the window's scale.
type fonts struct {
	regular, mono font.Face
	text, num     text.Face
}

func newFonts(scale float64) fonts {
	face := func(ttf []byte, size float64) font.Face {
		f, err := opentype.Parse(ttf)
		if err != nil {
			panic(err)
		}
		face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: size * scale, DPI: 72, Hinting: font.HintingFull})
		if err != nil {
			panic(err)
		}
		return face
	}
	f := fonts{regular: face(goregular.TTF, 12), mono: face(gomono.TTF, 12)}
	f.text, f.num = text.NewGoXFace(f.regular), text.NewGoXFace(f.mono)
	return f
}

// painter draws ebui's parts flat and dark.
type painter struct{ f *fonts }

func (p painter) Paint(dst *ebiten.Image, pt ebui.Paint) bool {
	r := pt.Rect
	switch pt.Part {
	case ebui.PartButton:
		var c color.Color = button
		if pt.Role == roleItem {
			c = nil
		}
		switch {
		case pt.State.Has(ebui.StateDisabled):
			c = panel
		case pt.State.Has(ebui.StatePressed):
			c = pressed
		case pt.Role == roleOn || pt.Role == roleItemOn:
			c = lit
		case pt.State.Has(ebui.StateHovered):
			c = hovered
		}
		if c != nil {
			fill(dst, r, c)
		}
	case ebui.PartButtonLabel:
		c := ink
		if pt.State.Has(ebui.StateDisabled) {
			c = faint
		}
		y := float64(r.Min.Y + r.Dy()/2)
		if pt.Role == roleItem || pt.Role == roleItemOn {
			drawText(dst, p.f.text, pt.Text, float64(r.Min.X)+float64(r.Dy())/4, y, c, text.AlignStart)
			break
		}
		drawText(dst, p.f.text, pt.Text, float64(r.Min.X+r.Dx()/2), y, c, text.AlignCenter)
	case ebui.PartScrollTrack:
		fill(dst, r, panel)
	case ebui.PartScrollThumb:
		fill(dst, r.Inset(max(r.Dx()/4, 1)), hovered)
	case ebui.PartFocusRing:
	default:
		return false
	}
	return true
}

func fill(dst *ebiten.Image, r image.Rectangle, c color.Color) {
	if r.Empty() {
		return
	}
	vector.FillRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), c, false)
}

func fillF(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	if w <= 0 || h <= 0 {
		return
	}
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), c, false)
}

func line(dst *ebiten.Image, x0, y0, x1, y1, width float64, c color.Color) {
	vector.StrokeLine(dst, float32(x0), float32(y0), float32(x1), float32(y1), float32(width), c, false)
}

// drawText draws s with its middle at y and aligned at x.
func drawText(dst *ebiten.Image, face text.Face, s string, x, y float64, c color.Color, align text.Align) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	op.PrimaryAlign = align
	op.SecondaryAlign = text.AlignCenter
	text.Draw(dst, s, face, op)
}

// fit cuts s to fit width, with an ellipsis.
func fit(face text.Face, s string, width float64) string {
	if text.Advance(s, face) <= width {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && text.Advance(string(r)+"…", face) > width {
		r = r[:len(r)-1]
	}
	if len(r) == 0 {
		return ""
	}
	return string(r) + "…"
}

// hue is a track's color, the same for its name every time.
func hue(name string, value float64) color.RGBA {
	h := fnv.New32a()
	h.Write([]byte(name))
	return hsv(float64(h.Sum32()%360), 0.45, value)
}

func hsv(h, s, v float64) color.RGBA {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	var r, g, b float64
	switch {
	case h < 60:
		r, g = c, x
	case h < 120:
		r, g = x, c
	case h < 180:
		g, b = c, x
	case h < 240:
		g, b = x, c
	case h < 300:
		r, b = x, c
	default:
		r, b = c, x
	}
	m := v - c
	return color.RGBA{uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255), 0xff}
}
