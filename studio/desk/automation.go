package desk

import (
	"fmt"
	"image"
	"math"
	"slices"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/studio/session"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// autoH is an automation lane's height, in points.
const autoH = 44

// automation is the timeline's automation lanes: the point last clicked, and the drag
// moving it.
type automation struct {
	lane  string // the selected point's lane, by its key
	point int
	drag  *pointDrag
}

type pointDrag struct {
	lane  session.Automation
	point int
}

// addLanes adds the lanes of the chain (a track's, a bus's, or the master's with
// neither) beneath what was added last.
func (t *timeline) addLanes(d *desk, add func(row), track, bus, group string) {
	if !d.view.Lanes {
		return
	}
	for k, l := range d.layout.Automation {
		if l.Track == track && l.Bus == bus {
			add(row{lane: -1, group: group, auto: &d.layout.Automation[k], h: d.px(autoH)})
		}
	}
}

// addOtherLanes adds the lanes of the buses that aren't groups, then the master's.
func (t *timeline) addOtherLanes(d *desk, add func(row)) {
	groups := d.layout.Groups()
	for _, b := range d.layout.Arrangement.Buses {
		if !slices.Contains(groups, b.Name) {
			t.addLanes(d, add, "", b.Name, "")
		}
	}
	t.addLanes(d, add, "", "", "")
}

func (t *timeline) beat(d *desk) float64 { return 60 / math.Max(d.layout.Tempo, 1e-9) }

func (t *timeline) pointX(d *desk, at float64) float64 { return t.x(d, at*t.beat(d)) }

// valueY is where v sits in the row, top for the param's Max.
func valueY(r row, p synth.Param, v float64, inset float64) float64 {
	return r.y + r.h - inset - session.Position(p, v, 0)*(r.h-2*inset)
}

// snap is the grid in beats a point lands on: the finest that keeps points apart.
func (t *timeline) snap(d *desk) float64 {
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		return 1.0 / 64
	}
	beatPx := t.beat(d) * t.pps(d)
	g := 1.0 / 16
	for g*beatPx < d.px(12) && g < 64 {
		g *= 2
	}
	return g
}

// pointAt is the point of the lane's row under (x, y).
func (t *timeline) pointAt(d *desk, r row, x, y float64) int {
	l := *r.auto
	inset := d.px(6)
	for i := len(l.Points) - 1; i >= 0; i-- {
		p := l.Points[i]
		if math.Abs(x-t.pointX(d, p.At)) <= d.px(grabPx+1) && math.Abs(y-valueY(r, l.Param, p.Value, inset)) <= d.px(grabPx+1) {
			return i
		}
	}
	return -1
}

// point is where (x, y) puts a point in the lane's row, snapped.
func (t *timeline) point(d *desk, r row, x, y float64) synth.Point {
	inset := d.px(6)
	at := session.Round(math.Max(0, t.at(d, x)/t.beat(d)), t.snap(d))
	pos := (r.y + r.h - inset - y) / (r.h - 2*inset)
	return synth.Point{At: at, Value: session.FromPosition(r.auto.Param, pos, 0)}
}

// automating handles the cursor over the automation lanes, reporting whether it's over
// one.
func (t *timeline) automating(d *desk, x, y float64) bool {
	left := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	var r row
	for _, rr := range t.rows(d) {
		if rr.auto != nil && y >= rr.y && y < rr.y+rr.h && y >= t.top(d) {
			r = rr
		}
	}
	if r.auto == nil {
		if left {
			t.auto.lane = ""
		}
		return false
	}
	if x < t.left(d) {
		return true
	}
	l := *r.auto
	path := l.Path(d.layout.Arrangement)
	switch {
	case d.doc == nil:
		t.hint = d.readOnly()
		return true
	case !d.editable(path...):
		t.hint = custom
		return true
	}
	i := t.pointAt(d, r, x, y)
	t.hint = "click adds a point · drag moves it (shift: off the grid) · delete removes it · A hides the lanes"
	if i >= 0 {
		ebiten.SetCursorShape(ebiten.CursorShapeMove)
		t.hint = t.pointHint(d, l.Param, l.Points[i]) + " · drag moves it (shift: off the grid) · delete removes it"
	}
	if !left {
		return true
	}
	d.doc.Hold()
	if i < 0 {
		p := t.point(d, r, x, y)
		if !d.arrange(func(a *synth.Arrangement) error {
			var err error
			i, err = session.AddPoint(a, l, p)
			return err
		}) {
			d.doc.Release()
			return true
		}
	}
	t.auto.lane, t.auto.point = l.Key(), i
	t.auto.drag = &pointDrag{lane: l, point: i}
	t.sel = playRef{}
	return true
}

