package desk

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"time"

	"github.com/cbodonnell/ebui"
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/studio/session"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Sizes in points, before the window's scale.
const (
	cellW      = 62
	cellH      = 54
	deviceHead = 22
	groupHead  = 18
	curveH     = 44
	deviceGap  = 8
	addW       = 34
	cellKnobR  = 12
)

// devices is the detail's other view: the chain of the selected track (its patch, then
// its inserts), bus or master, each device drawn from synth's parameter table.
type devices struct {
	scroll float64 // pixels the chain is scrolled left

	drag   *drag
	moved  bool
	typing *typing
	hint   string

	last     string // the control last clicked, and when, for a double-click
	lastTime time.Time
	track    string // the selected track last seen, so selecting another shows its chain

	carry   *carried // a patch or an effect dragged from the browser
	sources []source
	folded  map[string]bool // the browser's sections folded shut
	plots   map[string]*ebiten.Image
}

type carried struct{ patch, effect string }

type source struct {
	b *ebui.Button
	carried
}

// close forgets what was being done in the chain, as another song opens.
func (v *devices) close(d *desk) {
	v.scroll, v.drag, v.typing, v.carry = 0, nil, nil, nil
	v.track = d.view.Track
}

func (d *desk) toggleDevices() {
	d.view.Devices = !d.view.Devices
	d.dv.typing, d.ce.typing = nil, nil
}

// chain is the chain the devices show, and its name.
func (d *desk) chain() (session.Chain, string, bool) {
	a := d.layout.Arrangement
	switch {
	case d.view.Master:
		return session.Chain{Kind: session.MasterChain}, "master", true
	case d.view.Bus != "":
		if i := slices.IndexFunc(a.Buses, func(b synth.Bus) bool { return b.Name == d.view.Bus }); i >= 0 {
			return session.Chain{Kind: session.BusChain, Index: i}, d.view.Bus, true
		}
	}
	if i := slices.IndexFunc(a.Tracks, func(t synth.Track) bool { return t.Name == d.view.Track }); i >= 0 {
		return session.Chain{Kind: session.TrackChain, Index: i}, d.view.Track, true
	}
	return session.Chain{}, "", false
}

// detailHint is what the control under the cursor in the detail says.
func (d *desk) detailHint() string {
	if d.view.Devices {
		return d.dv.hint
	}
	return d.ce.hint
}

// The chain laid out: frames and headers, the controls, buttons and curves, and the
// places an effect can go.
type deviceView struct {
	chain   session.Chain
	name    string
	frames  []image.Rectangle
	heads   []devHead
	cells   []devCell
	buttons []devButton
	curves  []devCurve
	slots   []float64 // x before each effect, and after the last
	width   float64
}

type devHead struct {
	r           image.Rectangle
	title, note string
	group       bool
}

type devCell struct {
	r   image.Rectangle
	c   session.Control
	can bool
}

type devButton struct {
	r     image.Rectangle
	label string
	can   bool
	do    func()
}

type devCurve struct {
	r    image.Rectangle
	ys   []float64
	mark float64
}

func rect(x, y, w, h float64) image.Rectangle {
	return image.Rect(int(x), int(y), int(x+w), int(y+h))
}

