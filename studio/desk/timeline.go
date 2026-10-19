package desk

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"

	"github.com/cbodonnell/synth/studio/session"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Sizes in points, before the window's scale.
const (
	headerW = 170
	rulerH  = 40
	laneH   = 46
	groupH  = 22
	minZoom = 2 // points a second
	maxZoom = 4000
)

// timeline is the arrangement view: the section ruler over a lane per track, group
// buses folding their tracks under them, and the playhead across it all.
type timeline struct {
	r image.Rectangle // all of it, in window pixels
}

// row is a lane, or a group bus's header over its tracks.
type row struct {
	lane  int // -1 for a group
	group string
	y, h  float64
}

func (t *timeline) place(d *desk) {
	t.r = d.widgets.timeline
	if d.view.Zoom == 0 {
		t.fit(d)
	}
}

func (t *timeline) left(d *desk) float64  { return float64(t.r.Min.X) + d.px(headerW) }
func (t *timeline) top(d *desk) float64   { return float64(t.r.Min.Y) + d.px(rulerH) }
func (t *timeline) width(d *desk) float64 { return float64(t.r.Max.X) - t.left(d) }

func (t *timeline) pps(d *desk) float64 { return d.view.Zoom * d.scale }

func (t *timeline) x(d *desk, at float64) float64 {
	return t.left(d) + (at-d.view.Scroll)*t.pps(d)
}

func (t *timeline) at(d *desk, x float64) float64 {
	return d.view.Scroll + (x-t.left(d))/t.pps(d)
}

// fit zooms to show the whole song.
func (t *timeline) fit(d *desk) {
	w := t.width(d) - d.px(16)
	if w <= 0 {
		return
	}
	length := math.Max(d.layout.Length(), 0.1)
	d.view.Zoom = clampZoom(w / length / d.scale)
	d.view.Scroll = 0
}

func clampZoom(z float64) float64 { return math.Max(minZoom, math.Min(maxZoom, z)) }

func (t *timeline) rows(d *desk) []row {
	var rows []row
	y := t.top(d) - d.view.Top
	add := func(r row) {
		r.y = y
		y += r.h
		rows = append(rows, r)
	}
	done := map[string]bool{}
	for i, lane := range d.layout.Lanes {
		if lane.Out == "" {
			add(row{lane: i, h: d.px(laneH)})
			continue
		}
		if done[lane.Out] {
			continue
		}
		done[lane.Out] = true
		add(row{lane: -1, group: lane.Out, h: d.px(groupH)})
		if slices.Contains(d.view.Folded, lane.Out) {
			continue
		}
		for j, other := range d.layout.Lanes {
			if other.Out == lane.Out {
				add(row{lane: j, group: lane.Out, h: d.px(laneH)})
			}
		}
	}
	return rows
}

func (t *timeline) contentHeight(d *desk) float64 {
	rows := t.rows(d)
	if len(rows) == 0 {
		return 0
	}
	last := rows[len(rows)-1]
	return last.y + last.h - (t.top(d) - d.view.Top)
}

func (t *timeline) update(d *desk) {
	if d.sounding {
		// Turn the page when the playhead runs off it.
		if px := t.x(d, d.view.At); px > float64(t.r.Max.X) || px < t.left(d) {
			d.view.Scroll = math.Max(0, d.view.At-d.px(24)/t.pps(d))
		}
	}

	x, y := ebiten.CursorPosition()
	fx, fy := float64(x), float64(y)
	if !image.Pt(x, y).In(t.r) {
		return
	}

	if wx, wy := ebiten.Wheel(); wx != 0 || wy != 0 {
		command := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
		switch {
		case command:
			at := t.at(d, math.Max(fx, t.left(d)))
			d.view.Zoom = clampZoom(d.view.Zoom * math.Pow(1.1, wy))
			d.view.Scroll = at - (math.Max(fx, t.left(d))-t.left(d))/t.pps(d)
		case ebiten.IsKeyPressed(ebiten.KeyShift) || wx != 0:
			dx := wx
			if dx == 0 {
				dx = wy
			}
			d.view.Scroll -= dx * d.px(20) / t.pps(d)
		default:
			d.view.Top -= wy * d.px(20)
		}
		t.clampScroll(d)
	}

	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	if fy < t.top(d) {
		if fx >= t.left(d) {
			d.seek(t.at(d, fx))
		}
		return
	}
	for _, r := range t.rows(d) {
		if fy < r.y || fy >= r.y+r.h {
			continue
		}
		if r.lane < 0 {
			if i := slices.Index(d.view.Folded, r.group); i >= 0 {
				d.view.Folded = slices.Delete(d.view.Folded, i, i+1)
			} else {
				d.view.Folded = append(d.view.Folded, r.group)
			}
			t.clampScroll(d)
			return
		}
		name := d.layout.Lanes[r.lane].Name
		d.view.Track = name
		if s := t.soloBox(d, r); image.Pt(x, y).In(s) {
			d.toggleSolo(name)
		}
		return
	}
}

