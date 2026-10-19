package desk

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/cbodonnell/ebui"
	"github.com/cbodonnell/synth/studio/session"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Sizes in points, before the window's scale.
const (
	transportH = 40
	browserW   = 190
	statusH    = 24
	buttonH    = 26
	itemH      = 24
	outputRows = 14
)

// More roles, for the browser's list.
const (
	roleItem   = "item"
	roleItemOn = "item on"
)

// widgets are the controls ebui lays out: the transport along the top, the browser down
// the left, and the status bar along the bottom. The timeline fills what's left.
type widgets struct {
	root *ebui.BaseContainer

	play, loop, click                               *ebui.Button
	position, tempo, meter, key, loudness, progress *ebui.Label
	items                                           map[string]*ebui.Button
	status                                          *ebui.Label

	transport, browser, statusBar, timeline image.Rectangle
}

func buildWidgets(d *desk) *widgets {
	px := d.px
	w, h := float64(d.w), float64(d.h)
	ws := &widgets{root: ebui.NewBaseContainer(ebui.WithSize(w, h)), items: map[string]*ebui.Button{}}

	ws.transport = image.Rect(0, 0, d.w, int(px(transportH)))
	ws.statusBar = image.Rect(0, d.h-int(px(statusH)), d.w, d.h)
	left := 0
	if d.view.Browser {
		left = int(px(browserW))
		ws.browser = image.Rect(0, ws.transport.Max.Y, left, ws.statusBar.Min.Y)
	}
	ws.timeline = image.Rect(left, ws.transport.Max.Y, d.w, ws.statusBar.Min.Y)

	bar := ebui.NewLayoutContainer(
		ebui.WithSize(w, px(transportH)),
		ebui.WithPadding(0, px(8), 0, px(8)),
		ebui.WithLayout(ebui.NewHorizontalStackLayout(px(6), ebui.AlignStart)),
	)
	button := func(label string, width float64, do func()) *ebui.Button {
		b := ebui.NewButton(ebui.WithSize(px(width), px(buttonH)), ebui.WithLabelText(label), ebui.WithClickHandler(do))
		bar.AddChild(b)
		return b
	}
	label := func(width float64) *ebui.Label {
		l := ebui.NewLabel("", ebui.WithSize(px(width), px(buttonH)), ebui.WithColor(ink), ebui.WithJustify(ebui.JustifyLeft))
		bar.AddChild(l)
		return l
	}
	ws.play = button("Play", 56, d.togglePlay)
	button("Start", 52, func() { d.seek(0) })
	ws.loop = button("Loop", 52, d.toggleLoop)
	button("Seam", 52, func() {
		if !d.view.Loop {
			d.toggleLoop()
		}
		d.playBody(session.Seam(d.layout.Song))
	})
	ws.position = ebui.NewLabel("", ebui.WithSize(px(96), px(buttonH)), ebui.WithColor(ink), ebui.WithJustify(ebui.JustifyCenter), ebui.WithFont(d.fonts.mono))
	bar.AddChild(ws.position)
	ws.tempo = label(70)
	ws.meter = label(72)
	ws.key = label(120)
	ws.click = button("Click", 56, d.toggleClick)
	ws.loudness = label(150)
	ws.progress = label(140)
	button("Browser", 70, d.toggleBrowser)
	ws.root.AddChild(bar)

	if d.view.Browser {
		list := ebui.NewScrollableContainer(
			ebui.WithSize(px(browserW), float64(ws.browser.Dy())),
			ebui.WithPosition(ebui.Position{X: 0, Y: float64(ws.browser.Min.Y)}),
			ebui.WithPadding(px(6), px(6), px(6), px(6)),
			ebui.WithLayout(ebui.NewVerticalStackLayout(px(2), ebui.AlignStart)),
		)
		heading := func(s string) {
			list.AddChild(ebui.NewLabel(s, ebui.WithSize(px(browserW-24), px(itemH)), ebui.WithColor(dim), ebui.WithJustify(ebui.JustifyLeft)))
		}
		item := func(id, name string) {
			b := ebui.NewButton(
				ebui.WithSize(px(browserW-24), px(itemH)),
				ebui.WithLabelText(fit(d.fonts.text, name, px(browserW-36))),
				ebui.WithRole(roleItem),
				ebui.WithClickHandler(func() { d.open(id, true) }),
			)
			ws.items[id] = b
			list.AddChild(b)
		}
		heading("Songs")
		for _, a := range d.lib.Arrangements {
			item(a.ID, a.Name)
		}
		for _, s := range d.lib.Songs {
			item(s.ID, s.Name+" (code)")
		}
		if len(d.lib.Sounds) > 0 {
			heading("Sounds")
			for _, s := range d.lib.Sounds {
				item(s.ID, s.ID)
			}
		}
		ws.root.AddChild(list)
	}

	ws.status = ebui.NewLabel("",
		ebui.WithSize(w-px(16), px(statusH)),
		ebui.WithPosition(ebui.Position{X: px(8), Y: float64(ws.statusBar.Min.Y)}),
		ebui.WithColor(dim), ebui.WithJustify(ebui.JustifyLeft))
	ws.root.AddChild(ws.status)
	ws.refreshBrowser(d)
	return ws
}

