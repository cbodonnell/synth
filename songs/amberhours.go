package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Amber Hours": a harpsichord waltz in the spirit of Golden Brown, in a church with
// Interstellar's organ. D minor at 168.
//
// The phrase is thirteen beats, three bars of three and one of four. A is the hook,
// Dm Bb C A, its four-beat bar running the bass down to C# to fall back into the tonic.
// B moves to the relative major, Bb F/A Gm A, with a recorder singing over broken
// chords. Under it all a clock ticks each beat, and a pipe organ holds pedal tones,
// repeats a two-note ostinato against the moving harmony, and swells open through B.
//
// Two passes of A make the intro, harpsichord and clock and then the organ's pedal;
// the loop is A A B B A A, adding bass and a brushed kit, then the organ, the recorder
// taking up the hook on the way home. Each token is an eighth.
var amberHours = synth.Arrangement{
	ID: "amberhours", Name: "Amber Hours",
	Tempo: 168, Meter: []int{3, 3, 3, 4}, Key: "D minor",
	Sections: []synth.Section{
		{Name: "Intro", Bars: 4},
		{Name: "Pedal", Bars: 4},
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
			Name: "left hand", Patch: patches.Harpsichord(), Pan: -0.3, Out: "harpsichords",
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{
				{
					Name:     "A",
					Step:     0.5,
					Gate:     0.95,
					Humanize: 0.005,
					Seed:     1,
					Bars: []string{
						"D2 - F3,A3 . F3,A3 .",
						"Bb1 - F3,Bb3 . F3,Bb3 .",
						"C2 - E3,G3 . E3,G3 .",
						"A1 - E3,G3 . G2 F2 E2 C#2",
					},
				},
				{
					Name:     "B",
					Step:     0.5,
					Gate:     0.95,
					Humanize: 0.005,
					Seed:     1,
					Bars: []string{
						"Bb1 - F3,Bb3 . F3,Bb3 .",
						"A1 - F3,A3 . F3,A3 .",
						"G1 - G3,Bb3 . G3,Bb3 .",
						"A1 - E3,G3 . G2 F2 E2 C#2",
					},
				},
			},
			Play: []synth.Play{{Clip: "A", Bar: 1, Times: 4}, {Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			Name: "right hand", Patch: patches.Harpsichord(), Pan: -0.3, Out: "harpsichords",
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{
				{
					Name:     "A",
					Step:     0.5,
					Gate:     0.8,
					Strum:    0.005,
					Humanize: 0.005,
					Seed:     2,
					Bars: []string{
						". A4 D5 - E5 F5",
						"- D5 . Bb4 C5 D5",
						"E5 - C5 . G4 Bb4",
						"A4 - . E4 A4 C#5 E5 C#5",
					},
				},
				{
					Name:     "B",
					Step:     0.5,
					Gate:     0.8,
					Strum:    0.005,
					Humanize: 0.005,
					Seed:     2,
					Bars: []string{
						"D4 F4 Bb4 F4 D4 F4",
						"C4 F4 A4 F4 C4 F4",
						"D4 G4 Bb4 G4 D4 G4",
						"C#4 E4 A4 E4 C#4 E4 G4 E4",
					},
				},
			},
			Play: []synth.Play{{Clip: "A", Bar: 1, Times: 4}, {Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			Name: "double left", Patch: patches.Harpsichord(), Level: -1.9, Pan: 0.4, Out: "harpsichords",
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{
				{
					Name:     "A",
					Step:     0.5,
					Gate:     0.95,
					Humanize: 0.005,
					Seed:     3,
					Bars: []string{
						"D2 - F3,A3 . F3,A3 .",
						"Bb1 - F3,Bb3 . F3,Bb3 .",
						"C2 - E3,G3 . E3,G3 .",
						"A1 - E3,G3 . G2 F2 E2 C#2",
					},
				},
				{
					Name:     "B",
					Step:     0.5,
					Gate:     0.95,
					Humanize: 0.005,
					Seed:     3,
					Bars: []string{
						"Bb1 - F3,Bb3 . F3,Bb3 .",
						"A1 - F3,A3 . F3,A3 .",
						"G1 - G3,Bb3 . G3,Bb3 .",
						"A1 - E3,G3 . G2 F2 E2 C#2",
					},
				},
			},
			Play: []synth.Play{{Clip: "A", Bar: 5, Times: 3}, {Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			Name: "double right", Patch: patches.Harpsichord(), Level: -1.9, Pan: 0.4, Out: "harpsichords",
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{
				{
					Name:     "A",
					Step:     0.5,
					Gate:     0.8,
					Strum:    0.005,
					Humanize: 0.005,
					Seed:     4,
					Bars: []string{
						". A4 D5 - E5 F5",
						"- D5 . Bb4 C5 D5",
						"E5 - C5 . G4 Bb4",
						"A4 - . E4 A4 C#5 E5 C#5",
					},
				},
				{
					Name:     "B",
					Step:     0.5,
					Gate:     0.8,
					Strum:    0.005,
					Humanize: 0.005,
					Seed:     4,
					Bars: []string{
						"D4 F4 Bb4 F4 D4 F4",
						"C4 F4 A4 F4 C4 F4",
						"D4 G4 Bb4 G4 D4 G4",
						"C#4 E4 A4 E4 C#4 E4 G4 E4",
					},
				},
			},
			Play: []synth.Play{{Clip: "A", Bar: 5, Times: 3}, {Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			Name: "recorder", Patch: patches.Recorder(), Level: -1.9, Pan: 0.1,
			Sends: []synth.Send{{Bus: "echo", Level: -14}, {Bus: "room", Level: -10.5}},
			Clips: []synth.Clip{
				{
					Name:     "B",
					Step:     0.5,
					Gate:     0.92,
					Humanize: 0.006,
					Seed:     11,
					Bars: []string{
						"D5 - - C5 Bb4 C5",
						"A4 - - - F4 A4",
						"Bb4 - D5 - G5 F5",
						"E5 - - C#5 - E5 D5 C#5",
					},
				},
				{
					Name:     "hook",
					Step:     0.5,
					Gate:     0.92,
					Humanize: 0.006,
					Seed:     11,
					Bars: []string{
						". A4 D5 - E5 F5",
						"- D5 . Bb4 C5 D5",
						"E5 - C5 . G4 Bb4",
						"A4 - . E4 A4 C#5 E5 C#5",
					},
				},
			},
			Play: []synth.Play{{Clip: "B", Bar: 17, Times: 2}, {Clip: "hook", Bar: 25}},
		},
		{
			Name: "bass", Patch: patches.HofnerBass(), Level: -7.5,
			Inserts: []synth.Effect{synth.Compressor{Threshold: -20, Ratio: 4, Attack: 0.01, Makeup: 4}},
			Sends:   []synth.Send{{Bus: "room", Level: -28}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Gate: 0.9, Humanize: 0.004, Seed: 13, Bars: []string{
					"D2 - - A1 - -",
					"Bb1 - - F2 - -",
					"C2 - - G1 - -",
					"A1 - - E2 G1 F1 E1 C#2",
				}},
				{Name: "B", Step: 0.5, Gate: 0.9, Humanize: 0.004, Seed: 13, Bars: []string{
					"Bb1 - - F2 - -",
					"A1 - - F2 - -",
					"G1 - - D2 - -",
					"A1 - - E2 G1 F1 E1 C#2",
				}},
			},
			Play: []synth.Play{{Clip: "A", Bar: 9, Times: 2}, {Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			// The swell box opens across B and stays open for the return of the hook.
			Name: "organ", Patch: patches.PipeOrgan(), Level: -4.4,
			Inserts: []synth.Effect{
				synth.AutoFilter{Points: []synth.Point{{At: 52, Value: 500}, {At: 78, Value: 7000}}},
			},
			Sends: []synth.Send{{Bus: "nave", Level: -4.4}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Gate: 0.98, Bars: []string{
					"D3,F3,A3,D4 - - - - -",
					"D3,F3,Bb3,D4 - - - - -",
					"C3,E3,G3,C4 - - - - -",
					"C#3,E3,A3,E4 - - - - - - -",
				}},
				{Name: "B", Step: 0.5, Gate: 0.98, Bars: []string{
					"D3,F3,Bb3,D4 - - - - -",
					"C3,F3,A3,C4 - - - - -",
					"D3,G3,Bb3,D4 - - - - -",
					"C#3,E3,A3,E4 - - - - - - -",
				}},
			},
			Play: []synth.Play{{Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			Name: "pedal", Patch: patches.PipeOrgan(), Level: -1.4,
			Sends: []synth.Send{{Bus: "nave", Level: -6}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Gate: 0.98, Bars: []string{
					"D2 - - - - -",
					"Bb1 - - - - -",
					"C2 - - - - -",
					"A1 - - - - - - -",
				}},
				{Name: "B", Step: 0.5, Gate: 0.98, Bars: []string{
					"Bb1 - - - - -",
					"A1 - - - - -",
					"G1 - - - - -",
					"A1 - - - - - - -",
				}},
			},
			Play: []synth.Play{{Clip: "A", Bar: 5}, {Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			// Two notes held against the harmony as it moves: A and E are the ninth over
			// Dm, the major seventh and sharp eleventh over Bb.
			Name: "ostinato", Patch: patches.PipeOrgan(), Level: -2.5, Pan: -0.2,
			Inserts: []synth.Effect{
				synth.AutoFilter{Points: []synth.Point{{At: 52, Value: 500}, {At: 78, Value: 7000}}},
			},
			Sends: []synth.Send{{Bus: "nave", Level: -3.1}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Gate: 0.85, Vel: 0.7, Bars: []string{
					"A4 E5 A4 E5 A4 E5",
					"A4 E5 A4 E5 A4 E5",
					"A4 E5 A4 E5 A4 E5",
					"A4 E5 A4 E5 A4 E5 A4 E5",
				}},
				{Name: "B", Step: 0.5, Gate: 0.85, Vel: 0.7, Bars: []string{
					"D5 A5 D5 A5 D5 A5",
					"D5 A5 D5 A5 D5 A5",
					"D5 A5 D5 A5 D5 A5",
					"C#5 A5 C#5 A5 C#5 A5 C#5 A5",
				}},
			},
			Play: []synth.Play{{Clip: "B", Bar: 17, Times: 2}, {Clip: "A", Bar: 25, Times: 2}},
		},
		{
			Name: "clock", Patch: patches.Tick(), Level: 2.9, Pan: 0.5,
			Sends: []synth.Send{{Bus: "room", Level: -20}},
			Clips: []synth.Clip{{Name: "tick", Step: 0.5, Hit: "C5", Bars: []string{
				"X . x . x .",
				"X . x . x .",
				"X . x . x .",
				"X . x . x . x .",
			}}},
			Play: []synth.Play{{Clip: "tick", Bar: 1, Times: 8}},
		},
		{
			Name: "kick", Patch: patches.Kick(), Level: -12, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -16.5}},
			Clips: []synth.Clip{{Name: "beat", Step: 0.5, Humanize: 0.004, Seed: 17, Hit: "C2", Bars: []string{
				"x . . . . .",
				"x . . . . .",
				"x . . . . .",
				"x . . . x . . .",
			}}},
			Play: []synth.Play{{Clip: "beat", Bar: 9, Times: 6}},
		},
		{
			Name: "snare", Patch: brushedSnare(), Level: 0.8, Pan: 0.1, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -16.5}},
			Clips: []synth.Clip{
				{Name: "beat", Step: 0.5, Humanize: 0.004, Seed: 17, Hit: "D2", Bars: []string{
					". . o . x .",
					". . o . x .",
					". . o . x .",
					". . o . x . . o",
				}},
				{Name: "fill", Step: 0.5, Humanize: 0.004, Seed: 17, Hit: "D2", Bars: []string{
					". . o . x .",
					". . o . x .",
					". . o . x .",
					". . o . . . . .",
				}},
			},
			Play: []synth.Play{{Clip: "beat", Bar: 9, Times: 3}, {Clip: "fill", Bar: 21}, {Clip: "beat", Bar: 25}, {Clip: "fill", Bar: 29}},
		},
		{
			Name: "hats", Patch: patches.HiHat(), Level: -2.5, Pan: 0.3, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -16.5}},
			Clips: []synth.Clip{{Name: "beat", Step: 0.5, Humanize: 0.004, Seed: 17, Hit: "F#2", Bars: []string{
				"x o x o x o",
				"x o x o x o",
				"x o x o x o",
				"x o x o x o x o",
			}}},
			Play: []synth.Play{{Clip: "beat", Bar: 9, Times: 6}},
		},
		{
			Name: "toms", Patch: patches.Tom(), Level: -13.2, Pan: -0.15, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -16.5}},
			Clips: []synth.Clip{{Name: "fill", Step: 0.5, Humanize: 0.004, Seed: 17, Bars: []string{
				". . . . D3 A2 F2 D2!",
			}}},
			Play: []synth.Play{{Clip: "fill", Bar: 24}, {Clip: "fill", Bar: 32}},
		},
	},
	Buses: []synth.Bus{
		{Name: "harpsichords", Effects: harpsichordDesk()},
		{Name: "drums", Effects: []synth.Effect{
			synth.EQ{Bands: []synth.Band{{Kind: synth.HighCut, Freq: 8000}}},
			synth.Compressor{Threshold: -20, Ratio: 3, Attack: 0.006, Release: 0.12, Makeup: 3},
		}},
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.55, Damp: 0.5, PreDelay: 0.012, Width: 0.9}}},
		{Name: "nave", Effects: []synth.Effect{synth.Reverb{Size: 0.97, Damp: 0.3, PreDelay: 0.045}}},
		{Name: "echo", Effects: []synth.Effect{synth.Delay{Time: 1.5, Feedback: 0.3, Damp: 2500, PingPong: true}}},
	},
	Master: []synth.Effect{
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 28},
			{Kind: synth.HighShelf, Freq: 11000, Gain: -1.5},
		}},
		synth.Compressor{Threshold: -14, Ratio: 2, Attack: 0.015, Release: 0.2, Makeup: 1},
		synth.Limiter{},
	},
	Target: -14,
}

func AmberHours() synth.Song { return amberHours.Song() }

// A snare played with brushes: shorter, and darker.
func brushedSnare() synth.Patch {
	p := patches.Snare()
	p.Amp.Decay = 0.12
	p.Filter.Cutoff = 1800
	return p
}
