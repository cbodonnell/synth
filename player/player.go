// Package player plays rendered songs through Ebiten. A game holds one Mixer; the Mixer
// drives a Sink, which is a Player in the game and Silent in tests.
package player

import (
	"bytes"
	"io"
	"io/fs"
	"sync"
	"time"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/bake"
	ebaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

type Sink interface {
	Play(s synth.Song, gain float64)
	// Gain 0 means the song is current but inaudible: held, not discarded.
	PlayMusic(s synth.Song, gain float64)
	StopMusic()
	SetMusicGain(gain float64)
	// How long the music has played, looping included. False means no honest
	// position: nothing playing, paused, or still rendering.
	MusicPosition() (time.Duration, bool)
}

type Silent struct{}

func (Silent) Play(synth.Song, float64)             {}
func (Silent) PlayMusic(synth.Song, float64)        {}
func (Silent) StopMusic()                           {}
func (Silent) SetMusicGain(float64)                 {}
func (Silent) MusicPosition() (time.Duration, bool) { return 0, false }

// Past the cap the oldest effect is dropped rather than queued: a queued effect would arrive late.
const maxEffects = 16

type rendered struct {
	done        chan struct{}
	pcm         []byte
	ogg         []byte // baked music, decoded as it plays instead of held as pcm
	intro, body int64  // bytes; body 0 means a one-shot
}

const bytesPerSample = 8

type Option func(*Player)

// Baked plays songs from a bake.Bake directory where its bake is fresh, and renders the
// rest. Effects are decoded once into memory; music streams from its compressed file.
func Baked(fsys fs.FS) Option {
	return func(p *Player) {
		if m, err := bake.Load(fsys); err == nil {
			p.baked, p.manifest = fsys, m
		}
	}
}

// Player is the Ebiten-backed Sink. Exactly one per process: Ebiten panics on a second
// audio context.
type Player struct {
	ctx *ebaudio.Context

	mu      sync.Mutex
	cache   map[string]*rendered
	music   *ebaudio.Player
	gain    float64
	gen     int // bumped on every music change, so a render that lands late is dropped
	effects []*ebaudio.Player

	baked    fs.FS
	manifest bake.Manifest
}

func New(opts ...Option) *Player {
	p := &Player{
		ctx:   ebaudio.NewContext(synth.SampleRate),
		cache: make(map[string]*rendered),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Prepare renders songs in the background, so the first PlayMusic doesn't wait on them.
func (p *Player) Prepare(songs ...synth.Song) {
	for _, s := range songs {
		p.render(s)
	}
}

func (p *Player) render(s synth.Song) *rendered {
	p.mu.Lock()
	defer p.mu.Unlock()
	if r, ok := p.cache[s.ID]; ok {
		return r
	}
	r := &rendered{done: make(chan struct{})}
	p.cache[s.ID] = r
	e, baked := p.manifest.Fresh(s)
	go func() {
		defer close(r.done)
		if baked && p.load(r, s, e) == nil {
			return
		}
		if !s.Loop {
			r.pcm = synth.PCM(synth.Render(s))
			return
		}
		l := synth.RenderLoop(s)
		r.pcm = synth.PCM(synth.Concat(l.Intro, l.Body))
		r.intro = int64(len(r.pcm) - l.Body.Len()*bytesPerSample)
		r.body = int64(l.Body.Len() * bytesPerSample)
	}()
	return r
}

func (p *Player) load(r *rendered, s synth.Song, e bake.Entry) error {
	ogg, err := fs.ReadFile(p.baked, e.File)
	if err != nil {
		return err
	}
	if s.Loop {
		r.ogg = ogg
		r.intro = int64(e.Intro * bytesPerSample)
		r.body = int64(e.Body * bytesPerSample)
		return nil
	}
	d, err := vorbis.DecodeF32(bytes.NewReader(ogg))
	if err != nil {
		return err
	}
	r.pcm, err = io.ReadAll(d)
	return err
}

// stream is the music's source: a fresh decoder over a baked file, or the rendered pcm.
func (r *rendered) stream() (io.ReadSeeker, error) {
	if r.ogg == nil {
		return bytes.NewReader(r.pcm), nil
	}
	return vorbis.DecodeF32(bytes.NewReader(r.ogg))
}

// PlayMusic returns at once; playback starts when the render is ready.
func (p *Player) PlayMusic(s synth.Song, gain float64) {
	p.StopMusic()
	r := p.render(s)

	p.mu.Lock()
	p.gain = gain
	gen := p.gen
	p.mu.Unlock()

	go func() {
		<-r.done
		if len(r.pcm) == 0 && r.ogg == nil {
			return
		}

		p.mu.Lock()
		defer p.mu.Unlock()
		if gen != p.gen {
			return
		}

		src, err := r.stream()
		if err != nil {
			return
		}
		if r.body > 0 {
			src = ebaudio.NewInfiniteLoopWithIntroF32(src, r.intro, r.body)
		}
		pl, err := p.ctx.NewPlayerF32(src)
		if err != nil {
			return
		}
		pl.SetVolume(p.gain)
		p.music = pl
		if p.gain > 0 {
			pl.Play()
		}
	}()
}

// The render stays cached; only the Ebiten player is released.
func (p *Player) StopMusic() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.gen++
	if p.music == nil {
		return
	}
	p.music.Pause()
	p.music.Close()
	p.music = nil
}

// Pauses outright at 0 so a muted game isn't streaming a loop nobody can hear.
func (p *Player) SetMusicGain(gain float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.gain = gain
	if p.music == nil {
		return
	}
	p.music.SetVolume(gain)
	switch {
	case gain <= 0 && p.music.IsPlaying():
		p.music.Pause()
	case gain > 0 && !p.music.IsPlaying():
		p.music.Play()
	}
}

func (p *Player) MusicPosition() (time.Duration, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.music == nil || !p.music.IsPlaying() {
		return 0, false
	}
	return p.music.Position(), true
}

// Play renders an uncached effect on the spot. Effects are short, but Prepare them anyway
// if one must land on the frame it's triggered.
func (p *Player) Play(s synth.Song, gain float64) {
	r := p.render(s)
	<-r.done
	if len(r.pcm) == 0 {
		return
	}
	pl := p.ctx.NewPlayerF32FromBytes(r.pcm)
	pl.SetVolume(gain)
	pl.Play()

	p.mu.Lock()
	defer p.mu.Unlock()
	live := p.effects[:0]
	for _, e := range p.effects {
		if e.IsPlaying() {
			live = append(live, e)
		} else {
			e.Close()
		}
	}
	live = append(live, pl)
	for len(live) > maxEffects {
		live[0].Close()
		live = live[1:]
	}
	p.effects = live
}
