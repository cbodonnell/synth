package desk

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/studio/session"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	inspectorW  = 210
	repeatEvery = 300 * time.Millisecond
)

// soundPanel is what a sound effect adds to the desk: an inspector for how the player
// plays it (Vary, Voices, Priority) beside the clip editor, and auditions through the
// player as a game plays it, once or every repeatEvery.
type soundPanel struct {
	r        image.Rectangle
	id       string
	renderer *synth.Renderer
	began    time.Time
	repeat   bool
	next     time.Time

	drag     *drag
	moved    bool
	pending  *typing // what a click without a drag types into
	typing   *typing
	hint     string
	last     string
	lastTime time.Time
}

// place splits the detail between the clip editor and, for a sound, the inspector.
func (s *soundPanel) place(d *desk) {
	r := d.widgets.detail
	d.ce.r, s.r = r, image.Rectangle{}
	if d.sound() && !r.Empty() {
		split := r.Max.X - int(d.px(inspectorW))
		d.ce.r, s.r = image.Rect(r.Min.X, r.Min.Y, split, r.Max.Y), image.Rect(split, r.Min.Y, r.Max.X, r.Max.Y)
	}
}

// audition plays the sound as the game would: a fresh take, pitch and gain within its
// Vary, through its voice cap and priority.
func (s *soundPanel) audition(d *desk) {
	if s.renderer == nil {
		s.renderer = synth.NewRenderer()
	}
	song, r := session.Soloed(d.layout.Song, d.view.Solo), s.renderer
	go d.player.Audition(song, 1, func(take int) synth.Stereo { return r.Render(synth.Take(song, take)) })
	s.began = time.Now()
	d.view.At, d.frame, d.sounding = 0, 0, true
}

func (s *soundPanel) toggleRepeat(d *desk) {
	s.repeat = !s.repeat
	if s.repeat {
		s.next = time.Now()
	}
}

// follow moves the playhead along the last audition, by the clock.
func (s *soundPanel) follow(d *desk) {
	at := time.Since(s.began).Seconds()
	if d.take == nil || at >= d.take.Time(d.take.Buffer.Len()) {
		d.sounding = false
		return
	}
	d.frame, d.view.At = d.take.Frame(at), at
}

func (s *soundPanel) update(d *desk) {
	s.place(d)
	if s.id != d.layout.ID {
		s.id, s.repeat, s.drag, s.typing = d.layout.ID, false, nil, nil
	}
	if s.repeat && d.sound() && !time.Now().Before(s.next) {
		s.audition(d)
		s.next = s.next.Add(repeatEvery)
		if time.Now().After(s.next) {
			s.next = time.Now().Add(repeatEvery)
		}
	}
	s.hint = ""
	if s.drag != nil {
		g := s.drag
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || d.doc == nil {
			if !s.moved {
				s.typing = s.pending
			}
			s.drag, s.pending = nil, nil
			if d.doc != nil {
				d.doc.Release()
			}
			return
		}
		_, y := ebiten.CursorPosition()
		if math.Abs(float64(y)-g.y) > d.px(3) {
			s.moved = true
		}
		if !s.moved {
			return
		}
		dy := (g.y - float64(y)) / g.travel
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			dy /= 10
		}
		g.set(session.Round(math.Max(g.lo, math.Min(g.hi, g.from+dy*(g.hi-g.lo))), g.step))
		return
	}
	pt := image.Pt(ebiten.CursorPosition())
	if !pt.In(s.r) {
		return
	}
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if clicked {
		s.typing = nil
	}
	for i, row := range s.rows(d) {
		if !pt.In(s.rowRect(d, i)) {
			continue
		}
		if row.buttons != nil {
			for j, b := range row.buttons {
				if clicked && b.can && pt.In(s.buttonRect(d, i, j, len(row.buttons))) {
					b.do()
				}
			}
			s.hint = "space plays it as the game does, a fresh take each time · R repeats it every 300ms"
			return
		}
		if !d.editable(row.path...) {
			s.hint = custom
			if d.doc == nil {
				s.hint = d.readOnly()
			}
			return
		}
		s.hint = row.about + " · drag (shift: fine) · click to type · double-click resets"
		if !clicked {
			return
		}
		now := time.Now()
		if s.last == row.label && now.Sub(s.lastTime) < 400*time.Millisecond {
			s.last = ""
			row.set(0)
			return
		}
		s.last, s.lastTime = row.label, now
		s.moved = false
		s.pending = &typing{id: row.label, text: strconv.FormatFloat(row.value, 'f', -1, 64), chars: "0123456789.", set: func(text string) bool {
			v, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
			if err == nil {
				row.set(session.Round(math.Max(row.lo, math.Min(row.hi, v)), row.step))
			}
			return err == nil
		}}
		d.doc.Hold()
		s.drag = &drag{id: row.label, from: row.value, y: float64(pt.Y), set: row.set, travel: d.px(200), lo: row.lo, hi: row.hi, step: row.step}
		return
	}
}

// keys types into a value of the inspector, reporting whether it took the keys.
func (s *soundPanel) keys() bool {
	if s.typing == nil || s.drag != nil {
		return false
	}
	if typeInto(s.typing) {
		s.typing = nil
	}
	return true
}

// A soundRow is a value of how the player plays the sound, or the audition's buttons.
type soundRow struct {
	label, about string
	path         []any
	value        float64
	shown        string
	lo, hi, step float64
	set          func(v float64)
	buttons      []clipButton
}