// draggingPoint follows a point's drag until the button comes up.
func (t *timeline) draggingPoint(d *desk) {
	g := t.auto.drag
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || d.doc == nil {
		t.auto.drag = nil
		if d.doc != nil {
			d.doc.Release()
		}
		return
	}
	ebiten.SetCursorShape(ebiten.CursorShapeMove)
	var r row
	for _, rr := range t.rows(d) {
		if rr.auto != nil && rr.auto.Key() == g.lane.Key() {
			r = rr
		}
	}
	if r.auto == nil || g.point >= len(r.auto.Points) {
		return
	}
	x, y := ebiten.CursorPosition()
	p := t.point(d, r, float64(x), float64(y))
	if p == r.auto.Points[g.point] {
		return
	}
	moved := g.point
	if d.arrange(func(a *synth.Arrangement) error {
		var err error
		moved, err = session.MovePoint(a, g.lane, g.point, p)
		return err
	}) {
		g.point, t.auto.point = moved, moved
	}
	t.hint = t.pointHint(d, r.auto.Param, p)
}

// deletePoint deletes the selected point, reporting whether there was one.
func (t *timeline) deletePoint(d *desk) bool {
	k := slices.IndexFunc(d.layout.Automation, func(l session.Automation) bool { return l.Key() == t.auto.lane })
	if k < 0 || t.auto.point >= len(d.layout.Automation[k].Points) {
		t.auto.lane = ""
		return false
	}
	l := d.layout.Automation[k]
	if d.editable(l.Path(d.layout.Arrangement)...) {
		d.arrange(func(a *synth.Arrangement) error { return session.DeletePoint(a, l, t.auto.point) })
	}
	t.auto.lane = ""
	return true
}

func (d *desk) toggleLanes() {
	d.view.Lanes = !d.view.Lanes
	d.tl.clampScroll(d)
}

func (t *timeline) pointHint(d *desk, p synth.Param, pt synth.Point) string {
	return formatParam(p, pt.Value) + " at " + session.Format(d.layout.Song, pt.At*t.beat(d))
}

func formatParam(p synth.Param, v float64) string {
	s := fmt.Sprintf("%g", v)
	if u := p.Unit.String(); u != "" {
		s += " " + u
	}
	return s
}

// drawLane draws an automation lane: its name in the header, and its curve on the
// param's scale, held flat beyond its first and last points.
func (t *timeline) drawLane(dst *ebiten.Image, d *desk, r row) {
	l := *r.auto
	x := float64(t.r.Min.X)
	fillF(dst, x, r.y, d.px(headerW), r.h, ground)
	line(dst, x, r.y+r.h, float64(t.r.Max.X), r.y+r.h, 1, edge)
	indent := d.px(12)
	if r.group != "" {
		indent += d.px(10)
	}
	owner := l.Track
	switch {
	case l.Bus != "":
		owner = l.Bus
	case owner == "":
		owner = "master"
	}
	c := hue(owner, 0.6)
	if owner == "master" {
		c = dim
	}
	drawText(dst, d.fonts.text, fit(d.fonts.text, "↳ "+l.Name, d.px(headerW)-indent-d.px(8)), x+indent, r.y+d.px(12), c, text.AlignStart)
	drawText(dst, d.fonts.num, fit(d.fonts.num, formatParam(l.Param, l.Param.Max)+" · "+formatParam(l.Param, l.Param.Min), d.px(headerW)-indent-d.px(8)),
		x+indent, r.y+r.h-d.px(12), faint, text.AlignStart)

	body := dst.SubImage(image.Rect(int(t.left(d)), int(r.y), t.r.Max.X, int(r.y+r.h))).(*ebiten.Image)
	inset := d.px(6)
	if len(l.Points) == 0 {
		drawText(body, d.fonts.text, "click to add a point", t.left(d)+d.px(8), r.y+r.h/2, faint, text.AlignStart)
		return
	}
	curve := hue(owner, 0.8)
	if owner == "master" {
		curve = ink
	}
	prevX, prevY := t.left(d), valueY(r, l.Param, l.Points[0].Value, inset)
	for _, p := range l.Points {
		px, py := t.pointX(d, p.At), valueY(r, l.Param, p.Value, inset)
		line(body, prevX, prevY, px, py, d.px(1.5), curve)
		prevX, prevY = px, py
	}
	line(body, prevX, prevY, float64(t.r.Max.X), prevY, d.px(1.5), curve)
	s := d.px(3)
	for i, p := range l.Points {
		px, py := t.pointX(d, p.At), valueY(r, l.Param, p.Value, inset)
		pc := curve
		if t.auto.lane == l.Key() && t.auto.point == i {
			pc = playhead
		}
		fillF(body, px-s, py-s, 2*s, 2*s, pc)
	}
}
