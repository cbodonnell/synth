package synth

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
)

// Arrangement is a song laid out as a musician lays one out: sections of bars, and
// tracks whose clips are played at bars. It is the declarative form the studio edits,
// and compiles to a Song with Song.
//
// Time is musical: Tempo and Meter make the bars, and times inside effects that are
// song positions or durations (Point.At, TapeStop, Delay.Time, Sidechain.Hits) are in
// beats, so they follow the tempo.
type Arrangement struct {
	ID   string
	Name string

	Tempo float64 // beats per minute; 0 is a sound effect, timed in seconds
	Meter []int   // beats per bar, repeating; nil is 4/4
	Key   string  // for the eye: "D mixolydian"

	// The form in order: the ruler's markers and the length of the song.
	Sections []Section
	// The section the loop starts at; the song loops to its end. "" plays once.
	Loop string

	Tracks []Track
	Buses  []Bus
	Master []Effect

	Vary     Variation
	Voices   int
	Priority int
	Target   float64
}

type Section struct {
	Name string
	Bars int
}

// Track is a Part whose notes are its clips, played where Play says.
type Track struct {
	Name    string
	Patch   Patch
	Level   float64 // dB; 0 is unity
	Pan     float64
	Inserts []Effect
	Sends   []Send
	Out     string
	Mute    bool // leaves the track out of Song

	Clips []Clip
	Play  []Play
}

// Clip is a Phrase's notation and settings, a bar to a string. Step is in beats, and
// every bar string fills its bar exactly. With Hit set, the bars are drum lines in
// Pattern notation. A line that isn't on a grid holds Notes instead.
type Clip struct {
	Name string
	Step float64 // beats per token

	Gate     float64
	Vel      float64
	Strum    float64
	Swing    float64
	Glide    float64
	Humanize float64
	Seed     uint64

	// Holds every note to the next bar line, where a sustain pedal lifts.
	Pedal bool

	Hit  string
	Bars []string

	// At and Dur in beats from the clip's start, which spans Length bars (0 means 1).
	Notes  []Note
	Length int
}

// Play places a clip on its track at Bar, counted from 1, Times over back to back
// (0 means once).
type Play struct {
	Clip  string
	Bar   int
	Times int
}

// beatTimed is an effect holding song positions or durations, which an arrangement
// writes in beats.
type beatTimed interface{ inSeconds(beat float64) Effect }

func scalePoints(p []Point, k float64) []Point {
	out := make([]Point, len(p))
	for i, pt := range p {
		out[i] = Point{At: pt.At * k, Value: pt.Value}
	}
	return out
}

func (a AutoFilter) inSeconds(beat float64) Effect {
	a.Points = scalePoints(a.Points, beat)
	return a
}

func (f Fader) inSeconds(beat float64) Effect {
	f.Points = scalePoints(f.Points, beat)
	return f
}

func (p Panner) inSeconds(beat float64) Effect {
	p.Points = scalePoints(p.Points, beat)
	return p
}

func (ts TapeStop) inSeconds(beat float64) Effect {
	ts.At, ts.Length, ts.Until = ts.At*beat, ts.Length*beat, ts.Until*beat
	return ts
}

func (d Delay) inSeconds(beat float64) Effect {
	d.Time *= beat
	return d
}

func (sc Sidechain) inSeconds(beat float64) Effect {
	hits := make([]float64, len(sc.Hits))
	for i, h := range sc.Hits {
		hits[i] = h * beat
	}
	sc.Hits = hits
	return sc
}

// Pedal lifts this long before the bar line, so a held chord doesn't smear into the next.
const pedalLift = 0.02

func (a Arrangement) beat() float64 {
	if a.Tempo <= 0 {
		return 1
	}
	return 60 / a.Tempo
}

func (a Arrangement) meter() Meter { return Meter{Beat: a.beat(), Bars: a.Meter} }

// BarStart is where bar b (counted from 0) begins, in beats.
func (m Meter) BarStart(b int) float64 {
	bars := m.bars()
	beats := (b / len(bars)) * m.Cycle()
	for _, n := range bars[:b%len(bars)] {
		beats += n
	}
	return float64(beats)
}

func (m Meter) BarBeats(b int) int { return m.bars()[b%len(m.bars())] }

func (a Arrangement) bars() int {
	n := 0
	for _, s := range a.Sections {
		n += s.Bars
	}
	return n
}

// sectionBar is the bar (from 0) the named section starts at.
func (a Arrangement) sectionBar(name string) (int, bool) {
	bar := 0
	for _, s := range a.Sections {
		if s.Name == name {
			return bar, true
		}
		bar += s.Bars
	}
	return 0, false
}

