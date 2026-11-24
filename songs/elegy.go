package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Elegy", for solo piano, in D-flat major at 63.
//
// One eight-bar tune over a bass that walks down the scale, Db C Bb Ab Gb F Eb, and
// turns on Ab. It opens on a rising sixth and falls back by step. The answer climbs
// higher and darkens on bar six, where the IV turns minor over an A in the bass.
//
// It is played four times, each fuller: alone over a rocking left hand, with a second
// voice beneath and the left hand opened out, then after a middle section in Bb minor
// in filled octaves over a low octave and a sixteenth sweep. The last time a scale
// runs up into a deceptive cadence on A major, the bVI, struck at full force with A0
// in the bass, then falls through Gb minor to a slow cadence home. The coda is the
// opening, an octave higher and very soft, with the minor iv once more before the end.
//
// The tracks are what one pianist's hands play: the tune on top, the inner voices of
// the right hand, the left hand under a pedal that changes at every bar, and the low
// octaves the left hand strikes before each sweep.
var elegy = synth.Arrangement{
	ID: "elegy", Name: "Elegy",
	Tempo: 63, Key: "Db major",
	Sections: []synth.Section{
		{Name: "Intro", Bars: 2},
		{Name: "Tune", Bars: 8},
		{Name: "Answer", Bars: 8},
		{Name: "Middle", Bars: 8},
		{Name: "Octaves", Bars: 8},
		{Name: "Climax", Bars: 8},
		{Name: "Peak", Bars: 2},
		{Name: "Home", Bars: 3},
		{Name: "Coda", Bars: 6},
	},
	Tracks: []synth.Track{
		{
			Name: "tune", Patch: patches.GrandPiano(), Level: 3, Pan: 0.2,
			Sends: []synth.Send{{Bus: "hall", Level: -12}},
			Clips: []synth.Clip{
				{
					Name: "A", Step: 0.5, Vel: 0.5, Humanize: 0.008, Seed: 1,
					Bars: []string{
						"Ab4 - - - F5 - - -",
						"Eb5 - - - - - Db5 C5",
						"Db5 - - - - - C5 Bb4",
						"C5 - - - Ab4 - - -",
						"Bb4 - - - Gb5 - - -",
						"F5 - - - - - Eb5 Db5",
						"Gb5 - - - F5 - Eb5 -",
						"Eb5 - - - C5 - - -",
					},
				},
				{
					Name: "A2", Step: 0.5, Vel: 0.56, Humanize: 0.008, Seed: 1,
					Bars: []string{
						"Ab4 - - - F5 - - -",
						"Eb5 - - - - - Db5 C5",
						"Db5 - - - Bb5 - - -",
						"Ab5 - - - F5 - - -",
						"Bb5 - - - - - Ab5 Gb5",
						"A5 - - - - - Ab5 Gb5",
						"F5 - - - - - Eb5 Db5",
						"Eb5 - - - - - C5 -",
					},
				},
				{
					Name: "B1", Step: 0.5, Vel: 0.62, Humanize: 0.008, Seed: 1,
					Bars: []string{
						"Bb4 - C5 Db5 F5 - - -",
						"Gb5 - - - F5 - Db5 -",
						"Db5 - Eb5 F5 Ab5 - - -",
						"Ab5 - - - Eb5 - C5 -",
					},
				},
				{
					Name: "B2", Step: 0.5, Vel: 0.7, Humanize: 0.008, Seed: 1,
					Bars: []string{
						"Db5 - Eb5 F5 Bb5 - - -",
						"Bb5 - - - Db6 - - -",
						"Db6 - - - Bb5 - - Gb5",
						"Ab5 - - - Gb5 - Eb5 -",
					},
				},
				{
					Name: "A3", Step: 0.25, Vel: 0.85, Humanize: 0.008, Seed: 1,
					Bars: []string{
						"Ab5 - - - - - - - F6 - - - - - - -",
						"Eb6 - - - - - - - - - - - Db6 - C6 -",
						"Db6 - - - - - - - - - - - C6 - Bb5 -",
						"C6 - - - - - - - Ab5 - - - - - - -",
						"Bb5 - - - - - - - Gb6 - - - - - - -",
						"F6 - - - - - - - - - - - Eb6 - Db6 -",
						"Gb6 - - - - - - - F6 - - - Eb6 - - -",
						"Eb6 - - - - - - - C6 - - - - - - -",
					},
				},
				{
					Name: "A4", Step: 0.25, Vel: 0.92, Humanize: 0.008, Seed: 1,
					Bars: []string{
						"Ab5 - - - - - - - F6 - - - - - - -",
						"Eb6 - - - - - - - - - - - Db6 - C6 -",
						"Db6 - - - - - - - Bb6 - - - - - - -",
						"Ab6 - - - - - - - F6 - - - - - - -",
						"Bb6 - - - - - - - - - - - Ab6 - Gb6 -",
						"A6 - - - - - - - - - - - Ab6 - Gb6 -",
						"F6 - - - - - - - - - - - Eb6 - Db6 -",
						"Eb6 - - - - - - - C5 Db5 Eb5 F5 Gb5 Ab5 Bb5 C6",
					},
				},
				{
					Name: "peak", Step: 0.25, Vel: 0.95, Humanize: 0.008, Seed: 1,
					Bars: []string{
						"Db6! - - - - - - - - - - - B5 - A5 -",
						"A5 - - - - - - - - - - - Ab5 - Gb5 -",
					},
				},
				{
					Name: "home", Step: 0.5, Vel: 0.62, Humanize: 0.008, Seed: 1,
					Bars: []string{
						"F5 - - - - - - -",
						"Eb5 - - - - - - -",
						"Db5 - - - - - - -",
					},
				},
				{
					Name: "coda", Step: 0.5, Vel: 0.4, Humanize: 0.008, Seed: 1,
					Bars: []string{
						"Ab5 - - - F6 - - -",
						"Eb6 - - - - - Db6 C6",
						"A5 - - - - - - -",
						"Ab5 - - - - - - -",
						"- - - - - - - -",
						"- - - - - - - .",
					},
				},
			},
			Play: []synth.Play{{Clip: "A", Bar: 3}, {Clip: "A2", Bar: 11}, {Clip: "B1", Bar: 19}, {Clip: "B2", Bar: 23}, {Clip: "A3", Bar: 27}, {Clip: "A4", Bar: 35}, {Clip: "peak", Bar: 43}, {Clip: "home", Bar: 45}, {Clip: "coda", Bar: 48}},
		},
		{
			Name: "inner", Patch: patches.GrandPiano(), Level: -2.5, Pan: 0.1,
			Sends: []synth.Send{{Bus: "hall", Level: -12}},
			Clips: []synth.Clip{
				{
					Name: "A2", Step: 0.5, Vel: 0.34, Humanize: 0.008, Seed: 2,
					Bars: []string{
						"F4 - - - Db5 - - -",
						"C5 - - - - - Ab4 Ab4",
						"Bb4 - - - F5 - - -",
						"F5 - - - C5 - - -",
						"Gb5 - - - - - Db5 Db5",
						"Gb5 - - - - - Db5 Db5",
						"Db5 - - - - - Ab4 Ab4",
						"C5 - - - - - Ab4 -",
					},
				},
				{
					Name: "B1", Step: 0.5, Vel: 0.4, Humanize: 0.008, Seed: 2,
					Bars: []string{
						"Bb3,Db4,F4 - - - F4,Bb4,Db5 - - -",
						"Gb4,Bb4,Db5 - - - Gb4,Bb4,Db5 - - -",
						"Db4,F4,Ab4 - - - Ab4,Db5,F5 - - -",
						"Ab4,C5,Eb5 - - - Eb4,Ab4,C5 - - -",
					},
				},
				{
					Name: "B2", Step: 0.5, Vel: 0.46, Humanize: 0.008, Seed: 2,
					Bars: []string{
						"Db4,F4,Bb4 - - - Bb4,Db5,F5 - - -",
						"Bb4,Db5,Gb5 - - - Db5,Gb5,Bb5 - - -",
						"Eb5,Gb5,Bb5 - - - Db5,Eb5,Gb5 - - -",
						"Ab4,Db5,Eb5 - - - Ab4,Db5,Eb5 - - -",
					},
				},
				{
					Name: "A3", Step: 0.25, Vel: 0.6, Humanize: 0.008, Seed: 2,
					Bars: []string{
						"Ab4,Db5,F5 - - - - - - - F5,Ab5,Db6 - - - - - - -",
						"Eb5,Ab5,C6 - - - - - - - - - - - Db5,Eb5,Ab5 - C5,Eb5,Ab5 -",
						"Db5,F5,Bb5 - - - - - - - - - - - C5,Db5,F5 - Bb4,Db5,F5 -",
						"C5,F5,Ab5 - - - - - - - Ab4,C5,F5 - - - - - - -",
						"Bb4,Db5,Gb5 - - - - - - - Gb5,Bb5,Db6 - - - - - - -",
						"F5,Ab5,Db6 - - - - - - - - - - - Eb5,F5,Ab5 - Db5,F5,Ab5 -",
						"Gb5,Bb5,Db6,Eb6 - - - - - - - F5,Gb5,Bb5,Db6 - - - Eb5,Gb5,Bb5 - - -",
						"Eb5,Gb5,Ab5,C6 - - - - - - - C5,Eb5,Gb5,Ab5 - - - - - - -",
					},
				},
				{
					Name: "A4", Step: 0.25, Vel: 0.68, Humanize: 0.008, Seed: 2,
					Bars: []string{
						"Ab4,Db5,F5 - - - - - - - F5,Ab5,Db6 - - - - - - -",
						"Eb5,Ab5,C6 - - - - - - - - - - - Db5,Eb5,Ab5 - C5,Eb5,Ab5 -",
						"Db5,F5,Bb5 - - - - - - - Bb5,Db6,F6 - - - - - - -",
						"Ab5,C6,F6 - - - - - - - F5,Ab5,C6 - - - - - - -",
						"Bb5,Db6,Gb6 - - - - - - - - - - - Ab5,Bb5,Db6 - Gb5,Bb5,Db6 -",
						"A5,Db6,Gb6 - - - - - - - - - - - Ab5,A5,Db6 - Gb5,A5,Db6 -",
						"F5,Ab5,Db6 - - - - - - - - - - - Eb5,F5,Ab5 - Db5,F5,Ab5 -",
						"Eb5,Gb5,Ab5,C6 - - - - - - - . . . . . . . .",
					},
				},
				{
					Name: "peak", Step: 0.25, Vel: 0.75, Humanize: 0.008, Seed: 2,
					Bars: []string{
						"Db5,E5,A5 - - - - - - - - - - - B4,Db5,E5 - A4,Db5,E5 -",
						"A4,Db5,Gb5 - - - - - - - - - - - Ab4,A4,Db5 - Gb4,A4,Db5 -",
					},
				},
				{
					Name: "home", Step: 0.5, Vel: 0.42, Humanize: 0.008, Seed: 2,
					Bars: []string{
						"F4,Ab4,Db5 - - - - - - -",
						"Eb4,Gb4,Ab4,C5 - - - - - - -",
						"Db4,F4,Ab4 - - - - - - -",
					},
				},
			},
			Play: []synth.Play{{Clip: "A2", Bar: 11}, {Clip: "B1", Bar: 19}, {Clip: "B2", Bar: 23}, {Clip: "A3", Bar: 27}, {Clip: "A4", Bar: 35}, {Clip: "peak", Bar: 43}, {Clip: "home", Bar: 45}},
		},
		{
			Name: "left hand", Patch: patches.GrandPiano(), Level: 0, Pan: -0.2,
			Sends: []synth.Send{{Bus: "hall", Level: -12}},
			Clips: []synth.Clip{
				{
					Name: "intro", Step: 0.5, Vel: 0.3, Humanize: 0.008, Seed: 3, Pedal: true,
					Bars: []string{
						"Db3 Ab3 Db4 F4 Db4 Ab3 Db4 Ab3",
						"Db3 Ab3 Db4 F4 Db4 Ab3 Db4 Ab3",
					},
				},
				{
					Name: "A", Step: 0.5, Vel: 0.32, Humanize: 0.008, Seed: 3, Pedal: true,
					Bars: []string{
						"Db3 Ab3 Db4 F4 Db4 Ab3 Db4 Ab3",
						"C3 Ab3 C4 Eb4 C4 Ab3 C4 Ab3",
						"Bb2 F3 Bb3 Db4 Bb3 F3 Bb3 F3",
						"Ab2 F3 Ab3 C4 Ab3 F3 Ab3 F3",
						"Gb2 Db3 Gb3 Bb3 Gb3 Db3 Gb3 Db3",
						"F2 Db3 F3 Ab3 F3 Db3 F3 Db3",
						"Eb2 Bb2 Db3 Eb3 Db3 Bb2 Db3 Bb2",
						"Ab2 Eb3 Gb3 Ab3 Gb3 Eb3 Gb3 Eb3",
					},
				},
				{
					Name: "A2", Step: 0.5, Vel: 0.36, Humanize: 0.008, Seed: 3, Pedal: true,
					Bars: []string{
						"Db2 Ab2 Db3 F3 Ab3 Db4 Ab3 Db3",
						"C2 Ab2 C3 Eb3 Ab3 C4 Ab3 C3",
						"Bb1 F2 Bb2 Db3 F3 Bb3 F3 Bb2",
						"Ab1 F2 Ab2 C3 F3 Ab3 F3 Ab2",
						"Gb1 Db2 Gb2 Bb2 Db3 Gb3 Db3 Gb2",
						"A1 Gb2 A2 Db3 Gb3 A3 Gb3 A2",
						"Ab1 F2 Ab2 Db3 F3 Ab3 F3 Ab2",
						"Ab1 Eb2 Gb2 Ab2 C3 Eb3 C3 Gb2",
					},
				},
				{
					Name: "B1", Step: 0.5, Vel: 0.44, Humanize: 0.008, Seed: 3, Pedal: true,
					Bars: []string{
						"Bb1 F2 Bb2 Db3 F3 Bb3 F3 Bb2",
						"Gb1 Db2 Gb2 Bb2 Db3 Gb3 Db3 Gb2",
						"Db2 Ab2 Db3 F3 Ab3 Db4 Ab3 Db3",
						"Ab1 Eb2 Ab2 C3 Eb3 Ab3 Eb3 Ab2",
					},
				},
				{
					Name: "B2", Step: 0.5, Vel: 0.5, Humanize: 0.008, Seed: 3, Pedal: true,
					Bars: []string{
						"Bb1 F2 Bb2 Db3 F3 Bb3 F3 Bb2",
						"Gb1 Db2 Gb2 Bb2 Db3 Gb3 Db3 Gb2",
						"Eb2 Bb2 Db3 Eb3 Gb3 Bb3 Gb3 Db3",
						"Ab1 Eb2 Gb2 Ab2 Db3 Eb3 Db3 Gb2",
					},
				},
				{
					Name: "A3", Step: 0.25, Vel: 0.52, Humanize: 0.008, Seed: 3, Pedal: true,
					Bars: []string{
						". Ab2 Db3 F3 Ab3 Db4 F4 Ab4 F4 Db4 Ab3 F3 Ab3 F3 Db3 Ab2",
						". Ab2 C3 Eb3 Ab3 C4 Eb4 Ab4 Eb4 C4 Ab3 Eb3 Ab3 Eb3 C3 Ab2",
						". F2 Bb2 Db3 F3 Bb3 Db4 F4 Db4 Bb3 F3 Db3 F3 Db3 Bb2 F2",
						". F2 Ab2 C3 F3 Ab3 C4 F4 C4 Ab3 F3 C3 F3 C3 Ab2 F2",
						". Db2 Gb2 Bb2 Db3 Gb3 Bb3 Db4 Bb3 Gb3 Db3 Bb2 Db3 Bb2 Gb2 Db2",
						". Db2 F2 Ab2 Db3 F3 Ab3 Db4 Ab3 F3 Db3 Ab2 Db3 Ab2 F2 Db2",
						". Bb1 Db2 Eb2 Gb2 Bb2 Db3 Eb3 Db3 Bb2 Gb2 Eb2 Gb2 Eb2 Db2 Bb1",
						". Eb2 Gb2 Ab2 C3 Eb3 Gb3 Ab3 Gb3 Eb3 C3 Ab2 C3 Ab2 Gb2 Eb2",
					},
				},
				{
					Name: "A4", Step: 0.25, Vel: 0.6, Humanize: 0.008, Seed: 3, Pedal: true,
					Bars: []string{
						". Ab2 Db3 F3 Ab3 Db4 F4 Ab4 F4 Db4 Ab3 F3 Ab3 F3 Db3 Ab2",
						". Ab2 C3 Eb3 Ab3 C4 Eb4 Ab4 Eb4 C4 Ab3 Eb3 Ab3 Eb3 C3 Ab2",
						". F2 Bb2 Db3 F3 Bb3 Db4 F4 Db4 Bb3 F3 Db3 F3 Db3 Bb2 F2",
						". F2 Ab2 C3 F3 Ab3 C4 F4 C4 Ab3 F3 C3 F3 C3 Ab2 F2",
						". Db2 Gb2 Bb2 Db3 Gb3 Bb3 Db4 Bb3 Gb3 Db3 Bb2 Db3 Bb2 Gb2 Db2",
						". Gb2 A2 Db3 Gb3 A3 Db4 Gb4 Db4 A3 Gb3 Db3 Gb3 Db3 A2 Gb2",
						". F2 Ab2 Db3 F3 Ab3 Db4 F4 Db4 Ab3 F3 Db3 F3 Db3 Ab2 F2",
						". Eb2 Gb2 Ab2 C3 Eb3 Gb3 Ab3 Gb3 Eb3 C3 Ab2 C3 Ab2 Gb2 Eb2",
					},
				},
				{
					Name: "peak", Step: 0.25, Vel: 0.66, Humanize: 0.008, Seed: 3, Pedal: true,
					Bars: []string{
						". E2 A2 Db3 E3 A3 Db4 E4 Db4 A3 E3 Db3 E3 Db3 A2 E2",
						". Db2 Gb2 A2 Db3 Gb3 A3 Db4 A3 Gb3 Db3 A2 Db3 A2 Gb2 Db2",
					},
				},
				{
					Name: "home", Step: 0.5, Vel: 0.45, Humanize: 0.008, Seed: 3, Pedal: true,
					Bars: []string{
						"Ab2,Db3,F3,Ab3 - - - Db3,F3,Ab3 - - -",
						"Ab2,C3,Eb3,Gb3 - - - C3,Eb3,Gb3 - - -",
						"Db3,F3,Ab3 - - - - - - -",
					},
				},
				{
					Name: "coda", Step: 0.5, Vel: 0.24, Humanize: 0.008, Seed: 3, Pedal: true,
					Bars: []string{
						"Db3 Ab3 Db4 F4 Db4 Ab3 Db4 Ab3",
						"C3 Ab3 C4 Eb4 C4 Ab3 C4 Ab3",
						"Db3 A3 Db4 Gb4 Db4 A3 Db4 A3",
						"Db3 Ab3 Db4 F4 Db4 Ab3 Db4 Ab3",
					},
				},
				{
					Name: "end", Step: 0.5, Vel: 0.3, Strum: 0.045, Humanize: 0.008, Seed: 3,
					Bars: []string{
						"Db1,Db2,Ab2,F3,Ab3,Db4,F4 - - - - - - -",
						"- - - - - - - .",
					},
				},
			},
			Play: []synth.Play{{Clip: "intro", Bar: 1}, {Clip: "A", Bar: 3}, {Clip: "A2", Bar: 11}, {Clip: "B1", Bar: 19}, {Clip: "B2", Bar: 23}, {Clip: "A3", Bar: 27}, {Clip: "A4", Bar: 35}, {Clip: "peak", Bar: 43}, {Clip: "home", Bar: 45}, {Clip: "coda", Bar: 48}, {Clip: "end", Bar: 52}},
		},
		{
			Name: "bass", Patch: patches.GrandPiano(), Level: 0, Pan: -0.3,
			Sends: []synth.Send{{Bus: "hall", Level: -14}},
			Clips: []synth.Clip{
				{
					Name: "A3", Step: 0.25, Vel: 0.72, Humanize: 0.008, Seed: 4, Pedal: true,
					Bars: []string{
						"Db1,Db2 - - - - - - - - - - - - - - -",
						"C1,C2 - - - - - - - - - - - - - - -",
						"Bb0,Bb1 - - - - - - - - - - - - - - -",
						"Ab1,Ab2 - - - - - - - - - - - - - - -",
						"Gb1,Gb2 - - - - - - - - - - - - - - -",
						"F1,F2 - - - - - - - - - - - - - - -",
						"Eb1,Eb2 - - - - - - - - - - - - - - -",
						"Ab1,Ab2 - - - - - - - - - - - - - - -",
					},
				},
				{
					Name: "A4", Step: 0.25, Vel: 0.82, Humanize: 0.008, Seed: 4, Pedal: true,
					Bars: []string{
						"Db1,Db2 - - - - - - - - - - - - - - -",
						"C1,C2 - - - - - - - - - - - - - - -",
						"Bb0,Bb1 - - - - - - - - - - - - - - -",
						"Ab1,Ab2 - - - - - - - - - - - - - - -",
						"Gb1,Gb2 - - - - - - - - - - - - - - -",
						"A0,A1 - - - - - - - - - - - - - - -",
						"Ab1,Ab2 - - - - - - - - - - - - - - -",
						"Ab1,Ab2 - - - - - - - - - - - - - - -",
					},
				},
				{
					Name: "peak", Step: 0.25, Vel: 0.92, Humanize: 0.008, Seed: 4, Pedal: true,
					Bars: []string{
						"A0,A1 - - - - - - - - - - - - - - -",
						"Gb1,Gb2 - - - - - - - - - - - - - - -",
					},
				},
				{
					Name: "home", Step: 0.5, Vel: 0.55, Humanize: 0.008, Seed: 4, Pedal: true,
					Bars: []string{
						"Ab1,Ab2 - - - - - - -",
						"Ab1,Ab2 - - - - - - -",
						"Db1,Db2 - - - - - - -",
					},
				},
			},
			Play: []synth.Play{{Clip: "A3", Bar: 27}, {Clip: "A4", Bar: 35}, {Clip: "peak", Bar: 43}, {Clip: "home", Bar: 45}},
		},
	},
	Buses: []synth.Bus{
		{Name: "hall", Effects: []synth.Effect{synth.Reverb{Size: 0.86, Damp: 0.45, PreDelay: 0.03}}},
	},
	Master: []synth.Effect{
		synth.EQ{Bands: []synth.Band{{Kind: synth.LowCut, Freq: 25}}},
		synth.Compressor{Threshold: -18, Ratio: 1.5, Attack: 0.03, Release: 0.4, Makeup: 10},
		synth.Limiter{},
	},
	Target: -18,
}

func Elegy() synth.Song { return elegy.Song() }
