package songs

import (
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
var goldenBrown = synth.Arrangement{
	ID: "goldenbrown", Name: "Golden Brown (intro)",
	Tempo: 180, Meter: []int{3, 3, 3, 4}, Key: "Bb minor",
	Sections: []synth.Section{
		{Name: "Solo", Bars: 4},
		{Name: "Double", Bars: 4},
		{Name: "Lute", Bars: 12},
		{Name: "Turn", Bars: 4},
	},
	Loop: "Lute",
	Tracks: []synth.Track{
		{
			Name: "left hand", Patch: patches.Harpsichord(), Pan: -0.3, Out: "harpsichords",
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{
				{
					Name:     "first",
					Step:     0.5,
					Gate:     0.95,
					Humanize: 0.005,
					Seed:     1,
					Bars: []string{
						"Bb2 - . F3 - Bb2",
						"F2 - . C3 - F2",
						"Gb2 - . Db3 - Gb2",
						"F3 - Eb3 - Db3 - C3 -",
					},
				},
				{
					Name:     "second",
					Step:     0.5,
					Gate:     0.95,
					Humanize: 0.005,
					Seed:     1,
					Bars: []string{
						"Bb2 - . F3 - Bb2",
						"F2 - . C3 - F2",
						"Ab2 - . Eb3 - Ab2",
						"Gb2 - F2 - Eb2 - Db2 -",
					},
				},
			},
			Play: []synth.Play{{Clip: "first", Bar: 1, Times: 5}, {Clip: "second", Bar: 21}},
		},
		{
			Name: "right hand", Patch: patches.Harpsichord(), Pan: -0.3, Out: "harpsichords",
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{
				{
					Name:     "first",
					Step:     0.5,
					Gate:     0.7,
					Strum:    0.006,
					Humanize: 0.005,
					Seed:     2,
					Bars: []string{
						". Bb3,Db4,F4 . . Bb3,Db4,F4 -",
						". Ab3,C4,F4 . . Ab3,C4,F4 -",
						". Bb3,Db4,Gb4 . . Bb3,Db4,Gb4 -",
						". Ab3,C4,F4 . Ab3,C4,F4 . Ab3,C4,F4 . Ab3,C4,F4",
					},
				},
				{
					Name:     "second",
					Step:     0.5,
					Gate:     0.7,
					Strum:    0.006,
					Humanize: 0.005,
					Seed:     2,
					Bars: []string{
						". Bb3,Db4,F4 . . Bb3,Db4,F4 -",
						". Ab3,C4,F4 . . Ab3,C4,F4 -",
						". Ab3,C4,Eb4 . . Ab3,C4,Eb4 -",
						". Gb3,Bb3,Db4 . Gb3,Bb3,Db4 . Gb3,Bb3,Db4 . Gb3,Bb3,Db4",
					},
				},
			},
			Play: []synth.Play{{Clip: "first", Bar: 1, Times: 5}, {Clip: "second", Bar: 21}},
		},
		{
			Name: "double left", Patch: patches.Harpsichord(), Level: -1.9, Pan: 0.4, Out: "harpsichords",
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{
				{
					Name:     "first",
					Step:     0.5,
					Gate:     0.95,
					Humanize: 0.005,
					Seed:     3,
					Bars: []string{
						"Bb2 - . F3 - Bb2",
						"F2 - . C3 - F2",
						"Gb2 - . Db3 - Gb2",
						"F3 - Eb3 - Db3 - C3 -",
					},
				},
				{
					Name:     "second",
					Step:     0.5,
					Gate:     0.95,
					Humanize: 0.005,
					Seed:     3,
					Bars: []string{
						"Bb2 - . F3 - Bb2",
						"F2 - . C3 - F2",
						"Ab2 - . Eb3 - Ab2",
						"Gb2 - F2 - Eb2 - Db2 -",
					},
				},
			},
			Play: []synth.Play{{Clip: "first", Bar: 5, Times: 4}, {Clip: "second", Bar: 21}},
		},
		{
			Name: "double right", Patch: patches.Harpsichord(), Level: -1.9, Pan: 0.4, Out: "harpsichords",
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{
				{
					Name:     "first",
					Step:     0.5,
					Gate:     0.7,
					Strum:    0.006,
					Humanize: 0.005,
					Seed:     4,
					Bars: []string{
						". Bb3,Db4,F4 . . Bb3,Db4,F4 -",
						". Ab3,C4,F4 . . Ab3,C4,F4 -",
						". Bb3,Db4,Gb4 . . Bb3,Db4,Gb4 -",
						". Ab3,C4,F4 . Ab3,C4,F4 . Ab3,C4,F4 . Ab3,C4,F4",
					},
				},
				{
					Name:     "second",
					Step:     0.5,
					Gate:     0.7,
					Strum:    0.006,
					Humanize: 0.005,
					Seed:     4,
					Bars: []string{
						". Bb3,Db4,F4 . . Bb3,Db4,F4 -",
						". Ab3,C4,F4 . . Ab3,C4,F4 -",
						". Ab3,C4,Eb4 . . Ab3,C4,Eb4 -",
						". Gb3,Bb3,Db4 . Gb3,Bb3,Db4 . Gb3,Bb3,Db4 . Gb3,Bb3,Db4",
					},
				},
			},
			Play: []synth.Play{{Clip: "first", Bar: 5, Times: 4}, {Clip: "second", Bar: 21}},
		},
		{
			Name: "lute stop", Patch: buffStop(), Level: 1.6, Pan: 0.1,
			Sends: []synth.Send{{Bus: "room", Level: -10.5}},
			Clips: []synth.Clip{
				{Name: "first", Step: 0.5, Gate: 0.85, Humanize: 0.004, Seed: 5, Bars: []string{
					"Db5 Bb4 F4 Bb4 Db5 Bb4",
					"C5 Ab4 F4 Ab4 C5 Ab4",
					"Db5 Bb4 Gb4 Bb4 Db5 Bb4",
					"C5 Ab4 F4 Ab4 C5 Ab4 F4 Ab4",
				}},
				{Name: "second", Step: 0.5, Gate: 0.85, Humanize: 0.004, Seed: 5, Bars: []string{
					"Db5 Bb4 F4 Bb4 Db5 Bb4",
					"C5 Ab4 F4 Ab4 C5 Ab4",
					"C5 Ab4 Eb4 Ab4 C5 Ab4",
					"Db5 Bb4 Gb4 Bb4 Db5 Bb4 Gb4 F4",
				}},
			},
			Play: []synth.Play{{Clip: "first", Bar: 9, Times: 3}, {Clip: "second", Bar: 21}},
		},
		{
			Name: "bass", Patch: patches.HofnerBass(), Level: -7.5,
			Inserts: []synth.Effect{synth.Compressor{Threshold: -20, Ratio: 4, Attack: 0.01, Makeup: 4}},
			Sends:   []synth.Send{{Bus: "room", Level: -28}},
			Clips: []synth.Clip{
				{Name: "first", Step: 0.5, Gate: 0.9, Humanize: 0.004, Seed: 7, Bars: []string{
					"Bb1 - - . F2 Bb1",
					"F1 - - . C2 F1",
					"Gb1 - - . Db2 Gb1",
					"F1 - Eb1 - Db1 - C1 -",
				}},
				{Name: "second", Step: 0.5, Gate: 0.9, Humanize: 0.004, Seed: 7, Bars: []string{
					"Bb1 - - . F2 Bb1",
					"F1 - - . C2 F1",
					"Ab1 - - . Eb2 Ab1",
					"Gb1 - F1 - Eb1 - Db1 -",
				}},
			},
			Play: []synth.Play{{Clip: "first", Bar: 9, Times: 3}, {Clip: "second", Bar: 21}},
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
	Target: -14,
}

func GoldenBrown() synth.Song { return goldenBrown.Song() }

// The buff stop presses felt against the strings: a darker, shorter note, and no 4'.
func buffStop() synth.Patch {
	p := patches.Harpsichord()
	p.Oscs = []synth.Osc{p.Oscs[0], p.Oscs[1], p.Oscs[3]}
	for i := range 2 {
		p.Oscs[i].String.Decay = 1.4
		p.Oscs[i].String.Brightness = 0.45
	}
	return p
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
