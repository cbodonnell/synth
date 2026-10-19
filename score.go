package synth

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

const (
	restToken = "."
	holdToken = "-"
	barToken  = "|"
)

// Phrase reads a step-grid notation into Notes. Tokens are separated by whitespace and
// each takes one Step:
//
//	C4          a note
//	Bb3,Db4,F4  a chord
//	D5/E5       a note bent or slid to another, over Glide
//	-           holds the previous note or chord another step
//	.           a rest
//	|           a bar line, for the eye only; takes no time, and needs no spaces
//	F#2*3       a ratchet: three even hits inside the step (and any holds after it)
//
// A trailing ! accents a note or chord and a trailing ? ghosts it, after any ratchet.
type Phrase struct {
	At   float64 // seconds where the first token falls
	Step float64 // seconds per token

	Gate float64 // fraction of the written length that sounds; 0 means 1
	Vel  float64 // 0 means 0.8

	// Seconds between successive notes of a chord, lowest first: a harpsichordist's roll.
	Strum float64

	// Fraction of a step that every second step of the song is pushed late: 0.33 is a
	// triplet shuffle.
	Swing float64

	// Seconds a D5/E5 slide takes to arrive; 0 spreads it over the note.
	Glide float64

	// Up to this many seconds of reproducible timing and velocity drift per note, so a
	// machine-perfect line breathes a little. Seeded by Seed and the step the token falls
	// on, counted from the song's start, so a line written in pieces drifts as one.
	Humanize float64
	Seed     uint64
}

const (
	accentVel = 1.25
	ghostVel  = 0.55
)

func (p Phrase) Notes(bars ...string) []Note {
	text := strings.ReplaceAll(strings.Join(bars, " "), barToken, " ")
	tokens := strings.Fields(text)

	gate := p.Gate
	if gate <= 0 {
		gate = 1
	}
	vel := p.Vel
	if vel <= 0 {
		vel = 0.8
	}

	first := 0
	if p.Step > 0 {
		first = int(math.Round(p.At / p.Step))
	}

	var notes []Note
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == restToken || tok == holdToken {
			continue
		}
		steps := 1
		for i+steps < len(tokens) && tokens[i+steps] == holdToken {
			steps++
		}

		v := vel
		switch {
		case strings.HasSuffix(tok, "!"):
			v *= accentVel
			tok = tok[:len(tok)-1]
		case strings.HasSuffix(tok, "?"):
			v *= ghostVel
			tok = tok[:len(tok)-1]
		}

		at := p.At + float64(i)*p.Step
		step := first + i
		if step%2 != 0 {
			at += p.Swing * p.Step
		}
		if p.Humanize > 0 {
			r := newRNG(mixSeed(p.Seed, uint64(step)))
			at += p.Humanize * 0.5 * (r.float() + 1)
			v *= 1 + 0.15*r.float()
		}
		hits := 1
		if body, count, ok := strings.Cut(tok, "*"); ok {
			n, err := strconv.Atoi(count)
			if err != nil || n < 1 {
				panic(fmt.Sprintf("synth: bad ratchet %q", tok))
			}
			tok, hits = body, n
		}
		span := float64(steps) * p.Step / float64(hits)
		dur := span * gate

		for h := range hits {
			for k, name := range strings.Split(tok, ",") {
				n := Note{
					At:        at + float64(h)*span + float64(k)*p.Strum,
					Dur:       math.Max(dur-float64(k)*p.Strum, span*0.25),
					Vel:       math.Min(v, 1),
					SlideTime: p.Glide,
				}
				from, to, slides := strings.Cut(name, "/")
				n.Key = MustKey(from)
				if slides {
					n.Slide = MustKey(to) - n.Key
				}
				notes = append(notes, n)
			}
		}
		i += steps - 1
	}
	return notes
}

// Pattern writes a rhythm for one sound, for drum lines and strummed chords:
// x plays sound, X accents it, o ghosts it, any of them may ratchet (x*3), and every
// other token passes through.
//
//	Pattern("C2", "X . x . | . x*3 o x")
func Pattern(sound, rhythm string) string {
	var b strings.Builder
	for f := range strings.FieldsSeq(strings.ReplaceAll(rhythm, barToken, " | ")) {
		hit, ratchet := f, ""
		if i := strings.IndexByte(f, '*'); i > 0 {
			hit, ratchet = f[:i], f[i:]
		}
		switch hit {
		case "x":
			f = sound + ratchet
		case "X":
			f = sound + ratchet + "!"
		case "o":
			f = sound + ratchet + "?"
		}
		b.WriteString(f)
		b.WriteByte(' ')
	}
	return b.String()
}

// Onsets lists when notes start, once each, in order: the hits a Sidechain ducks on.
func Onsets(notes []Note) []float64 {
	var at []float64
	for _, n := range notes {
		at = append(at, n.At)
	}
	slices.Sort(at)
	return slices.Compact(at)
}

func Merge(lines ...[]Note) []Note {
	var all []Note
	for _, l := range lines {
		all = append(all, l...)
	}
	return all
}

// Repeat plays notes times over, each pass period seconds after the last.
func Repeat(notes []Note, times int, period float64) []Note {
	out := make([]Note, 0, len(notes)*max(times, 0))
	for k := range times {
		out = append(out, Shift(notes, float64(k)*period)...)
	}
	return out
}

func Shift(notes []Note, by float64) []Note {
	out := make([]Note, len(notes))
	for i, n := range notes {
		n.At += by
		out[i] = n
	}
	return out
}

func Transpose(notes []Note, semis float64) []Note {
	out := make([]Note, len(notes))
	for i, n := range notes {
		n.Key += semis
		out[i] = n
	}
	return out
}

// Hit is a single note through a patch: the usual shape of a sound effect.
func Hit(id string, p Patch, key, dur float64) Song {
	return Song{
		ID:    id,
		Name:  id,
		Parts: []Part{{Name: id, Patch: p, Notes: []Note{{Key: key, Dur: dur, Vel: 1}}}},
	}
}