func (s *soundPanel) rows(d *desk) []soundRow {
	a := d.layout.Arrangement
	edit := func(f func(a *synth.Arrangement, v float64)) func(v float64) {
		return func(v float64) { d.edit(func(a *synth.Arrangement) { f(a, v) }) }
	}
	count := func(n int, zero string) string {
		if n == 0 {
			return zero
		}
		return strconv.Itoa(n)
	}
	return []soundRow{
		{buttons: []clipButton{
			{"Play", func() { s.audition(d) }, true, false},
			{"Repeat", func() { s.toggleRepeat(d) }, true, s.repeat},
		}},
		{
			label: "pitch ±", about: "semitones each play moves at random", path: []any{"Vary", "Pitch"},
			value: a.Vary.Pitch, shown: fmt.Sprintf("%g st", a.Vary.Pitch), hi: 12, step: 0.1,
			set: edit(func(a *synth.Arrangement, v float64) { a.Vary.Pitch = v }),
		},
		{
			label: "gain ±", about: "dB each play moves at random", path: []any{"Vary", "Gain"},
			value: a.Vary.Gain, shown: fmt.Sprintf("%g dB", a.Vary.Gain), hi: 12, step: 0.1,
			set: edit(func(a *synth.Arrangement, v float64) { a.Vary.Gain = v }),
		},
		{
			label: "takes", about: "renders with their noise seeded apart, never the same twice running", path: []any{"Vary", "Takes"},
			value: float64(a.Vary.Takes), shown: count(a.Vary.Takes, "1"), hi: 8, step: 1,
			set: edit(func(a *synth.Arrangement, v float64) { a.Vary.Takes = int(v) }),
		},
		{
			label: "voices", about: "plays sounding at once; a new one stops the oldest", path: []any{"Voices"},
			value: float64(a.Voices), shown: count(a.Voices, "any"), hi: 16, step: 1,
			set: edit(func(a *synth.Arrangement, v float64) { a.Voices = int(v) }),
		},
		{
			label: "priority", about: "who gives way when the player runs out of voices", path: []any{"Priority"},
			value: float64(a.Priority), shown: strconv.Itoa(a.Priority), hi: 9, step: 1,
			set: edit(func(a *synth.Arrangement, v float64) { a.Priority = int(v) }),
		},
	}
}

func (s *soundPanel) rowRect(d *desk, i int) image.Rectangle {
	y := s.r.Min.Y + int(d.px(clipHeadH+6)) + i*int(d.px(settingH))
	return image.Rect(s.r.Min.X+int(d.px(10)), y, s.r.Max.X-int(d.px(10)), y+int(d.px(settingH))-1)
}

func (s *soundPanel) buttonRect(d *desk, i, j, n int) image.Rectangle {
	r := s.rowRect(d, i)
	w := (r.Dx() - int(d.px(4))*(n-1)) / n
	x := r.Min.X + j*(w+int(d.px(4)))
	return image.Rect(x, r.Min.Y+1, x+w, r.Max.Y-1)
}

func (s *soundPanel) draw(dst *ebiten.Image, d *desk) {
	if s.r.Empty() {
		return
	}
	if _, ok := d.target(); !ok {
		r := d.ce.r
		fill(dst, r, ground)
		line(dst, float64(r.Min.X), float64(r.Min.Y), float64(r.Max.X), float64(r.Min.Y), 1, edge)
		drawText(dst, d.fonts.text, "click a layer's clip to edit its notes", float64(r.Min.X)+d.px(12), float64(r.Min.Y)+d.px(20), dim, text.AlignStart)
	}
	fill(dst, s.r, panel)
	line(dst, float64(s.r.Min.X), float64(s.r.Min.Y), float64(s.r.Max.X), float64(s.r.Min.Y), 1, edge)
	line(dst, float64(s.r.Min.X), float64(s.r.Min.Y), float64(s.r.Min.X), float64(s.r.Max.Y), 1, edge)
	drawText(dst, d.fonts.text, "as the game plays it", float64(s.r.Min.X)+d.px(10), float64(s.r.Min.Y)+d.px(clipHeadH/2), dim, text.AlignStart)
	for i, row := range s.rows(d) {
		r := s.rowRect(d, i)
		if r.Max.Y > s.r.Max.Y {
			break
		}
		if row.buttons != nil {
			for j, b := range row.buttons {
				box(dst, d, s.buttonRect(d, i, j, len(row.buttons)), b.label, b.on, lit, b.can)
			}
			continue
		}
		mid := float64(r.Min.Y + r.Dy()/2)
		drawText(dst, d.fonts.text, row.label, float64(r.Min.X), mid, dim, text.AlignStart)
		v := image.Rect(r.Min.X+int(d.px(70)), r.Min.Y, r.Max.X, r.Max.Y)
		value, col := row.shown, color.Color(ink)
		if !d.editable(row.path...) {
			value, col = row.shown+" (code)", faint
		}
		if s.typing != nil && s.typing.id == row.label {
			fill(dst, v, edge)
			value, col = s.typing.text+"▏", playhead
		}
		drawText(dst, d.fonts.num, fit(d.fonts.num, value, float64(v.Dx())), float64(v.Min.X), mid, col, text.AlignStart)
	}
}
