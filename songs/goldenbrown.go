package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// The Stranglers, "Golden Brown" (1981), the harpsichord intro, in B-flat minor.
//
// The phrase is thirteen beats: three bars of three and one of four. The left hand
// alternates root and fifth under right-hand chords on the off-beats, and the four-beat
// bar walks the bass down F, Eb, Db, C back to the tonic. Each token is an eighth.
const (
	gbBeat  = 60.0 / 140
	gbStep  = gbBeat / 2
	gbCycle = 13 * gbBeat
)

func GoldenBrown() synth.Song {
	harpsichord := patches.Harpsichord()

	left := synth.Phrase{Step: gbStep, Gate: 0.95, Humanize: 0.004, Seed: 1}.Notes(
		"Bb2 - F3  - Bb2 - |",
		"F2  - C3  - F2  - |",
		"Gb2 - Db3 - Gb2 - |",
		"F3  - Eb3 - Db3 - C3 - |",
	)

	right := synth.Phrase{Step: gbStep, Gate: 0.8, Strum: 0.007, Humanize: 0.004, Seed: 2}.Notes(
		". Bb3,Db4,F4  . Bb3,Db4,F4  . Bb3,Db4,F4  |",
		". Ab3,C4,F4   . Ab3,C4,F4   . Ab3,C4,F4   |",
		". Bb3,Db4,Gb4 . Bb3,Db4,Gb4 . Bb3,Db4,Gb4 |",
		". Ab3,Db4,F4  . Ab3,Db4,F4  . Ab3,Db4,F4  . Ab3,Db4,F4 |",
	)

	// The second keyboard comes in on the third time round, doubling the riff's top
	// line in octaves from above.
	octaves := synth.Phrase{At: 2 * gbCycle, Step: gbStep, Gate: 0.97, Vel: 0.7}.Notes(
		"F4,F5   - - - - - | - - - - - - |",
		"Gb4,Gb5 - - - - - |",
		"F4,F5   - - - - - - - |",
	)

	room := synth.Send{Bus: "room", Level: 0.22}
	hall := synth.Send{Bus: "hall", Level: 0.45}

	return synth.Song{
		ID:   "goldenbrown",
		Name: "Golden Brown (intro)",
		Parts: []synth.Part{
			{
				Name: "harpsichord left", Patch: harpsichord, Pan: -0.2,
				Notes:   synth.Repeat(left, 4, gbCycle),
				Inserts: harpsichordChain(),
				Sends:   []synth.Send{room},
			},
			{
				Name: "harpsichord right", Patch: harpsichord, Pan: 0.2,
				Notes:   synth.Repeat(right, 4, gbCycle),
				Inserts: harpsichordChain(),
				Sends:   []synth.Send{room},
			},
			{
				Name: "octave synth", Patch: patches.OctaveSynth(),
				Notes: synth.Repeat(octaves, 2, gbCycle),
				Inserts: []synth.Effect{
					synth.Chorus{Rate: 0.6, Depth: 0.004, Mix: 0.5},
					synth.Delay{Time: 3 * gbStep, Feedback: 0.3, Damp: 2500, PingPong: true, Mix: 0.18},
				},
				Sends: []synth.Send{hall},
			},
		},
		Buses: []synth.Bus{
			{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.5, Damp: 0.55, PreDelay: 0.01, Width: 0.9}}},
			{Name: "hall", Effects: []synth.Effect{synth.Reverb{Size: 0.85, Damp: 0.35, PreDelay: 0.03}}},
		},
		Master: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowCut, Freq: 40},
				{Kind: synth.HighShelf, Freq: 11000, Gain: -1.5},
			}},
			synth.Compressor{Threshold: -14, Ratio: 2.5, Attack: 0.01, Release: 0.2, Makeup: 2},
		},
		Length:    4 * gbCycle,
		Loop:      true,
		LoopStart: 2 * gbCycle,
		Meter:     synth.Meter{Beat: gbBeat, Bars: []int{3, 3, 3, 4}},
	}
}

// A close-miked harpsichord on a 1981 record: thinned below, a presence lift where
// the quill bites, and a touch of grit from the desk.
func harpsichordChain() []synth.Effect {
	return []synth.Effect{
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 85},
			{Kind: synth.Peak, Freq: 250, Gain: -2, Q: 0.8},
			{Kind: synth.Peak, Freq: 2800, Gain: 3, Q: 0.9},
		}},
		synth.Drive{Amount: 1.2, Tone: 9000, Mix: 0.3},
	}
}
