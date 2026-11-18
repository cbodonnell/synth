package desk

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/studio/session"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// notesEdit is the clip editor on a clip of Notes: a piano roll without a grid, its notes
// at free times in the clip's own time (beats, or seconds in a sound effect), snapped to
// the step chosen.
type notesEdit struct {
	picked []int // the selected notes, by index
	snap   int   // into session.Snaps
	length float64
	drag   *notesDrag
}

// notesDrag is a drag in the roll, from the notes as they were when it began, over the
// span as it was then, so the roll doesn't rescale under the cursor.
type notesDrag struct {
	kind  dragKind
	from  []synth.Note
	span  float64
	note  int
	at    float64
	key   int
	moved int // the key last auditioned
}

func (d *desk) sound() bool { return !session.Timed(d.layout.Song) }

func (c *clipEditor) snapStep(d *desk) float64 { return session.Snaps(d.sound())[c.nt.snap] }

// nudge is how far an arrow moves a note with snapping off.
func (c *clipEditor) nudge(d *desk) float64 {
	if d.sound() {
		return 0.005
	}
	return 1.0 / 32
}

// span is the time the roll shows, from the clip's start.
func (c *clipEditor) span(d *desk, t clipTarget) float64 {
	if g := c.nt.drag; g != nil {
		return g.span
	}
	return session.NotesSpan(d.layout.Arrangement, t.c, c.firstBar(d, t))
}

func (c *clipEditor) firstBar(d *desk, t clipTarget) int {
	for _, p := range d.layout.Arrangement.Tracks[t.track].Play {
		if p.Clip == t.c.Name {
			return p.Bar - 1
		}
	}
	return 0
}

func (c *clipEditor) nx(d *desk, t clipTarget, at float64) float64 {
	g := c.grid(d)
	return float64(g.Min.X) + at/c.span(d, t)*float64(g.Dx())
}

func (c *clipEditor) nat(d *desk, t clipTarget, x float64) float64 {
	g := c.grid(d)
	return (x - float64(g.Min.X)) / float64(g.Dx()) * c.span(d, t)
}

func (c *clipEditor) canWriteNotes(d *desk, t clipTarget) (bool, string) {
	switch {
	case d.doc == nil:
		return false, d.readOnly()
	case !d.editable("Tracks", t.track, "Clips", t.clip, "Notes"):
		return false, custom
	}
	return true, ""
}

func (c *clipEditor) writeNotes(d *desk, t clipTarget, notes []synth.Note) {
	d.edit(func(a *synth.Arrangement) { a.Tracks[t.track].Clips[t.clip].Notes = notes })
}

// noteUnder is the note under (x, y), and whether (x, y) is on its end.
func (c *clipEditor) noteUnder(d *desk, t clipTarget, x, y float64) (int, bool) {
	key := c.keyAt(d, y)
	for i := len(t.c.Notes) - 1; i >= 0; i-- {
		n := t.c.Notes[i]
		x0, x1 := c.nx(d, t, n.At), math.Max(c.nx(d, t, n.At+n.Dur), c.nx(d, t, n.At)+d.px(4))
		if int(n.Key) == key && x >= x0 && x < x1+d.px(2) {
			return i, x >= x1-math.Min(d.px(6), (x1-x0)/3)
		}
	}
	return -1, false
}

func noteVel(n synth.Note) float64 {
	if n.Vel <= 0 {
		return 1
	}
	return math.Min(n.Vel, 1)
}

func (c *clipEditor) updateNotes(d *desk, t clipTarget) {
	c.nt.picked = slices.DeleteFunc(c.nt.picked, func(i int) bool { return i >= len(t.c.Notes) })
	if c.drag != nil {
		c.dragging(d, t)
		return
	}
	if c.nt.drag != nil {
		c.draggingNotes(d, t)
		return
	}
	pt := image.Pt(ebiten.CursorPosition())
	if !pt.In(c.r) {
		return
	}
	if pt.In(c.settings(d)) {
		c.updateSettings(d, t, pt)
		return
	}
	if _, wy := ebiten.Wheel(); wy != 0 {
		c.top += wy * 2
		rows := float64(c.grid(d).Dy()) / d.px(keyRowH)
		c.top = math.Max(rows-1, math.Min(c.top, 127))
	}
	can, why := c.canWriteNotes(d, t)
	c.hint = "click draws · drag moves · drag its end stretches · alt+drag slides · shift+click picks more · drag the lane for velocity · arrows move · snap in the settings"
	if !can {
		c.hint = why
		return
	}
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	c.typing = nil
	x, y := float64(pt.X), float64(pt.Y)
	if pt.In(c.lane(d)) {
		d.doc.Hold()
		c.nt.drag = &notesDrag{kind: velocities, span: c.span(d, t)}
		c.draggingNotes(d, t)
		return
	}
	if !pt.In(c.grid(d)) {
		return
	}
	span, step := c.span(d, t), c.snapStep(d)
	if i, end := c.noteUnder(d, t, x, y); i >= 0 {
		n := t.c.Notes[i]
		g := &notesDrag{from: t.c.Notes, span: span, note: i, at: c.nat(d, t, x), key: c.keyAt(d, y), moved: int(n.Key)}
		switch {
		case ebiten.IsKeyPressed(ebiten.KeyAlt):
			g.kind = sliding
		case end:
			g.kind = stretching
			c.nt.picked = []int{i}
		case ebiten.IsKeyPressed(ebiten.KeyShift):
			if k := slices.Index(c.nt.picked, i); k >= 0 {
				c.nt.picked = slices.Delete(c.nt.picked, k, k+1)
			} else {
				c.nt.picked = append(c.nt.picked, i)
			}
			return
		default:
			g.kind = moving
			if !slices.Contains(c.nt.picked, i) {
				c.nt.picked = []int{i}
			}
			d.audition(t, []int{int(n.Key)}, noteVel(n))
		}
		d.doc.Hold()
		c.nt.drag = g
		return
	}
	at := session.Snap(c.nat(d, t, x), 0)
	if step > 0 {
		at = math.Floor(c.nat(d, t, x)/step) * step
	}
	length := c.nt.length
	if length <= 0 {
		length = math.Max(step, c.nudge(d)*8)
	}
	key := c.keyAt(d, y)
	notes := session.DrawNote(t.c.Notes, synth.Note{At: math.Max(at, 0), Dur: length, Key: float64(key)})
	d.doc.Hold()
	c.writeNotes(d, t, notes)
	i := len(notes) - 1
	c.nt.picked = []int{i}
	c.nt.drag = &notesDrag{kind: stretching, from: notes, span: span, note: i}
	d.audition(t, []int{key}, 1)
}

