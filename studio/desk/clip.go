package desk

import (
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/studio/session"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Sizes in points, before the window's scale.
const (
	detailH   = 280
	clipHeadH = 24
	settingsW = 220
	keysW     = 48
	velocityH = 56
	keyRowH   = 11
	settingH  = 22
	minStepW  = 6
	drumRowH  = 40
)

// clipEditor is the detail under the arrangement: the selected clip in a piano roll, or
// a Hit clip in a drum grid, over a velocity lane, beside the clip's settings.
type clipEditor struct {
	r image.Rectangle

	top    float64 // the key at the piano roll's top edge
	scroll float64 // steps scrolled off the left
	sel    []session.Ref
	length int    // steps a drawn note holds
	shown  string // the clip last shown, to find its notes when another is opened

	drag   *clipDrag
	typing *typing
	step   *stepInput
	hint   string

	last     string // the setting last clicked, and when, for a double-click
	lastTime time.Time
}

type dragKind int

const (
	moving dragKind = iota
	stretching
	sliding
	velocities
	setting
)

// clipDrag is a drag in the editor. Notes are dragged from the grid as it was when the
// drag began, so passing over another note doesn't cut it for good.
type clipDrag struct {
	kind dragKind
	from session.Grid
	at   session.Ref // the step and key it was grabbed at
	refs []session.Ref
	key  int // the key last auditioned

	value *drag // a setting's
	moved bool
	typed *typing // what a click on the setting without a drag types into
}

// stepInput writes notes from the keyboard at a cursor that steps on.
type stepInput struct {
	cursor int
	octave int
	accent int
	chord  int // the step the chord being held was struck on
	held   []ebiten.Key
}

// The keys musical typing plays, and the semitones above the octave's C they play.
var typingKeys = []struct {
	key   ebiten.Key
	semis int
}{
	{ebiten.KeyA, 0}, {ebiten.KeyW, 1}, {ebiten.KeyS, 2}, {ebiten.KeyE, 3}, {ebiten.KeyD, 4},
	{ebiten.KeyF, 5}, {ebiten.KeyT, 6}, {ebiten.KeyG, 7}, {ebiten.KeyY, 8}, {ebiten.KeyH, 9},
	{ebiten.KeyU, 10}, {ebiten.KeyJ, 11}, {ebiten.KeyK, 12},
}

// clipTarget is the clip open in the editor.
type clipTarget struct {
	track, clip int
	c           synth.Clip
	g           session.Grid
	err         error // why it has no grid
}

// target is the selected track's open clip, if there is one.
func (d *desk) target() (clipTarget, bool) {
	if d.view.Clip == "" {
		return clipTarget{}, false
	}
	a := d.layout.Arrangement
	ti := slices.IndexFunc(a.Tracks, func(t synth.Track) bool { return t.Name == d.view.Track })
	if ti < 0 {
		return clipTarget{}, false
	}
	ci := slices.IndexFunc(a.Tracks[ti].Clips, func(c synth.Clip) bool { return c.Name == d.view.Clip })
	if ci < 0 {
		return clipTarget{}, false
	}
	t := clipTarget{track: ti, clip: ci, c: a.Tracks[ti].Clips[ci]}
	t.g, t.err = session.ReadGrid(t.c)
	return t, true
}

// openClip selects the track and opens its clip in the editor; "" closes it.
func (d *desk) openClip(track, clip string) {
	d.view.Track, d.view.Clip = track, clip
	d.ce.sel, d.ce.step, d.ce.typing = nil, nil, nil
}

// detailOpen reports whether the window has room for the clip editor.
func (d *desk) detailOpen() bool {
	_, ok := d.target()
	return ok
}

func (d *desk) musicalScale() session.Scale {
	s, _ := session.ParseScale(d.layout.Key)
	return s
}

// canWrite reports whether the target's bars can be written, or why not.
func (c *clipEditor) canWrite(d *desk, t clipTarget) (bool, string) {
	switch {
	case d.doc == nil:
		return false, d.readOnly()
	case t.err != nil:
		return false, "notes off the grid: edit them in code"
	case !d.editable("Tracks", t.track, "Clips", t.clip, "Bars"):
		return false, custom
	}
	for k := range t.c.Bars {
		if !d.editable("Tracks", t.track, "Clips", t.clip, "Bars", k) {
			return false, custom
		}
	}
	return true, ""
}

func (c *clipEditor) write(d *desk, t clipTarget, g session.Grid) {
	bars := g.Write(d.musicalScale())
	d.edit(func(a *synth.Arrangement) { a.Tracks[t.track].Clips[t.clip].Bars = bars })
}

// Areas of the editor, in window pixels.
func (c *clipEditor) head(d *desk) image.Rectangle {
	return image.Rect(c.r.Min.X, c.r.Min.Y, c.r.Max.X, c.r.Min.Y+int(d.px(clipHeadH)))
}

func (c *clipEditor) settings(d *desk) image.Rectangle {
	return image.Rect(c.r.Max.X-int(d.px(settingsW)), c.head(d).Max.Y, c.r.Max.X, c.r.Max.Y)
}

func (c *clipEditor) keyboard(d *desk) image.Rectangle {
	return image.Rect(c.r.Min.X, c.head(d).Max.Y, c.r.Min.X+int(d.px(keysW)), c.r.Max.Y-int(d.px(velocityH)))
}

func (c *clipEditor) grid(d *desk) image.Rectangle {
	k := c.keyboard(d)
	return image.Rect(k.Max.X, k.Min.Y, c.settings(d).Min.X, k.Max.Y)
}

func (c *clipEditor) lane(d *desk) image.Rectangle {
	g := c.grid(d)
	return image.Rect(g.Min.X, g.Max.Y, g.Max.X, c.r.Max.Y)
}

func (c *clipEditor) drumRow(d *desk) image.Rectangle {
	g := c.grid(d)
	return image.Rect(g.Min.X, g.Min.Y+int(d.px(8)), g.Max.X, g.Min.Y+int(d.px(8+drumRowH)))
}

func (c *clipEditor) stepW(d *desk, t clipTarget) float64 {
	return math.Max(float64(c.grid(d).Dx())/float64(max(t.g.Steps(), 1)), d.px(minStepW))
}

func (c *clipEditor) x(d *desk, t clipTarget, step float64) float64 {
	return float64(c.grid(d).Min.X) + (step-c.scroll)*c.stepW(d, t)
}

func (c *clipEditor) stepAt(d *desk, t clipTarget, x float64) int {
	s := int(math.Floor(c.scroll + (x-float64(c.grid(d).Min.X))/c.stepW(d, t)))
	return max(0, min(s, t.g.Steps()-1))
}

func (c *clipEditor) y(d *desk, key int) float64 {
	return float64(c.grid(d).Min.Y) + (c.top-float64(key))*d.px(keyRowH)
}

func (c *clipEditor) keyAt(d *desk, y float64) int {
	k := int(math.Ceil(c.top - (y-float64(c.grid(d).Min.Y))/d.px(keyRowH)))
	return max(0, min(k, 127))
}

// show scrolls a clip into view the first time it's shown.
func (c *clipEditor) show(d *desk, t clipTarget) {
	id := d.view.Track + "\x00" + d.view.Clip
	if c.shown == id {
		return
	}
	c.shown, c.scroll, c.length = id, 0, 1
	lo, hi := 128, -1
	for _, e := range t.g.Events {
		for _, p := range e.Keys {
			lo, hi = min(lo, p.Key, p.Key+p.Slide), max(hi, p.Key, p.Key+p.Slide)
		}
	}
	mid := 60.0
	if hi >= 0 {
		mid = float64(lo+hi) / 2
	}
	rows := float64(c.grid(d).Dy()) / d.px(keyRowH)
	c.top = math.Round(mid + rows/2)
	c.clampScroll(d, t)
}

func (c *clipEditor) clampScroll(d *desk, t clipTarget) {
	rows := float64(c.grid(d).Dy()) / d.px(keyRowH)
	c.top = math.Max(rows-1, math.Min(c.top, 127))
	visible := float64(c.grid(d).Dx()) / c.stepW(d, t)
	c.scroll = math.Max(0, math.Min(c.scroll, float64(t.g.Steps())-visible))
}

// noteAt is the note sounding at step on key.
func noteAt(g session.Grid, step, key int) (session.Ref, bool) {
	i := g.At(step)
	if i < 0 {
		return session.Ref{}, false
	}
	e := g.Events[i]
	if slices.ContainsFunc(e.Keys, func(p session.Pitch) bool { return p.Key == key }) {
		return session.Ref{Step: e.Step, Key: key}, true
	}
	return session.Ref{}, false
}

func (c *clipEditor) update(d *desk) {
	c.hint = ""
	t, ok := d.target()
	if !ok {
		c.drag = nil
		return
	}
	c.show(d, t)
	if t.err == nil {
		c.sel = slices.DeleteFunc(c.sel, func(r session.Ref) bool { return !t.g.Has(r) })
	}
	if c.step != nil && t.err == nil {
		c.step.cursor = min(c.step.cursor, t.g.Steps()-1)
	}
	if c.drag != nil {
		c.dragging(d, t)
		return
	}
	pt := image.Pt(ebiten.CursorPosition())
	if !pt.In(c.r) {
		return
	}
	can, why := c.canWrite(d, t)
	if pt.In(c.settings(d)) {
		c.updateSettings(d, t, pt)
		return
	}
	if t.err != nil {
		c.hint = why
		return
	}
	if wx, wy := ebiten.Wheel(); wx != 0 || wy != 0 {
		if ebiten.IsKeyPressed(ebiten.KeyShift) || wx != 0 {
			dx := wx
			if dx == 0 {
				dx = wy
			}
			c.scroll -= dx * 2
		} else if !t.g.Hit {
			c.top += wy * 2
		}
		c.clampScroll(d, t)
	}
	switch {
	case t.g.Hit:
		c.hint = "click a cell to cycle it . x X o · alt+click ratchets it · drag the lane for accents · cmd+E step input"
	default:
		c.hint = "click draws · drag moves · drag its end stretches · alt+drag slides · shift+click a triad, alt+click a seventh · arrows move · [ ] ratchet · cmd+E step input"
	}
	if !can {
		c.hint = why
	}
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	c.typing = nil
	if pt.In(c.lane(d)) && can {
		d.doc.Hold()
		c.drag = &clipDrag{kind: velocities}
		c.dragging(d, t)
		return
	}
	if !pt.In(c.grid(d)) || !can {
		return
	}
	step := c.stepAt(d, t, float64(pt.X))
	alt := ebiten.IsKeyPressed(ebiten.KeyAlt)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	if t.g.Hit {
		if !pt.In(c.drumRow(d)) {
			return
		}
		g := t.g.Cycle(step)
		if alt {
			g = t.g
			if i := g.Start(step); i >= 0 {
				hits := []int{2, 3, 4, 6, 1}
				g = g.Ratchet(step, hits[(slices.Index(hits, max(g.Events[i].Hits, 1))+1)%len(hits)])
			}
		}
		c.write(d, t, g)
		if i := g.Start(step); i >= 0 {
			d.audition(t, nil, g.Events[i].Velocity(t.c.Vel))
		}
		return
	}

	key := c.keyAt(d, float64(pt.Y))
	if ref, ok := noteAt(t.g, step, key); ok {
		e := t.g.Events[t.g.Start(ref.Step)]
		end := c.x(d, t, float64(e.Step+e.Len))
		switch {
		case alt:
			c.begin(d, &clipDrag{kind: sliding, from: t.g, at: ref, key: key})
		case float64(pt.X) >= end-math.Min(d.px(6), c.stepW(d, t)/3):
			c.sel = []session.Ref{ref}
			c.begin(d, &clipDrag{kind: stretching, from: t.g, at: ref})
		case shift:
			if i := slices.Index(c.sel, ref); i >= 0 {
				c.sel = slices.Delete(c.sel, i, i+1)
			} else {
				c.sel = append(c.sel, ref)
			}
		default:
			if !slices.Contains(c.sel, ref) {
				c.sel = []session.Ref{ref}
			}
			c.begin(d, &clipDrag{kind: moving, from: t.g, at: session.Ref{Step: step, Key: key}, refs: slices.Clone(c.sel), key: key})
			d.audition(t, []int{key}, e.Velocity(t.c.Vel))
		}
		return
	}

	keys := []int{key}
	switch {
	case shift:
		keys = d.musicalScale().Chord(key, 3)
	case alt:
		keys = d.musicalScale().Chord(key, 4)
	}
	g := t.g.Draw(step, c.length, keys...)
	c.sel = nil
	for _, k := range keys {
		c.sel = append(c.sel, session.Ref{Step: step, Key: k})
	}
	d.doc.Hold()
	c.write(d, t, g)
	c.drag = &clipDrag{kind: stretching, from: g, at: session.Ref{Step: step, Key: key}}
	d.audition(t, keys, g.Events[g.Start(step)].Velocity(t.c.Vel))
}

func (c *clipEditor) begin(d *desk, g *clipDrag) {
	d.doc.Hold()
	c.drag = g
}

// dragging follows the cursor with the drag under way, and ends it when the button lifts.
func (c *clipEditor) dragging(d *desk, t clipTarget) {
	g := c.drag
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || d.doc == nil {
		if g.kind == setting && !g.moved && g.typed != nil {
			c.typing = g.typed
		}
		c.drag = nil
		if d.doc != nil {
			d.doc.Release()
		}
		return
	}
	pt := image.Pt(ebiten.CursorPosition())
	x, y := float64(pt.X), float64(pt.Y)
	if g.kind == setting {
		v := g.value
		if math.Abs(y-v.y) > d.px(3) {
			g.moved = true
		}
		if !g.moved {
			return
		}
		dy := (v.y - y) / v.travel
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			dy /= 10
		}
		v.set(session.Round(math.Max(v.lo, math.Min(v.hi, v.from+dy*(v.hi-v.lo))), v.step))
		return
	}
	if t.err != nil {
		return
	}
	switch g.kind {
	case moving:
		key := c.keyAt(d, y)
		moved, refs := g.from.Move(g.refs, c.stepAt(d, t, x)-g.at.Step, key-g.at.Key)
		c.sel = refs
		c.write(d, t, moved)
		if key != g.key {
			g.key = key
			if len(refs) > 0 {
				if i := moved.Start(refs[0].Step); i >= 0 {
					d.audition(t, []int{refs[0].Key}, moved.Events[i].Velocity(t.c.Vel))
				}
			}
		}
	case stretching:
		n := c.stepAt(d, t, x) - g.at.Step + 1
		c.length = max(n, 1)
		c.write(d, t, g.from.Stretch(g.at.Step, n))
	case sliding:
		c.write(d, t, g.from.Slide(g.at, c.keyAt(d, y)))
	case velocities:
		if i := t.g.At(c.stepAt(d, t, x)); i >= 0 {
			l := c.lane(d)
			level := 1 - int(3*(y-float64(l.Min.Y))/float64(l.Dy()))
			c.write(d, t, t.g.Accent(t.g.Events[i].Step, max(-1, min(1, level))))
		}
	}
}

