package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Untitled": <genre> in <key> at <tempo>. <Instruments.> <Form: what each section does.>
var untitled = synth.Arrangement{
	ID: "untitled", Name: "Untitled",
	Tempo: 100, Key: "C major",
	Sections: []synth.Section{
		{Name: "Intro", Bars: 4},
		{Name: "A", Bars: 8},
		{Name: "B", Bars: 8},
	},
	Loop: "A",
	Tracks: []synth.Track{
		{
			Name: "kick", Patch: untitledKick(), Level: -2, Out: "drums",
			Clips: []synth.Clip{{Name: "beat", Step: 0.25, Hit: "C2", Bars: []string{
				"X . . . x . . . X . . . x . . .",
			}}},
			Play: []synth.Play{{Clip: "beat", Bar: 5, Times: 16}},
		},
		{
			Name: "snare", Patch: patches.Snare(), Level: 5, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -12}},
			Clips: []synth.Clip{{Name: "backbeat", Step: 0.25, Hit: "C3", Bars: []string{
				". . . . X . . . . . . . X . . o",
			}}},
			Play: []synth.Play{{Clip: "backbeat", Bar: 5, Times: 16}},
		},
		{
			Name: "hihat", Patch: patches.HiHat(), Level: 1, Pan: 0.3, Out: "drums",
			Clips: []synth.Clip{{Name: "eighths", Step: 0.25, Swing: 0.05, Humanize: 0.003, Seed: 1, Hit: "C5", Bars: []string{
				"x . o . x . o . x . o . x . o .",
			}}},
			Play: []synth.Play{{Clip: "eighths", Bar: 5, Times: 16}},
		},
		{
			Name: "bass", Patch: patches.Bass(), Level: 3,
			Inserts: []synth.Effect{synth.Sidechain{Key: "kick", Depth: 0.25, Release: 0.2}},
			Clips: []synth.Clip{{Name: "A", Step: 0.5, Gate: 0.85, Bars: []string{
				"C2 - - C2 . G1 - -",
				"A1 - - A1 . E2 - -",
				"F1 - - F1 . C2 - -",
				"G1 - - G1 . D2 - -",
			}}},
			Play: []synth.Play{{Clip: "A", Bar: 5, Times: 4}},
		},
		{
			Name: "chords", Patch: patches.SawPad(), Level: 3,
			Sends: []synth.Send{{Bus: "hall", Level: -8}},
			Clips: []synth.Clip{{Name: "A", Step: 1, Gate: 0.98, Vel: 0.7, Bars: []string{
				"E3,G3,C4 - - -",
				"E3,A3,C4 - - -",
				"F3,A3,C4 - - -",
				"D3,G3,B3 - - -",
			}}},
			Play: []synth.Play{{Clip: "A", Bar: 1, Times: 5}},
		},
		{
			Name: "lead", Patch: patches.Lead(), Level: 0, Pan: 0.1,
			Sends: []synth.Send{{Bus: "echo", Level: -10}, {Bus: "hall", Level: -12}},
			Clips: []synth.Clip{{Name: "tune", Step: 0.5, Gate: 0.9, Humanize: 0.005, Seed: 2, Bars: []string{
				"E5 - D5 C5 - - G4 -",
				"A4 - - C5 - - E5 -",
				"F5 - E5 D5 - C5 - -",
				"D5 - - - . . . .",
			}}},
			Play: []synth.Play{{Clip: "tune", Bar: 13, Times: 2}},
		},
	},
	Buses: []synth.Bus{
		{Name: "drums", Effects: []synth.Effect{
			synth.Compressor{Threshold: -18, Ratio: 4, Attack: 0.01, Release: 0.12, Makeup: 4},
		}},
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.3, Damp: 0.5}}},
		{Name: "hall", Effects: []synth.Effect{synth.Reverb{Size: 0.85, Damp: 0.45, PreDelay: 0.03}}},
		{Name: "echo", Effects: []synth.Effect{
			synth.Delay{Time: 0.75, Feedback: 0.4, Damp: 2500, PingPong: true},
		}},
	},
	Master: []synth.Effect{
		synth.EQ{Bands: []synth.Band{{Kind: synth.LowCut, Freq: 30}}},
		synth.Compressor{Threshold: -16, Ratio: 2, Attack: 0.02, Release: 0.25, Makeup: 3.5},
		synth.Limiter{},
	},
	Target: -14,
}

func Untitled() synth.Song { return untitled.Song() }

func untitledKick() synth.Patch {
	p := patches.Kick()
	p.Amp.Decay = 0.3
	return p
}
