package desk

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// prompt asks a question over the whole window, which waits for the answer.
type prompt struct {
	text    string
	choices []choice
}

type choice struct {
	label string
	do    func()
}

const promptW, promptH, choiceW = 460, 110, 140

func (p *prompt) box(d *desk) image.Rectangle {
	w, h := int(d.px(promptW)), int(d.px(promptH))
	x, y := (d.w-w)/2, (d.h-h)/3
	return image.Rect(x, y, x+w, y+h)
}

func (p *prompt) choice(d *desk, i int) image.Rectangle {
	b := p.box(d)
	w, h, gap := int(d.px(choiceW)), int(d.px(buttonH)), int(d.px(8))
	x := b.Max.X - gap - (len(p.choices)-i)*(w+gap) + gap
	y := b.Max.Y - gap - h
	return image.Rect(x, y, x+w, y+h)
}

func (p *prompt) update(d *desk) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		d.ask = nil
		return
	}
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	pt := image.Pt(ebiten.CursorPosition())
	for i, c := range p.choices {
		if pt.In(p.choice(d, i)) {
			d.ask = nil
			c.do()
			return
		}
	}
}

func (p *prompt) draw(dst *ebiten.Image, d *desk) {
	fill(dst, image.Rect(0, 0, d.w, d.h), shade)
	b := p.box(d)
	fill(dst, b, panel)
	line(dst, float64(b.Min.X), float64(b.Min.Y), float64(b.Max.X), float64(b.Min.Y), d.px(2), playhead)
	pad := d.px(16)
	drawText(dst, d.fonts.text, fit(d.fonts.text, p.text, float64(b.Dx())-2*pad), float64(b.Min.X)+pad, float64(b.Min.Y)+pad+d.px(8), ink, text.AlignStart)
	pt := image.Pt(ebiten.CursorPosition())
	for i, c := range p.choices {
		r := p.choice(d, i)
		bg := button
		if pt.In(r) {
			bg = hovered
		}
		fill(dst, r, bg)
		drawText(dst, d.fonts.text, c.label, float64(r.Min.X+r.Dx()/2), float64(r.Min.Y+r.Dy()/2), ink, text.AlignCenter)
	}
}