// keys handles the editor's keys, reporting whether it took them.
func (c *clipEditor) keys(d *desk) bool {
	if c.typing != nil {
		if typeInto(c.typing) {
			c.typing = nil
		}
		return true
	}
	t, ok := d.target()
	if !ok {
		c.step = nil
		return false
	}
	command := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	pressed := inpututil.IsKeyJustPressed
	can, _ := c.canWrite(d, t)
	if c.step == nil && !command && pressed(ebiten.KeyEscape) {
		if len(c.sel) > 0 {
			c.sel = nil
		} else {
			d.openClip(d.view.Track, "")
		}
		return true
	}
	if command && pressed(ebiten.KeyE) {
		if c.step != nil || !can {
			c.step = nil
			return true
		}
		c.step = &stepInput{octave: 4, chord: -1}
		if len(c.sel) > 0 {
			c.step.cursor = c.sel[0].Step
		}
		return true
	}
	if c.step != nil {
		if !can {
			c.step = nil
			return false
		}
		if command {
			return false
		}
		c.stepKeys(d, t)
		return true
	}
	if !can || command {
		return false
	}
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	octave := 1
	if shift {
		octave = 12
	}
	move := func(steps, semis int) {
		g, refs := t.g.Move(c.sel, steps, semis)
		c.sel = refs
		c.write(d, t, g)
	}
	switch {
	case len(c.sel) == 0:
		return false
	case pressed(ebiten.KeyUp):
		move(0, octave)
	case pressed(ebiten.KeyDown):
		move(0, -octave)
	case pressed(ebiten.KeyLeft):
		move(-1, 0)
	case pressed(ebiten.KeyRight):
		move(1, 0)
	case pressed(ebiten.KeyDelete) || pressed(ebiten.KeyBackspace):
		c.write(d, t, t.g.Delete(c.sel...))
		c.sel = nil
	case pressed(ebiten.KeyBracketLeft) || pressed(ebiten.KeyBracketRight):
		by := 1
		if pressed(ebiten.KeyBracketLeft) {
			by = -1
		}
		g := t.g
		for _, r := range c.sel {
			if i := g.Start(r.Step); i >= 0 {
				g = g.Ratchet(r.Step, max(g.Events[i].Hits, 1)+by)
			}
		}
		c.write(d, t, g)
	default:
		return false
	}
	return true
}

