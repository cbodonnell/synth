package desk

import (
	"fmt"
	"image"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/studio/session"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// grabPx is how near an edge, in points, the cursor grabs it.
const grabPx = 5

// playRef is a play on the timeline: its track's name and its index in the track's Play.
type playRef struct {
	track string
	play  int
}

type laneDragKind int

const (
	movingPlay laneDragKind = iota
	repeatingPlay
	resizingSection
	movingLoop
)

// laneDrag is a drag on the timeline, held as one step to undo.
type laneDrag struct {
	kind  laneDragKind
	track int
	play  int // which a copy changes to the copy's
	from  synth.Play
	bar   int  // grabbed at, from 0
	copy  bool // copies the play on its first move
	// The section being resized, its bars when grabbed, and whether the resize is applied
	// (and undone before the next, so bars cut and grown back come back).
	section, bars int
	applied       bool
	tried         int // the bar, count or length last tried, so a refused one isn't retried every frame
}

// arrange edits the song's arrangement, keeping only an edit that validates, and says
// why not in the status bar.
func (d *desk) arrange(f func(a *synth.Arrangement) error) bool {
	if d.doc == nil {
		d.note = d.readOnly()
		return false
	}
	if err := d.doc.Arrange(f); err != nil {
		first, _, _ := strings.Cut(err.Error(), "\n")
		d.note = "can't: " + first
		return false
	}
	d.changed()
	return true
}

// bar is the bar, from 0 and with its fraction, at x.
func (t *timeline) bar(d *desk, x float64) float64 {
	m := d.layout.Song.Meter
	at := math.Max(0, t.at(d, x))
	b := session.BarAt(m, at)
	from, to := session.BarTime(m, b), session.BarTime(m, b+1)
	return float64(b) + (at-from)/(to-from)
}

// selected is the play last clicked, while it's still there and its clip is open.
func (t *timeline) selected(d *desk) (track, play int, ok bool) {
	a := d.layout.Arrangement
	track = slices.IndexFunc(a.Tracks, func(tr synth.Track) bool { return tr.Name == t.sel.track })
	if track < 0 || t.sel.track != d.view.Track || t.sel.play >= len(a.Tracks[track].Play) ||
		a.Tracks[track].Play[t.sel.play].Clip != d.view.Clip {
		return 0, 0, false
	}
	return track, t.sel.play, true
}

// sectionAt is the ruler's section under x, and whether x is on its end.
func (t *timeline) sectionAt(d *desk, x float64) (int, bool) {
	for i, s := range d.layout.Sections {
		if math.Abs(x-t.x(d, s.To)) <= d.px(grabPx) {
			return i, true
		}
	}
	for i, s := range d.layout.Sections {
		if x >= t.x(d, s.From) && x < t.x(d, s.To) {
			return i, false
		}
	}
	return -1, false
}

// blockAt is the lane's block under x, and whether x is on its end.
func (t *timeline) blockAt(d *desk, lane int, x float64) (int, bool) {
	for i, b := range d.layout.Lanes[lane].Blocks {
		x0, x1 := t.x(d, b.From), t.x(d, b.To)
		if x1-x0 > 3*d.px(grabPx) && math.Abs(x-x1) <= d.px(grabPx) {
			return i, true
		}
		if x >= x0 && x < x1 {
			return i, false
		}
	}
	return -1, false
}

// onLoop reports whether (x, y) grabs the loop brace's start.
func (t *timeline) onLoop(d *desk, x, y float64) bool {
	s := d.layout.Song
	mid := float64(t.r.Min.Y) + d.px(rulerH/2)
	return s.Loop && y >= mid && y < t.top(d) && math.Abs(x-t.x(d, s.LoopStart)) <= d.px(grabPx*1.5)
}

// loopTarget is the section whose start is nearest x, where a dropped loop brace lands.
func (t *timeline) loopTarget(d *desk, x float64) int {
	best, dist := -1, math.Inf(1)
	for i, s := range d.layout.Sections {
		if dx := math.Abs(x - t.x(d, s.From)); dx < dist {
			best, dist = i, dx
		}
	}
	return best
}

func (t *timeline) double(id string) bool {
	now := time.Now()
	dbl := t.last == id && now.Sub(t.lastTime) < 400*time.Millisecond
	t.last, t.lastTime = id, now
	if dbl {
		t.last = ""
	}
	return dbl
}

func (t *timeline) begin(d *desk, g *laneDrag) {
	d.doc.Hold()
	t.drag = g
}

// arranging handles the cursor over the timeline for what arranges the song, reporting
// whether it took the click.
func (t *timeline) arranging(d *desk, x, y float64) bool {
	timed := session.Timed(d.layout.Song)
	left := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	right := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
	pt := image.Pt(int(x), int(y))
	can := d.doc != nil

	if y < t.top(d) {
		if x < t.left(d) || !timed {
			return false
		}
		mid := float64(t.r.Min.Y) + d.px(rulerH/2)
		if t.onLoop(d, x, y) {
			t.hint = "drag the loop's start onto a section to loop from there"
			ebiten.SetCursorShape(ebiten.CursorShapeEWResize)
			if left && can {
				t.begin(d, &laneDrag{kind: movingLoop})
				return true
			}
			return false
		}
		if y >= mid {
			return false
		}
		i, edge := t.sectionAt(d, x)
		if i < 0 {
			return false
		}
		t.hint = "double-click renames · drag its end to resize · right-click for more"
		if !can {
			t.hint = d.readOnly()
		}
		if edge {
			ebiten.SetCursorShape(ebiten.CursorShapeEWResize)
		}
		switch {
		case right && can:
			t.sectionMenu(d, i, pt)
			return true
		case left && can && edge:
			bars := d.layout.Arrangement.Sections[i].Bars
			t.begin(d, &laneDrag{kind: resizingSection, section: i, bars: bars, tried: bars})
			return true
		case left && can && t.double(fmt.Sprint("section ", i)):
			t.rename(d, i)
			return true
		}
		return false
	}

	if x < t.left(d) {
		return false
	}
	for _, r := range t.rows(d) {
		if y < r.y || y >= r.y+r.h || r.lane < 0 {
			continue
		}
		lane := d.layout.Lanes[r.lane]
		track := slices.IndexFunc(d.layout.Arrangement.Tracks, func(tr synth.Track) bool { return tr.Name == lane.Name })
		k, edge := t.blockAt(d, r.lane, x)
		if k < 0 {
			if !timed || d.layout.Custom {
				return false
			}
			t.hint = "double-click adds a clip"
			if left && t.double(fmt.Sprint("lane ", lane.Name)) {
				if !can {
					t.hint = d.readOnly()
					return false
				}
				bar := int(math.Floor(t.bar(d, x))) + 1
				var name string
				if d.arrange(func(a *synth.Arrangement) error { name = session.AddClip(a, track, bar); return nil }) {
					d.openClip(lane.Name, name)
					t.sel = playRef{lane.Name, len(d.layout.Arrangement.Tracks[track].Play) - 1}
				}
				return true
			}
			return false
		}
		if d.layout.Custom {
			return false
		}
		b := lane.Blocks[k]
		t.hint = "drag to move · alt+drag copies · drag its end to repeat · right-click for more"
		if edge {
			ebiten.SetCursorShape(ebiten.CursorShapeEWResize)
			t.hint = "drag to play it more or fewer times over"
		}
		if !can {
			t.hint = d.readOnly()
		}
		if !left && !right {
			return false
		}
		d.openClip(lane.Name, b.Clip)
		t.sel = playRef{lane.Name, b.Play}
		if !can {
			return true
		}
		if right {
			t.playMenu(d, track, b.Play, pt)
			return true
		}
		p := d.layout.Arrangement.Tracks[track].Play[b.Play]
		g := &laneDrag{kind: movingPlay, track: track, play: b.Play, from: p, bar: int(math.Floor(t.bar(d, x))), tried: p.Bar}
		if edge {
			g.kind, g.tried = repeatingPlay, max(p.Times, 1)
		}
		g.copy = ebiten.IsKeyPressed(ebiten.KeyAlt)
		t.begin(d, g)
		return true
	}
	return false
}

// dragging follows a drag on the timeline until the button comes up.
func (t *timeline) dragging(d *desk) {
	g := t.drag
	x, _ := ebiten.CursorPosition()
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || d.doc == nil {
		t.drag = nil
		if d.doc == nil {
			return
		}
		if g.kind == movingLoop {
			if i := t.loopTarget(d, float64(x)); i >= 0 {
				name := d.layout.Arrangement.Sections[i].Name
				d.arrange(func(a *synth.Arrangement) error { a.Loop = name; return nil })
			}
		}
		d.doc.Release()
		return
	}
	fb := t.bar(d, float64(x))
	a := d.layout.Arrangement
	switch g.kind {
	case movingLoop:
		ebiten.SetCursorShape(ebiten.CursorShapeEWResize)
	case movingPlay:
		ebiten.SetCursorShape(ebiten.CursorShapeMove)
		to := g.from.Bar + int(math.Floor(fb)) - g.bar
		if to == g.tried {
			return
		}
		g.tried = to
		if g.copy {
			copied := -1
			if d.arrange(func(a *synth.Arrangement) error { copied = session.CopyPlay(a, g.track, g.play, to); return nil }) {
				g.play, g.copy = copied, false
				t.sel.play = copied
			}
			return
		}
		d.arrange(func(a *synth.Arrangement) error { a.Tracks[g.track].Play[g.play].Bar = to; return nil })
	case repeatingPlay:
		ebiten.SetCursorShape(ebiten.CursorShapeEWResize)
		n := 1
		if i := slices.IndexFunc(a.Tracks[g.track].Clips, func(c synth.Clip) bool { return c.Name == g.from.Clip }); i >= 0 {
			c := a.Tracks[g.track].Clips[i]
			n = max(len(c.Bars), c.Length, 1)
		}
		times := max(1, int(math.Round((fb-float64(g.from.Bar-1))/float64(n))))
		if times == g.tried {
			return
		}
		g.tried = times
		d.arrange(func(a *synth.Arrangement) error { session.SetTimes(a, g.track, g.play, times); return nil })
	case resizingSection:
		ebiten.SetCursorShape(ebiten.CursorShapeEWResize)
		step := session.SectionBars(a)
		start := session.SectionBar(a, g.section) - 1
		bars := max(step, int(math.Round((fb-float64(start))/float64(step)))*step)
		if bars == g.tried {
			return
		}
		g.tried = bars
		if g.applied {
			d.doc.Undo()
		}
		switch {
		case d.arrange(func(a *synth.Arrangement) error { return session.ResizeSection(a, g.section, bars) }):
			g.applied = bars != g.bars
		case g.applied:
			d.doc.Redo()
		}
	}
}

func (t *timeline) playMenu(d *desk, track, play int, at image.Point) {
	tr := d.layout.Arrangement.Tracks[track]
	p := tr.Play[play]
	shared := slices.ContainsFunc(tr.Play, func(q synth.Play) bool { return q != p && q.Clip == p.Clip })
	items := []item{
		{"Repeat after itself   ⌘D", func() { t.repeat(d) }},
	}
	if shared {
		items = append(items, item{"Make unique", func() {
			var name string
			if d.arrange(func(a *synth.Arrangement) error { name = session.MakeUnique(a, track, play); return nil }) {
				d.openClip(tr.Name, name)
			}
		}})
	}
	items = append(items, item{"Delete   ⌫", func() { t.deletePlay(d) }})
	d.openMenu(at, items)
}

func (t *timeline) sectionMenu(d *desk, i int, at image.Point) {
	a := d.layout.Arrangement
	s := a.Sections[i]
	insert := func(at int) func() {
		return func() {
			d.arrange(func(a *synth.Arrangement) error { session.InsertSection(a, at, session.NewSectionBars(*a)); return nil })
		}
	}
	items := []item{
		{"Rename", func() { t.rename(d, i) }},
		{"Insert a section before", insert(i)},
		{"Insert a section after", insert(i + 1)},
		{"Duplicate", func() {
			d.arrange(func(a *synth.Arrangement) error { session.DuplicateSection(a, i); return nil })
		}},
		{"Delete", func() {
			d.arrange(func(a *synth.Arrangement) error { return session.DeleteSection(a, i) })
		}},
	}
	if a.Loop != s.Name {
		items = append(items, item{"Loop from here", func() {
			d.arrange(func(a *synth.Arrangement) error { a.Loop = s.Name; return nil })
		}})
	}
	if a.Loop != "" {
		items = append(items, item{"Play once, without a loop", func() {
			d.arrange(func(a *synth.Arrangement) error { a.Loop = ""; return nil })
		}})
	}
	d.openMenu(at, items)
}

// rename types a new name into section i.
func (t *timeline) rename(d *desk, i int) {
	d.typing = &typing{id: fmt.Sprint("section ", i), text: d.layout.Arrangement.Sections[i].Name, chars: "*", set: func(s string) bool {
		d.arrange(func(a *synth.Arrangement) error { return session.RenameSection(a, i, s) })
		return true
	}}
}

// repeat plays the selected play again right after itself.
func (t *timeline) repeat(d *desk) {
	track, play, ok := t.selected(d)
	if !ok {
		return
	}
	copied := -1
	if d.arrange(func(a *synth.Arrangement) error { copied = session.RepeatPlay(a, track, play); return nil }) {
		t.sel.play = copied
	}
}

func (t *timeline) deletePlay(d *desk) {
	track, play, ok := t.selected(d)
	if !ok {
		return
	}
	if d.arrange(func(a *synth.Arrangement) error { session.DeletePlay(a, track, play); return nil }) {
		t.sel = playRef{}
		if _, ok := d.target(); !ok {
			d.openClip(d.view.Track, "")
		}
	}
}

// editTempo types a tempo into the transport.
func (d *desk) editTempo() {
	if !d.typable("Tempo") {
		return
	}
	d.typing = &typing{id: "tempo", text: fmt.Sprintf("%g", d.layout.Tempo), chars: "0123456789.", set: func(s string) bool {
		v, ok := session.ParseTempo(s)
		if ok {
			d.arrange(func(a *synth.Arrangement) error { a.Tempo = v; return nil })
		}
		return ok
	}}
}

// editMeter types a meter into the transport: 3/4, or 3+3+3+4.
func (d *desk) editMeter() {
	if !d.typable("Meter") {
		return
	}
	d.typing = &typing{id: "meter", text: session.FormatMeter(d.layout.Meter), chars: "0123456789+/", set: func(s string) bool {
		bars, ok := session.ParseMeter(s)
		if ok {
			d.arrange(func(a *synth.Arrangement) error { a.Meter = bars; return nil })
		}
		return ok
	}}
}

// typable reports whether a field of the song can be typed into, saying why not.
func (d *desk) typable(field string) bool {
	switch {
	case !session.Timed(d.layout.Song):
		return false
	case d.doc == nil:
		d.note = d.readOnly()
		return false
	case !d.editable(field):
		d.note = custom
		return false
	}
	return true
}

// typeKeys types into the transport or the ruler, reporting whether it took the keys.
func (d *desk) typeKeys() bool {
	if d.typing == nil {
		return false
	}
	if typeInto(d.typing) {
		d.typing = nil
	}
	return true
}

// typed is what's being typed into the field with the id, with a caret.
func (d *desk) typed(id string) (string, bool) {
	if d.typing == nil || d.typing.id != id {
		return "", false
	}
	return d.typing.text + "▏", true
}