func (v *devices) layout(d *desk) (deviceView, bool) {
	ch, name, ok := d.chain()
	if !ok {
		return deviceView{}, false
	}
	a := d.layout.Arrangement
	px := d.px
	r := d.widgets.detail
	top := float64(r.Min.Y) + px(clipHeadH) + px(6)
	bottom := float64(r.Max.Y) - px(6)
	left := float64(r.Min.X) + px(8)
	x := left - v.scroll
	out := deviceView{chain: ch, name: name}

	place := func(cs []session.Control, x0, y0 float64, rows, cols int) {
		for k, c := range cs {
			col, row := k/rows, k%rows
			if col >= cols {
				break
			}
			out.cells = append(out.cells, devCell{rect(x0+float64(col)*px(cellW), y0+float64(row)*px(cellH), px(cellW), px(cellH)), c, d.editable(c.Path...)})
		}
	}
	columns := func(n, rows int) int { return max(2, (n+rows-1)/rows) }
	literal := func(r image.Rectangle, path []any) {
		w := px(84)
		out.buttons = append(out.buttons, devButton{rect(float64(r.Max.X)-w-px(4), float64(r.Min.Y)+px(3), w, px(deviceHead-6)), "Make literal", d.doc != nil, func() {
			if d.doc.MakeLiteral(path...) {
				d.changed()
			}
		}})
	}

	if ch.Kind == session.TrackChain {
		t := a.Tracks[ch.Index]
		path := []any{"Tracks", ch.Index, "Patch"}
		start := x
		gx := x + px(6)
		for _, g := range session.PatchGroups(t.Patch, path) {
			gy := top + px(deviceHead) + px(4)
			space := 0.0
			if g.Curve != session.NoCurve {
				space = px(curveH) + px(4)
			}
			rows := max(1, int((bottom-gy-px(groupHead)-space)/px(cellH)))
			cols := columns(len(g.Controls), rows)
			w := float64(cols) * px(cellW)
			out.heads = append(out.heads, devHead{r: rect(gx, gy, w, px(groupHead)), title: g.Title, group: true})
			if g.Curve != session.NoCurve {
				c := devCurve{r: rect(gx, gy+px(groupHead), w, px(curveH)), mark: -1}
				if val, ok := session.Get(a, g.Path); ok {
					n := max(2, int(w/2))
					switch g.Curve {
					case session.EnvelopeCurve:
						c.ys, c.mark = session.EnvelopeShape(val.Interface().(synth.Envelope), n)
					case session.ResponseCurve:
						c.ys = session.FilterResponse(val.Interface().(synth.Filter), n)
					}
				}
				out.curves = append(out.curves, c)
			}
			place(g.Controls, gx, gy+px(groupHead)+space, rows, cols)
			gx += w + px(10)
		}
		frame := rect(start, top, gx-start, bottom-top)
		note := t.Patch.Name
		if d.doc != nil && d.doc.Custom(path...) {
			note = "written as code: shown, not editable"
		}
		out.frames = append(out.frames, frame)
		head := rect(start, top, gx-start, px(deviceHead))
		out.heads = append(out.heads, devHead{r: head, title: "Patch", note: note})
		if d.doc != nil && d.doc.Custom(path...) {
			literal(head, path)
		}
		x = gx + px(deviceGap)
	}

	chainPath := ch.Path()
	fx := *ch.In(&a)
	chainCan := d.editable(chainPath...)
	if d.doc != nil && d.doc.Custom(chainPath...) && len(fx) > 0 {
		w := px(150)
		head := rect(x, top, w, px(deviceHead))
		out.frames = append(out.frames, rect(x, top, w, bottom-top))
		out.heads = append(out.heads, devHead{r: head, title: "Chain", note: "written as code"})
		literal(rect(x, top+px(deviceHead), w, px(deviceHead)), chainPath)
		x += w + px(deviceGap)
	}
	rows := max(1, int((bottom-top-px(deviceHead)-px(8))/px(cellH)))
	for j, e := range fx {
		out.slots = append(out.slots, x)
		path := slices.Concat(chainPath, []any{j})
		cs := session.Controls(e, path)
		cols := columns(len(cs), rows)
		w := float64(cols)*px(cellW) + px(12)
		head := rect(x, top, w, px(deviceHead))
		out.frames = append(out.frames, rect(x, top, w, bottom-top))
		custom := d.doc != nil && d.doc.Custom(path...)
		note := ""
		if custom {
			note = "code"
		}
		out.heads = append(out.heads, devHead{r: head, title: effectName(e), note: note})
		switch {
		case custom && chainCan:
			literal(head, path)
		case chainCan:
			b := px(16)
			bx := float64(head.Max.X) - px(4) - 3*b - px(4)
			for k, it := range []struct {
				label string
				do    func(a *synth.Arrangement)
			}{
				{"<", func(a *synth.Arrangement) { session.MoveEffect(a, ch, j, j-1) }},
				{">", func(a *synth.Arrangement) { session.MoveEffect(a, ch, j, j+1) }},
				{"x", func(a *synth.Arrangement) { session.RemoveEffect(a, ch, j) }},
			} {
				can := (k != 0 || j > 0) && (k != 1 || j < len(fx)-1)
				out.buttons = append(out.buttons, devButton{rect(bx+float64(k)*(b+px(2)), float64(head.Min.Y)+px(3), b, px(deviceHead-6)), it.label, can, func() { d.edit(it.do) }})
			}
		}
		place(cs, x+px(6), top+px(deviceHead)+px(4), rows, cols)
		x += w + px(deviceGap)
	}
	out.slots = append(out.slots, x)
	if chainCan {
		out.buttons = append(out.buttons, devButton{rect(x, top, px(addW), px(deviceHead)), "+", true, func() {
			at := image.Pt(ebiten.CursorPosition())
			var items []item
			for _, name := range session.EffectNames() {
				items = append(items, item{"add " + name, func() { d.dv.addEffect(d, ch, len(fx), name) }})
			}
			d.openMenu(at, items)
		}})
		x += px(addW)
	}
	out.width = x + v.scroll - left + px(8)
	return out, true
}