// stepKeys writes this frame's keys at the cursor.
func (c *clipEditor) stepKeys(d *desk, t clipTarget) {
	s := c.step
	pressed := inpututil.IsKeyJustPressed
	n := t.g.Steps()
	if n == 0 {
		return
	}
	g := t.g
	var played []int
	var held []ebiten.Key
	for _, k := range typingKeys {
		if !ebiten.IsKeyPressed(k.key) {
			continue
		}
		held = append(held, k.key)
		if !pressed(k.key) {
			continue
		}
		key := (s.octave+1)*12 + k.semis
		var keys []int
		if !g.Hit {
			keys = []int{key}
		}
		if len(s.held) > 0 && s.chord >= 0 {
			g = g.Draw(s.chord, 1, keys...)
		} else {
			s.chord = s.cursor
			g = g.Rest(s.cursor).Draw(s.cursor, 1, keys...).Accent(s.cursor, s.accent)
			s.cursor = (s.cursor + 1) % n
		}
		s.held = append(s.held, k.key)
		played = append(played, keys...)
	}
	s.held = held
	if len(held) == 0 {
		s.chord = -1
	}
	switch {
	case pressed(ebiten.KeyEscape):
		c.step = nil
	case pressed(ebiten.KeySpace):
		d.togglePlay()
	case pressed(ebiten.KeyTab):
		g = g.Rest(s.cursor)
		s.cursor = (s.cursor + 1) % n
	case pressed(ebiten.KeyMinus):
		g = g.Hold(s.cursor)
		s.cursor = (s.cursor + 1) % n
	case pressed(ebiten.KeyBackspace) || pressed(ebiten.KeyLeft):
		s.cursor = (s.cursor + n - 1) % n
	case pressed(ebiten.KeyRight):
		s.cursor = (s.cursor + 1) % n
	case pressed(ebiten.KeyZ):
		s.octave = max(s.octave-1, 0)
	case pressed(ebiten.KeyX):
		s.octave = min(s.octave+1, 8)
	case pressed(ebiten.KeyC):
		s.accent = max(s.accent-1, -1)
	case pressed(ebiten.KeyV):
		s.accent = min(s.accent+1, 1)
	}
	if len(played) > 0 {
		if i := g.Start(s.chord); i >= 0 {
			d.audition(t, played, g.Events[i].Velocity(t.c.Vel))
		}
	}
	c.write(d, t, g)
}