func (c Clip) bars() int {
	if len(c.Bars) > 0 {
		return len(c.Bars)
	}
	return max(c.Length, 1)
}

func (t Track) clip(name string) (Clip, bool) {
	for _, c := range t.Clips {
		if c.Name == name {
			return c, true
		}
	}
	return Clip{}, false
}

// notes plays the clip from bar (from 0).
func (c Clip) notes(m Meter, bar int) []Note {
	at := m.BarStart(bar) * m.Beat
	var notes []Note
	if len(c.Bars) > 0 {
		bars := c.Bars
		if c.Hit != "" {
			bars = make([]string, len(c.Bars))
			for i, b := range c.Bars {
				bars[i] = Pattern(c.Hit, b)
			}
		}
		notes = Phrase{
			At: at, Step: c.Step * m.Beat,
			Gate: c.Gate, Vel: c.Vel, Strum: c.Strum, Swing: c.Swing, Glide: c.Glide,
			Humanize: c.Humanize, Seed: c.Seed,
		}.Notes(bars...)
	} else {
		for _, n := range c.Notes {
			n.At = at + n.At*m.Beat
			n.Dur *= m.Beat
			notes = append(notes, n)
		}
	}
	if c.Pedal {
		for i, n := range notes {
			b := bar
			for m.BarStart(b+1)*m.Beat <= n.At+1e-6 {
				b++
			}
			notes[i].Dur = max(n.Dur, m.BarStart(b+1)*m.Beat-n.At-pedalLift)
		}
	}
	return notes
}

func (t Track) notes(m Meter) []Note {
	var notes []Note
	for _, p := range t.Play {
		c, ok := t.clip(p.Clip)
		if !ok {
			continue
		}
		for k := range max(p.Times, 1) {
			notes = append(notes, c.notes(m, p.Bar-1+k*c.bars())...)
		}
	}
	return notes
}

func (a Arrangement) track(name string) (Track, bool) {
	for _, t := range a.Tracks {
		if t.Name == name {
			return t, true
		}
	}
	return Track{}, false
}

// keys lists the tracks a Sidechain is keyed by.
func keys(key string) []string {
	var names []string
	for k := range strings.SplitSeq(key, ",") {
		names = append(names, strings.TrimSpace(k))
	}
	return names
}

func (a Arrangement) chain(fx []Effect, notes map[string][]Note) []Effect {
	if fx == nil {
		return nil
	}
	out := make([]Effect, len(fx))
	for i, e := range fx {
		if b, ok := e.(beatTimed); ok && a.Tempo > 0 {
			e = b.inSeconds(a.beat())
		}
		if sc, ok := e.(Sidechain); ok && sc.Key != "" {
			var on []Note
			for _, k := range keys(sc.Key) {
				on = append(on, notes[k]...)
			}
			sc.Hits, sc.Key = Onsets(on), ""
			e = sc
		}
		out[i] = e
	}
	return out
}

// Song compiles the arrangement to the Song the player and the bake take.
func (a Arrangement) Song() Song {
	m := a.meter()
	notes := make(map[string][]Note, len(a.Tracks))
	for _, t := range a.Tracks {
		notes[t.Name] = t.notes(m)
	}

	s := Song{
		ID: a.ID, Name: a.Name,
		Length: m.BarStart(a.bars()) * m.Beat,
		Meter:  m,
		Vary:   a.Vary, Voices: a.Voices, Priority: a.Priority,
		Target: a.Target,
		Master: a.chain(a.Master, notes),
	}
	if a.Tempo <= 0 {
		s.Meter, s.Length = Meter{}, 0
	}
	if bar, ok := a.sectionBar(a.Loop); ok && a.Loop != "" {
		s.Loop, s.LoopStart = true, m.BarStart(bar)*m.Beat
	}
	for _, t := range a.Tracks {
		if t.Mute {
			continue
		}
		s.Parts = append(s.Parts, Part{
			Name: t.Name, Patch: t.Patch, Notes: notes[t.Name],
			Level: t.Level, Pan: t.Pan, Out: t.Out,
			Inserts: a.chain(t.Inserts, notes),
			Sends:   t.Sends,
		})
	}
	for _, b := range a.Buses {
		b.Effects = a.chain(b.Effects, notes)
		s.Buses = append(s.Buses, b)
	}
	return s
}

