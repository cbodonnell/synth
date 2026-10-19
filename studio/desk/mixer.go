package desk

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/studio/session"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Sizes in points, before the window's scale.
const (
	stripW   = 100
	rowH     = 18
	inserts  = 3
	knobR    = 11
	readoutH = 18
)

// mixer is the mixer view: a strip per track, then each bus, then the master.
type mixer struct {
	scroll float64 // pixels the strips are scrolled left

	drag   *drag
	typing *typing
	menu   *menu
	hint   string // what the control under the cursor can't do

	last     string // the control last clicked, and when, for a double-click
	lastTime time.Time
}

// A drag moves a value from where it was when the button went down.
type drag struct {
	id     string
	from   float64
	y      float64
	set    func(v float64)
	travel float64 // pixels from one end of the value's range to the other
	fader  bool    // moves along the fader's taper rather than linearly
	lo, hi float64
	step   float64
}

// typing is a value being typed into a readout.
type typing struct {
	id   string
	text string
	set  func(s string) bool
	// The characters it takes; "" takes a level's or a pan's.
	chars string
}

type menu struct {
	at    image.Point
	items []item
}

type item struct {
	label string
	do    func()
}

type stripKind int

const (
	trackStrip stripKind = iota
	busStrip
	masterStrip
)

type strip struct {
	kind  stripKind
	index int
	name  string
	r     image.Rectangle
}

func (m *mixer) strips(d *desk, r image.Rectangle) []strip {
	a := d.layout.Arrangement
	var out []strip
	x := float64(r.Min.X) - m.scroll
	w := d.px(stripW)
	add := func(k stripKind, i int, name string) {
		out = append(out, strip{k, i, name, image.Rect(int(x), r.Min.Y, int(x+w)-1, r.Max.Y)})
		x += w
	}
	for i, t := range a.Tracks {
		add(trackStrip, i, t.Name)
	}
	for i, b := range a.Buses {
		add(busStrip, i, b.Name)
	}
	add(masterStrip, 0, "master")
	return out
}

// sendRows is how many sends the strips leave room for: the most any track has, and one
// to add another.
func (m *mixer) sendRows(d *desk) int {
	n := 0
	for _, t := range d.layout.Arrangement.Tracks {
		n = max(n, len(t.Sends))
	}
	return n + 1
}

// The rows of a strip, top to bottom, in window pixels.
func (m *mixer) header(d *desk, s strip) image.Rectangle  { return m.row(d, s, 4, 22) }
func (m *mixer) buttons(d *desk, s strip) image.Rectangle { return m.row(d, s, 28, 16) }
func (m *mixer) out(d *desk, s strip) image.Rectangle     { return m.row(d, s, 48, 16) }
func (m *mixer) insert(d *desk, s strip, i int) image.Rectangle {
	return m.row(d, s, 70+float64(i)*14, 14)
}
func (m *mixer) send(d *desk, s strip, i int) image.Rectangle {
	return m.row(d, s, 70+inserts*14+6+float64(i)*rowH, rowH)
}
func (m *mixer) panTop(d *desk) float64 { return 70 + inserts*14 + 6 + float64(m.sendRows(d))*rowH + 4 }
func (m *mixer) pan(d *desk, s strip) image.Rectangle {
	return m.row(d, s, m.panTop(d), 2*knobR+4)
}
func (m *mixer) panReadout(d *desk, s strip) image.Rectangle {
	return m.row(d, s, m.panTop(d)+2*knobR+4, 14)
}

// fader is the fader's travel: its bottom is -60 dB, its top +12.
func (m *mixer) fader(d *desk, s strip) image.Rectangle {
	top := float64(s.r.Min.Y) + d.px(m.panTop(d)+2*knobR+4+14+12)
	bottom := float64(s.r.Max.Y) - d.px(2*readoutH+12)
	x := float64(s.r.Min.X) + d.px(30)
	return image.Rect(int(x), int(top), int(x+d.px(8)), int(math.Max(bottom, top+d.px(20))))
}