func (t *timeline) clampScroll(d *desk) {
	end := d.layout.Length()
	d.view.Scroll = math.Max(0, math.Min(d.view.Scroll, end-t.width(d)/t.pps(d)/2))
	view := float64(t.r.Max.Y) - t.top(d)
	d.view.Top = math.Max(0, math.Min(d.view.Top, t.contentHeight(d)-view))
}

// selectNext moves the selection by n tracks, as they're shown.
func (t *timeline) selectNext(d *desk, n int) {
	var shown []string
	for _, r := range t.rows(d) {
		if r.lane >= 0 {
			shown = append(shown, d.layout.Lanes[r.lane].Name)
		}
	}
	if len(shown) == 0 {
		return
	}
	i := slices.Index(shown, d.view.Track)
	if i < 0 {
		d.view.Track = shown[0]
		return
	}
	d.view.Track = shown[max(0, min(len(shown)-1, i+n))]
}

func (t *timeline) soloBox(d *desk, r row) image.Rectangle {
	s := int(d.px(16))
	x := t.r.Min.X + int(d.px(headerW-48))
	y := int(r.y + d.px(8))
	return image.Rect(x, y, x+s, y+s)
}

func (t *timeline) muteBox(d *desk, r row) image.Rectangle {
	return t.soloBox(d, r).Add(image.Pt(int(d.px(20)), 0))
}

func (t *timeline) draw(dst *ebiten.Image, d *desk) {
	if t.r.Empty() {
		return
	}
	area := dst.SubImage(t.r).(*ebiten.Image)
	fill(area, t.r, ground)
	lanes := area.SubImage(image.Rect(t.r.Min.X, int(t.top(d)), t.r.Max.X, t.r.Max.Y)).(*ebiten.Image)
	body := lanes.SubImage(image.Rect(int(t.left(d)), int(t.top(d)), t.r.Max.X, t.r.Max.Y)).(*ebiten.Image)

	rows := t.rows(d)
	for i, r := range rows {
		c := laneEven
		if i%2 == 1 {
			c = laneOdd
		}
		if r.lane >= 0 && d.layout.Lanes[r.lane].Name == d.view.Track {
			c = selected
		}
		fillF(lanes, float64(t.r.Min.X), r.y, float64(t.r.Dx()), r.h, c)
	}
	t.drawGrid(body, d, t.top(d), float64(t.r.Max.Y))
	for _, r := range rows {
		if r.lane >= 0 {
			t.drawBlocks(body, d, r)
		}
		t.drawHeader(lanes, d, r)
	}
	line(lanes, t.left(d), t.top(d), t.left(d), float64(t.r.Max.Y), 1, edge)
	t.drawRuler(area, d)

	if px := t.x(d, d.view.At); px >= t.left(d) && px <= float64(t.r.Max.X) {
		line(area, px, float64(t.r.Min.Y), px, float64(t.r.Max.Y), d.px(1.5), playhead)
	}
}

// grid is where the timeline's lines fall: bars and beats for a song, or a tick in
// seconds for a sound.
func (t *timeline) drawGrid(dst *ebiten.Image, d *desk, top, bottom float64) {
	s := d.layout.Song
	from, to := t.at(d, t.left(d)), t.at(d, float64(t.r.Max.X))
	if !session.Timed(s) {
		tick := t.tick(d)
		for at := math.Floor(from/tick) * tick; at <= to; at += tick {
			line(dst, t.x(d, at), top, t.x(d, at), bottom, 1, barLine)
		}
		return
	}
	beatPx := s.Meter.Beat * t.pps(d)
	every := t.barEvery(d)
	for b := session.BarAt(s.Meter, math.Max(from, 0)); ; b++ {
		at := session.BarTime(s.Meter, b)
		if at > to {
			break
		}
		if b%every == 0 {
			line(dst, t.x(d, at), top, t.x(d, at), bottom, 1, barLine)
		}
		if beatPx >= d.px(10) {
			for k := 1; k < s.Meter.BarBeats(b); k++ {
				x := t.x(d, at+float64(k)*s.Meter.Beat)
				line(dst, x, top, x, bottom, 1, beatLine)
			}
		}
	}
}