// audition plays keys through the track's patch, or the clip's Hit with no keys.
func (d *desk) audition(t clipTarget, keys []int, vel float64) {
	tr := d.layout.Arrangement.Tracks[t.track]
	if len(keys) == 0 {
		for name := range strings.SplitSeq(t.c.Hit, ",") {
			if k, err := synth.ParseKey(strings.TrimSpace(name)); err == nil {
				keys = append(keys, int(k))
			}
		}
	}
	if len(keys) == 0 {
		return
	}
	var notes []synth.Note
	for _, k := range keys {
		notes = append(notes, synth.Note{Key: float64(k), Dur: 0.3, Vel: vel})
	}
	h := fnv.New64a()
	fmt.Fprintf(h, "%v", tr.Patch)
	s := synth.Song{
		ID:    fmt.Sprintf("studio audition %x %v %.2f", h.Sum64(), keys, vel),
		Parts: []synth.Part{{Name: tr.Name, Patch: tr.Patch, Level: tr.Level, Pan: tr.Pan, Notes: notes}},
	}
	go d.player.Play(s, 1)
}

// A clipSetting is a row of the clip's settings: a value to drag or type, or buttons.
type clipSetting struct {
	label, value string
	field        string // the Clip field it writes, for whether it's code
	value0       float64
	lo, hi, step float64
	set          func(v float64)
	parse        func(s string) (float64, bool)
	text         func(s string) bool // a value typed as text
	chars        string
	buttons      []clipButton
}