func (m *mixer) meter(d *desk, s strip) image.Rectangle {
	f := m.fader(d, s)
	x := f.Max.X + int(d.px(30))
	return image.Rect(x, f.Min.Y, x+int(d.px(8)), f.Max.Y)
}

func (m *mixer) readout(d *desk, s strip) image.Rectangle {
	return image.Rect(s.r.Min.X+int(d.px(6)), s.r.Max.Y-int(d.px(2*readoutH+6)), s.r.Max.X-int(d.px(6)), s.r.Max.Y-int(d.px(readoutH+6)))
}

func (m *mixer) loudness(d *desk, s strip) image.Rectangle {
	return m.readout(d, s).Add(image.Pt(0, int(d.px(readoutH))))
}

func (m *mixer) row(d *desk, s strip, y, h float64) image.Rectangle {
	top := float64(s.r.Min.Y) + d.px(y)
	return image.Rect(s.r.Min.X+int(d.px(6)), int(top), s.r.Max.X-int(d.px(6)), int(top+d.px(h)))
}

// editable reports whether the value at the path in the song can be edited here.
func (d *desk) editable(path ...any) bool { return d.doc != nil && !d.doc.Custom(path...) }

const custom = "custom: edit it in code"

func (m *mixer) update(d *desk, r image.Rectangle) {
	m.hint = ""
	x, y := ebiten.CursorPosition()
	pt := image.Pt(x, y)

	if m.drag != nil {
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || d.doc == nil {
			m.drag = nil
			if d.doc != nil {
				d.doc.Release()
			}
			return
		}
		g := m.drag
		dy := (g.y - float64(y)) / g.travel
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			dy /= 10
		}
		var v float64
		if g.fader {
			v = session.FaderDB(session.FaderPosition(g.from) + dy)
		} else {
			v = g.from + dy*(g.hi-g.lo)
		}
		g.set(session.Round(math.Max(g.lo, math.Min(g.hi, v)), g.step))
		return
	}

	if !pt.In(r) {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			m.menu, m.typing = nil, nil
		}
		return
	}
	if wx, wy := ebiten.Wheel(); wx != 0 || wy != 0 {
		dx := wx
		if dx == 0 {
			dx = wy
		}
		m.scroll -= dx * d.px(24)
		m.clampScroll(d, r)
	}

	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if m.menu != nil {
		if !clicked {
			return
		}
		for i, it := range m.menu.items {
			if pt.In(m.itemRect(d, i)) {
				it.do()
			}
		}
		m.menu = nil
		return
	}

	for _, s := range m.strips(d, r) {
		if !pt.In(s.r) {
			continue
		}
		m.strip(d, s, pt, clicked)
		return
	}
	if clicked {
		m.typing = nil
	}
}