// Validate checks what Song can't express: plays of clips that don't exist, bars that
// don't fill their meter, plays that overlap, and names that point nowhere; then the
// song it compiles to.
func (a Arrangement) Validate() error {
	var errs []error
	fail := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	if a.Tempo < 0 {
		fail("negative Tempo")
	}
	if slices.Contains(a.Meter, 0) || slices.ContainsFunc(a.Meter, func(n int) bool { return n < 0 }) {
		fail("Meter %v has a bar without beats", a.Meter)
	}
	sections := map[string]bool{}
	for _, s := range a.Sections {
		if sections[s.Name] {
			fail("section %q named twice", s.Name)
		}
		sections[s.Name] = true
		if s.Bars <= 0 {
			fail("section %q has no bars", s.Name)
		}
	}
	if a.Loop != "" && !sections[a.Loop] {
		fail("Loop %q names no section", a.Loop)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	m := a.meter()
	total := a.bars()
	buses := map[string]bool{}
	for _, b := range a.Buses {
		buses[b.Name] = true
	}
	tracks := map[string]bool{}
	for _, t := range a.Tracks {
		if tracks[t.Name] {
			fail("track %q named twice", t.Name)
		}
		tracks[t.Name] = true
	}
	checkChain := func(where string, fx []Effect) {
		for i, e := range fx {
			sc, ok := e.(Sidechain)
			if !ok || sc.Key == "" {
				continue
			}
			if len(sc.Hits) > 0 {
				fail("%s effect %d: a Sidechain takes Key or Hits, not both", where, i)
			}
			for _, k := range keys(sc.Key) {
				if !tracks[k] {
					fail("%s effect %d: Sidechain keyed by unknown track %q", where, i, k)
				}
			}
		}
	}
	checkChain("master", a.Master)
	for _, b := range a.Buses {
		checkChain(fmt.Sprintf("bus %q", b.Name), b.Effects)
	}

	for _, t := range a.Tracks {
		where := fmt.Sprintf("track %q", t.Name)
		checkChain(where, t.Inserts)
		for _, s := range t.Sends {
			if !buses[s.Bus] {
				fail("%s: send to unknown bus %q", where, s.Bus)
			}
		}
		if t.Out != "" && !buses[t.Out] {
			fail("%s: out to unknown bus %q", where, t.Out)
		}

		clips := map[string]bool{}
		for _, c := range t.Clips {
			if clips[c.Name] {
				fail("%s: clip %q named twice", where, c.Name)
			}
			clips[c.Name] = true
			if len(c.Bars) > 0 && len(c.Notes) > 0 {
				fail("%s clip %q: holds both Bars and Notes", where, c.Name)
			}
			if len(c.Bars) > 0 && c.Step <= 0 {
				fail("%s clip %q: Bars need a Step", where, c.Name)
			}
			if err := c.parse(); err != nil {
				fail("%s clip %q: %v", where, c.Name, err)
			}
		}

		type span struct{ from, to int }
		var spans []span
		for _, p := range t.Play {
			c, ok := t.clip(p.Clip)
			if !ok {
				fail("%s: play of unknown clip %q", where, p.Clip)
				continue
			}
			if p.Bar < 1 || p.Times < 0 {
				fail("%s: play of %q at bar %d, %d times", where, p.Clip, p.Bar, p.Times)
				continue
			}
			from, to := p.Bar-1, p.Bar-1+max(p.Times, 1)*c.bars()
			if to > total {
				fail("%s: play of %q at bar %d runs past the song's %d bars", where, p.Clip, p.Bar, total)
			}
			spans = append(spans, span{from, to})
			if len(c.Bars) == 0 || c.Step <= 0 {
				continue
			}
			for k := range max(p.Times, 1) {
				for i, b := range c.Bars {
					bar := from + k*c.bars() + i
					steps := float64(len(strings.Fields(strings.ReplaceAll(b, barToken, " "))))
					if want := float64(m.BarBeats(bar)); math.Abs(steps*c.Step-want) > 1e-9 {
						fail("%s clip %q bar %d: %g beats where bar %d has %g", where, c.Name, i+1, steps*c.Step, bar+1, want)
					}
				}
			}
		}
		slices.SortFunc(spans, func(x, y span) int { return x.from - y.from })
		for i := 1; i < len(spans); i++ {
			if spans[i].from < spans[i-1].to {
				fail("%s: plays at bars %d and %d overlap", where, spans[i-1].from+1, spans[i].from+1)
			}
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return Validate(a.Song())
}

// parse reads the clip's bars, reporting a typo as an error rather than a panic.
func (c Clip) parse() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	c.notes(Meter{Beat: 1}, 0)
	return nil
}

// InBeats is the effect with its song positions and durations in beats at tempo, as an
// arrangement writes them.
func InBeats(e Effect, tempo float64) Effect {
	if b, ok := e.(beatTimed); ok && tempo > 0 {
		return b.inSeconds(tempo / 60)
	}
	return e
}