func (v *devices) update(d *desk) {
	v.hint = ""
	x, y := ebiten.CursorPosition()
	pt := image.Pt(x, y)
	if v.drag != nil {
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || d.doc == nil {
			v.drag = nil
			if d.doc != nil {
				d.doc.Release()
			}
			return
		}
		g := v.drag
		if float64(y) == g.y && !v.moved {
			return
		}
		v.moved = true
		dy := (g.y - float64(y)) / g.travel
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			dy /= 10
		}
		g.set(g.from + dy)
		return
	}
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	r := d.widgets.detail
	if !pt.In(r) {
		if clicked {
			v.typing = nil
		}
		return
	}
	if clicked {
		v.typing = nil
	}
	view, ok := v.layout(d)
	if !ok {
		return
	}
	if wx, wy := ebiten.Wheel(); wx != 0 || wy != 0 {
		dx := wx
		if dx == 0 {
			dx = wy
		}
		v.scroll = math.Max(0, math.Min(v.scroll-dx*d.px(24), view.width-float64(r.Dx())))
	}
	v.hint = "drag a knob (shift: fine) · double-click resets · click a value to type it · click a switch to change it · shift+tab clip editor"
	cant := func() {
		v.hint = custom
		if d.doc == nil {
			v.hint = d.readOnly()
		}
	}
	for _, b := range view.buttons {
		if pt.In(b.r) {
			if !b.can {
				cant()
			} else if clicked {
				b.do()
			}
			return
		}
	}
	for _, c := range view.cells {
		if !pt.In(c.r) {
			continue
		}
		switch {
		case !c.can:
			cant()
		case c.c.Kind == session.Points:
			v.hint = "automation: edit its points in the lane under the track"
		case clicked:
			v.press(d, c, pt)
		}
		return
	}
}

// press takes a click on a control.
func (v *devices) press(d *desk, cell devCell, pt image.Point) {
	c := cell.c
	id := fmt.Sprint(c.Path...)
	set := func(x any) { d.edit(func(a *synth.Arrangement) { session.Set(a, c.Path, x) }) }
	now := time.Now()
	double := v.last == id && now.Sub(v.lastTime) < 400*time.Millisecond
	v.last, v.lastTime = id, now
	if double {
		v.last = ""
	}
	switch c.Kind {
	case session.Knob:
		if pt.Y >= v.valueRect(d, cell.r).Min.Y {
			v.typing = &typing{id: id, text: session.FormatParam(c.Param, session.Effective(c.Param, c.Value)), chars: "*", set: func(s string) bool {
				x, ok := session.ParseParam(c.Param, s)
				if ok {
					set(x)
				}
				return ok
			}}
			return
		}
		if double {
			set(0.0)
			return
		}
		hi := d.songBeats()
		d.doc.Hold()
		v.moved = false
		v.drag = &drag{id: id, from: session.Position(c.Param, c.Value, hi), y: float64(pt.Y), travel: d.px(160), set: func(pos float64) {
			set(session.FromPosition(c.Param, pos, hi))
		}}
	case session.Cycle:
		step := 1
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			step = len(c.Names) - 1
		}
		set((int(c.Value) + step) % len(c.Names))
	case session.Toggle:
		set(!c.On)
	case session.Text:
		v.typing = &typing{id: id, text: c.Text, chars: "*", set: func(s string) bool {
			set(s)
			return true
		}}
	}
}