// strip handles the cursor over a strip.
func (m *mixer) strip(d *desk, s strip, pt image.Point, clicked bool) {
	a := d.layout.Arrangement
	on := func(r image.Rectangle, can bool) bool {
		if !pt.In(r) {
			return false
		}
		if !can {
			m.hint = custom
			if d.doc == nil {
				m.hint = d.readOnly()
			}
			return false
		}
		return clicked
	}
	double := func(id string) bool {
		now := time.Now()
		dbl := m.last == id && now.Sub(m.lastTime) < 400*time.Millisecond
		m.last, m.lastTime = id, now
		if dbl {
			m.last = ""
		}
		return dbl
	}
	if clicked {
		m.typing = nil
	}

	switch s.kind {
	case trackStrip:
		i := s.index
		t := a.Tracks[i]
		if clicked && pt.In(m.header(d, s)) {
			d.view.Track = t.Name
		}
		mute, solo := m.muteSolo(d, s)
		if pt.In(solo) && clicked {
			d.view.Track = t.Name
			d.toggleSolo(t.Name)
			return
		}
		if on(mute, d.editable("Tracks", i, "Mute")) {
			d.toggleMute(t.Name)
			return
		}
		if on(m.out(d, s), d.editable("Tracks", i, "Out")) {
			m.outMenu(d, i, pt)
			return
		}
		for j, send := range t.Sends {
			row := m.send(d, s, j)
			name, knob := m.sendParts(d, row)
			if on(name, d.editable("Tracks", i, "Sends")) {
				m.open(d, pt, []item{{"remove send to " + send.Bus, func() {
					d.edit(func(a *synth.Arrangement) {
						a.Tracks[i].Sends = slices.Delete(a.Tracks[i].Sends, j, j+1)
					})
				}}})
				return
			}
			id := fmt.Sprintf("send %d %d", i, j)
			set := func(v float64) { d.edit(func(a *synth.Arrangement) { a.Tracks[i].Sends[j].Level = v }) }
			if on(knob, d.editable("Tracks", i, "Sends", j, "Level")) {
				if double(id) {
					set(0)
					return
				}
				m.begin(d, &drag{id: id, from: send.Level, y: float64(pt.Y), set: set, travel: d.px(200), lo: -60, hi: 6, step: session.DBStep})
				return
			}
		}
		if len(t.Sends) < m.sendRows(d) && on(m.send(d, s, len(t.Sends)), d.editable("Tracks", i, "Sends")) {
			m.sendMenu(d, i, pt)
			return
		}
		pan := func(v float64) { d.edit(func(a *synth.Arrangement) { a.Tracks[i].Pan = v }) }
		if on(m.pan(d, s), d.editable("Tracks", i, "Pan")) {
			if double(fmt.Sprint("pan ", i)) {
				pan(0)
				return
			}
			m.begin(d, &drag{id: fmt.Sprint("pan ", i), from: t.Pan, y: float64(pt.Y), set: pan, travel: d.px(200), lo: -1, hi: 1, step: session.PanStep})
			return
		}
		if on(m.panReadout(d, s), d.editable("Tracks", i, "Pan")) {
			m.typing = &typing{id: fmt.Sprint("pan ", i), text: session.FormatPan(t.Pan), set: func(s string) bool {
				v, ok := session.ParsePan(s)
				if ok {
					pan(v)
				}
				return ok
			}}
			return
		}
		m.level(d, s, fmt.Sprint("level ", i), t.Level, d.editable("Tracks", i, "Level"), on, double,
			func(v float64) { d.edit(func(a *synth.Arrangement) { a.Tracks[i].Level = v }) })
	case busStrip:
		k := s.index
		m.level(d, s, fmt.Sprint("bus ", k), a.Buses[k].Level, d.editable("Buses", k, "Level"), on, double,
			func(v float64) { d.edit(func(a *synth.Arrangement) { a.Buses[k].Level = v }) })
	}
}

// level handles a strip's fader and its readout.
func (m *mixer) level(d *desk, s strip, id string, db float64, can bool, on func(image.Rectangle, bool) bool, double func(string) bool, set func(float64)) {
	f := m.fader(d, s)
	grab := image.Rect(f.Min.X-int(d.px(12)), f.Min.Y-int(d.px(6)), f.Max.X+int(d.px(12)), f.Max.Y+int(d.px(6)))
	if on(grab, can) {
		if double(id) {
			set(0)
			return
		}
		pt := image.Pt(ebiten.CursorPosition())
		from := db
		// A click off the thumb jumps it there first.
		if ty := m.thumbY(f, db); math.Abs(float64(pt.Y)-ty) > d.px(8) {
			from = session.Round(session.FaderDB((float64(f.Max.Y)-float64(pt.Y))/float64(f.Dy())), session.DBStep)
		}
		m.begin(d, &drag{id: id, from: from, y: float64(pt.Y), set: set, travel: float64(f.Dy()), fader: true, lo: -60, hi: 12, step: session.DBStep})
		set(from)
		return
	}
	if on(m.readout(d, s), can) {
		m.typing = &typing{id: id, text: fmt.Sprintf("%.1f", db), set: func(s string) bool {
			v, ok := session.ParseDB(s)
			if ok {
				set(v)
			}
			return ok
		}}
	}
}