func (c *clipEditor) draggingNotes(d *desk, t clipTarget) {
	g := c.nt.drag
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || d.doc == nil {
		c.nt.drag = nil
		if d.doc != nil {
			d.doc.Release()
		}
		return
	}
	pt := image.Pt(ebiten.CursorPosition())
	x, y := float64(pt.X), float64(pt.Y)
	step := c.snapStep(d)
	at := c.nat(d, t, x)
	switch g.kind {
	case moving:
		n := g.from[g.note]
		by := session.Snap(n.At+at-g.at, step) - n.At
		key := c.keyAt(d, y)
		c.writeNotes(d, t, session.MoveNotes(g.from, c.nt.picked, by, float64(key-g.key), g.span))
		if key != g.moved {
			g.moved = key
			d.audition(t, []int{int(n.Key) + key - g.key}, noteVel(n))
		}
	case stretching:
		n := g.from[g.note]
		end := session.Snap(at, step)
		if step > 0 {
			end = math.Ceil(at/step-1e-9) * step
		}
		notes := session.StretchNote(g.from, g.note, end-n.At, step)
		c.nt.length = notes[g.note].Dur
		c.writeNotes(d, t, notes)
	case sliding:
		notes := slices.Clone(g.from)
		notes[g.note].Slide = float64(c.keyAt(d, y)) - notes[g.note].Key
		c.writeNotes(d, t, notes)
	case velocities:
		best, dist := -1, math.Inf(1)
		for i, n := range t.c.Notes {
			if dx := math.Abs(c.nx(d, t, n.At) - x); dx < dist && dx < d.px(12) {
				best, dist = i, dx
			}
		}
		if best >= 0 {
			l := c.lane(d)
			v := session.Round(math.Max(0.01, math.Min(1, 1-(y-float64(l.Min.Y))/float64(l.Dy()))), 0.01)
			notes := slices.Clone(t.c.Notes)
			notes[best].Vel = v
			c.writeNotes(d, t, notes)
		}
	}
}

func (c *clipEditor) notesKeys(d *desk, t clipTarget) bool {
	command := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	pressed := inpututil.IsKeyJustPressed
	if !command && pressed(ebiten.KeyEscape) {
		if len(c.nt.picked) > 0 {
			c.nt.picked = nil
		} else {
			d.openClip(d.view.Track, "")
		}
		return true
	}
	can, _ := c.canWriteNotes(d, t)
	if !can || command || len(c.nt.picked) == 0 {
		return false
	}
	octave := 1.0
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		octave = 12
	}
	nudge := c.snapStep(d)
	if nudge == 0 {
		nudge = c.nudge(d)
	}
	move := func(by, semis float64) {
		c.writeNotes(d, t, session.MoveNotes(t.c.Notes, c.nt.picked, by, semis, c.span(d, t)))
	}
	switch {
	case pressed(ebiten.KeyUp):
		move(0, octave)
	case pressed(ebiten.KeyDown):
		move(0, -octave)
	case pressed(ebiten.KeyLeft):
		move(-nudge, 0)
	case pressed(ebiten.KeyRight):
		move(nudge, 0)
	case pressed(ebiten.KeyDelete) || pressed(ebiten.KeyBackspace):
		c.writeNotes(d, t, session.DeleteNotes(t.c.Notes, c.nt.picked))
		c.nt.picked = nil
	default:
		return false
	}
	return true
}