// songBeats is how long the song is in beats, or in seconds for a sound: how far a
// song position's control reaches.
func (d *desk) songBeats() float64 {
	if d.layout.Tempo == 0 {
		return d.layout.Song.Length
	}
	return d.layout.Song.Length * d.layout.Tempo / 60
}

func (v *devices) valueRect(d *desk, r image.Rectangle) image.Rectangle {
	return image.Rect(r.Min.X, r.Max.Y-int(d.px(16)), r.Max.X, r.Max.Y)
}

// keys types into the value being typed into, or switches the detail, reporting whether
// it took the keys.
func (v *devices) keys(d *desk) bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) && ebiten.IsKeyPressed(ebiten.KeyShift) {
		d.toggleDevices()
		return true
	}
	if v.typing == nil || !d.view.Devices {
		return false
	}
	if typeInto(v.typing) {
		v.typing = nil
	}
	return true
}

func (v *devices) draw(dst *ebiten.Image, d *desk) {
	r := d.widgets.detail
	if r.Empty() {
		return
	}
	area := dst.SubImage(r).(*ebiten.Image)
	fill(area, r, ground)
	line(area, float64(r.Min.X), float64(r.Min.Y), float64(r.Max.X), float64(r.Min.Y), 1, edge)
	head := image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+int(d.px(clipHeadH)))
	fill(area, head, panel)
	view, ok := v.layout(d)
	about := "devices · select a track, a bus or the master"
	if ok {
		about = "devices · " + view.name
		if view.chain.Kind == session.TrackChain {
			fill(area, image.Rect(head.Min.X, head.Min.Y, head.Min.X+int(d.px(4)), head.Max.Y), hue(view.name, 0.75))
		}
	}
	drawText(area, d.fonts.text, about+"   ·   shift+tab clip editor · drag patches and effects in from the browser", float64(head.Min.X)+d.px(12), float64(head.Min.Y+head.Dy()/2), ink, text.AlignStart)
	if !ok {
		return
	}
	for _, f := range view.frames {
		fill(area, f, laneEven)
	}
	for _, h := range view.heads {
		if !h.group {
			fill(area, h.r, panel)
		}
		c := color.Color(ink)
		if h.group {
			c = dim
		}
		drawText(area, d.fonts.text, h.title, float64(h.r.Min.X)+d.px(6), float64(h.r.Min.Y+h.r.Dy()/2), c, text.AlignStart)
		if h.note != "" {
			x := float64(h.r.Min.X) + d.px(10) + text.Advance(h.title, d.fonts.text)
			drawText(area, d.fonts.text, fit(d.fonts.text, h.note, float64(h.r.Max.X)-x-d.px(100)), x, float64(h.r.Min.Y+h.r.Dy()/2), faint, text.AlignStart)
		}
	}
	for _, c := range view.curves {
		v.drawCurve(area, d, c)
	}
	for _, c := range view.cells {
		v.drawCell(area, d, c)
	}
	for _, b := range view.buttons {
		box(area, d, b.r, b.label, false, lit, b.can)
	}
	if v.carry != nil && v.carry.effect != "" && image.Pt(ebiten.CursorPosition()).In(r) {
		cx, _ := ebiten.CursorPosition()
		x := view.slots[v.slot(view, float64(cx))]
		line(area, x-d.px(deviceGap)/2, float64(r.Min.Y)+d.px(clipHeadH), x-d.px(deviceGap)/2, float64(r.Max.Y), d.px(2), playhead)
	}
}