func (m *mixer) begin(d *desk, g *drag) {
	d.doc.Hold()
	m.drag = g
}

func (m *mixer) outMenu(d *desk, i int, at image.Point) {
	a := d.layout.Arrangement
	set := func(out string) func() {
		return func() { d.edit(func(a *synth.Arrangement) { a.Tracks[i].Out = out }) }
	}
	items := []item{{"master", set("")}}
	for _, b := range a.Buses {
		items = append(items, item{b.Name, set(b.Name)})
	}
	m.open(d, at, items)
}

func (m *mixer) sendMenu(d *desk, i int, at image.Point) {
	a := d.layout.Arrangement
	t := a.Tracks[i]
	var items []item
	for _, b := range a.Buses {
		if b.Name == t.Out || slices.ContainsFunc(t.Sends, func(s synth.Send) bool { return s.Bus == b.Name }) {
			continue
		}
		items = append(items, item{"send to " + b.Name, func() {
			d.edit(func(a *synth.Arrangement) {
				a.Tracks[i].Sends = append(a.Tracks[i].Sends, synth.Send{Bus: b.Name, Level: -12})
			})
		}})
	}
	if d.editable("Buses") {
		items = append(items, item{"send to a new reverb", func() {
			d.edit(func(a *synth.Arrangement) { session.AddReturn(a, i) })
		}})
	}
	m.open(d, at, items)
}

// open opens a menu at the cursor, moved as far as it must be to fit the window.
func (m *mixer) open(d *desk, at image.Point, items []item) {
	if len(items) == 0 {
		return
	}
	m.menu = &menu{items: items}
	r := m.itemRect(d, len(items)-1)
	at.X = min(at.X, d.w-r.Dx())
	at.Y = min(at.Y, d.h-r.Max.Y)
	m.menu.at = at
}

func (m *mixer) muteSolo(d *desk, s strip) (mute, solo image.Rectangle) {
	b := m.buttons(d, s)
	w := (b.Dx() - int(d.px(4))) / 2
	return image.Rect(b.Min.X, b.Min.Y, b.Min.X+w, b.Max.Y), image.Rect(b.Max.X-w, b.Min.Y, b.Max.X, b.Max.Y)
}

func (m *mixer) sendParts(d *desk, row image.Rectangle) (name, knob image.Rectangle) {
	split := row.Max.X - int(d.px(40))
	return image.Rect(row.Min.X, row.Min.Y, split, row.Max.Y), image.Rect(split, row.Min.Y, row.Max.X, row.Max.Y)
}

func (m *mixer) thumbY(f image.Rectangle, db float64) float64 {
	return float64(f.Max.Y) - session.FaderPosition(db)*float64(f.Dy())
}

func (m *mixer) clampScroll(d *desk, r image.Rectangle) {
	n := len(d.layout.Arrangement.Tracks) + len(d.layout.Arrangement.Buses) + 1
	m.scroll = math.Max(0, math.Min(m.scroll, float64(n)*d.px(stripW)-float64(r.Dx())))
}

// keys types into the readout being typed into, reporting whether it took the keys.
func (m *mixer) keys(d *desk) bool {
	if m.typing == nil {
		return false
	}
	if typeInto(m.typing) {
		m.typing = nil
	}
	return true
}

