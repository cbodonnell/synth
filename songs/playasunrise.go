package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Playa Sunrise": Golden Brown's lilt carried out to the desert at dawn. E mixolydian
// at 148, swung a little, over a tanpura and a didgeridoo that never leave E.
//
// The phrase is thirteen beats, three bars of three and one of four. A is E D A, the
// handpan keeping Golden Brown's interlock of root, off-beat, and answer; its four-beat
// bar falls E D C#m B in parallel chords, a sitar singing tenths above the bass. B
// borrows C for C D E, the same fall with C natural, and a bansuri takes the tune.
// The didgeridoo breathes in threes against the waltz, 3+3+3+3+3+3+3+3+2 to the pass.
//
// The intro is wind, tanpura, and handpan, then the didgeridoo and a shaker; the loop
// is A A B B A A, adding djembes, bass, and sitar, then a pad that opens through B and
// wind chimes, the flute joining the sitar on the hook for the last pass.
// Each token is an eighth.
var playaSunrise = synth.Arrangement{
	ID: "playasunrise", Name: "Playa Sunrise",
	Tempo: 148, Meter: []int{3, 3, 3, 4}, Key: "E mixolydian",
	Sections: []synth.Section{
		{Name: "Dawn", Bars: 4},
		{Name: "Didgeridoo", Bars: 4},
		{Name: "A", Bars: 4},
		{Name: "A'", Bars: 4},
		{Name: "B", Bars: 4},
		{Name: "B'", Bars: 4},
		{Name: "A''", Bars: 4},
		{Name: "A'''", Bars: 4},
	},
	Loop: "A",
	Tracks: []synth.Track{
		{
			Name: "wind", Patch: patches.Wind(),
			Sends: []synth.Send{{Bus: "sky", Level: -10.5}},
			Clips: []synth.Clip{{Name: "gust", Step: 0.5, Bars: []string{
				"E3 - - - - -",
				"- - - - - -",
				"- - - - - -",
				"- - - - - - - -",
			}}},
			Play: []synth.Play{{Clip: "gust", Bar: 1, Times: 8}},
		},
		{
			// Pa and three Sa, the tanpura's cycle, laid on the bars so each one starts on the fifth.
			Name: "tanpura", Patch: patches.Tanpura(), Level: 1.9, Pan: -0.35,
			Sends: []synth.Send{{Bus: "sky", Level: -10.5}},
			Clips: []synth.Clip{{Name: "drone", Step: 0.5, Swing: 0.12, Humanize: 0.008, Seed: 3, Bars: []string{
				"B2 - E3 - E3 -",
				"B2 - E3 - E3 -",
				"B2 - E3 - E3 -",
				"B2 - E3 - E3 - E2 -",
			}}},
			Play: []synth.Play{{Clip: "drone", Bar: 1, Times: 8}},
		},
		{
			Name: "handpan", Patch: patches.Handpan(), Level: -6.9, Pan: 0.2,
			Sends: []synth.Send{{Bus: "sky", Level: -9.1}, {Bus: "dub", Level: -18.4}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Swing: 0.12, Humanize: 0.005, Seed: 1, Bars: []string{
					"E3 G#4 . B3,B4 . E4",
					"D3 F#4 . A3,A4 . D4",
					"A2 C#4 . E3,E4 . A3",
					"E3 G#4,B4 D3 F#4,A4 C#3 E4,G#4 B2 D#4,F#4",
				}},
				{Name: "B", Step: 0.5, Swing: 0.12, Humanize: 0.005, Seed: 1, Bars: []string{
					"C3 E4 . G3,G4 . C4",
					"D3 F#4 . A3,A4 . D4",
					"E3 G#4 . B3,B4 . E4",
					"E3 G#4,B4 D3 F#4,A4 C3 E4,G4 B2 D#4,F#4",
				}},
			},
			Play: []synth.Play{{Clip: "A", Bar: 1, Times: 4}, {Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			// Breathing in threes against the waltz, 3+3+3+3+3+3+3+3+2 to the pass.
			Name: "didgeridoo", Patch: patches.Didgeridoo(), Level: -1.9,
			Sends: []synth.Send{{Bus: "sky", Level: -16.5}},
			Clips: []synth.Clip{{Name: "drone", Step: 0.5, Gate: 0.92, Swing: 0.12, Humanize: 0.006, Seed: 13, Bars: []string{
				"E2! - - E2 - -",
				"E2! - - E2 - -",
				"E2! - - E2 - -",
				"E2! - - E2 - - G#3 -",
			}}},
			Play: []synth.Play{{Clip: "drone", Bar: 5, Times: 7}},
		},
		{
			Name: "bass", Patch: patches.HofnerBass(), Level: -7.5,
			Inserts: []synth.Effect{synth.Compressor{Threshold: -20, Ratio: 4, Attack: 0.01, Makeup: 4}},
			Sends:   []synth.Send{{Bus: "sky", Level: -28}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Gate: 0.9, Swing: 0.12, Humanize: 0.004, Seed: 5, Bars: []string{
					"E2 - - . B1 E2",
					"D2 - - . A1 D2",
					"A1 - - . E2 A1",
					"E2 - D2 - C#2 - B1 -",
				}},
				{Name: "B", Step: 0.5, Gate: 0.9, Swing: 0.12, Humanize: 0.004, Seed: 5, Bars: []string{
					"C2 - - . G1 C2",
					"D2 - - . A1 D2",
					"E2 - - . B1 E2",
					"E2 - D2 - C2 - B1 -",
				}},
			},
			Play: []synth.Play{{Clip: "A", Bar: 9, Times: 2}, {Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			Name: "sitar", Patch: patches.Sitar(), Level: 3.5, Pan: -0.15,
			Sends: []synth.Send{{Bus: "sky", Level: -10.5}, {Bus: "dub", Level: -12}},
			Clips: []synth.Clip{
				{
					Name:     "A",
					Step:     0.5,
					Gate:     0.95,
					Swing:    0.12,
					Glide:    0.12,
					Humanize: 0.006,
					Seed:     7,
					Bars: []string{
						". B4 E5 - F#5 G#5",
						"A5 - F#5 . E5 D5",
						"E5 - C#5 . A4 B4",
						"F#5/G#5 - F#5 - E5 - D#5 -",
					},
				},
			},
			Play: []synth.Play{{Clip: "A", Bar: 9, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			Name: "bansuri", Patch: bansuri(), Level: -3.7, Pan: 0.25,
			Sends: []synth.Send{{Bus: "sky", Level: -8}, {Bus: "dub", Level: -10.5}},
			Clips: []synth.Clip{
				{
					Name:     "A",
					Step:     0.5,
					Gate:     0.92,
					Swing:    0.12,
					Glide:    0.1,
					Humanize: 0.006,
					Seed:     9,
					Bars: []string{
						". B4 E5 - F#5 G#5",
						"A5 - F#5 . E5 D5",
						"E5 - C#5 . A4 B4",
						"F#5/G#5 - F#5 - E5 - D#5 -",
					},
				},
				{
					Name:     "B",
					Step:     0.5,
					Gate:     0.92,
					Swing:    0.12,
					Glide:    0.1,
					Humanize: 0.006,
					Seed:     9,
					Bars: []string{
						"G5 - - E5 - G5",
						"A5 - - F#5 - A5",
						"B5 - - - A5 G#5",
						"G#5 - F#5 - E5 - D#5 -",
					},
				},
			},
			Play: []synth.Play{{Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 29}},
		},
		{
			Name: "pad", Patch: patches.SawPad(),
			Inserts: []synth.Effect{
				// The swell opens across B and stays open for the return of the hook.
				synth.AutoFilter{Points: []synth.Point{{At: 52, Value: 500}, {At: 78, Value: 6000}}},
				synth.Phaser{Rate: 0.15, Depth: 2, Feedback: 0.5},
			},
			Sends: []synth.Send{{Bus: "sky", Level: -6}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Swing: 0.12, Bars: []string{
					"E3,B3,E4,G#4 - - - - -",
					"D3,A3,D4,F#4 - - - - -",
					"E3,A3,C#4,E4 - - - - -",
					"E3,B3,E4,G#4 - - - B2,F#3,B3,D#4 - - -",
				}},
				{Name: "B", Step: 0.5, Swing: 0.12, Bars: []string{
					"C3,G3,C4,E4 - - - - -",
					"D3,A3,D4,F#4 - - - - -",
					"E3,B3,E4,G#4 - - - - -",
					"E3,B3,E4,G#4 - - - B2,F#3,B3,D#4 - - -",
				}},
			},
			Play: []synth.Play{{Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			Name: "chimes", Patch: patches.Bell(), Level: -9.6, Pan: 0.5,
			Sends: []synth.Send{{Bus: "sky", Level: -4.4}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Swing: 0.12, Humanize: 0.03, Seed: 11, Bars: []string{
					". . . . . .",
					". . B6 . . .",
					". . . . . .",
					"E7 . . . . . C#7 .",
				}},
				{Name: "B", Step: 0.5, Swing: 0.12, Humanize: 0.03, Seed: 11, Bars: []string{
					". . . . G#6 .",
					". . . . . .",
					". B6 . . . .",
					". . E7 . . . . .",
				}},
			},
			Play: []synth.Play{{Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 29}},
		},
		{
			Name: "shaker", Patch: patches.Shaker(), Level: 0.8, Pan: 0.35, Out: "drums",
			Sends: []synth.Send{{Bus: "sky", Level: -21.9}},
			Clips: []synth.Clip{{Name: "beat", Step: 0.5, Swing: 0.12, Humanize: 0.004, Seed: 17, Hit: "C5", Bars: []string{
				"x o x o x o",
				"x o x o x o",
				"x o x o x o",
				"x o x o x o x o",
			}}},
			Play: []synth.Play{{Clip: "beat", Bar: 5, Times: 7}},
		},
		{
			Name: "djembe", Patch: patches.HandDrum(), Level: -3.1, Pan: -0.1, Out: "drums",
			Sends: []synth.Send{{Bus: "sky", Level: -21.9}},
			Clips: []synth.Clip{
				{Name: "beat", Step: 0.5, Swing: 0.12, Humanize: 0.004, Seed: 17, Bars: []string{
					"D2! . D4 D4? . D4",
					"D2! . D4 D4? . D4",
					"D2! . D4 D4? . D4",
					"D2! . D4 D2 . D4 D4? D4",
				}},
				{Name: "roll", Step: 0.5, Swing: 0.12, Humanize: 0.004, Seed: 17, Bars: []string{
					"D2! . D4 D4? . D4",
					"D2! . D4 D4? . D4",
					"D2! . D4 D4? . D4",
					"D2! D4*2 D4 F3 D3*3 D3 A2? D2!",
				}},
			},
			Play: []synth.Play{{Clip: "beat", Bar: 9, Times: 3}, {Clip: "roll", Bar: 21}, {Clip: "beat", Bar: 25}, {Clip: "roll", Bar: 29}},
		},
		{
			Name: "slap", Patch: patches.Slap(), Level: 6.8, Pan: -0.1, Out: "drums",
			Sends: []synth.Send{{Bus: "sky", Level: -21.9}},
			Clips: []synth.Clip{
				{Name: "beat", Step: 0.5, Swing: 0.12, Humanize: 0.004, Seed: 17, Hit: "C4", Bars: []string{
					". . . . x .",
					". . . . x .",
					". . . . x .",
					". . . . x . . .",
				}},
				{Name: "roll", Step: 0.5, Swing: 0.12, Humanize: 0.004, Seed: 17, Hit: "C4", Bars: []string{
					". . . . x .",
					". . . . x .",
					". . . . x .",
				}},
			},
			Play: []synth.Play{{Clip: "beat", Bar: 9, Times: 3}, {Clip: "roll", Bar: 21}, {Clip: "beat", Bar: 25}, {Clip: "roll", Bar: 29}},
		},
	},
	Buses: []synth.Bus{
		{Name: "drums", Effects: []synth.Effect{
			synth.Compressor{Threshold: -18, Ratio: 2.5, Attack: 0.008, Release: 0.12, Makeup: 2},
		}},
		{Name: "sky", Effects: []synth.Effect{synth.Reverb{Size: 0.85, Damp: 0.4, PreDelay: 0.03}}},
		{Name: "dub", Effects: []synth.Effect{synth.Delay{Time: 1.5, Feedback: 0.4, Damp: 2200, PingPong: true}}},
	},
	Master: []synth.Effect{
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 30},
			{Kind: synth.HighShelf, Freq: 11000, Gain: -1.5},
		}},
		synth.Compressor{Threshold: -14, Ratio: 2, Attack: 0.015, Release: 0.2, Makeup: -0.3},
		synth.Limiter{},
	},
	Target: -14,
}

func PlayaSunrise() synth.Song { return playaSunrise.Song() }

// A recorder breathier, darker, and sung with a slow vibrato.
func bansuri() synth.Patch {
	p := patches.Recorder()
	p.Name = "bansuri"
	p.Oscs[2].Level = 0.07
	p.Amp.Attack = 0.05
	p.Filter.Cutoff = 2800
	p.LFOs = []synth.LFO{{Target: synth.Vibrato, Rate: 5.2, Depth: 22, Delay: 0.3}}
	return p
}
