package desk

import (
	"fmt"
	"maps"
	"reflect"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/player"
	"github.com/cbodonnell/synth/studio/session"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// The keys musical typing plays, and the semitones above the octave's C they play.
var typingKeys = []typingKey{
	{ebiten.KeyA, 0}, {ebiten.KeyW, 1}, {ebiten.KeyS, 2}, {ebiten.KeyE, 3}, {ebiten.KeyD, 4},
	{ebiten.KeyF, 5}, {ebiten.KeyT, 6}, {ebiten.KeyG, 7}, {ebiten.KeyY, 8}, {ebiten.KeyH, 9},
	{ebiten.KeyU, 10}, {ebiten.KeyJ, 11}, {ebiten.KeyK, 12},
}

type typingKey struct {
	key   ebiten.Key
	semis int
}

// heldFor is how long a typed note is rendered: a key held longer hears its release.
const heldFor = 4.0

// keyboard is musical typing, with Caps Lock: the typing keys play the selected track's
// patch for as long as they're held, Z X shift the octave and C V the velocity. Step
// input shares its octave and velocity.
type keyboard struct {
	out    *player.Player
	on     bool
	octave int
	accent int // ghosted -1, plain 0, accented 1

	// The part the notes are rendered for, and what's rendered or rendering for it, by
	// key. A change of patch, level, pan or velocity drops them.
	part    synth.Part
	vel     float64
	mu      sync.Mutex
	renders map[int]*typedNote
	playing map[ebiten.Key]*press
}

type typedNote struct {
	done chan struct{}
	buf  synth.Stereo
}

// press is a typing key held down: its note once it's rendered, or that it was let go
// before then.
type press struct {
	held *player.Held
	up   bool
}

func newKeyboard(out *player.Player) keyboard {
	return keyboard{out: out, octave: 4, renders: map[int]*typedNote{}, playing: map[ebiten.Key]*press{}}
}

// key is the MIDI key the typing key plays in the keyboard's octave.
func (k *keyboard) key(semis int) int { return (k.octave+1)*12 + semis }

// velocity is what the keyboard plays at in a clip of base velocity vel.
func (k *keyboard) velocity(vel float64) float64 {
	return session.Event{Accent: k.accent}.Velocity(vel)
}

func (k *keyboard) accentName() string {
	return map[int]string{1: "accented", 0: "plain", -1: "ghosted"}[k.accent]
}

// hint is the status bar's line while musical typing is on.
func (k *keyboard) hint(d *desk) string {
	s := fmt.Sprintf("musical typing (caps lock): A–K white keys, W E T Y U black · Z X octave (C%d) · C V velocity (%s)", k.octave, k.accentName())
	if _, ok := d.typingTrack(); !ok {
		s = "musical typing: select a track to play its patch · " + s
	}
	return s
}

// lock follows Caps Lock. On macOS the key reads as pressed while the lock is on, and a
// second tap arrives as its release; elsewhere each tap is a press.
func (k *keyboard) lock() {
	on := k.on
	if runtime.GOOS == "darwin" {
		on = ebiten.IsKeyPressed(ebiten.KeyCapsLock)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyCapsLock) {
		on = !on
	}
	if on != k.on {
		k.on = on
		k.releaseAll()
	}
}

// shift takes Z X for the octave and C V for the velocity, reporting whether it did.
func (k *keyboard) shift() bool {
	pressed := inpututil.IsKeyJustPressed
	switch {
	case pressed(ebiten.KeyZ):
		k.octave = max(k.octave-1, 0)
	case pressed(ebiten.KeyX):
		k.octave = min(k.octave+1, 8)
	case pressed(ebiten.KeyC):
		k.accent = max(k.accent-1, -1)
	case pressed(ebiten.KeyV):
		k.accent = min(k.accent+1, 1)
	default:
		return false
	}
	return true
}

// keys plays the typing keys while musical typing is on, reporting whether it took the
// keys. Step input writes what's played at its cursor; the command shortcuts still work.
func (k *keyboard) keys(d *desk) bool {
	if !k.on || d.ce.typing != nil || d.mx.typing != nil {
		return false
	}
	if ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta) {
		return false
	}
	k.play(d)
	switch {
	case d.ce.step != nil:
		d.ce.keys(d)
	case k.shift():
	case inpututil.IsKeyJustPressed(ebiten.KeySpace):
		d.togglePlay()
	}
	return true
}