// typeInto types this frame's keys into t, reporting whether it's done: entered, or
// given up with Escape.
func typeInto(t *typing) bool {
	chars := t.chars
	if chars == "" {
		chars = "0123456789.-+ dDbBlLrRcC"
	}
	for _, r := range ebiten.AppendInputChars(nil) {
		if chars == "*" || strings.ContainsRune(chars, r) {
			t.text += string(r)
		}
	}
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		return true
	case inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(t.text) > 0:
		_, n := utf8.DecodeLastRuneInString(t.text)
		t.text = t.text[:len(t.text)-n]
	case inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
		return t.set(t.text)
	}
	return false
}

func (m *mixer) itemRect(d *desk, i int) image.Rectangle {
	w, h := int(d.px(170)), int(d.px(itemH))
	return image.Rect(m.menu.at.X, m.menu.at.Y+i*h, m.menu.at.X+w, m.menu.at.Y+(i+1)*h)
}

func (m *mixer) draw(dst *ebiten.Image, d *desk, r image.Rectangle) {
	if r.Empty() {
		return
	}
	area := dst.SubImage(r).(*ebiten.Image)
	fill(area, r, ground)
	for n, s := range m.strips(d, r) {
		if s.r.Max.X < r.Min.X || s.r.Min.X > r.Max.X {
			continue
		}
		bg := laneEven
		if n%2 == 1 {
			bg = laneOdd
		}
		if s.kind == trackStrip && s.name == d.view.Track {
			bg = selected
		}
		if s.kind == masterStrip {
			bg = panel
		}
		fill(area, s.r, bg)
		m.drawStrip(area, d, s)
	}
	if m.menu != nil {
		for i, it := range m.menu.items {
			b := m.itemRect(d, i)
			c := button
			if image.Pt(ebiten.CursorPosition()).In(b) {
				c = hovered
			}
			fill(dst, b, c)
			drawText(dst, d.fonts.text, it.label, float64(b.Min.X)+d.px(8), float64(b.Min.Y+b.Dy()/2), ink, text.AlignStart)
		}
	}
}