type clipButton struct {
	label   string
	do      func()
	can, on bool
}

func ms(s float64) string { return strconv.FormatFloat(math.Round(s*1e6)/1e3, 'f', -1, 64) + " ms" }

func number(s string) (float64, bool) {
	x, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "ms")), 64)
	return x, err == nil && !math.IsNaN(x) && !math.IsInf(x, 0)
}

func (c *clipEditor) settingRows(d *desk, t clipTarget) []clipSetting {
	cl := t.c
	edit := func(f func(c *synth.Clip)) {
		d.edit(func(a *synth.Arrangement) { f(&a.Tracks[t.track].Clips[t.clip]) })
	}
	frac := func(label, field string, v, def float64, lo, hi float64, set func(c *synth.Clip, v float64)) clipSetting {
		shown := v
		if shown == 0 {
			shown = def
		}
		return clipSetting{
			label: label, value: strconv.FormatFloat(shown, 'f', -1, 64), field: field,
			value0: shown, lo: lo, hi: hi, step: 0.01,
			set:   func(v float64) { edit(func(c *synth.Clip) { set(c, v) }) },
			parse: number,
		}
	}
	seconds := func(label, field string, v, hi float64, set func(c *synth.Clip, v float64)) clipSetting {
		return clipSetting{
			label: label, value: ms(v), field: field,
			value0: v, lo: 0, hi: hi, step: 0.001,
			set: func(v float64) { edit(func(c *synth.Clip) { set(c, v) }) },
			parse: func(s string) (float64, bool) {
				x, ok := number(s)
				return x / 1000, ok && x >= 0
			},
		}
	}

	rows := []clipSetting{{
		label: "name", value: cl.Name, field: "Name", chars: "*",
		text: func(s string) bool {
			s = strings.TrimSpace(s)
			tr := d.layout.Arrangement.Tracks[t.track]
			if s == "" || s != cl.Name && slices.ContainsFunc(tr.Clips, func(c synth.Clip) bool { return c.Name == s }) {
				return false
			}
			d.edit(func(a *synth.Arrangement) {
				tr := &a.Tracks[t.track]
				tr.Clips[t.clip].Name = s
				for i := range tr.Play {
					if tr.Play[i].Clip == cl.Name {
						tr.Play[i].Clip = s
					}
				}
			})
			d.view.Clip = s
			c.shown = d.view.Track + "\x00" + s
			return true
		},
	}}
	if t.err != nil {
		return rows
	}
	steps := d.editable("Tracks", t.track, "Clips", t.clip, "Step")
	can, _ := c.canWrite(d, t)
	_, doubles := t.g.Double()
	rows = append(rows, clipSetting{
		label: "step", value: beats(cl.Step), field: "Step",
		buttons: []clipButton{
			{"÷2", func() {
				bars := t.g.Halve().Write(d.musicalScale())
				edit(func(c *synth.Clip) { c.Step, c.Bars = c.Step/2, bars })
			}, steps && can, false},
			{"×2", func() {
				g, _ := t.g.Double()
				bars := g.Write(d.musicalScale())
				edit(func(c *synth.Clip) { c.Step, c.Bars = c.Step*2, bars })
			}, steps && can && doubles, false},
		},
	})
	if cl.Hit != "" {
		rows = append(rows, clipSetting{
			label: "hit", value: cl.Hit, field: "Hit", chars: "ABCDEFGabcdefg#0123456789-,",
			text: func(s string) bool {
				s = strings.ReplaceAll(strings.TrimSpace(s), " ", "")
				for name := range strings.SplitSeq(s, ",") {
					if _, err := synth.ParseKey(name); err != nil {
						return false
					}
				}
				edit(func(c *synth.Clip) { c.Hit = s })
				return true
			},
		})
	}
	rows = append(rows,
		frac("gate", "Gate", cl.Gate, 1, 0.05, 1, func(c *synth.Clip, v float64) { c.Gate = v }),
		frac("velocity", "Vel", cl.Vel, 0.8, 0.05, 1, func(c *synth.Clip, v float64) { c.Vel = v }),
		frac("swing", "Swing", cl.Swing, 0, 0, 0.5, func(c *synth.Clip, v float64) { c.Swing = v }),
		seconds("strum", "Strum", cl.Strum, 0.06, func(c *synth.Clip, v float64) { c.Strum = v }),
		seconds("glide", "Glide", cl.Glide, 0.5, func(c *synth.Clip, v float64) { c.Glide = v }),
		seconds("humanize", "Humanize", cl.Humanize, 0.03, func(c *synth.Clip, v float64) { c.Humanize = v }),
		clipSetting{
			label: "seed", value: strconv.FormatUint(cl.Seed, 10), field: "Seed",
			value0: float64(cl.Seed), lo: 0, hi: 999, step: 1,
			set: func(v float64) { edit(func(c *synth.Clip) { c.Seed = uint64(v) }) },
			parse: func(s string) (float64, bool) {
				n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
				return float64(n), err == nil
			},
		},
		clipSetting{
			label: "pedal", field: "Pedal",
			buttons: []clipButton{{map[bool]string{true: "held to the bar", false: "off"}[cl.Pedal], func() {
				edit(func(c *synth.Clip) { c.Pedal = !c.Pedal })
			}, true, cl.Pedal}},
		},
	)
	return rows
}