// play starts and releases the notes of the typing keys pressed and let go this frame.
func (k *keyboard) play(d *desk) {
	tr, ok := d.typingTrack()
	if !ok {
		k.releaseAll()
		return
	}
	if len(k.renders) == 0 || slices.ContainsFunc(typingKeys, func(tk typingKey) bool { return inpututil.IsKeyJustPressed(tk.key) }) {
		k.prepare(tr)
	}
	for _, tk := range typingKeys {
		switch {
		case inpututil.IsKeyJustPressed(tk.key):
			k.release(tk.key)
			k.press(tk.key, k.key(tk.semis))
		case inpututil.IsKeyJustReleased(tk.key):
			k.release(tk.key)
		}
	}
}

// typingTrack is the selected track, whose patch the keyboard plays.
func (d *desk) typingTrack() (synth.Track, bool) {
	a := d.layout.Arrangement
	i := slices.IndexFunc(a.Tracks, func(t synth.Track) bool { return t.Name == d.view.Track })
	if i < 0 {
		return synth.Track{}, false
	}
	return a.Tracks[i], true
}

// prepare renders the octave's keys for the track ahead of their being typed, letting
// go of any other octave's, and of all of them when the part or the velocity changes.
func (k *keyboard) prepare(tr synth.Track) {
	part := synth.Part{Name: tr.Name, Patch: tr.Patch, Level: tr.Level, Pan: tr.Pan}
	vel := k.velocity(0)
	k.mu.Lock()
	defer k.mu.Unlock()
	if vel != k.vel || !reflect.DeepEqual(part, k.part) {
		k.part, k.vel, k.renders = part, vel, map[int]*typedNote{}
	}
	lo, hi := k.key(0), k.key(typingKeys[len(typingKeys)-1].semis)
	maps.DeleteFunc(k.renders, func(key int, _ *typedNote) bool { return key < lo || key > hi })
	for _, tk := range typingKeys {
		k.render(k.key(tk.semis))
	}
}

// render renders key in the background, if it isn't already. Hold k.mu.
func (k *keyboard) render(key int) *typedNote {
	if n, ok := k.renders[key]; ok {
		return n
	}
	n := &typedNote{done: make(chan struct{})}
	k.renders[key] = n
	part := k.part
	part.Notes = []synth.Note{{Key: float64(key), Dur: heldFor, Vel: k.vel}}
	go func() {
		n.buf = synth.Render(synth.Song{ID: fmt.Sprint("typed ", key), Parts: []synth.Part{part}})
		close(n.done)
	}()
	return n
}

func (k *keyboard) press(tk ebiten.Key, key int) {
	p := &press{}
	k.mu.Lock()
	k.playing[tk] = p
	n := k.render(key)
	k.mu.Unlock()
	start := func() {
		k.mu.Lock()
		defer k.mu.Unlock()
		if !p.up {
			p.held = k.out.Hold(n.buf, 1)
		}
	}
	select {
	case <-n.done:
		start()
	default:
		go func() {
			<-n.done
			start()
		}()
	}
}

// release lets the key's note go over its patch's release.
func (k *keyboard) release(tk ebiten.Key) {
	k.mu.Lock()
	defer k.mu.Unlock()
	p, ok := k.playing[tk]
	if !ok {
		return
	}
	delete(k.playing, tk)
	p.up = true
	if p.held != nil {
		p.held.Release(time.Duration(k.part.Patch.Amp.Release * float64(time.Second)))
	}
}

func (k *keyboard) releaseAll() {
	for tk := range k.playing {
		k.release(tk)
	}
}
