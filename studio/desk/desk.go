// Package desk is the studio's window: a transport, a browser of the game's songs and
// sounds, and the arrangement on a timeline, playing through synth's own render.
package desk

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/cbodonnell/ebui"
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/player"
	"github.com/cbodonnell/synth/studio/session"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ErrReload ends the window once a new build has been handed off.
var ErrReload = errors.New("reload")

type Config struct {
	Library session.Library
	View    session.View
	// Where to leave a handoff after rebuilding; "" runs without reloading.
	Handoff string
	// Bytes of memory renders may keep; the parts and buses beyond it go to mapped
	// files, and the notes are rendered again.
	Cache int64
}

// Run opens the window and returns when it closes: ErrReload when it closed to hand
// off to a new build.
func Run(c Config) error {
	d := &desk{
		lib: c.Library, view: c.View, handoff: c.Handoff,
		renders: session.NewRenders(synth.Budget(c.Cache), synth.Spill(os.TempDir())),
	}
	d.player = player.New()
	d.preview = d.player.Preview()
	if d.handoff == "" {
		d.status = "no reload: run with -edit"
	} else if w, err := session.Watch(filepath.Dir(d.handoff)); err != nil {
		d.status = "no reload: " + err.Error()
	} else {
		d.watcher = w
		go w.Run()
	}
	id := d.view.ID
	if id == "" {
		if all := d.lib.All(); len(all) > 0 {
			id = all[0].ID
		}
	}
	d.open(id, false)

	ebiten.SetWindowTitle("synth studio")
	ebiten.SetWindowSize(1440, 900)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(d)
}

type desk struct {
	lib    session.Library
	layout session.Layout
	view   session.View

	player  *player.Player
	preview *player.Preview
	renders *session.Renders
	take    *session.Take
	// The preview is playing the take, rather than waiting for its render, and the
	// frame of it being heard.
	sounding bool
	frame    int
	// The song's loudness as last rendered with nothing soloed.
	loudness float64
	// Waiting for a looping render to play the seam from inside the body.
	seam bool

	handoff string
	watcher *session.Watcher
	build   session.Build
	status  string

	scale   float64
	w, h    int
	fonts   fonts
	ui      *ebui.Manager
	widgets *widgets
	tl      timeline
	stale   bool // laid out for a view that has changed
}

// open shows the song or sound with the ID: fresh, or as the view has it, as after a
// reload.
func (d *desk) open(id string, fresh bool) {
	l, ok := d.lib.Open(id)
	if !ok {
		if all := d.lib.All(); len(all) > 0 && id != all[0].ID {
			d.open(all[0].ID, true)
		}
		return
	}
	playing := d.view.Playing && !fresh
	d.stop()
	d.layout, d.take, d.loudness, d.seam = l, nil, math.Inf(-1), false
	d.view.ID, d.view.Playing = id, playing
	if fresh {
		d.view = session.View{ID: id, Loop: d.view.Loop, Click: d.view.Click, Browser: d.view.Browser}
		d.tl.fit(d)
	}
	if !session.Timed(l.Song) {
		d.view.Loop = false
	}
	d.view.Solo = slices.DeleteFunc(d.view.Solo, func(name string) bool { return !d.hasLane(name) })
	if d.widgets != nil {
		d.widgets.refreshBrowser(d)
	}
	d.render()
}

func (d *desk) hasLane(name string) bool {
	return slices.ContainsFunc(d.layout.Lanes, func(l session.Lane) bool { return l.Name == name })
}

// render asks for the song as it's now heard.
func (d *desk) render() {
	d.renders.Request(d.layout.Song, session.Mix{Loop: d.view.Loop, Solo: slices.Clone(d.view.Solo), Click: d.view.Click})
}

// receive takes a finished render, swapping it in at the playhead if it's playing.
func (d *desk) receive(t session.Take) {
	old := d.take
	d.take = &t
	if len(d.view.Solo) == 0 {
		d.loudness = t.Loudness
	}
	start, end := t.Region()
	switch {
	case d.seam && t.Loops:
		d.playBody(session.Seam(d.layout.Song))
	case d.sounding && old != nil:
		d.preview.Swap(t.Buffer, start, end, func(f int) int { return t.Follow(*old, f) })
	case d.view.Playing:
		d.preview.Play(t.Buffer, t.Frame(d.view.At), start, end)
		d.sounding = true
	}
}

func (d *desk) play() {
	d.view.Playing = true
	if d.take == nil {
		return
	}
	if !d.take.Loops && d.take.Frame(d.view.At) >= d.take.Buffer.Len() {
		d.view.At = 0
	}
	start, end := d.take.Region()
	d.preview.Play(d.take.Buffer, d.take.Frame(d.view.At), start, end)
	d.sounding = true
}

// playBody plays from song time at inside the loop's body, where the seam is heard
// as the game hears it every time round.
func (d *desk) playBody(at float64) {
	d.view.At = at
	if d.take == nil || !d.take.Loops {
		d.seam = d.view.Loop
		d.play()
		return
	}
	d.seam = false
	d.view.Playing = true
	start, end := d.take.Region()
	from := d.take.Intro + d.take.Frame(at) - d.take.Frame(d.layout.Song.LoopStart)
	d.preview.Play(d.take.Buffer, from, start, end)
	d.sounding = true
}

func (d *desk) stop() {
	d.view.Playing = false
	if d.sounding {
		d.preview.Stop()
		d.sounding = false
	}
}