// slot is where an effect dropped at x goes in the chain.
func (v *devices) slot(view deviceView, x float64) int {
	n := len(view.slots) - 1
	for i := range n {
		if x < (view.slots[i]+view.slots[i+1])/2 {
			return i
		}
	}
	return n
}

func (v *devices) drawCurve(dst *ebiten.Image, d *desk, c devCurve) {
	fill(dst, c.r, edge)
	if len(c.ys) == 0 || c.r.Dx() < 2 || c.r.Dy() < 2 {
		return
	}
	key := fmt.Sprint(c.r.Dx(), c.r.Dy(), c.mark, c.ys)
	img, ok := v.plots[key]
	if !ok {
		if len(v.plots) > 64 || v.plots == nil {
			for _, p := range v.plots {
				p.Deallocate()
			}
			v.plots = map[string]*ebiten.Image{}
		}
		img = ebiten.NewImageFromImage(session.Plot(c.ys, c.mark, c.r.Dx(), c.r.Dy(), max(1, int(d.px(1.5))), lit, color.RGBA{0x2a, 0x3a, 0x48, 0xff}))
		v.plots[key] = img
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(c.r.Min.X), float64(c.r.Min.Y))
	dst.DrawImage(img, op)
}

func (v *devices) drawCell(dst *ebiten.Image, d *desk, cell devCell) {
	c, r := cell.c, cell.r
	ink, label := color.Color(ink), color.Color(dim)
	if !cell.can {
		ink, label = faint, faint
	}
	cx := float64(r.Min.X + r.Dx()/2)
	drawText(dst, d.fonts.text, fit(d.fonts.text, c.Label, float64(r.Dx())-d.px(4)), cx, float64(r.Min.Y)+d.px(7), label, text.AlignCenter)
	val := v.valueRect(d, r)
	mid := image.Rect(r.Min.X+int(d.px(4)), r.Min.Y+int(d.px(15)), r.Max.X-int(d.px(4)), val.Min.Y)
	value := ""
	switch c.Kind {
	case session.Knob:
		x := session.Effective(c.Param, c.Value)
		value = session.FormatParam(c.Param, x)
		if c.Value == 0 && c.Param.Zero != 0 {
			ink = faint
		}
		pos := session.Position(c.Param, c.Value, d.songBeats())
		kx, ky := float32(cx), float32(mid.Min.Y+mid.Dy()/2)
		rad := float32(d.px(cellKnobR))
		vector.FillCircle(dst, kx, ky, rad, button, true)
		from := 0.75 * math.Pi
		var path vector.Path
		path.Arc(kx, ky, rad-float32(d.px(1)), float32(from), float32(from+pos*1.5*math.Pi), vector.Clockwise)
		so := &vector.StrokeOptions{Width: float32(d.px(2))}
		col := color.Color(lit)
		if !cell.can {
			col = faint
		}
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(col)
		vector.StrokePath(dst, &path, so, op)
		angle := from + pos*1.5*math.Pi
		vector.StrokeLine(dst, kx, ky, kx+float32(math.Cos(angle))*rad, ky+float32(math.Sin(angle))*rad, float32(d.px(2)), ink, true)
	case session.Cycle:
		if i := int(c.Value); i >= 0 && i < len(c.Names) {
			value = c.Names[i]
		}
		fill(dst, mid.Inset(int(d.px(3))), button)
		drawText(dst, d.fonts.text, fit(d.fonts.text, value, float64(mid.Dx())-d.px(8)), cx, float64(mid.Min.Y+mid.Dy()/2), ink, text.AlignCenter)
		value = ""
	case session.Toggle:
		label, bg := "off", color.Color(button)
		if c.On {
			label, bg = "on", lit
		}
		fill(dst, mid.Inset(int(d.px(3))), bg)
		drawText(dst, d.fonts.text, label, cx, float64(mid.Min.Y+mid.Dy()/2), ink, text.AlignCenter)
	case session.Text:
		fill(dst, mid.Inset(int(d.px(3))), edge)
		drawText(dst, d.fonts.text, fit(d.fonts.text, c.Text, float64(mid.Dx())-d.px(8)), cx, float64(mid.Min.Y+mid.Dy()/2), ink, text.AlignCenter)
	case session.Points, session.Note:
		drawText(dst, d.fonts.text, fit(d.fonts.text, c.Text, float64(r.Dx())-d.px(4)), cx, float64(mid.Min.Y+mid.Dy()/2), faint, text.AlignCenter)
	}
	id := fmt.Sprint(c.Path...)
	if v.typing != nil && v.typing.id == id {
		fill(dst, val, edge)
		value, ink = v.typing.text+"▏", playhead
	}
	if value != "" {
		drawText(dst, d.fonts.num, fit(d.fonts.num, value, float64(r.Dx())), cx, float64(val.Min.Y+val.Dy()/2), ink, text.AlignCenter)
	}
}