// barEvery is how many bars apart the ruler numbers them, so numbers don't crowd.
func (t *timeline) barEvery(d *desk) int {
	s := d.layout.Song
	barPx := s.Meter.Beat * float64(s.Meter.Cycle()) / float64(max(len(s.Meter.Bars), 1)) * t.pps(d)
	every := 1
	for barPx*float64(every) < d.px(28) {
		every *= 2
	}
	return every
}

// tick is the seconds between a sound's ruler marks.
func (t *timeline) tick(d *desk) float64 {
	for _, tick := range []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10} {
		if tick*t.pps(d) >= d.px(60) {
			return tick
		}
	}
	return 30
}

func (t *timeline) drawRuler(dst *ebiten.Image, d *desk) {
	s := d.layout.Song
	top, mid, bottom := float64(t.r.Min.Y), float64(t.r.Min.Y)+d.px(rulerH/2), t.top(d)
	fillF(dst, float64(t.r.Min.X), top, float64(t.r.Dx()), bottom-top, panel)
	ruler := dst.SubImage(image.Rect(int(t.left(d)), t.r.Min.Y, t.r.Max.X, int(bottom))).(*ebiten.Image)

	for i, sec := range d.layout.Sections {
		x0, x1 := t.x(d, sec.From), t.x(d, sec.To)
		c := laneEven
		if i%2 == 1 {
			c = laneOdd
		}
		fillF(ruler, x0, top, x1-x0, mid-top, c)
		line(ruler, x0, top, x0, mid, 1, barLine)
		drawText(ruler, d.fonts.text, fit(d.fonts.text, sec.Name, x1-x0-d.px(8)), x0+d.px(5), (top+mid)/2, ink, text.AlignStart)
	}
	if s.Loop {
		x0, x1 := t.x(d, s.LoopStart), t.x(d, s.Length)
		c := faint
		if d.view.Loop {
			c = loopColor
		}
		y := mid + d.px(3)
		line(ruler, x0, y, x1, y, d.px(2), c)
		line(ruler, x0, y, x0, y+d.px(6), d.px(2), c)
		line(ruler, x1, y, x1, y+d.px(6), d.px(2), c)
	}

	if !session.Timed(s) {
		tick := t.tick(d)
		from, to := t.at(d, t.left(d)), t.at(d, float64(t.r.Max.X))
		for at := math.Max(0, math.Floor(from/tick)*tick); at <= to; at += tick {
			x := t.x(d, at)
			line(ruler, x, bottom-d.px(6), x, bottom, 1, dim)
			drawText(ruler, d.fonts.num, fmt.Sprintf("%.0fms", at*1000), x+d.px(3), (mid+bottom)/2+d.px(2), dim, text.AlignStart)
		}
	} else {
		every := t.barEvery(d)
		for b := 0; b < session.Bars(s); b += every {
			at := session.BarTime(s.Meter, b)
			x := t.x(d, at)
			if x > float64(t.r.Max.X) {
				break
			}
			line(ruler, x, bottom-d.px(6), x, bottom, 1, dim)
			drawText(ruler, d.fonts.num, fmt.Sprint(b+1), x+d.px(3), (mid+bottom)/2+d.px(2), dim, text.AlignStart)
		}
	}
	line(dst, float64(t.r.Min.X), bottom, float64(t.r.Max.X), bottom, 1, edge)
	x := float64(t.r.Min.X) + d.px(8)
	drawText(dst, d.fonts.text, fit(d.fonts.text, d.layout.Name, d.px(headerW-16)), x, (top+mid)/2, ink, text.AlignStart)
	t.drawMaster(dst, d, x, (mid+bottom)/2)
}

// drawMaster meters the mix at the playhead: its peak, and its short-term loudness.
func (t *timeline) drawMaster(dst *ebiten.Image, d *desk, x, y float64) {
	w := d.px(70)
	fillF(dst, x, y-d.px(2), w, d.px(4), edge)
	if d.take == nil || !d.sounding {
		return
	}
	g := d.take.Master
	db := 20 * math.Log10(float64(g.Peak(d.frame, 0.05)))
	c := meterLow
	if db > -1 {
		c = meterHigh
	}
	fillF(dst, x, y-d.px(2), w*math.Max(0, math.Min(1, (db+60)/60)), d.px(4), c)
	if lufs := g.ShortTerm(d.frame); !math.IsInf(lufs, -1) {
		drawText(dst, d.fonts.num, fmt.Sprintf("%.1f", lufs), x+w+d.px(6), y, dim, text.AlignStart)
	}
}

