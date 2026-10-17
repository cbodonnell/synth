package songs

import (
	"strings"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// The Stranglers, "Golden Brown" (1981), the harpsichord intro, in B-flat minor at 180.
//
// The phrase is thirteen beats, bars of three, three, three, and four. The left hand
// plays root, a fifth tied across the beat, and root again; the right hand answers on
// the off-beat and lands on three. The four-beat bar walks the bass F, Eb, Db, C under
// off-beat chords. The last pass takes the second ending, Ab and then Gb over a bass
// that falls to Db, which turns back to the top.
//
// The first pass is one harpsichord, the second double-tracks it, and the loop adds a
// second manual on the lute stop running eighths, and the bass. Each token is an eighth.
const (
	gbBeat   = 60.0 / 180
	gbStep   = gbBeat / 2
	gbCycle  = 13 * gbBeat
	gbPasses = 6
)

// A hand's part: two bars every pass plays, then the first ending or, on the last pass, the second.
type gbHand struct{ head, first, second string }

// from is the first pass the hand plays; it rests through the ones before.
func (h gbHand) score(from int) string {
	var b strings.Builder
	b.WriteString(silence(from * 26))
	for pass := from; pass < gbPasses; pass++ {
		ending := h.first
		if pass == gbPasses-1 {
			ending = h.second
		}
		b.WriteString(h.head)
		b.WriteString(ending)
	}
	return b.String()
}

var (
	gbLeft = gbHand{
		head:   "Bb2 - . F3 - Bb2 | F2 - . C3 - F2 |",
		first:  "Gb2 - . Db3 - Gb2 | F3 - Eb3 - Db3 - C3 - |",
		second: "Ab2 - . Eb3 - Ab2 | Gb2 - F2 - Eb2 - Db2 - |",
	}
	gbRight = gbHand{
		head:   ". Bb3,Db4,F4 . . Bb3,Db4,F4 - | . Ab3,C4,F4 . . Ab3,C4,F4 - |",
		first:  ". Bb3,Db4,Gb4 . . Bb3,Db4,Gb4 - |" + synth.Pattern("Ab3,C4,F4", ". x . x . x . x |"),
		second: ". Ab3,C4,Eb4 . . Ab3,C4,Eb4 - |" + synth.Pattern("Gb3,Bb3,Db4", ". x . x . x . x |"),
	}
	gbLute = gbHand{
		head:   "Db5 Bb4 F4 Bb4 Db5 Bb4 | C5 Ab4 F4 Ab4 C5 Ab4 |",
		first:  "Db5 Bb4 Gb4 Bb4 Db5 Bb4 | C5 Ab4 F4 Ab4 C5 Ab4 F4 Ab4 |",
		second: "C5 Ab4 Eb4 Ab4 C5 Ab4 | Db5 Bb4 Gb4 Bb4 Db5 Bb4 Gb4 F4 |",
	}
	gbBass = gbHand{
		head:   "Bb1 - - . F2 Bb1 | F1 - - . C2 F1 |",
		first:  "Gb1 - - . Db2 Gb1 | F1 - Eb1 - Db1 - C1 - |",
		second: "Ab1 - - . Eb2 Ab1 | Gb1 - F1 - Eb1 - Db1 - |",
	}
)

func GoldenBrown() synth.Song {
	// One player's take on both hands: the left legato, the right detached and rolled.
	take := func(from int, seed uint64) []synth.Note {
		return synth.Merge(
			synth.Phrase{Step: gbStep, Gate: 0.95, Humanize: 0.005, Seed: seed}.Notes(gbLeft.score(from)),
			synth.Phrase{Step: gbStep, Gate: 0.7, Strum: 0.006, Humanize: 0.005, Seed: seed + 1}.Notes(gbRight.score(from)),
		)
	}
	lute := synth.Phrase{Step: gbStep, Gate: 0.85, Humanize: 0.004, Seed: 5}.Notes(gbLute.score(2))
	bass := synth.Phrase{Step: gbStep, Gate: 0.9, Humanize: 0.004, Seed: 7}.Notes(gbBass.score(2))

	// The buff stop presses felt against the strings: a darker, shorter note, and no 4'.
	lutePatch := patches.Harpsichord()
	lutePatch.Oscs = []synth.Osc{lutePatch.Oscs[0], lutePatch.Oscs[1], lutePatch.Oscs[3]}
	for i := range 2 {
		lutePatch.Oscs[i].String.Decay = 1.4
		lutePatch.Oscs[i].String.Brightness = 0.45
	}

	room := func(l float64) synth.Send { return synth.Send{Bus: "room", Level: l} }

	return synth.Song{
		ID:   "goldenbrown",
		Name: "Golden Brown (intro)",
		Parts: []synth.Part{
			{Name: "harpsichord", Patch: patches.Harpsichord(), Notes: take(0, 1), Pan: -0.3, Out: "harpsichords", Sends: []synth.Send{room(0.2)}},
			{Name: "harpsichord double", Patch: patches.Harpsichord(), Notes: take(1, 3), Pan: 0.4, Gain: 0.8, Out: "harpsichords", Sends: []synth.Send{room(0.2)}},
			{Name: "lute stop", Patch: lutePatch, Notes: lute, Pan: 0.1, Gain: 1.2, Sends: []synth.Send{room(0.3)}},
			{
				Name: "bass", Patch: patches.HofnerBass(), Notes: bass, Gain: 0.42,
				Inserts: []synth.Effect{synth.Compressor{Threshold: -20, Ratio: 4, Attack: 0.01, Makeup: 4}},
				Sends:   []synth.Send{room(0.04)},
			},
		},
		Buses: []synth.Bus{
			{Name: "harpsichords", Effects: harpsichordDesk()},
			{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.55, Damp: 0.5, PreDelay: 0.012, Width: 0.9}}},
		},
		Master: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowCut, Freq: 35},
				{Kind: synth.HighShelf, Freq: 11000, Gain: -1.5},
			}},
			synth.Compressor{Threshold: -14, Ratio: 2, Attack: 0.015, Release: 0.2, Makeup: 1},
			synth.Limiter{},
		},
		Length:    gbPasses * gbCycle,
		Loop:      true,
		LoopStart: 2 * gbCycle,
		Meter:     synth.Meter{Beat: gbBeat, Bars: []int{3, 3, 3, 4}},
		Target:    -14,
	}
}

// A close-miked harpsichord on a 1981 record: thinned below, a presence lift where the
// quill bites, a touch of grit from the desk, and held together.
func harpsichordDesk() []synth.Effect {
	return []synth.Effect{
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 85},
			{Kind: synth.Peak, Freq: 250, Gain: -2, Q: 0.8},
			{Kind: synth.Peak, Freq: 2800, Gain: 3, Q: 0.9},
		}},
		synth.Drive{Amount: 1.2, Tone: 9000, Mix: 0.3},
		synth.Compressor{Threshold: -18, Ratio: 2.5, Attack: 0.008, Release: 0.15, Makeup: 2},
	}
}