// beats writes a step as a fraction of a beat where it is one: 1/4 beat, 2 beats.
func beats(step float64) string {
	for _, den := range []float64{1, 2, 3, 4, 6, 8, 12, 16, 24, 32} {
		if num := step * den; math.Abs(num-math.Round(num)) < 1e-9 {
			switch {
			case den == 1 && num == 1:
				return "1 beat"
			case den == 1:
				return fmt.Sprintf("%g beats", num)
			}
			return fmt.Sprintf("%g/%g beat", math.Round(num), den)
		}
	}
	return fmt.Sprintf("%g beat", step)
}

func (c *clipEditor) settingRect(d *desk, i int) image.Rectangle {
	s := c.settings(d)
	y := s.Min.Y + int(d.px(6)) + i*int(d.px(settingH))
	return image.Rect(s.Min.X+int(d.px(10)), y, s.Max.X-int(d.px(10)), y+int(d.px(settingH))-1)
}

func (c *clipEditor) valueRect(d *desk, i int) image.Rectangle {
	r := c.settingRect(d, i)
	return image.Rect(r.Min.X+int(d.px(70)), r.Min.Y, r.Max.X, r.Max.Y)
}

// buttonRect is button j of n on row i, beside the row's value if it has one.
func (c *clipEditor) buttonRect(d *desk, s clipSetting, i, j int) image.Rectangle {
	v := c.valueRect(d, i)
	n := len(s.buttons)
	if s.value != "" {
		v.Min.X = v.Max.X - int(d.px(70))
	}
	w := (v.Dx() - int(d.px(4))*(n-1)) / n
	x := v.Min.X + j*(w+int(d.px(4)))
	return image.Rect(x, v.Min.Y+1, x+w, v.Max.Y-1)
}

func (c *clipEditor) updateSettings(d *desk, t clipTarget, pt image.Point) {
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if clicked {
		c.typing = nil
	}
	for i, s := range c.settingRows(d, t) {
		if !pt.In(c.settingRect(d, i)) {
			continue
		}
		can := d.editable("Tracks", t.track, "Clips", t.clip, s.field)
		if !can {
			c.hint = custom
			if d.doc == nil {
				c.hint = d.readOnly()
			}
			return
		}
		switch {
		case s.buttons != nil:
			c.hint = ""
			for j, b := range s.buttons {
				if pt.In(c.buttonRect(d, s, i, j)) && clicked && b.can {
					b.do()
				}
			}
			if s.field == "Step" {
				c.hint = "halving the step rewrites the bars, losing nothing; doubling it needs every note on the longer step"
			}
		case s.text != nil:
			c.hint = "click to type it"
			if clicked && pt.In(c.valueRect(d, i)) {
				c.typing = &typing{id: s.label, text: s.value, set: s.text, chars: s.chars}
			}
		default:
			c.hint = "drag (shift: fine) · click to type · double-click resets"
			if !clicked || !pt.In(c.valueRect(d, i)) {
				return
			}
			now := time.Now()
			if c.last == s.label && now.Sub(c.lastTime) < 400*time.Millisecond {
				c.last = ""
				s.set(0)
				return
			}
			c.last, c.lastTime = s.label, now
			typed := &typing{id: s.label, text: strings.TrimSuffix(s.value, " ms"), chars: "0123456789.", set: func(text string) bool {
				v, ok := s.parse(text)
				if ok {
					s.set(session.Round(math.Max(s.lo, math.Min(s.hi, v)), s.step))
				}
				return ok
			}}
			c.begin(d, &clipDrag{kind: setting, typed: typed, value: &drag{
				id: s.label, from: s.value0, y: float64(pt.Y), set: s.set, travel: d.px(200), lo: s.lo, hi: s.hi, step: s.step,
			}})
		}
		return
	}
}

func (c *clipEditor) draw(dst *ebiten.Image, d *desk) {
	t, ok := d.target()
	if !ok || c.r.Empty() {
		return
	}
	area := dst.SubImage(c.r).(*ebiten.Image)
	fill(area, c.r, ground)
	line(area, float64(c.r.Min.X), float64(c.r.Min.Y), float64(c.r.Max.X), float64(c.r.Min.Y), 1, edge)
	c.drawHead(area, d, t)
	c.drawSettings(area, d, t)
	if t.err != nil {
		g := c.grid(d)
		drawText(area, d.fonts.text, fmt.Sprintf("%d notes off the grid: edit them in code", len(t.c.Notes)),
			float64(g.Min.X)+d.px(12), float64(g.Min.Y)+d.px(20), dim, text.AlignStart)
		return
	}
	if t.g.Hit {
		c.drawDrums(area, d, t)
	} else {
		c.drawRoll(area, d, t)
	}
	c.drawLane(area, d, t)
}

