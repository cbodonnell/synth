package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Daybreak": the last set on the beach as the sun comes up. Downtempo in D major at
// 108, a sister to Long Light with none of its instruments: a cassette-warm pulse pad,
// glass plucks rolling sixteenths, a picked bass singing up the neck, a gliding
// monosynth tune, vibraphone, and surf and birds under all of it.
//
// Eight bars of intro (the sea, the birds, the pad, and the plucks opening up), then a
// 32-bar loop: A is the groove on Dmaj9 A/C# Bm9 Gmaj9#11; A' brings in the lead; B
// lets the drums go for Em9 Gmaj9 D/F# Asus4-A under the vibraphone and builds back
// on a four-to-the-floor and a snare roll; A” is everything, the vibraphone answering.
var daybreak = synth.Arrangement{
	ID: "daybreak", Name: "Daybreak",
	Tempo: 108, Key: "D major",
	Sections: []synth.Section{
		{Name: "Intro", Bars: 8},
		{Name: "A", Bars: 8},
		{Name: "A'", Bars: 8},
		{Name: "B", Bars: 8},
		{Name: "A''", Bars: 8},
	},
	Loop: "A",
	Tracks: []synth.Track{
		{
			Name: "surf", Patch: patches.Surf(), Level: 2, Pan: -0.2,
			Sends: []synth.Send{{Bus: "haze", Level: -14}},
			Clips: []synth.Clip{{Name: "waves", Step: 1, Bars: []string{
				"D3 - - -",
				"- - - -",
				"- - - -",
				"- - - -",
			}}},
			Play: []synth.Play{{Clip: "waves", Bar: 1, Times: 10}},
		},
		{
			Name: "birds", Patch: patches.Bird(), Level: -3, Pan: 0.45,
			Sends: []synth.Send{{Bus: "haze", Level: -8}},
			Clips: []synth.Clip{{Name: "dawn", Step: 0.25, Humanize: 0.01, Seed: 11, Bars: []string{
				". . . . . . . . . . E7 . . C7 . .",
				". . . . . . . . . . . . . . . .",
				". . G7*3 . . . . . . . . . . . . .",
				". . . . . . D7 . A6 . . . . . . .",
			}}},
			Play: []synth.Play{{Clip: "dawn", Bar: 1, Times: 2}, {Clip: "dawn", Bar: 25, Times: 2}},
		},
		{
			Name: "kick", Patch: daybreakKick(), Level: -9, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -24}},
			Clips: []synth.Clip{
				{Name: "groove", Step: 0.25, Swing: 0.06, Humanize: 0.003, Seed: 8, Hit: "C2", Bars: []string{
					"X . . . . . . x X . x . . . . .",
					"X . . . . . . x X . . . . . x .",
				}},
				{Name: "four", Step: 0.25, Swing: 0.06, Humanize: 0.003, Seed: 8, Hit: "C2", Bars: []string{
					"x . . . x . . . x . . . x . . .",
				}},
			},
			Play: []synth.Play{
				{Clip: "groove", Bar: 9, Times: 8},
				{Clip: "four", Bar: 29, Times: 4},
				{Clip: "groove", Bar: 33, Times: 4},
			},
		},
		{
			Name: "snare", Patch: daybreakSnare(), Pan: 0.05, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -12}, {Bus: "haze", Level: -18}},
			Clips: []synth.Clip{{Name: "backbeat", Step: 0.25, Swing: 0.06, Humanize: 0.003, Seed: 8, Hit: "C3", Bars: []string{
				". . . . X . . o . . . . X . . .",
				". . . . X . . o . . o . X . . o",
			}}},
			Play: []synth.Play{{Clip: "backbeat", Bar: 9, Times: 8}, {Clip: "backbeat", Bar: 33, Times: 4}},
		},
		{
			// Off the drum bus, so its compressor can't flatten the crescendo.
			Name: "roll", Patch: daybreakSnare(), Pan: 0.05,
			Sends: []synth.Send{{Bus: "room", Level: -12}, {Bus: "haze", Level: -14}},
			Clips: []synth.Clip{{Name: "build", Step: 0.25, Hit: "C3", Bars: []string{
				"o . o . o . o . x x x x X X X X",
			}}},
			Play: []synth.Play{{Clip: "build", Bar: 32}},
		},
		{
			Name: "hihat", Patch: daybreakHat(), Level: -2, Pan: 0.3, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -20}},
			Clips: []synth.Clip{
				{Name: "sixteenths", Step: 0.25, Swing: 0.06, Humanize: 0.003, Seed: 8, Hit: "C5", Bars: []string{
					"x o x o x o x o x o x o x o . o",
				}},
				{Name: "eighths", Step: 0.25, Swing: 0.06, Humanize: 0.003, Seed: 8, Hit: "C5", Bars: []string{
					"x . x . x . x . x . x . x . . .",
				}},
			},
			Play: []synth.Play{
				{Clip: "eighths", Bar: 9, Times: 8},
				{Clip: "sixteenths", Bar: 17, Times: 8},
				{Clip: "eighths", Bar: 29, Times: 4},
				{Clip: "sixteenths", Bar: 33, Times: 8},
			},
		},
		{
			Name: "open hat", Patch: patches.OpenHat(), Level: -2.5, Pan: 0.3, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -18}},
			Clips: []synth.Clip{{Name: "lift", Step: 0.25, Swing: 0.06, Hit: "C5", Bars: []string{
				". . . . . . . . . . . . . . x .",
			}}},
			Play: []synth.Play{{Clip: "lift", Bar: 9, Times: 16}, {Clip: "lift", Bar: 33, Times: 8}},
		},
		{
			Name: "shaker", Patch: patches.Shaker(), Level: -2, Pan: -0.35,
			Sends: []synth.Send{{Bus: "room", Level: -16}, {Bus: "haze", Level: -18}},
			Clips: []synth.Clip{{Name: "shake", Step: 0.25, Swing: 0.06, Humanize: 0.004, Seed: 9, Hit: "C4", Bars: []string{
				"o o x o o o x o o o x o o o x o",
			}}},
			Play: []synth.Play{{Clip: "shake", Bar: 25, Times: 16}},
		},
		{
			Name: "pad", Patch: patches.TapePad(), Level: -6.5,
			Inserts: []synth.Effect{
				synth.Chorus{Rate: 0.3, Depth: 0.005, Mix: 0.5},
				synth.Sidechain{Key: "kick", Depth: 0.3, Release: 0.35},
			},
			Sends: []synth.Send{{Bus: "haze", Level: -8}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Gate: 0.98, Vel: 0.7, Bars: []string{
					"F#3,A3,C#4,E4 - - - - - - -",
					"E3,A3,B3,C#4 - - - - - - -",
					"F#3,A3,C#4,D4 - - - - - - -",
					"F#3,A3,B3,C#4 - - - - - - -",
				}},
				{Name: "B", Step: 0.5, Gate: 0.98, Vel: 0.7, Bars: []string{
					"G3,B3,D4,F#4 - - - - - - -",
					"F#3,A3,B3,D4 - - - - - - -",
					"F#3,A3,D4,E4 - - - - - - -",
					"E3,A3,D4,E4 - - - E3,A3,C#4,E4 - - -",
				}},
			},
			Play: []synth.Play{
				{Clip: "A", Bar: 1, Times: 6},
				{Clip: "B", Bar: 25, Times: 2},
				{Clip: "A", Bar: 33, Times: 2},
			},
		},
		{
			Name: "plucks", Patch: patches.GlassPluck(), Level: -6.5, Pan: -0.25,
			Inserts: []synth.Effect{
				synth.AutoFilter{Resonance: 0.2, Points: []synth.Point{{At: 0, Value: 400}, {At: 32, Value: 7000}}},
				synth.Sidechain{Key: "kick", Depth: 0.2, Release: 0.3},
			},
			Sends: []synth.Send{{Bus: "echo", Level: -10}, {Bus: "haze", Level: -14}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.25, Vel: 0.65, Swing: 0.06, Humanize: 0.003, Seed: 2, Bars: []string{
					"D4 C#5 E5 A4 E5 F#5 C#5 E5 D4 C#5 E5 A4 E5 F#5 C#5 A4",
					"C#4 B4 C#5 A4 C#5 E5 B4 C#5 C#4 B4 C#5 A4 C#5 E5 B4 A4",
					"B3 A4 C#5 F#4 C#5 D5 A4 C#5 B3 A4 C#5 F#4 C#5 D5 A4 F#4",
					"G3 F#4 A4 D4 A4 C#5 F#4 A4 G3 F#4 A4 D4 A4 C#5 F#4 D4",
				}},
				{Name: "B", Step: 0.25, Vel: 0.65, Swing: 0.06, Humanize: 0.003, Seed: 2, Bars: []string{
					"E4 D5 F#5 B4 F#5 G5 D5 F#5 E4 D5 F#5 B4 F#5 G5 D5 B4",
					"G3 F#4 A4 D4 A4 B4 F#4 A4 G3 F#4 A4 D4 A4 B4 F#4 D4",
					"D4 D5 E5 A4 E5 F#5 D5 E5 D4 D5 E5 A4 E5 F#5 D5 A4",
					"A3 D5 E5 E4 E5 A5 D5 E5 A3 C#5 E5 E4 E5 A5 C#5 E4",
				}},
			},
			Play: []synth.Play{
				{Clip: "A", Bar: 1, Times: 6},
				{Clip: "B", Bar: 25, Times: 2},
				{Clip: "A", Bar: 33, Times: 2},
			},
		},
		{
			Name: "bass", Patch: patches.PickBass(), Level: -8,
			Inserts: []synth.Effect{
				synth.Compressor{Threshold: -20, Ratio: 4, Attack: 0.01, Makeup: 4},
				synth.Sidechain{Key: "kick", Depth: 0.25, Release: 0.25},
			},
			Sends: []synth.Send{{Bus: "room", Level: -22}},
			Clips: []synth.Clip{
				{Name: "pickup", Step: 0.5, Gate: 0.85, Humanize: 0.006, Seed: 3, Bars: []string{
					". . . . . . A1 C#2",
				}},
				{Name: "A", Step: 0.5, Gate: 0.85, Humanize: 0.006, Seed: 3, Bars: []string{
					"D2 - D2 A2 . D3 C#3 A2",
					"C#2 - C#2 E2 . A2 B2 C#3",
					"B1 - B1 F#2 . B2 A2 F#2",
					"G1 - G1 D2 . G2 F#2 E2",
				}},
				{Name: "B long", Step: 0.5, Gate: 0.9, Humanize: 0.006, Seed: 3, Bars: []string{
					"E2 - - - - - B1 -",
					"G1 - - - - - D2 -",
					"F#1 - - - - - A1 -",
					"A1 - - - - - E2 -",
				}},
				{Name: "B", Step: 0.5, Gate: 0.85, Humanize: 0.006, Seed: 3, Bars: []string{
					"E2 - E2 B2 . E3 D3 B2",
					"G1 - G1 D2 . G2 F#2 D2",
					"F#1 - F#1 A1 . D2 E2 F#2",
					"A1 - A1 E2 . A2 C#3 A2",
				}},
			},
			Play: []synth.Play{
				{Clip: "pickup", Bar: 8},
				{Clip: "A", Bar: 9, Times: 4},
				{Clip: "B long", Bar: 25},
				{Clip: "B", Bar: 29},
				{Clip: "A", Bar: 33, Times: 2},
			},
		},
		{
			Name: "lead", Patch: patches.GlideLead(), Level: 4, Pan: 0.1,
			Inserts: []synth.Effect{synth.Chorus{Rate: 0.6, Mix: 0.4}},
			Sends:   []synth.Send{{Bus: "echo", Level: -7}, {Bus: "haze", Level: -10}},
			Clips: []synth.Clip{{Name: "tune", Step: 0.5, Gate: 0.95, Glide: 0.08, Humanize: 0.006, Seed: 4, Bars: []string{
				"F#5 - - E5 - - A4 -",
				"E5 - - - C#5 - B4 -",
				"D5 - - C#5 - - F#4 -",
				"A4 - - - - - . .",
				"F#5 - - E5 - - A5 -",
				"B5 - - A5 - - E5 -",
				"F#5 - - E5 - - D5 C#5",
				"D5 - - - - - . .",
			}}},
			Play: []synth.Play{{Clip: "tune", Bar: 17}, {Clip: "tune", Bar: 33}},
		},
		{
			Name: "vibes", Patch: patches.Vibraphone(), Level: -4.5, Pan: 0.35,
			Sends: []synth.Send{{Bus: "haze", Level: -9}, {Bus: "echo", Level: -12}},
			Clips: []synth.Clip{
				{Name: "tune", Step: 0.5, Vel: 0.7, Humanize: 0.006, Seed: 7, Bars: []string{
					"B4 - - D5 - - F#5 -",
					"D5 - - - B4 - A4 -",
					"A4 - - F#4 - - D5 -",
					"E5 - - - C#5 - - -",
				}},
				{Name: "answer", Step: 0.5, Vel: 0.55, Humanize: 0.006, Seed: 7, Bars: []string{
					". . . . A5 . F#5 E5",
					". . . . . . . .",
					". . . . F#5 . D5 C#5",
					"D5 . . . . . . .",
				}},
			},
			Play: []synth.Play{{Clip: "tune", Bar: 25, Times: 2}, {Clip: "answer", Bar: 33, Times: 2}},
		},
	},
	Buses: []synth.Bus{
		{Name: "drums", Effects: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowShelf, Freq: 80, Gain: 2},
				{Kind: synth.Peak, Freq: 400, Gain: -3, Q: 0.9},
				{Kind: synth.HighCut, Freq: 10000},
			}},
			synth.Drive{Amount: 1.2, Mix: 0.25},
			synth.Compressor{Threshold: -18, Ratio: 4, Attack: 0.01, Release: 0.12, Makeup: 4},
		}},
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.3, Damp: 0.55, PreDelay: 0.01}}},
		{Name: "haze", Effects: []synth.Effect{synth.Reverb{Size: 0.92, Damp: 0.4, PreDelay: 0.05}}},
		{Name: "echo", Effects: []synth.Effect{
			synth.Delay{Time: 0.75, Feedback: 0.5, Damp: 2200, PingPong: true},
			synth.Reverb{Size: 0.7, Damp: 0.5, Mix: 0.35},
		}},
	},
	Master: []synth.Effect{
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 30},
			{Kind: synth.LowShelf, Freq: 140, Gain: 1},
			{Kind: synth.HighShelf, Freq: 9000, Gain: -2},
		}},
		synth.Drive{Amount: 1, Tone: 11000, Mix: 0.25},
		synth.Compressor{Threshold: -16, Ratio: 2, Attack: 0.02, Release: 0.25},
		synth.Limiter{},
	},
}

func Daybreak() synth.Song { return daybreak.Song() }

func daybreakKick() synth.Patch {
	p := patches.Kick()
	p.Oscs[0].Fixed = 54
	p.Sweep.Semis = 24
	p.Amp.Decay = 0.3
	return p
}

func daybreakSnare() synth.Patch {
	p := patches.Snare()
	p.Filter.Cutoff = 1900
	p.Amp.Decay = 0.22
	return p
}

func daybreakHat() synth.Patch {
	p := patches.HiHat()
	p.Filter.Cutoff = 9000
	p.Amp.Decay = 0.035
	return p
}
