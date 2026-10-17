// Package player plays rendered songs through Ebiten. A game holds one Mixer; the Mixer
// drives a Sink, which is a Player in the game and Silent in tests.
package player

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"math"
	"math/rand/v2"
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

// Past the cap the lowest priority gives way rather than queueing: a queued effect would arrive late.
const maxEffects = 16

// The effects stream runs continuously, so a new effect waits behind whatever is
// buffered. Ebiten's default half second would be heard; this is short but safe.
const effectsBuffer = 40 * time.Millisecond

type rendered struct {
	done        chan struct{}
	pcm         []float32 // interleaved stereo
	ogg         []byte    // baked music, decoded as it plays instead of held as pcm
	intro, body int       // frames; body 0 means a one-shot
}

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
// audio context. Music and effects each play through one stream, which mixes their
// voices itself.
type Player struct {
	ctx            *ebaudio.Context
	music, effects *stream
	musicOut       *ebaudio.Player

	mu      sync.Mutex
	cache   map[string]*rendered
	current *voice
	gain    float64
	gen     int // bumped on every music change, so a render that lands late is dropped
	rng     *rand.Rand
	last    map[string]int // the take each sound played last

	baked    fs.FS
	manifest bake.Manifest
}

func New(opts ...Option) *Player {
	p := &Player{
		ctx:     ebaudio.NewContext(synth.SampleRate),
		music:   newStream(0, 0),
		effects: newStream(1, maxEffects),
		cache:   make(map[string]*rendered),
		rng:     rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
		last:    make(map[string]int),
	}
	for _, o := range opts {
		o(p)
	}
	var err error
	if p.musicOut, err = p.ctx.NewPlayerF32(p.music); err != nil {
		panic(err)
	}
	effectsOut, err := p.ctx.NewPlayerF32(p.effects)
	if err != nil {
		panic(err)
	}
	effectsOut.SetBufferSize(effectsBuffer)
	effectsOut.Play()
	return p
}

// Prepare renders songs, and every take of a sound, in the background, so the first
// play doesn't wait on them.
func (p *Player) Prepare(songs ...synth.Song) {
	for _, s := range songs {
		for take := range takes(s) {
			p.render(s, take)
		}
	}
}

func takes(s synth.Song) int { return max(s.Vary.Takes, 1) }

func (p *Player) render(s synth.Song, take int) *rendered {
	key := s.ID
	if take > 0 {
		key = fmt.Sprintf("%s#%d", s.ID, take)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if r, ok := p.cache[key]; ok {
		return r
	}
	r := &rendered{done: make(chan struct{})}
	p.cache[key] = r
	e, baked := p.manifest.Fresh(s)
	go func() {
		defer close(r.done)
		if baked && take < len(e.Files) && p.load(r, s, e, take) == nil {
			return
		}
		s := synth.Take(s, take)
		if !s.Loop {
			r.pcm = interleave(synth.Render(s))
			return
		}
		l := synth.RenderLoop(s)
		r.pcm = interleave(synth.Concat(l.Intro, l.Body))
		r.intro, r.body = l.Intro.Len(), l.Body.Len()
	}()
	return r
}

func (p *Player) load(r *rendered, s synth.Song, e bake.Entry, take int) error {
	ogg, err := fs.ReadFile(p.baked, e.Files[take])
	if err != nil {
		return err
	}
	if s.Loop {
		r.ogg = ogg
		r.intro, r.body = e.Intro, e.Body
		return nil
	}
	d, err := vorbis.DecodeF32(bytes.NewReader(ogg))
	if err != nil {
		return err
	}
	b, err := io.ReadAll(d)
	if err != nil {
		return err
	}
	r.pcm = make([]float32, len(b)/4)
	floats(r.pcm, b)
	return nil
}

func interleave(s synth.Stereo) []float32 {
	out := make([]float32, s.Len()*2)
	for i := range s.L {
		out[i*2], out[i*2+1] = s.L[i], s.R[i]
	}
	return out
}

// source is a fresh read of the audio: a decoder over a baked file, or the rendered pcm.
func (r *rendered) source() (source, error) {
	var src seekable = &memory{pcm: r.pcm}
	if r.ogg != nil {
		d, err := vorbis.DecodeF32(bytes.NewReader(r.ogg))
		if err != nil {
			return nil, err
		}
		src = &decoded{rs: d}
	}
	if r.body == 0 {
		return src, nil
	}
	return &looped{src: src, start: r.intro, end: r.intro + r.body}, nil
}

// PlayMusic returns at once; playback starts when the render is ready.
func (p *Player) PlayMusic(s synth.Song, gain float64) {
	p.StopMusic()
	r := p.render(s, 0)
	p.SetMusicGain(gain)

	p.mu.Lock()
	gen := p.gen
	p.mu.Unlock()

	go func() {
		<-r.done
		if len(r.pcm) == 0 && r.ogg == nil {
			return
		}
		src, err := r.source()
		if err != nil {
			return
		}

		p.mu.Lock()
		defer p.mu.Unlock()
		if gen != p.gen {
			return
		}
		p.current = newVoice(s.ID, src, 1, 1)
		p.music.play(p.current, 0)
		if p.gain > 0 {
			p.musicOut.Play()
		}
	}()
}

// The render stays cached; the music fades out over a few milliseconds.
func (p *Player) StopMusic() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.gen++
	if p.current == nil {
		return
	}
	p.music.stop(p.current)
	p.current = nil
}

// Pauses outright at 0 so a muted game isn't streaming a loop nobody can hear.
func (p *Player) SetMusicGain(gain float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.gain = gain
	p.music.setGain(gain)
	switch {
	case gain <= 0:
		p.musicOut.Pause()
	case p.current != nil:
		p.musicOut.Play()
	}
}

func (p *Player) MusicPosition() (time.Duration, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current == nil || !p.musicOut.IsPlaying() {
		return 0, false
	}
	start, ok := p.music.started(p.current)
	if !ok {
		return 0, false
	}
	pos := p.musicOut.Position() - time.Duration(start)*time.Second/synth.SampleRate
	return pos, pos >= 0
}

// Play renders an uncached effect on the spot. Effects are short, but Prepare them anyway
// if one must land on the frame it's triggered.
func (p *Player) Play(s synth.Song, gain float64) {
	p.mu.Lock()
	take, rate, g := vary(s.Vary, p.rng, p.lastTake(s.ID))
	p.last[s.ID] = take
	p.mu.Unlock()

	r := p.render(s, take)
	<-r.done
	if len(r.pcm) == 0 {
		return
	}
	v := newVoice(s.ID, &memory{pcm: r.pcm}, rate, gain*g)
	v.priority = s.Priority
	p.effects.play(v, s.Voices)
}

func (p *Player) lastTake(id string) int {
	if t, ok := p.last[id]; ok {
		return t
	}
	return -1
}

// vary picks a play: a take other than the last, and a playback rate and gain within
// the variation's range. With no variation it's the song as rendered.
func vary(v synth.Variation, rng *rand.Rand, last int) (take int, rate, gain float64) {
	if n := max(v.Takes, 1); n > 1 {
		if last < 0 || last >= n {
			take = rng.IntN(n)
		} else if take = rng.IntN(n - 1); take >= last {
			take++
		}
	}
	rate = math.Exp2(v.Pitch * (2*rng.Float64() - 1) / 12)
	gain = math.Pow(10, v.Gain*(2*rng.Float64()-1)/20)
	return take, rate, gain
}