func (c *clipEditor) drawHead(dst *ebiten.Image, d *desk, t clipTarget) {
	h := c.head(d)
	fill(dst, h, panel)
	tr := d.layout.Arrangement.Tracks[t.track]
	fill(dst, image.Rect(h.Min.X, h.Min.Y, h.Min.X+int(d.px(4)), h.Max.Y), hue(tr.Name, 0.75))
	plays := 0
	for _, p := range tr.Play {
		if p.Clip == t.c.Name {
			plays += max(p.Times, 1)
		}
	}
	about := fmt.Sprintf("%s · %s · %d bars, played %d times", t.c.Name, tr.Name, len(t.c.Bars), plays)
	if t.err != nil {
		about = fmt.Sprintf("%s · %s · notes, played %d times", t.c.Name, tr.Name, plays)
	}
	if s := c.step; s != nil {
		accent := map[int]string{1: "accented", 0: "plain", -1: "ghosted"}[s.accent]
		about += fmt.Sprintf("   ·   step input at %d: A–K white keys, W E T Y U black · Z X octave (C%d) · C V %s · tab rest · - hold · esc done",
			s.cursor+1, s.octave, accent)
	}
	drawText(dst, d.fonts.text, fit(d.fonts.text, about, float64(h.Dx())-d.px(20)), float64(h.Min.X)+d.px(12), float64(h.Min.Y+h.Dy()/2), ink, text.AlignStart)
}

// drawColumns shades the grid's bars, beats and steps, the step input's cursor and
// the playhead.
func (c *clipEditor) drawColumns(dst *ebiten.Image, d *desk, t clipTarget, top, bottom float64) {
	sw := c.stepW(d, t)
	perBeat := 1 / t.c.Step
	beatSteps := 0
	if r := math.Round(perBeat); r >= 1 && math.Abs(perBeat-r) < 1e-9 {
		beatSteps = int(r)
	}
	bar, barStep := 0, 0
	for s := 0; s <= t.g.Steps(); s++ {
		x := c.x(d, t, float64(s))
		switch {
		case s == barStep:
			line(dst, x, top, x, bottom, d.px(1.5), barLine)
			if bar < len(t.g.Bars) {
				drawText(dst, d.fonts.num, fmt.Sprint(bar+1), x+d.px(3), top+d.px(7), faint, text.AlignStart)
				barStep += t.g.Bars[bar]
			}
			bar++
		case beatSteps > 0 && (s-t.g.BarStep(bar-1))%beatSteps == 0:
			line(dst, x, top, x, bottom, 1, barLine)
		case sw >= d.px(10):
			line(dst, x, top, x, bottom, 1, beatLine)
		}
	}
	if c.step != nil {
		x := c.x(d, t, float64(c.step.cursor))
		fillF(dst, x, top, sw, bottom-top, color.RGBA{0xff, 0xa0, 0x40, 0x30})
		line(dst, x, top, x, bottom, d.px(1.5), playhead)
	}
	if at, ok := c.playing(d, t); ok {
		x := c.x(d, t, at)
		line(dst, x, top, x, bottom, d.px(1.5), playhead)
	}
}

// playing is the step the playhead is at, when it's in a play of the clip.
func (c *clipEditor) playing(d *desk, t clipTarget) (float64, bool) {
	if !d.sounding || d.layout.Tempo <= 0 {
		return 0, false
	}
	for _, l := range d.layout.Lanes {
		if l.Name != d.view.Track {
			continue
		}
		for _, b := range l.Blocks {
			if b.Clip == t.c.Name && d.view.At >= b.From && d.view.At < b.To {
				step := (d.view.At - b.From) / (t.c.Step * 60 / d.layout.Tempo)
				return math.Mod(step, float64(t.g.Steps())), true
			}
		}
	}
	return 0, false
}

func (c *clipEditor) drawRoll(dst *ebiten.Image, d *desk, t clipTarget) {
	g := c.grid(d)
	k := c.keyboard(d)
	roll := dst.SubImage(g).(*ebiten.Image)
	keys := dst.SubImage(k).(*ebiten.Image)
	sc, scaled := session.ParseScale(d.layout.Key)
	rowH := d.px(keyRowH)
	lo := c.keyAt(d, float64(g.Max.Y))
	hi := c.keyAt(d, float64(g.Min.Y))
	black := []bool{false, true, false, true, false, false, true, false, true, false, true, false}
	for key := lo; key <= hi; key++ {
		y := c.y(d, key)
		bg := laneEven
		if scaled && sc.Has(key) {
			bg = laneOdd
		}
		if scaled && key%12 == sc.Tonic {
			bg = selected
		}
		fillF(roll, float64(g.Min.X), y, float64(g.Dx()), rowH, bg)
		if key%12 == 0 {
			line(roll, float64(g.Min.X), y+rowH, float64(g.Max.X), y+rowH, 1, barLine)
		}
		kc := color.Color(ink)
		if black[key%12] {
			kc = edge
		}
		if scaled && sc.Has(key) {
			kc = hsv(200, 0.25, 0.85)
			if black[key%12] {
				kc = hsv(200, 0.35, 0.45)
			}
		}
		fillF(keys, float64(k.Min.X), y+1, float64(k.Dx())-d.px(2), rowH-1, kc)
		if key%12 == 0 {
			drawText(keys, d.fonts.num, synth.KeyName(key), float64(k.Max.X)-d.px(5), y+rowH/2, ground, text.AlignEnd)
		}
	}
	c.drawColumns(roll, d, t, float64(g.Min.Y), float64(g.Max.Y))

	tr := d.layout.Arrangement.Tracks[t.track]
	sw := c.stepW(d, t)
	for _, e := range t.g.Events {
		x0, x1 := c.x(d, t, float64(e.Step)), c.x(d, t, float64(e.Step+e.Len))
		v := e.Velocity(t.c.Vel)
		for _, p := range e.Keys {
			y := c.y(d, p.Key)
			on := slices.Contains(c.sel, session.Ref{Step: e.Step, Key: p.Key})
			body := hue(tr.Name, 0.45+0.45*v)
			if on {
				body = color.RGBA{0xff, 0xc0, 0x70, 0xff}
			}
			fillF(roll, x0+1, y+1, x1-x0-2, rowH-2, body)
			if hits := max(e.Hits, 1); hits > 1 {
				span := (x1 - x0) / float64(hits)
				for h := 1; h < hits; h++ {
					line(roll, x0+span*float64(h), y+1, x0+span*float64(h), y+rowH-1, d.px(1.5), ground)
				}
			}
			if p.Slide != 0 {
				ty := c.y(d, p.Key+p.Slide)
				line(roll, x0+math.Min(sw/2, d.px(6)), y+rowH/2, x1-1, ty+rowH/2, d.px(1.5), ink)
				fillF(roll, x1-d.px(4), ty+rowH/4, d.px(4), rowH/2, ink)
			}
		}
	}
}