// browse adds the browser's patches and effects, which drag onto a track, the
// arrangement or a chain.
func (v *devices) browse(d *desk, list *ebui.ScrollableContainer) {
	v.sources = nil
	if v.folded == nil {
		v.folded = map[string]bool{}
	}
	section := func(title string, names []string, carry func(string) carried) {
		label := title
		if v.folded[title] {
			label += " …"
		}
		list.AddChild(ebui.NewButton(
			ebui.WithSize(d.px(browserW-24), d.px(itemH)),
			ebui.WithLabelText(label),
			ebui.WithRole(roleItem),
			ebui.WithClickHandler(func() {
				v.folded[title] = !v.folded[title]
				d.stale = true
			}),
		))
		if v.folded[title] {
			return
		}
		for _, name := range names {
			b := ebui.NewButton(
				ebui.WithSize(d.px(browserW-24), d.px(itemH)),
				ebui.WithLabelText("  "+fit(d.fonts.text, name, d.px(browserW-48))),
				ebui.WithRole(roleItem),
			)
			list.AddChild(b)
			v.sources = append(v.sources, source{b, carry(name)})
		}
	}
	section("Patches", session.PatchNames(), func(n string) carried { return carried{patch: n} })
	section("Effects", session.EffectNames(), func(n string) carried { return carried{effect: n} })
}

// carrying follows what's dragged from the browser and drops it where the button goes
// up, reporting whether something is being carried. It also sees a newly selected
// track, whose chain the devices then show.
func (v *devices) carrying(d *desk) bool {
	if d.view.Track != v.track {
		v.track = d.view.Track
		d.view.Bus, d.view.Master = "", false
	}
	x, y := ebiten.CursorPosition()
	pt := image.Pt(x, y)
	if v.carry == nil {
		if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || !pt.In(d.widgets.browser) {
			return false
		}
		for _, s := range v.sources {
			if s.b.Contains(float64(x), float64(y)) {
				c := s.carried
				v.carry = &c
				return true
			}
		}
		return false
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		return true
	}
	c := *v.carry
	v.carry = nil
	if d.doc == nil {
		d.note = d.readOnly()
		return true
	}
	v.drop(d, c, pt)
	return true
}