func (ws *widgets) refreshBrowser(d *desk) {
	for id, b := range ws.items {
		if id == d.layout.ID {
			b.SetRole(roleItemOn)
		} else {
			b.SetRole(roleItem)
		}
	}
}

func toggle(b *ebui.Button, on bool) {
	if on {
		b.SetRole(roleOn)
	} else {
		b.SetRole("")
	}
}

// refresh shows the session as it is now.
func (ws *widgets) refresh(d *desk) {
	l, s := d.layout, d.layout.Song
	if d.view.Playing {
		ws.play.SetLabel("Stop")
	} else {
		ws.play.SetLabel("Play")
	}
	toggle(ws.play, d.view.Playing)
	toggle(ws.loop, d.view.Loop)
	toggle(ws.click, d.view.Click)
	ws.position.SetText(session.Format(s, d.view.At))

	timed := session.Timed(s)
	ws.tempo.SetText("")
	ws.meter.SetText("")
	if timed {
		ws.tempo.SetText(fmt.Sprintf("%g bpm", math.Round(l.Tempo*100)/100))
		ws.meter.SetText(meterText(l.Meter))
	}
	ws.key.SetText(l.Key)

	switch {
	case math.IsInf(d.loudness, -1):
		ws.loudness.SetText("")
	case s.Target != 0:
		ws.loudness.SetText(fmt.Sprintf("%.1f / %g LUFS", d.loudness, s.Target))
	default:
		ws.loudness.SetText(fmt.Sprintf("%.1f LUFS", d.loudness))
	}
	if busy, done := d.renders.Busy(); busy {
		ws.progress.SetText(fmt.Sprintf("rendering %.0f%%", 100*done))
	} else {
		ws.progress.SetText("")
	}

	hint := "space play · return start · C loop · shift+C seam · K click · S solo · Z fit · B browser · cmd+wheel zoom"
	if l.Custom {
		hint = "written as code: shown, not editable · " + hint
	}
	ws.status.SetText(d.buildStatus() + "   ·   " + hint)
}

// meterText writes beats per bar, irregular bars joined: 4/4, or 3+3+3+4.
func meterText(bars []int) string {
	if len(bars) == 0 {
		return "4/4"
	}
	if len(bars) == 1 {
		return fmt.Sprintf("%d/4", bars[0])
	}
	var parts []string
	for _, b := range bars {
		parts = append(parts, fmt.Sprint(b))
	}
	return strings.Join(parts, "+")
}

// drawUnder fills the panels the controls sit on.
func (ws *widgets) drawUnder(dst *ebiten.Image) {
	fill(dst, ws.transport, panel)
	fill(dst, ws.browser, panel)
	fill(dst, ws.statusBar, panel)
	line(dst, 0, float64(ws.transport.Max.Y), float64(ws.transport.Max.X), float64(ws.transport.Max.Y), 1, edge)
	if !ws.browser.Empty() {
		line(dst, float64(ws.browser.Max.X), float64(ws.browser.Min.Y), float64(ws.browser.Max.X), float64(ws.browser.Max.Y), 1, edge)
	}
}

// drawOver shows a failed build's output over the bottom of the timeline.
func (ws *widgets) drawOver(dst *ebiten.Image, d *desk) {
	if d.build.State != session.Failed {
		return
	}
	lines := strings.Split(d.build.Output, "\n")
	if len(lines) > outputRows {
		lines = append(lines[:outputRows-1], fmt.Sprintf("… %d more", len(lines)-outputRows+1))
	}
	rowH := d.px(16)
	h := rowH*float64(len(lines)) + d.px(16)
	r := ws.timeline
	box := image.Rect(r.Min.X, r.Max.Y-int(h), r.Max.X, r.Max.Y)
	fill(dst, box, edge)
	line(dst, float64(box.Min.X), float64(box.Min.Y), float64(box.Max.X), float64(box.Min.Y), d.px(2), failColor)
	for i, s := range lines {
		y := float64(box.Min.Y) + d.px(8) + rowH*(float64(i)+0.5)
		drawText(dst, d.fonts.num, fit(d.fonts.num, s, float64(box.Dx())-d.px(16)), float64(box.Min.X)+d.px(8), y, failColor, text.AlignStart)
	}
}
