package desk

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// menu is a list of choices opened at the cursor, over everything else.
type menu struct {
	at    image.Point
	items []item
}

type item struct {
	label string
	do    func()
}

// openMenu opens a menu at the cursor, moved as far as it must be to fit the window.
func (d *desk) openMenu(at image.Point, items []item) {
	if len(items) == 0 {
		return
	}
	m := &menu{items: items}
	r := m.rect(d, len(items)-1)
	at.X = min(at.X, d.w-r.Dx())
	at.Y = min(at.Y, d.h-r.Max.Y)
	m.at = at
	d.menu = m
}

func (m *menu) rect(d *desk, i int) image.Rectangle {
	w, h := int(d.px(190)), int(d.px(itemH))
	return image.Rect(m.at.X, m.at.Y+i*h, m.at.X+w, m.at.Y+(i+1)*h)
}

// update takes a click on a choice, or anywhere else, which closes the menu.
func (m *menu) update(d *desk) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		d.menu = nil
		return
	}
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	d.menu = nil
	pt := image.Pt(ebiten.CursorPosition())
	for i, it := range m.items {
		if pt.In(m.rect(d, i)) {
			it.do()
		}
	}
}

func (m *menu) draw(dst *ebiten.Image, d *desk) {
	pt := image.Pt(ebiten.CursorPosition())
	for i, it := range m.items {
		b := m.rect(d, i)
		c := button
		if pt.In(b) {
			c = hovered
		}
		fill(dst, b, c)
		drawText(dst, d.fonts.text, it.label, float64(b.Min.X)+d.px(8), float64(b.Min.Y+b.Dy()/2), ink, text.AlignStart)
	}
}