func (d *desk) togglePlay() {
	if d.view.Playing {
		d.stop()
	} else {
		d.play()
	}
}

// seek moves the playhead, playing on from there if it's playing.
func (d *desk) seek(at float64) {
	d.view.At = math.Max(0, math.Min(at, d.layout.Length()))
	if d.view.Playing {
		d.play()
	}
}

func (d *desk) toggleLoop() {
	if !session.Timed(d.layout.Song) || !d.layout.Song.Loop {
		return
	}
	d.view.Loop = !d.view.Loop
	d.render()
}

func (d *desk) toggleClick() {
	d.view.Click = !d.view.Click
	d.render()
}

func (d *desk) toggleSolo(name string) {
	if i := slices.Index(d.view.Solo, name); i >= 0 {
		d.view.Solo = slices.Delete(d.view.Solo, i, i+1)
	} else {
		d.view.Solo = append(d.view.Solo, name)
	}
	d.render()
}

// toggleBrowser shows or hides the browser, laying the window out again before the
// next update rather than under the controls handling this one.
func (d *desk) toggleBrowser() {
	d.view.Browser = !d.view.Browser
	d.stale = true
}

func (d *desk) Update() error {
	if err := d.receiveBuilds(); err != nil {
		return err
	}
	if d.stale {
		d.stale = false
		d.relayout()
	}
	if t, ok := d.renders.Done(); ok {
		d.receive(t)
	}
	if d.sounding {
		if f, ok := d.preview.Position(); ok {
			d.frame, d.view.At = f, d.take.Time(f)
		} else {
			d.view.Playing, d.sounding = false, false
		}
	}
	if d.ui == nil {
		return nil
	}
	if err := d.ui.Update(); err != nil {
		return err
	}
	d.keys()
	d.tl.update(d)
	d.widgets.refresh(d)
	return nil
}

func (d *desk) keys() {
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	command := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if command {
		return
	}
	pressed := inpututil.IsKeyJustPressed
	switch {
	case pressed(ebiten.KeySpace):
		d.togglePlay()
	case pressed(ebiten.KeyEnter):
		d.seek(0)
	case pressed(ebiten.KeyC) && shift:
		if !d.view.Loop {
			d.toggleLoop()
		}
		d.playBody(session.Seam(d.layout.Song))
	case pressed(ebiten.KeyC):
		d.toggleLoop()
	case pressed(ebiten.KeyK):
		d.toggleClick()
	case pressed(ebiten.KeyS):
		if d.view.Track != "" {
			d.toggleSolo(d.view.Track)
		}
	case pressed(ebiten.KeyZ):
		d.tl.fit(d)
	case pressed(ebiten.KeyB):
		d.toggleBrowser()
	case pressed(ebiten.KeyUp):
		d.tl.selectNext(d, -1)
	case pressed(ebiten.KeyDown):
		d.tl.selectNext(d, 1)
	}
}

// receiveBuilds follows the rebuilds, and hands off to a new one when it's built.
func (d *desk) receiveBuilds() error {
	if d.watcher == nil {
		return nil
	}
	for {
		select {
		case b := <-d.watcher.Builds:
			d.build = b
			switch b.State {
			case session.Failed:
				fmt.Fprintln(os.Stderr, b.Output)
			case session.Built:
				h := session.Handoff{Binary: b.Binary, Args: []string{"-edit", "-view", d.view.String()}}
				if err := h.Write(d.handoff); err != nil {
					d.build = session.Build{State: session.Failed, Output: "handoff: " + err.Error()}
					continue
				}
				// Fade out rather than cut off.
				d.preview.Stop()
				time.Sleep(20 * time.Millisecond)
				return ErrReload
			}
		default:
			return nil
		}
	}
}

// buildStatus is a line on the rebuild, for the status bar.
func (d *desk) buildStatus() string {
	switch {
	case d.watcher == nil:
		return d.status
	case d.build.State == session.Building:
		return "building…"
	case d.build.State == session.Failed:
		first, _, _ := strings.Cut(d.build.Output, "\n")
		return "build failed, still playing the last build: " + first
	}
	return "reloads when its source changes"
}

func (d *desk) Layout(w, h int) (int, int) {
	scale := ebiten.Monitor().DeviceScaleFactor()
	pw, ph := int(float64(w)*scale), int(float64(h)*scale)
	if pw != d.w || ph != d.h || scale != d.scale {
		d.w, d.h, d.scale = pw, ph, scale
		d.relayout()
	}
	return pw, ph
}

// relayout builds the controls for the window's size and scale.
func (d *desk) relayout() {
	d.fonts = newFonts(d.scale)
	ebui.DefaultPainter = painter{&d.fonts}
	ebui.DefaultFont = d.fonts.regular
	d.widgets = buildWidgets(d)
	d.ui = ebui.NewManager(d.widgets.root, ebui.WithNavInput(nil))
	d.ui.DisableFocus()
	d.tl.place(d)
}

func (d *desk) Draw(screen *ebiten.Image) {
	screen.Fill(ground)
	if d.ui == nil {
		return
	}
	d.widgets.drawUnder(screen)
	d.tl.draw(screen, d)
	d.ui.Draw(screen)
	d.widgets.drawOver(screen, d)
}

// px is a length in the window's pixels.
func (d *desk) px(x float64) float64 { return x * d.scale }
