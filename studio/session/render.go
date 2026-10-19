package session

import (
	"math"
	"slices"
	"sync"

	"github.com/cbodonnell/synth"
)

// Mix is how the studio is listening: with the loop or through to the end, to some
// tracks alone, with or without the metronome. None of it reaches the song.
type Mix struct {
	Loop  bool
	Solo  []string
	Click bool
}

// Take is a song rendered for the preview: as the game plays it when it loops (the
// intro, then the body forever), or once through to the end of its tail.
type Take struct {
	Song   synth.Song // as heard: soloed
	Buffer synth.Stereo
	// Frames of Buffer before the body; all of it when it doesn't loop.
	Intro int
	Loops bool
	// Each part's and bus's meter, by name, in song frames; the master's, in buffer
	// frames.
	Parts, Buses map[string]synth.Gauge
	Master       synth.Gauge
	// Integrated LUFS of the mix without the metronome, as Song.Target measures it.
	Loudness float64
}

// Render renders s as heard through m.
func Render(r *synth.Renderer, s synth.Song, m Mix) Take {
	if len(m.Solo) > 0 {
		var parts []synth.Part
		for _, p := range s.Parts {
			if slices.Contains(m.Solo, p.Name) {
				parts = append(parts, p)
			}
		}
		s.Parts = parts
	}
	t := Take{Song: s, Loops: m.Loop && s.Loop}
	if !t.Loops {
		s.Loop = false
	}
	var click Take
	if m.Click && Timed(s) {
		click = Render(r, metronome(s), Mix{Loop: t.Loops})
	}

	if t.Loops {
		l := r.RenderLoop(s)
		t.Buffer, t.Intro = synth.Concat(l.Intro, l.Body), l.Intro.Len()
	} else {
		t.Buffer = r.Render(s)
		t.Intro = t.Buffer.Len()
	}
	t.Loudness, _ = synth.Loudness(t.Buffer)
	if click.Buffer.Len() > 0 {
		t.Buffer = mixIn(t.Buffer, click.Buffer)
	}

	t.Master = synth.NewGauge(t.Buffer)
	t.Parts, t.Buses = map[string]synth.Gauge{}, map[string]synth.Gauge{}
	parts, buses := r.Gauges(s)
	for i, g := range parts {
		t.Parts[s.Parts[i].Name] = g
	}
	for i, g := range buses {
		t.Buses[s.Buses[i].Name] = g
	}
	return t
}

func mixIn(dst, src synth.Stereo) synth.Stereo {
	out := synth.Concat(dst)
	for i := range min(out.Len(), src.Len()) {
		out.L[i] = clamp(out.L[i] + src.L[i])
		out.R[i] = clamp(out.R[i] + src.R[i])
	}
	return out
}

func clamp(x float32) float32 { return max(-1, min(1, x)) }

// metronome is a click on every beat of s, the bar's first higher.
func metronome(s synth.Song) synth.Song {
	click := synth.Patch{
		Oscs: []synth.Osc{{Wave: synth.Sine}},
		Amp:  synth.Envelope{Attack: 0.001, Decay: 0.03, Release: 0.01},
	}
	var notes []synth.Note
	beats := int(math.Ceil(s.Length/s.Meter.Beat - 1e-9))
	for b := range beats {
		at := float64(b) * s.Meter.Beat
		key := 84.0
		if s.Meter.At(at).Beat == 0 {
			key = 96
		}
		notes = append(notes, synth.Note{At: at, Dur: 0.02, Key: key, Vel: 0.6})
	}
	return synth.Song{
		ID: s.ID + " click", Length: s.Length, Loop: s.Loop, LoopStart: s.LoopStart, Meter: s.Meter,
		Parts: []synth.Part{{Name: "click", Patch: click, Notes: notes, Level: -14}},
	}
}

// Region is the loop in buffer frames, or 0, 0 when it plays through.
func (t Take) Region() (start, end int) {
	if !t.Loops {
		return 0, 0
	}
	return t.Intro, t.Buffer.Len()
}

// Frame is the buffer frame to play from to hear song time at.
func (t Take) Frame(at float64) int {
	return min(max(int(math.Round(at*synth.SampleRate)), 0), t.Buffer.Len())
}

// SongFrame is where buffer frame f falls in the song.
func (t Take) SongFrame(f int) int {
	if !t.Loops || f < t.Intro {
		return f
	}
	return int(math.Round(t.Song.LoopStart*synth.SampleRate)) + f - t.Intro
}

// Follow maps frame f of the take this one replaces onto this one: through the loop's
// body when both loop, so a re-render heard while playing stays in the body.
func (t Take) Follow(old Take, f int) int {
	if t.Loops && old.Loops && f >= old.Intro {
		body := t.Buffer.Len() - t.Intro
		if body <= 0 {
			return t.Intro
		}
		return t.Intro + (f-old.Intro)%body
	}
	return t.Frame(old.Time(f))
}

// Time is where buffer frame f falls in the song, in seconds.
func (t Take) Time(f int) float64 { return float64(t.SongFrame(f)) / synth.SampleRate }

// Renders renders on a goroutine of its own, one song at a time and always the latest
// asked for, reusing what the last render shares with the next.
type Renders struct {
	r *synth.Renderer

	mu      sync.Mutex
	next    *request
	running bool
	done    *Take
}

type request struct {
	song synth.Song
	mix  Mix
}

func NewRenders(opts ...synth.RendererOption) *Renders {
	return &Renders{r: synth.NewRenderer(opts...)}
}

// Request renders s as heard through m, once what's rendering now is done. A render
// finished but not yet taken is dropped.
func (b *Renders) Request(s synth.Song, m Mix) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.next, b.done = &request{s, m}, nil
	if !b.running {
		b.running = true
		go b.run()
	}
}

func (b *Renders) run() {
	for {
		b.mu.Lock()
		req := b.next
		b.next = nil
		if req == nil {
			b.running = false
			b.mu.Unlock()
			return
		}
		b.mu.Unlock()
		t := Render(b.r, req.song, req.mix)
		b.mu.Lock()
		if b.next == nil {
			b.done = &t
		}
		b.mu.Unlock()
	}
}

// Done is the latest render, once, when it's finished.
func (b *Renders) Done() (Take, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.done == nil {
		return Take{}, false
	}
	t := *b.done
	b.done = nil
	return t, true
}

// Busy reports whether a render is running, and how far it has come.
func (b *Renders) Busy() (bool, float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.running, b.r.Progress()
}