func (m *mixer) drawStrip(dst *ebiten.Image, d *desk, s strip) {
	a := d.layout.Arrangement
	small := func(r image.Rectangle, s string, c color.Color, align text.Align) {
		x := float64(r.Min.X)
		if align == text.AlignCenter {
			x = float64(r.Min.X + r.Dx()/2)
		}
		drawText(dst, d.fonts.text, fit(d.fonts.text, s, float64(r.Dx())), x, float64(r.Min.Y+r.Dy()/2), c, align)
	}
	h := m.header(d, s)
	ink := color.Color(ink)
	switch s.kind {
	case trackStrip:
		fill(dst, image.Rect(h.Min.X, h.Min.Y, h.Max.X, h.Min.Y+int(d.px(3))), hue(s.name, 0.75))
		if a.Tracks[s.index].Mute {
			ink = faint
		}
	case busStrip:
		fill(dst, image.Rect(h.Min.X, h.Min.Y, h.Max.X, h.Min.Y+int(d.px(3))), dim)
	}
	small(h.Add(image.Pt(0, int(d.px(2)))), s.name, ink, text.AlignStart)

	var level float64
	var levelOK bool
	var fx []synth.Effect
	var fxCustom bool
	var gauge *synth.Gauge
	frame := d.frame
	if d.take != nil {
		frame = d.take.SongFrame(d.frame)
	}
	switch s.kind {
	case trackStrip:
		i := s.index
		t := a.Tracks[i]
		mute, solo := m.muteSolo(d, s)
		box(dst, d, mute, "M", t.Mute, muteColor, d.editable("Tracks", i, "Mute"))
		box(dst, d, solo, "S", slices.Contains(d.view.Solo, t.Name), soloColor, true)
		out := "→ master"
		if t.Out != "" {
			out = "→ " + t.Out
		}
		small(m.out(d, s), out, enabled(d.editable("Tracks", i, "Out")), text.AlignStart)
		for j, send := range t.Sends {
			name, knob := m.sendParts(d, m.send(d, s, j))
			small(name, send.Bus, enabled(d.editable("Tracks", i, "Sends")), text.AlignStart)
			m.drawSend(dst, d, knob, send.Level, d.editable("Tracks", i, "Sends", j, "Level"), fmt.Sprintf("send %d %d", i, j))
		}
		if len(t.Sends) < m.sendRows(d) && d.editable("Tracks", i, "Sends") {
			small(m.send(d, s, len(t.Sends)), "+ send", faint, text.AlignStart)
		}
		m.drawPan(dst, d, s, t.Pan, d.editable("Tracks", i, "Pan"), fmt.Sprint("pan ", i))
		level, levelOK = t.Level, d.editable("Tracks", i, "Level")
		fx, fxCustom = t.Inserts, d.doc != nil && d.doc.Custom("Tracks", i, "Inserts")
		if d.take != nil {
			if g, ok := d.take.Parts[t.Name]; ok {
				gauge = &g
			}
		}
		m.drawLevel(dst, d, s, level, levelOK, fmt.Sprint("level ", i))
	case busStrip:
		k := s.index
		b := a.Buses[k]
		role := "return"
		if slices.Contains(d.layout.Groups(), b.Name) {
			role = "group"
		}
		small(m.buttons(d, s), role, dim, text.AlignStart)
		small(m.out(d, s), "→ master", dim, text.AlignStart)
		fx, fxCustom = b.Effects, d.doc != nil && d.doc.Custom("Buses", k, "Effects")
		if d.take != nil {
			if g, ok := d.take.Buses[b.Name]; ok {
				gauge = &g
			}
		}
		m.drawLevel(dst, d, s, b.Level, d.editable("Buses", k, "Level"), fmt.Sprint("bus ", k))
	case masterStrip:
		fx = a.Master
		frame = d.frame
		if d.take != nil {
			gauge = &d.take.Master
		}
	}

	for i := range inserts {
		r := m.insert(d, s, i)
		switch {
		case i == inserts-1 && len(fx) > inserts:
			small(r, fmt.Sprintf("+%d more", len(fx)-i), faint, text.AlignStart)
		case i < len(fx):
			name := effectName(fx[i])
			if fxCustom {
				name += " (code)"
			}
			small(r, name, dim, text.AlignStart)
		}
	}

	if gauge != nil && d.sounding {
		mt := m.meter(d, s)
		fill(dst, mt, edge)
		db := 20 * math.Log10(float64(gauge.Peak(frame, 0.05)))
		hgt := session.FaderPosition(db) * float64(mt.Dy())
		c := meterLow
		if db > -1 {
			c = meterHigh
		}
		fillF(dst, float64(mt.Min.X), float64(mt.Max.Y)-hgt, float64(mt.Dx()), hgt, c)
		if lufs := gauge.ShortTerm(frame); !math.IsInf(lufs, -1) {
			small(m.loudness(d, s), fmt.Sprintf("%.1f LUFS", lufs), dim, text.AlignCenter)
		}
	} else {
		fill(dst, m.meter(d, s), edge)
	}
	if s.kind == masterStrip {
		r := m.readout(d, s)
		switch {
		case math.IsInf(d.loudness, -1):
		case a.Target != 0:
			small(r, fmt.Sprintf("%.1f / %g", d.loudness, a.Target), ink, text.AlignCenter)
		default:
			small(r, fmt.Sprintf("%.1f LUFS", d.loudness), ink, text.AlignCenter)
		}
	}
}

func enabled(can bool) color.Color {
	if can {
		return dim
	}
	return faint
}

func box(dst *ebiten.Image, d *desk, b image.Rectangle, label string, on bool, onColor color.Color, can bool) {
	bg, fg := color.Color(button), color.Color(dim)
	if !can {
		bg, fg = panel, faint
	}
	if on {
		bg, fg = onColor, ground
	}
	fill(dst, b, bg)
	drawText(dst, d.fonts.text, label, float64(b.Min.X+b.Dx()/2), float64(b.Min.Y+b.Dy()/2), fg, text.AlignCenter)
}