// drop puts a patch onto the track under pt, or a new track, and an effect into the
// chain under pt.
func (v *devices) drop(d *desk, c carried, pt image.Point) {
	a := d.layout.Arrangement
	track := func(name string) int {
		return slices.IndexFunc(a.Tracks, func(t synth.Track) bool { return t.Name == name })
	}
	switch {
	case pt.In(d.widgets.detail) && d.view.Devices:
		view, ok := v.layout(d)
		switch {
		case !ok:
		case c.effect != "":
			v.addEffect(d, view.chain, v.slot(view, float64(pt.X)), c.effect)
		case view.chain.Kind == session.TrackChain:
			v.setPatch(d, view.chain.Index, c.patch)
		}
	case pt.In(d.widgets.timeline) && d.view.Mixer:
		for _, s := range d.mx.strips(d, d.widgets.timeline) {
			if !pt.In(s.r) {
				continue
			}
			ch := session.Chain{Kind: session.MasterChain}
			switch s.kind {
			case trackStrip:
				ch = session.Chain{Kind: session.TrackChain, Index: s.index}
			case busStrip:
				ch = session.Chain{Kind: session.BusChain, Index: s.index}
			}
			switch {
			case c.effect != "":
				v.addEffect(d, ch, math.MaxInt, c.effect)
			case s.kind == trackStrip:
				v.setPatch(d, s.index, c.patch)
			}
			return
		}
		if c.patch != "" {
			v.addTrack(d, c.patch)
		}
	case pt.In(d.widgets.timeline) && pt.X >= d.tl.r.Min.X && float64(pt.Y) >= d.tl.top(d):
		for _, r := range d.tl.rows(d) {
			if float64(pt.Y) < r.y || float64(pt.Y) >= r.y+r.h {
				continue
			}
			if r.lane < 0 {
				return
			}
			i := track(d.layout.Lanes[r.lane].Name)
			if c.effect != "" {
				v.addEffect(d, session.Chain{Kind: session.TrackChain, Index: i}, math.MaxInt, c.effect)
			} else {
				v.setPatch(d, i, c.patch)
			}
			return
		}
		if c.patch != "" {
			v.addTrack(d, c.patch)
		}
	}
}

func (v *devices) setPatch(d *desk, track int, name string) {
	if !d.editable("Tracks", track) {
		d.note = "the track is " + custom
		return
	}
	if d.doc.SetPatch(track, name) {
		d.view.Track = d.doc.Now().Tracks[track].Name
		d.changed()
	}
}

func (v *devices) addTrack(d *desk, patch string) {
	if !d.editable("Tracks") {
		d.note = "the tracks are " + custom
		return
	}
	if name, ok := d.doc.AddTrack(patch); ok {
		d.view.Track = name
		d.changed()
	}
}

func (v *devices) addEffect(d *desk, ch session.Chain, at int, name string) {
	e, ok := session.NewEffect(name)
	if !ok {
		return
	}
	if !d.editable(ch.Path()...) {
		d.note = "the chain is " + custom
		return
	}
	d.edit(func(a *synth.Arrangement) { session.AddEffect(a, ch, at, e) })
}

// drawCarry draws what's being dragged from the browser at the cursor.
func (v *devices) drawCarry(dst *ebiten.Image, d *desk) {
	if v.carry == nil {
		return
	}
	name := v.carry.patch + v.carry.effect
	x, y := ebiten.CursorPosition()
	w := text.Advance(name, d.fonts.text) + d.px(16)
	r := rect(float64(x)+d.px(12), float64(y)+d.px(8), w, d.px(itemH))
	fill(dst, r, lit)
	drawText(dst, d.fonts.text, name, float64(r.Min.X)+d.px(8), float64(r.Min.Y+r.Dy()/2), ink, text.AlignStart)
}

// pickStrip selects the chain of a mixer strip whose name or inserts were clicked,
// showing the devices for inserts, and reports whether it took the click.
func (v *devices) pickStrip(d *desk, m *mixer, s strip, pt image.Point) bool {
	fx := image.Rect(s.r.Min.X, m.insert(d, s, 0).Min.Y, s.r.Max.X, m.insert(d, s, inserts-1).Max.Y)
	if !pt.In(fx) && !pt.In(m.header(d, s)) {
		return false
	}
	switch s.kind {
	case trackStrip:
		d.view.Track, d.view.Bus, d.view.Master = s.name, "", false
		v.track = s.name
	case busStrip:
		d.view.Bus, d.view.Master = s.name, false
	case masterStrip:
		d.view.Bus, d.view.Master = "", true
	}
	if pt.In(fx) {
		d.view.Devices = true
		return true
	}
	return s.kind != trackStrip
}