func (c *clipEditor) drawDrums(dst *ebiten.Image, d *desk, t clipTarget) {
	g := c.grid(d)
	area := dst.SubImage(g).(*ebiten.Image)
	k := c.keyboard(d)
	fill(dst, k, panel)
	row := c.drumRow(d)
	drawText(dst, d.fonts.text, fit(d.fonts.text, t.c.Hit, float64(k.Dx())-d.px(8)), float64(k.Min.X)+d.px(6), float64(row.Min.Y+row.Dy()/2), dim, text.AlignStart)
	fill(area, g, laneEven)
	fill(area, row, laneOdd)
	c.drawColumns(area, d, t, float64(g.Min.Y), float64(g.Max.Y))

	tr := d.layout.Arrangement.Tracks[t.track]
	sw := c.stepW(d, t)
	top, h := float64(row.Min.Y), float64(row.Dy())
	for _, e := range t.g.Events {
		x0, x1 := c.x(d, t, float64(e.Step)), c.x(d, t, float64(e.Step+e.Len))
		inset := map[int]float64{1: 0.08, 0: 0.2, -1: 0.36}[e.Accent] * h
		col := hue(tr.Name, 0.45+0.45*e.Velocity(t.c.Vel))
		if len(e.Keys) > 0 {
			col = dim
		}
		if e.Len > 1 {
			fillF(area, x0+sw/2, top+h/2-d.px(1), x1-x0-sw/2, d.px(2), col)
		}
		hits := max(e.Hits, 1)
		span := (x1 - x0) / float64(hits)
		for n := range hits {
			hx := x0 + span*float64(n)
			fillF(area, hx+2, top+inset, math.Min(span, sw)-4, h-2*inset, col)
		}
	}
}

func (c *clipEditor) drawLane(dst *ebiten.Image, d *desk, t clipTarget) {
	l := c.lane(d)
	fill(dst, l, panel)
	area := dst.SubImage(l).(*ebiten.Image)
	for _, level := range []float64{1, 2} {
		y := float64(l.Min.Y) + float64(l.Dy())*level/3
		line(area, float64(l.Min.X), y, float64(l.Max.X), y, 1, beatLine)
	}
	drawText(dst, d.fonts.text, "velocity", float64(c.r.Min.X)+d.px(6), float64(l.Min.Y+l.Dy()/2), faint, text.AlignStart)
	tr := d.layout.Arrangement.Tracks[t.track]
	w := math.Max(math.Min(c.stepW(d, t)-d.px(2), d.px(6)), d.px(2))
	for _, e := range t.g.Events {
		x := c.x(d, t, float64(e.Step))
		v := e.Velocity(t.c.Vel)
		hgt := v * (float64(l.Dy()) - d.px(6))
		fillF(area, x+1, float64(l.Max.Y)-hgt, w, hgt, hue(tr.Name, 0.45+0.45*v))
	}
}

func (c *clipEditor) drawSettings(dst *ebiten.Image, d *desk, t clipTarget) {
	s := c.settings(d)
	fill(dst, s, panel)
	line(dst, float64(s.Min.X), float64(s.Min.Y), float64(s.Min.X), float64(s.Max.Y), 1, edge)
	for i, row := range c.settingRows(d, t) {
		r := c.settingRect(d, i)
		if r.Max.Y > s.Max.Y {
			break
		}
		can := d.editable("Tracks", t.track, "Clips", t.clip, row.field)
		mid := float64(r.Min.Y + r.Dy()/2)
		drawText(dst, d.fonts.text, row.label, float64(r.Min.X), mid, dim, text.AlignStart)
		if row.buttons != nil {
			for j, b := range row.buttons {
				box(dst, d, c.buttonRect(d, row, i, j), b.label, b.on, lit, can && b.can)
			}
			if row.value != "" {
				v := c.valueRect(d, i)
				drawText(dst, d.fonts.num, row.value, float64(v.Min.X), mid, ink, text.AlignStart)
			}
			continue
		}
		v := c.valueRect(d, i)
		value, col := row.value, color.Color(ink)
		if !can {
			value, col = row.value+" (code)", faint
		}
		if c.typing != nil && c.typing.id == row.label {
			fill(dst, v, edge)
			value, col = c.typing.text+"▏", playhead
		}
		drawText(dst, d.fonts.num, fit(d.fonts.num, value, float64(v.Dx())), float64(v.Min.X), mid, col, text.AlignStart)
	}
}
