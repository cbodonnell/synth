package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Golden Hour": the main stage as the sun goes down, the field still talking. Downtempo
// in A major at 110, a third sibling to Long Light and Daybreak with none of their
// instruments: a reedy FM lead singing a hook on a 3-3-2 rhythm, a kalimba ticking the
// same rhythm on the notes every chord shares, a clean guitar picking eighths into
// dotted-eighth echoes, a pad that blooms on each chord, a round synth bass, and a
// tight live kit.
//
// Eight bars of intro (the crowd, the pad, the kalimba opening up, then the guitar),
// then a 32-bar loop: A is the hook over Amaj9 E/G# F#m9 Dmaj9; A' adds the kalimba,
// sixteenths and tambourine; B lets the drums go for Dmaj9 E C#m7 F#m9 Bm9 C#m7 Dmaj9
// Esus4-E under a slower tune and the crowd, and builds back on a roll; A” is
// everything, the kalimba doubling the hook.
var goldenHour = synth.Arrangement{
	ID: "goldenhour", Name: "Golden Hour",
	Tempo: 110, Key: "A major",
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
			Name: "crowd", Patch: patches.Crowd(), Level: -3, Pan: 0.1,
			Sends: []synth.Send{{Bus: "dusk", Level: -10}},
			Clips: []synth.Clip{{Name: "field", Step: 1, Bars: []string{
				"A4 - - -",
				"- - - -",
				"- - - -",
				"- - - -",
			}}},
			Play: []synth.Play{{Clip: "field", Bar: 1, Times: 2}, {Clip: "field", Bar: 25}},
		},
		{
			Name: "kick", Patch: goldenHourKick(), Level: -9, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -24}},
			Clips: []synth.Clip{
				{Name: "groove", Step: 0.25, Swing: 0.05, Humanize: 0.003, Seed: 8, Hit: "C2", Bars: []string{
					"X . . . . . . . . . X . . . . .",
					"X . . x . . . . . . X . . . x .",
				}},
				{Name: "four", Step: 0.25, Swing: 0.05, Humanize: 0.003, Seed: 8, Hit: "C2", Bars: []string{
					"x . . . x . . . x . . . x . . .",
				}},
			},
			Play: []synth.Play{
				{Clip: "groove", Bar: 9, Times: 8},
				{Clip: "four", Bar: 31, Times: 2},
				{Clip: "groove", Bar: 33, Times: 4},
			},
		},
		{
			Name: "snare", Patch: patches.TightSnare(), Level: -8.5, Pan: 0.05, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -12}, {Bus: "dusk", Level: -20}},
			Clips: []synth.Clip{{Name: "backbeat", Step: 0.25, Swing: 0.05, Humanize: 0.003, Seed: 8, Hit: "C3", Bars: []string{
				". . . . X . . o . . . o X . . .",
				". . . . X . . o . . . . X . o o",
			}}},
			Play: []synth.Play{{Clip: "backbeat", Bar: 9, Times: 8}, {Clip: "backbeat", Bar: 33, Times: 4}},
		},
		{
			// Off the drum bus, so its compressor can't flatten the crescendo.
			Name: "roll", Patch: patches.TightSnare(), Level: -4, Pan: 0.05,
			Sends: []synth.Send{{Bus: "room", Level: -12}, {Bus: "dusk", Level: -16}},
			Clips: []synth.Clip{{Name: "build", Step: 0.25, Hit: "C3", Bars: []string{
				"o . o . o . o . o o x x x x X X",
			}}},
			Play: []synth.Play{{Clip: "build", Bar: 32}},
		},
		{
			Name: "hihat", Patch: goldenHourHat(), Level: -10, Pan: 0.3, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -20}},
			Clips: []synth.Clip{
				{Name: "eighths", Step: 0.25, Swing: 0.05, Humanize: 0.003, Seed: 8, Hit: "C5", Bars: []string{
					"x . o . x . o . x . o . x . o .",
				}},
				{Name: "sixteenths", Step: 0.25, Swing: 0.05, Humanize: 0.003, Seed: 8, Hit: "C5", Bars: []string{
					"x o o o x o o o x o o o x o x o",
				}},
			},
			Play: []synth.Play{
				{Clip: "eighths", Bar: 9, Times: 8},
				{Clip: "sixteenths", Bar: 17, Times: 8},
				{Clip: "eighths", Bar: 29, Times: 2},
				{Clip: "sixteenths", Bar: 33, Times: 8},
			},
		},
		{
			Name: "tambourine", Patch: patches.Tambourine(), Level: -2, Pan: -0.35,
			Sends: []synth.Send{{Bus: "room", Level: -16}, {Bus: "dusk", Level: -20}},
			Clips: []synth.Clip{{Name: "offbeats", Step: 0.25, Swing: 0.05, Humanize: 0.004, Seed: 9, Hit: "C4", Bars: []string{
				". . x . . . x . . . x . . . x o",
			}}},
			Play: []synth.Play{{Clip: "offbeats", Bar: 17, Times: 8}, {Clip: "offbeats", Bar: 33, Times: 8}},
		},
		{
			Name: "riser", Patch: patches.Riser(8 * 60.0 / 110), Level: -8,
			Sends: []synth.Send{{Bus: "dusk", Level: -10}},
			Clips: []synth.Clip{{Name: "build", Step: 1, Bars: []string{
				"A3 - - -",
				"- - - -",
			}}},
			Play: []synth.Play{{Clip: "build", Bar: 31}},
		},
		{
			Name: "pad", Patch: patches.SwellPad(),
			Inserts: []synth.Effect{
				synth.Chorus{Rate: 0.4, Depth: 0.004, Mix: 0.5},
				synth.Sidechain{Key: "kick", Depth: 0.35, Release: 0.3},
			},
			Sends: []synth.Send{{Bus: "dusk", Level: -8}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Gate: 0.98, Vel: 0.7, Bars: []string{
					"C#4,E4,G#4,B4 - - - - - - -",
					"B3,E4,F#4,G#4 - - - - - - -",
					"A3,C#4,E4,G#4 - - - - - - -",
					"A3,C#4,E4,F#4 - - - - - - -",
				}},
				{Name: "B", Step: 0.5, Gate: 0.98, Vel: 0.7, Bars: []string{
					"A3,C#4,E4,F#4 - - - - - - -",
					"G#3,B3,E4,F#4 - - - - - - -",
					"B3,C#4,E4,G#4 - - - - - - -",
					"A3,C#4,E4,G#4 - - - - - - -",
					"A3,C#4,D4,F#4 - - - - - - -",
					"B3,C#4,E4,G#4 - - - - - - -",
					"A3,C#4,E4,F#4 - - - - - - -",
					"A3,B3,E4 - - - G#3,B3,E4 - - -",
				}},
			},
			Play: []synth.Play{
				{Clip: "A", Bar: 1, Times: 6},
				{Clip: "B", Bar: 25},
				{Clip: "A", Bar: 33, Times: 2},
			},
		},
		{
			Name: "kalimba", Patch: patches.Kalimba(), Level: -6, Pan: -0.3,
			Inserts: []synth.Effect{
				synth.AutoFilter{Resonance: 0.2, Points: []synth.Point{{At: 0, Value: 500}, {At: 32, Value: 9000}}},
			},
			Sends: []synth.Send{{Bus: "echo", Level: -10}, {Bus: "dusk", Level: -14}},
			Clips: []synth.Clip{
				{Name: "tick", Step: 0.25, Vel: 0.6, Swing: 0.05, Humanize: 0.004, Seed: 2, Bars: []string{
					"E5 . . C#5 . . B4 . E5 . . C#5 . . B4 E5",
				}},
				{Name: "hook", Step: 0.25, Vel: 0.6, Humanize: 0.004, Seed: 2, Bars: []string{
					"C#6 . . E6 . . A6 . . . G#6 . E6 . . .",
					"B5 . . . . . . . . . G#5 . A5 . B5 .",
					"C#6 . . E6 . . A6 . . . G#6 . A6 . . .",
					"F#6 . . . E6 . . . C#6 . . . . . . .",
					"C#6 . . E6 . . A6 . . . G#6 . E6 . . .",
					"B6 . . . A6 . . . G#6 . . . E6 . . .",
					"A6 . . G#6 . . E6 . . . C#6 . . . . .",
					"D6 . . C#6 . . B5 . . . . . . . . .",
				}},
			},
			Play: []synth.Play{
				{Clip: "tick", Bar: 1, Times: 8},
				{Clip: "tick", Bar: 17, Times: 8},
				{Clip: "hook", Bar: 33},
			},
		},
		{
			Name: "guitar", Patch: patches.CleanGuitar(), Level: 5, Pan: 0.3,
			Inserts: []synth.Effect{
				synth.Chorus{Rate: 0.9, Depth: 0.003, Mix: 0.45},
				synth.Sidechain{Key: "kick", Depth: 0.2, Release: 0.25},
			},
			Sends: []synth.Send{{Bus: "echo", Level: -8}, {Bus: "dusk", Level: -14}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Vel: 0.6, Humanize: 0.005, Seed: 5, Bars: []string{
					"A3 E4 G#4 B4 C#5 B4 G#4 E4",
					"G#3 E4 F#4 B4 E5 B4 F#4 E4",
					"F#3 C#4 E4 G#4 A4 G#4 E4 C#4",
					"D3 A3 C#4 E4 F#4 E4 C#4 A3",
				}},
				{Name: "B", Step: 0.5, Vel: 0.6, Humanize: 0.005, Seed: 5, Bars: []string{
					"D3 A3 C#4 E4 F#4 E4 C#4 A3",
					"E3 B3 E4 F#4 G#4 F#4 E4 B3",
					"C#3 G#3 B3 E4 G#4 E4 B3 G#3",
					"F#3 C#4 E4 G#4 A4 G#4 E4 C#4",
					"B2 F#3 A3 C#4 D4 C#4 A3 F#3",
					"C#3 G#3 B3 E4 G#4 E4 B3 G#3",
					"D3 A3 C#4 E4 F#4 E4 C#4 A3",
					"E3 B3 E4 A4 B4 A4 G#4 E4",
				}},
			},
			Play: []synth.Play{
				{Clip: "A", Bar: 5, Times: 5},
				{Clip: "B", Bar: 25},
				{Clip: "A", Bar: 33, Times: 2},
			},
		},
		{
			Name: "bass", Patch: patches.RoundBass(), Level: -9.5,
			Inserts: []synth.Effect{
				synth.Compressor{Threshold: -20, Ratio: 4, Attack: 0.01, Makeup: 4},
				synth.Sidechain{Key: "kick", Depth: 0.3, Release: 0.25},
			},
			Clips: []synth.Clip{
				{Name: "pickup", Step: 0.5, Gate: 0.85, Humanize: 0.006, Seed: 3, Bars: []string{
					". . . . . . E1 G#1",
				}},
				{Name: "A", Step: 0.5, Gate: 0.85, Glide: 0.04, Humanize: 0.006, Seed: 3, Bars: []string{
					"A1 - - E2 - A1 C#2 E2",
					"G#1 - - E2 - G#1 B1 E2",
					"F#1 - - C#2 - F#1 A1 C#2",
					"D2 - - A2 - D2 E2 F#2",
				}},
				{Name: "B long", Step: 0.5, Gate: 0.9, Humanize: 0.006, Seed: 3, Bars: []string{
					"D2 - - - - - A1 -",
					"E2 - - - - - B1 -",
					"C#2 - - - - - G#1 -",
					"F#1 - - - - - C#2 -",
				}},
				{Name: "B", Step: 0.5, Gate: 0.85, Glide: 0.04, Humanize: 0.006, Seed: 3, Bars: []string{
					"B1 - - F#2 - B1 D2 F#2",
					"C#2 - - G#2 - C#2 E2 G#2",
					"D2 - - A2 - D2 F#2 A2",
					"E2 - - B1 - E2 D2 B1",
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
			Name: "lead", Patch: patches.SunLead(), Level: -2, Pan: -0.05,
			Inserts: []synth.Effect{synth.Chorus{Rate: 0.5, Mix: 0.35}},
			Sends:   []synth.Send{{Bus: "echo", Level: -8}, {Bus: "dusk", Level: -11}},
			Clips: []synth.Clip{
				{Name: "hook", Step: 0.25, Gate: 0.95, Glide: 0.06, Humanize: 0.004, Seed: 4, Bars: []string{
					"C#5 - - E5 - - A5 - - - G#5 - E5 - - -",
					"B4 - - - - - - - . . G#4 - A4 - B4 -",
					"C#5 - - E5 - - A5 - - - G#5 - A5 - - -",
					"F#5 - - - E5 - - - C#5 - - - . . . .",
					"C#5 - - E5 - - A5 - - - G#5 - E5 - - -",
					"B5 - - - A5 - - - G#5 - - - E5 - - -",
					"A5 - - G#5 - - E5 - - - C#5 - - - - -",
					"D5 - - C#5 - - B4 - - - - - . . . .",
				}},
				{Name: "high", Step: 0.5, Gate: 0.95, Glide: 0.1, Vel: 0.8, Humanize: 0.006, Seed: 4, Bars: []string{
					"F#5 - - - E5 - C#5 -",
					"E5 - - - - - B4 -",
					"C#5 - - - E5 - G#5 -",
					"A5 - - - - - . .",
					"F#5 - - - E5 - D5 -",
					"C#5 - - - E5 - G#5 -",
					"A5 - - - G#5 - F#5 -",
					"E5 - - - - - - .",
				}},
			},
			Play: []synth.Play{
				{Clip: "hook", Bar: 9, Times: 2},
				{Clip: "high", Bar: 25},
				{Clip: "hook", Bar: 33},
			},
		},
	},
	Buses: []synth.Bus{
		{Name: "drums", Effects: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowShelf, Freq: 80, Gain: 2},
				{Kind: synth.Peak, Freq: 420, Gain: -3, Q: 0.9},
				{Kind: synth.HighShelf, Freq: 6000, Gain: 1.5},
			}},
			synth.Drive{Amount: 1.5, Mix: 0.3},
			synth.Compressor{Threshold: -18, Ratio: 4, Attack: 0.006, Release: 0.1, Makeup: 4},
		}},
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.3, Damp: 0.5, PreDelay: 0.01}}},
		{Name: "dusk", Effects: []synth.Effect{synth.Reverb{Size: 0.88, Damp: 0.45, PreDelay: 0.04}}},
		{Name: "echo", Effects: []synth.Effect{
			synth.Delay{Time: 0.75, Feedback: 0.42, Damp: 2600, PingPong: true},
			synth.Reverb{Size: 0.6, Damp: 0.5, Mix: 0.3},
		}},
	},
	Master: []synth.Effect{
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 30},
			{Kind: synth.LowShelf, Freq: 140, Gain: 1},
			{Kind: synth.HighShelf, Freq: 9000, Gain: -2},
		}},
		synth.Drive{Amount: 1.2, Tone: 11000, Mix: 0.3},
		synth.Compressor{Threshold: -16, Ratio: 2, Attack: 0.02, Release: 0.25},
		synth.Limiter{},
	},
}

func GoldenHour() synth.Song { return goldenHour.Song() }

func goldenHourKick() synth.Patch {
	p := patches.Kick()
	p.Oscs[0].Fixed = 52
	p.Amp.Decay = 0.26
	p.Oscs[1].Level = 0.45
	return p
}

func goldenHourHat() synth.Patch {
	p := patches.HiHat()
	p.Filter.Cutoff = 8500
	p.Amp.Decay = 0.03
	return p
}