func (c *clipEditor) notesRows(d *desk, t clipTarget) []clipSetting {
	var buttons []clipButton
	for i, s := range session.Snaps(d.sound()) {
		buttons = append(buttons, clipButton{snapLabel(s, d.sound()), func() { c.nt.snap = i }, true, c.nt.snap == i})
	}
	return []clipSetting{{label: "snap", field: "Notes", buttons: buttons}}
}

func snapLabel(step float64, sound bool) string {
	switch {
	case step == 0:
		return "off"
	case sound:
		return fmt.Sprintf("%gms", step*1000)
	}
	return beats(step)
}

func (c *clipEditor) drawNotes(dst *ebiten.Image, d *desk, t clipTarget) {
	g := c.grid(d)
	roll := dst.SubImage(g).(*ebiten.Image)
	c.drawKeys(dst, d)
	c.drawTimes(roll, d, t, float64(g.Min.Y), float64(g.Max.Y))

	tr := d.layout.Arrangement.Tracks[t.track]
	rowH := d.px(keyRowH)
	for i, n := range t.c.Notes {
		x0 := c.nx(d, t, n.At)
		x1 := math.Max(c.nx(d, t, n.At+n.Dur), x0+d.px(4))
		y := c.y(d, int(n.Key))
		body := hue(tr.Name, 0.45+0.45*noteVel(n))
		if slices.Contains(c.nt.picked, i) {
			body = color.RGBA{0xff, 0xc0, 0x70, 0xff}
		}
		fillF(roll, x0+1, y+1, x1-x0-2, rowH-2, body)
		if n.Slide != 0 {
			ty := c.y(d, int(math.Round(n.Key+n.Slide)))
			line(roll, x0+d.px(3), y+rowH/2, x1-1, ty+rowH/2, d.px(1.5), ink)
			fillF(roll, x1-d.px(4), ty+rowH/4, d.px(4), rowH/2, ink)
		}
	}

	l := c.lane(d)
	fill(dst, l, panel)
	lane := dst.SubImage(l).(*ebiten.Image)
	drawText(dst, d.fonts.text, "velocity", float64(c.r.Min.X)+d.px(6), float64(l.Min.Y+l.Dy()/2), faint, text.AlignStart)
	for _, n := range t.c.Notes {
		v := noteVel(n)
		h := v * (float64(l.Dy()) - d.px(6))
		fillF(lane, c.nx(d, t, n.At)+1, float64(l.Max.Y)-h, d.px(4), h, hue(tr.Name, 0.45+0.45*v))
	}
}

// drawTimes rules the roll: milliseconds in a sound, bars and beats in a song, and the
// playhead.
func (c *clipEditor) drawTimes(dst *ebiten.Image, d *desk, t clipTarget, top, bottom float64) {
	span := c.span(d, t)
	g := c.grid(d)
	perPx := span / float64(g.Dx())
	if d.sound() {
		tick := 0.01
		for _, s := range []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1} {
			if tick = s; s/perPx >= d.px(50) {
				break
			}
		}
		for k := 0; float64(k)*tick < span; k++ {
			x := c.nx(d, t, float64(k)*tick)
			line(dst, x, top, x, bottom, 1, barLine)
			drawText(dst, d.fonts.num, fmt.Sprintf("%.0fms", float64(k)*tick*1000), x+d.px(3), top+d.px(7), faint, text.AlignStart)
		}
		if step := c.snapStep(d); step > 0 && step/perPx >= d.px(6) {
			for k := 0; float64(k)*step < span; k++ {
				x := c.nx(d, t, float64(k)*step)
				line(dst, x, top, x, bottom, 1, beatLine)
			}
		}
	} else {
		m := synth.Meter{Beat: 1, Bars: d.layout.Meter}
		first := c.firstBar(d, t)
		start := m.BarStart(first)
		for b := first; m.BarStart(b)-start < span-1e-9; b++ {
			x := c.nx(d, t, m.BarStart(b)-start)
			line(dst, x, top, x, bottom, d.px(1.5), barLine)
			drawText(dst, d.fonts.num, fmt.Sprint(b-first+1), x+d.px(3), top+d.px(7), faint, text.AlignStart)
			for k := 1; k < m.BarBeats(b); k++ {
				bx := c.nx(d, t, m.BarStart(b)-start+float64(k))
				line(dst, bx, top, bx, bottom, 1, beatLine)
			}
		}
	}
	if at, ok := c.notesPlaying(d, t); ok && at < span {
		x := c.nx(d, t, at)
		line(dst, x, top, x, bottom, d.px(1.5), playhead)
	}
}

// notesPlaying is where the playhead is in the clip's time, when it's in a play of it.
func (c *clipEditor) notesPlaying(d *desk, t clipTarget) (float64, bool) {
	if !d.sounding {
		return 0, false
	}
	if d.sound() {
		return d.view.At, true
	}
	for _, l := range d.layout.Lanes {
		if l.Name != d.view.Track {
			continue
		}
		for _, b := range l.Blocks {
			if b.Clip == t.c.Name && d.view.At >= b.From && d.view.At < b.To {
				return math.Mod((d.view.At-b.From)*d.layout.Tempo/60, c.span(d, t)), true
			}
		}
	}
	return 0, false
}