func (m *mixer) drawSend(dst *ebiten.Image, d *desk, r image.Rectangle, db float64, can bool, id string) {
	frac := (math.Max(-60, math.Min(6, db)) + 60) / 66
	y := float64(r.Max.Y) - d.px(4)
	fillF(dst, float64(r.Min.X), y, float64(r.Dx()), d.px(2), edge)
	c := lit
	if !can {
		c = faint
	}
	fillF(dst, float64(r.Min.X), y, frac*float64(r.Dx()), d.px(2), c)
	m.drawValue(dst, d, image.Rect(r.Min.X, r.Min.Y, r.Max.X, int(y)), id, fmt.Sprintf("%.1f", db), can, text.AlignEnd)
}

func (m *mixer) drawPan(dst *ebiten.Image, d *desk, s strip, pan float64, can bool, id string) {
	r := m.pan(d, s)
	cx, cy := float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2)
	rad := float32(d.px(knobR))
	c := color.Color(dim)
	if !can {
		c = faint
	}
	vector.FillCircle(dst, cx, cy, rad, button, true)
	vector.StrokeCircle(dst, cx, cy, rad, float32(d.px(1)), edge, true)
	angle := pan * 0.75 * math.Pi
	ex, ey := cx+float32(math.Sin(angle))*rad, cy-float32(math.Cos(angle))*rad
	vector.StrokeLine(dst, cx, cy, ex, ey, float32(d.px(2)), c, true)
	m.drawValue(dst, d, m.panReadout(d, s), id, session.FormatPan(pan), can, text.AlignCenter)
}

func (m *mixer) drawLevel(dst *ebiten.Image, d *desk, s strip, db float64, can bool, id string) {
	f := m.fader(d, s)
	for _, mark := range session.FaderMarks() {
		y := m.thumbY(f, mark)
		line(dst, float64(f.Min.X)-d.px(6), y, float64(f.Min.X)-d.px(2), y, 1, faint)
		if mark == 0 || mark == -12 || mark == -40 || mark == 12 {
			drawText(dst, d.fonts.num, fmt.Sprint(mark), float64(f.Min.X)-d.px(8), y, faint, text.AlignEnd)
		}
	}
	fill(dst, f.Inset(int(d.px(2))), edge)
	ty := m.thumbY(f, db)
	c := color.Color(hovered)
	if m.drag != nil && m.drag.id == id {
		c = pressed
	}
	if !can {
		c = button
	}
	fillF(dst, float64(f.Min.X)-d.px(6), ty-d.px(5), float64(f.Dx())+d.px(12), d.px(10), c)
	line(dst, float64(f.Min.X)-d.px(6), ty, float64(f.Max.X)+d.px(6), ty, 1, ink)
	m.drawValue(dst, d, m.readout(d, s), id, fmt.Sprintf("%.1f dB", db), can, text.AlignCenter)
}

// drawValue draws a value's readout, or what's being typed into it.
func (m *mixer) drawValue(dst *ebiten.Image, d *desk, r image.Rectangle, id, value string, can bool, align text.Align) {
	c := color.Color(ink)
	if !can {
		c = faint
	}
	if m.typing != nil && m.typing.id == id {
		fill(dst, r, edge)
		value, c, align = m.typing.text+"▏", playhead, text.AlignCenter
	}
	x := float64(r.Min.X + r.Dx()/2)
	switch align {
	case text.AlignStart:
		x = float64(r.Min.X)
	case text.AlignEnd:
		x = float64(r.Max.X)
	}
	drawText(dst, d.fonts.num, value, x, float64(r.Min.Y+r.Dy()/2), c, align)
}

// effectName is an effect's type, as the mixer lists it.
func effectName(e synth.Effect) string {
	name := fmt.Sprintf("%T", e)
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		name = name[i+1:]
	}
	return name
}