func (t *timeline) drawHeader(dst *ebiten.Image, d *desk, r row) {
	x := float64(t.r.Min.X)
	w := d.px(headerW)
	mid := r.y + r.h/2
	if r.lane < 0 {
		mark := "▾"
		if slices.Contains(d.view.Folded, r.group) {
			mark = "▸"
		}
		fillF(dst, x, r.y, w, r.h, panel)
		drawText(dst, d.fonts.text, mark+" "+r.group, x+d.px(8), mid, dim, text.AlignStart)
		return
	}
	lane := d.layout.Lanes[r.lane]
	indent := 0.0
	if r.group != "" {
		indent = d.px(10)
	}
	fillF(dst, x+indent, r.y+d.px(4), d.px(4), r.h-d.px(8), hue(lane.Name, 0.75))
	c := ink
	if lane.Mute {
		c = faint
	}
	drawText(dst, d.fonts.text, fit(d.fonts.text, lane.Name, w-d.px(64)-indent), x+indent+d.px(12), r.y+d.px(16), c, text.AlignStart)

	box := func(b image.Rectangle, label string, on bool, onColor color.Color) {
		bg, fg := color.Color(button), color.Color(dim)
		if on {
			bg, fg = onColor, ground
		}
		fill(dst, b, bg)
		drawText(dst, d.fonts.text, label, float64(b.Min.X+b.Dx()/2), float64(b.Min.Y+b.Dy()/2), fg, text.AlignCenter)
	}
	box(t.soloBox(d, r), "S", slices.Contains(d.view.Solo, lane.Name), soloColor)
	box(t.muteBox(d, r), "M", lane.Mute, muteColor)

	t.drawMeter(dst, d, lane.Name, x+indent+d.px(12), r.y+r.h-d.px(12), w-d.px(64)-indent)
}

// drawMeter draws the track's peak at the playhead, -60 to 0 dB.
func (t *timeline) drawMeter(dst *ebiten.Image, d *desk, name string, x, y, w float64) {
	h := d.px(4)
	fillF(dst, x, y, w, h, edge)
	if d.take == nil || !d.sounding {
		return
	}
	g, ok := d.take.Parts[name]
	if !ok {
		return
	}
	db := 20 * math.Log10(float64(g.Peak(d.take.SongFrame(d.frame), 0.05)))
	level := math.Max(0, math.Min(1, (db+60)/60))
	c := meterLow
	if db > -1 {
		c = meterHigh
	}
	fillF(dst, x, y, w*level, h, c)
}

func (t *timeline) drawBlocks(dst *ebiten.Image, d *desk, r row) {
	lane := d.layout.Lanes[r.lane]
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, b := range lane.Blocks {
		for _, n := range b.Notes {
			lo, hi = math.Min(lo, n.Key), math.Max(hi, n.Key+n.Slide)
			lo = math.Min(lo, n.Key+n.Slide)
		}
	}
	span := math.Max(hi-lo, 11)
	right := float64(t.r.Max.X)

	body, label, note := hue(lane.Name, 0.42), hue(lane.Name, 0.62), hue(lane.Name, 0.92)
	if lane.Mute {
		body, label, note = hsv(0, 0, 0.25), hsv(0, 0, 0.32), hsv(0, 0, 0.5)
	}
	for _, b := range lane.Blocks {
		x0, x1 := t.x(d, b.From), t.x(d, b.To)
		if x1 < t.left(d) || x0 > right {
			continue
		}
		y0, y1 := r.y+d.px(2), r.y+r.h-d.px(2)
		fillF(dst, x0+1, y0, x1-x0-1, y1-y0, body)
		strip := d.px(14)
		fillF(dst, x0+1, y0, x1-x0-1, strip, label)
		drawText(dst, d.fonts.text, fit(d.fonts.text, b.Clip, x1-x0-d.px(8)), x0+d.px(4), y0+strip/2, ground, text.AlignStart)

		top, bottom := y0+strip+d.px(2), y1-d.px(2)
		keyH := math.Max((bottom-top)/(span+1), 1)
		for _, n := range b.Notes {
			nx0, nx1 := t.x(d, n.At), t.x(d, n.At+n.Dur)
			if nx1 < t.left(d) || nx0 > right {
				continue
			}
			ny := bottom - (n.Key-lo+1)*(bottom-top)/(span+1)
			if b.Hit {
				fillF(dst, nx0, ny, math.Max(d.px(2), math.Min(nx1-nx0, d.px(4))), keyH, note)
				continue
			}
			fillF(dst, nx0, ny, math.Max(nx1-nx0-1, 1), keyH, note)
			if n.Slide != 0 {
				ey := bottom - (n.Key+n.Slide-lo+1)*(bottom-top)/(span+1)
				line(dst, nx0, ny+keyH/2, nx1, ey+keyH/2, 1, note)
			}
		}
	}
}
