package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "The Woodlands": Firefly, a June weekend in the woods outside Dover, Delaware.
// Stomp-clap indie folk in G at 116, the main stage at golden hour, until the sun
// goes down and the night stage's synths take over among the trees.
//
// The intro is dusk: crickets, fireflies, a guitar strummed by a campfire, then a
// banjo and a stomp. The loop is A A B B C C B B. A is the verse, G Em C D, whistled
// over banjo rolls; B is the singalong, C G Em D, a gang of voices with mandolin and
// glockenspiel, ending each pass on a shout. C is nightfall, Em C G D: the band drops
// out for an arp and a pad, crickets and fireflies come back, and the second pass
// builds into the last chorus, which keeps the night's arp under the band.
// Each token is an eighth, except the banjo's and arp's sixteenths.
var woodlands = synth.Arrangement{
	ID: "woodlands", Name: "The Woodlands",
	Tempo: 116, Key: "G major",
	Sections: []synth.Section{
		{Name: "Dusk", Bars: 4},
		{Name: "Campfire", Bars: 4},
		{Name: "Verse", Bars: 4},
		{Name: "Verse'", Bars: 4},
		{Name: "Chorus", Bars: 4},
		{Name: "Chorus'", Bars: 4},
		{Name: "Nightfall", Bars: 4},
		{Name: "Build", Bars: 4},
		{Name: "Last Chorus", Bars: 4},
		{Name: "Last Chorus'", Bars: 4},
	},
	Loop: "Verse",
	Tracks: []synth.Track{
		{
			Name: "crickets", Patch: patches.Cricket(), Pan: -0.6,
			Sends: []synth.Send{{Bus: "woods", Level: -6}},
			Clips: []synth.Clip{{Name: "chirps", Step: 0.5, Humanize: 0.05, Seed: 17, Bars: []string{
				"C4 . . . C4 . C4 .",
				". . C4 . . . . .",
			}}},
			Play: []synth.Play{{Clip: "chirps", Bar: 1, Times: 4}, {Clip: "chirps", Bar: 25, Times: 4}},
		},
		{
			Name: "fireflies", Patch: patches.Firefly(), Level: -1.9,
			Sends: []synth.Send{{Bus: "woods", Level: -4.4}, {Bus: "echo", Level: -10.5}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Humanize: 0.12, Seed: 15, Bars: []string{
					". . E7 . . . . .",
					". . . . . B6 . .",
					". D7 . . . . . .",
					". . . . . . A6 .",
				}},
				{Name: "B", Step: 0.5, Humanize: 0.12, Seed: 15, Bars: []string{
					". . . . . G6 . .",
					". . D7 . . . . .",
					". . . . . . . E7",
					". . . B6 . . . .",
				}},
			},
			Play: []synth.Play{
				{Clip: "A", Bar: 1}, {Clip: "B", Bar: 5},
				{Clip: "A", Bar: 25}, {Clip: "B", Bar: 29}, {Clip: "A", Bar: 33}, {Clip: "B", Bar: 37},
			},
		},
		{
			Name: "acoustic guitar", Patch: patches.AcousticGuitar(), Pan: -0.35,
			Inserts: []synth.Effect{synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowCut, Freq: 90},
				{Kind: synth.Peak, Freq: 250, Gain: -2},
				{Kind: synth.HighShelf, Freq: 6000, Gain: 2},
			}}},
			Sends: []synth.Send{{Bus: "barn", Level: -16.5}},
			Clips: []synth.Clip{
				{Name: "dusk", Step: 0.5, Gate: 0.95, Strum: 0.012, Humanize: 0.006, Seed: 1, Bars: []string{
					"G2,B2,D3,G3,B3,G4! - - - - - - -",
					"E2,B2,E3,G3,B3,E4! - - - - - - -",
					"C3,E3,G3,C4,E4! - - - - - - -",
					"D3,A3,D4,F#4! - - - - - - -",
				}},
				{Name: "verse", Step: 0.5, Gate: 0.95, Strum: 0.012, Humanize: 0.006, Seed: 1, Bars: []string{
					"G2,B2,D3,G3,B3,G4! . G2,B2,D3,G3,B3,G4 G2,B2,D3,G3,B3,G4 . G2,B2,D3,G3,B3,G4 G2,B2,D3,G3,B3,G4 G2,B2,D3,G3,B3,G4",
					"E2,B2,E3,G3,B3,E4! . E2,B2,E3,G3,B3,E4 E2,B2,E3,G3,B3,E4 . E2,B2,E3,G3,B3,E4 E2,B2,E3,G3,B3,E4 E2,B2,E3,G3,B3,E4",
					"C3,E3,G3,C4,E4! . C3,E3,G3,C4,E4 C3,E3,G3,C4,E4 . C3,E3,G3,C4,E4 C3,E3,G3,C4,E4 C3,E3,G3,C4,E4",
					"D3,A3,D4,F#4! . D3,A3,D4,F#4 D3,A3,D4,F#4 . D3,A3,D4,F#4 D3,A3,D4,F#4 D3,A3,D4,F#4",
				}},
				{Name: "chorus", Step: 0.5, Gate: 0.95, Strum: 0.012, Humanize: 0.006, Seed: 1, Bars: []string{
					"C3,E3,G3,C4,E4! . C3,E3,G3,C4,E4 C3,E3,G3,C4,E4 . C3,E3,G3,C4,E4 C3,E3,G3,C4,E4 C3,E3,G3,C4,E4",
					"G2,B2,D3,G3,B3,G4! . G2,B2,D3,G3,B3,G4 G2,B2,D3,G3,B3,G4 . G2,B2,D3,G3,B3,G4 G2,B2,D3,G3,B3,G4 G2,B2,D3,G3,B3,G4",
					"E2,B2,E3,G3,B3,E4! . E2,B2,E3,G3,B3,E4 E2,B2,E3,G3,B3,E4 . E2,B2,E3,G3,B3,E4 E2,B2,E3,G3,B3,E4 E2,B2,E3,G3,B3,E4",
					"D3,A3,D4,F#4! . D3,A3,D4,F#4 D3,A3,D4,F#4 . D3,A3,D4,F#4 D3,A3,D4,F#4 D3,A3,D4,F#4",
				}},
				{Name: "nightfall", Step: 0.5, Gate: 0.95, Strum: 0.012, Humanize: 0.006, Seed: 1, Bars: []string{
					"E2,B2,E3,G3,B3,E4! - - - - - - -",
					"C3,E3,G3,C4,E4! - - - - - - -",
					"G2,B2,D3,G3,B3,G4! - - - - - - -",
					"D3,A3,D4,F#4! - - - - - - -",
				}},
				{Name: "build", Step: 0.5, Gate: 0.95, Strum: 0.012, Humanize: 0.006, Seed: 1, Bars: []string{
					"E2,B2,E3,G3,B3,E4! E2,B2,E3,G3,B3,E4 E2,B2,E3,G3,B3,E4 E2,B2,E3,G3,B3,E4 E2,B2,E3,G3,B3,E4! E2,B2,E3,G3,B3,E4 E2,B2,E3,G3,B3,E4 E2,B2,E3,G3,B3,E4",
					"C3,E3,G3,C4,E4! C3,E3,G3,C4,E4 C3,E3,G3,C4,E4 C3,E3,G3,C4,E4 C3,E3,G3,C4,E4! C3,E3,G3,C4,E4 C3,E3,G3,C4,E4 C3,E3,G3,C4,E4",
					"G2,B2,D3,G3,B3,G4! G2,B2,D3,G3,B3,G4 G2,B2,D3,G3,B3,G4 G2,B2,D3,G3,B3,G4 G2,B2,D3,G3,B3,G4! G2,B2,D3,G3,B3,G4 G2,B2,D3,G3,B3,G4 G2,B2,D3,G3,B3,G4",
					"D3,A3,D4,F#4! D3,A3,D4,F#4 D3,A3,D4,F#4 D3,A3,D4,F#4 D3,A3,D4,F#4! D3,A3,D4,F#4 D3,A3,D4,F#4 D3,A3,D4,F#4",
				}},
			},
			Play: []synth.Play{
				{Clip: "dusk", Bar: 1}, {Clip: "verse", Bar: 5, Times: 3}, {Clip: "chorus", Bar: 17, Times: 2},
				{Clip: "nightfall", Bar: 25}, {Clip: "build", Bar: 29}, {Clip: "chorus", Bar: 33, Times: 2},
			},
		},
		{
			// Forward rolls, thumb, index, middle; the second half of each bar puts the
			// thumb on the short fifth string's high G, the drone every banjo tune leans on.
			Name: "banjo", Patch: patches.Banjo(), Level: 4.6, Pan: 0.4,
			Sends: []synth.Send{{Bus: "barn", Level: -18.4}},
			Clips: []synth.Clip{
				{Name: "verse", Step: 0.25, Humanize: 0.004, Seed: 3, Bars: []string{
					"G3! B3 D4 G3 B3 D4 G3 D4 G4! B3 D4 G4 B3 D4 G4 D4",
					"G3! B3 E4 G3 B3 E4 G3 E4 G4! B3 E4 G4 B3 E4 G4 E4",
					"G3! C4 E4 G3 C4 E4 G3 E4 G4! C4 E4 G4 C4 E4 G4 E4",
					"F#3! A3 D4 F#3 A3 D4 F#3 D4 G4! A3 D4 G4 A3 D4 G4 D4",
				}},
				{Name: "chorus", Step: 0.25, Humanize: 0.004, Seed: 3, Bars: []string{
					"G3! C4 E4 G3 C4 E4 G3 E4 G4! C4 E4 G4 C4 E4 G4 E4",
					"G3! B3 D4 G3 B3 D4 G3 D4 G4! B3 D4 G4 B3 D4 G4 D4",
					"G3! B3 E4 G3 B3 E4 G3 E4 G4! B3 E4 G4 B3 E4 G4 E4",
					"F#3! A3 D4 F#3 A3 D4 F#3 D4 G4! A3 D4 G4 A3 D4 G4 D4",
				}},
			},
			Play: []synth.Play{{Clip: "verse", Bar: 5, Times: 3}, {Clip: "chorus", Bar: 17, Times: 2}, {Clip: "chorus", Bar: 33, Times: 2}},
		},
		{
			Name: "upright bass", Patch: patches.UprightBass(),
			Inserts: []synth.Effect{synth.Compressor{Threshold: -20, Ratio: 3, Attack: 0.01, Makeup: 3}},
			Sends:   []synth.Send{{Bus: "barn", Level: -28}},
			Clips: []synth.Clip{
				{Name: "verse", Step: 0.5, Gate: 0.85, Humanize: 0.004, Seed: 5, Bars: []string{
					"G1 - . G1 D2 - . D2",
					"E1 - . E1 B1 - . B1",
					"C2 - . C2 G1 - . G1",
					"D2 - . D2 A1 - . A1",
				}},
				{Name: "chorus", Step: 0.5, Gate: 0.85, Humanize: 0.004, Seed: 5, Bars: []string{
					"C2 - . C2 G1 - . G1",
					"G1 - . G1 D2 - . D2",
					"E1 - . E1 B1 - . B1",
					"D2 - . D2 A1 - . A1",
				}},
				{Name: "nightfall", Step: 0.5, Gate: 0.85, Humanize: 0.004, Seed: 5, Bars: []string{
					"E1 - - - - - - -",
					"C2 - - - - - - -",
					"G1 - - - - - - -",
					"D2 - - - - - - -",
				}},
				{Name: "build", Step: 0.5, Gate: 0.85, Humanize: 0.004, Seed: 5, Bars: []string{
					"E1 E1 E1 E1 E1 E1 E1 E1",
					"C2 C2 C2 C2 C2 C2 C2 C2",
					"G1 G1 G1 G1 G1 G1 G1 G1",
					"D2 D2 D2 D2 D2 D2 D2 D2",
				}},
			},
			Play: []synth.Play{
				{Clip: "verse", Bar: 9, Times: 2}, {Clip: "chorus", Bar: 17, Times: 2},
				{Clip: "nightfall", Bar: 25}, {Clip: "build", Bar: 29}, {Clip: "chorus", Bar: 33, Times: 2},
			},
		},
		{
			Name: "whistle", Patch: patches.Whistle(), Level: -1.5, Pan: 0.1,
			Sends: []synth.Send{{Bus: "woods", Level: -10.5}, {Bus: "echo", Level: -14}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Gate: 0.92, Glide: 0.08, Humanize: 0.006, Seed: 7, Bars: []string{
					"D5 - G5 - B5 - A5 G5",
					"E5 - - - G5 - E5 D5",
					"E5 - G5 - C6 - B5 A5",
					"A5 - - - F#5/G5 - - .",
				}},
				{Name: "A'", Step: 0.5, Gate: 0.92, Glide: 0.08, Humanize: 0.006, Seed: 7, Bars: []string{
					"D5 - G5 - B5 - A5 G5",
					"E5 - - - G5 - B5 D6",
					"C6 - B5 - A5 - G5 E5",
					"F#5 - A5 - G5 - - .",
				}},
			},
			Play: []synth.Play{{Clip: "A", Bar: 9}, {Clip: "A'", Bar: 13}},
		},
		{
			Name: "gang vocals", Patch: patches.GangVocals(), Level: 4.1,
			Sends: []synth.Send{{Bus: "barn", Level: -10.5}, {Bus: "woods", Level: -14}},
			Clips: []synth.Clip{{Name: "chorus", Step: 0.5, Gate: 0.95, Humanize: 0.01, Seed: 11, Bars: []string{
				"G3,G4 - - - E3,E4 - G3,G4 -",
				"B3,B4 - - - A3,A4 - G3,G4 -",
				"B3,B4 - - - D4,D5 - B3,B4 -",
				"A3,A4 - - - - - . .",
			}}},
			Play: []synth.Play{{Clip: "chorus", Bar: 17, Times: 2}, {Clip: "chorus", Bar: 33, Times: 2}},
		},
		{
			Name: "shout", Patch: patches.Shout(), Level: 6,
			Sends: []synth.Send{{Bus: "barn", Level: -9.1}},
			Clips: []synth.Clip{{Name: "hey", Step: 0.5, Humanize: 0.01, Seed: 13, Bars: []string{". . . . . . D4,A4! ."}}},
			Play:  []synth.Play{{Clip: "hey", Bar: 20}, {Clip: "hey", Bar: 24}, {Clip: "hey", Bar: 36}, {Clip: "hey", Bar: 40}},
		},
		{
			// Tremolo: each held note a run of picks across its length, four to an eighth.
			Name: "mandolin", Patch: patches.Mandolin(), Level: 2.9, Pan: -0.5,
			Sends: []synth.Send{{Bus: "barn", Level: -16.5}},
			Clips: []synth.Clip{{Name: "chorus", Step: 0.5, Bars: []string{
				"G4*16 - - - E4*8 - G4*8 -",
				"B4*16 - - - A4*8 - G4*8 -",
				"B4*16 - - - D5*8 - B4*8 -",
				"A4*24 - - - - - . .",
			}}},
			Play: []synth.Play{{Clip: "chorus", Bar: 17, Times: 2}, {Clip: "chorus", Bar: 33, Times: 2}},
		},
		{
			Name: "glockenspiel", Patch: patches.Glockenspiel(), Level: -1.9, Pan: 0.3,
			Sends: []synth.Send{{Bus: "woods", Level: -10.5}},
			Clips: []synth.Clip{{Name: "chorus", Step: 0.5, Humanize: 0.003, Seed: 9, Bars: []string{
				"G5 - - - E5 - G5 -",
				"B5 - - - A5 - G5 -",
				"B5 - - - D6 - B5 -",
				"A5 - - - - - . .",
			}}},
			Play: []synth.Play{{Clip: "chorus", Bar: 17, Times: 2}, {Clip: "chorus", Bar: 33, Times: 2}},
		},
		{
			Name: "arp", Patch: patches.SynthArp(), Level: 4.1, Pan: -0.2,
			Inserts: []synth.Effect{synth.Phaser{Rate: 0.12, Depth: 1.6, Feedback: 0.5}},
			Sends:   []synth.Send{{Bus: "echo", Level: -10.5}, {Bus: "woods", Level: -12}},
			Clips: []synth.Clip{
				{Name: "nightfall", Step: 0.25, Gate: 0.7, Vel: 0.75, Bars: []string{
					"E4 G4 B4 D5 F#5 D5 B4 G4 E4 G4 B4 D5 F#5 D5 B4 G4",
					"C4 E4 G4 B4 D5 B4 G4 E4 C4 E4 G4 B4 D5 B4 G4 E4",
					"G4 B4 D5 F#5 A5 F#5 D5 B4 G4 B4 D5 F#5 A5 F#5 D5 B4",
					"D4 F#4 A4 C#5 E5 C#5 A4 F#4 D4 F#4 A4 C#5 E5 C#5 A4 F#4",
				}},
				{Name: "chorus", Step: 0.25, Gate: 0.7, Vel: 0.75, Bars: []string{
					"C4 E4 G4 B4 D5 B4 G4 E4 C4 E4 G4 B4 D5 B4 G4 E4",
					"G4 B4 D5 F#5 A5 F#5 D5 B4 G4 B4 D5 F#5 A5 F#5 D5 B4",
					"E4 G4 B4 D5 F#5 D5 B4 G4 E4 G4 B4 D5 F#5 D5 B4 G4",
					"D4 F#4 A4 C#5 E5 C#5 A4 F#4 D4 F#4 A4 C#5 E5 C#5 A4 F#4",
				}},
			},
			Play: []synth.Play{{Clip: "nightfall", Bar: 25, Times: 2}, {Clip: "chorus", Bar: 33, Times: 2}},
		},
		{
			Name: "pad", Patch: patches.SawPad(), Level: -4,
			Sends: []synth.Send{{Bus: "woods", Level: -6}},
			Clips: []synth.Clip{
				{Name: "nightfall", Step: 0.5, Bars: []string{
					"E3,G3,B3,D4 - - - - - - -",
					"C3,G3,B3,E4 - - - - - - -",
					"G3,B3,D4,F#4 - - - - - - -",
					"D3,F#3,A3,E4 - - - - - - -",
				}},
				{Name: "chorus", Step: 0.5, Bars: []string{
					"C3,G3,B3,E4 - - - - - - -",
					"G3,B3,D4,F#4 - - - - - - -",
					"E3,G3,B3,D4 - - - - - - -",
					"D3,F#3,A3,E4 - - - - - - -",
				}},
			},
			Play: []synth.Play{{Clip: "nightfall", Bar: 25, Times: 2}, {Clip: "chorus", Bar: 33, Times: 2}},
		},
		{
			Name: "riser", Patch: patches.Riser(8 * 60.0 / 116),
			Sends: []synth.Send{{Bus: "woods", Level: -10.5}},
			Clips: []synth.Clip{{Name: "rise", Notes: []synth.Note{{Dur: 8, Key: 50, Slide: 12}}, Length: 2}},
			Play:  []synth.Play{{Clip: "rise", Bar: 31}},
		},
		{
			Name: "roll", Patch: wlSnare(), Pan: 0.05,
			Sends: []synth.Send{{Bus: "barn", Level: -14}},
			Clips: []synth.Clip{{Name: "build", Step: 0.5, Humanize: 0.004, Seed: 19, Hit: "D2", Bars: []string{
				"o o x o x x x x",
				"x*2 x*2 x*2 x*2 X*4 - X*4 -",
			}}},
			Play: []synth.Play{{Clip: "build", Bar: 31}},
		},
		{
			Name: "stomp", Patch: wlStomp(), Level: -1.5, Out: "drums",
			Sends: []synth.Send{{Bus: "barn", Level: -20}},
			Clips: []synth.Clip{
				{Name: "four", Step: 0.5, Humanize: 0.004, Seed: 19, Hit: "C2", Bars: []string{"X . . . x . . ."}},
				{Name: "stomp", Step: 0.5, Humanize: 0.004, Seed: 19, Hit: "C2", Bars: []string{"X . x . X . x ."}},
				{Name: "push", Step: 0.5, Humanize: 0.004, Seed: 19, Hit: "C2", Bars: []string{"X x X x X x X x"}},
			},
			Play: []synth.Play{
				{Clip: "four", Bar: 5, Times: 4}, {Clip: "stomp", Bar: 9, Times: 16},
				{Clip: "stomp", Bar: 29, Times: 3}, {Clip: "push", Bar: 32}, {Clip: "stomp", Bar: 33, Times: 8},
			},
		},
		{
			Name: "claps", Patch: patches.Clap(), Level: 6, Out: "drums",
			Sends: []synth.Send{{Bus: "barn", Level: -12}},
			Clips: []synth.Clip{
				{Name: "backbeat", Step: 0.5, Strum: 0.009, Humanize: 0.004, Seed: 19, Hit: "D#3,D#3,D#3", Bars: []string{". . X . . . X ."}},
				{Name: "push", Step: 0.5, Strum: 0.009, Humanize: 0.004, Seed: 19, Hit: "D#3,D#3,D#3", Bars: []string{"X . X . X x X x"}},
			},
			Play: []synth.Play{
				{Clip: "backbeat", Bar: 5, Times: 20},
				{Clip: "backbeat", Bar: 31}, {Clip: "push", Bar: 32}, {Clip: "backbeat", Bar: 33, Times: 8},
			},
		},
		{
			Name: "tambourine", Patch: patches.Tambourine(), Level: 1.6, Pan: 0.45, Out: "drums",
			Sends: []synth.Send{{Bus: "barn", Level: -20}},
			Clips: []synth.Clip{{Name: "shake", Step: 0.5, Humanize: 0.004, Seed: 19, Hit: "C5", Bars: []string{"o x X x o x X x"}}},
			Play:  []synth.Play{{Clip: "shake", Bar: 13, Times: 12}, {Clip: "shake", Bar: 33, Times: 8}},
		},
		{
			Name: "crash", Patch: patches.Cymbal(2.4), Pan: -0.3, Out: "drums",
			Sends: []synth.Send{{Bus: "barn", Level: -16.5}},
			Clips: []synth.Clip{{Name: "crash", Step: 0.5, Humanize: 0.004, Seed: 19, Hit: "C#3", Bars: []string{"X . . . . . . ."}}},
			Play:  []synth.Play{{Clip: "crash", Bar: 9}, {Clip: "crash", Bar: 17}, {Clip: "crash", Bar: 33}},
		},
	},
	Buses: []synth.Bus{
		{Name: "drums", Effects: []synth.Effect{
			synth.EQ{Bands: []synth.Band{{Kind: synth.LowShelf, Freq: 100, Gain: 2}}},
			synth.Compressor{Threshold: -18, Ratio: 3, Attack: 0.006, Release: 0.12, Makeup: 2},
		}},
		// A wooden room for the band; the woods are wider, a little darker, and longer.
		{Name: "barn", Effects: []synth.Effect{synth.Reverb{Size: 0.55, Damp: 0.5, PreDelay: 0.015}}},
		{Name: "woods", Effects: []synth.Effect{synth.Reverb{Size: 0.9, Damp: 0.55, PreDelay: 0.04}}},
		{Name: "echo", Effects: []synth.Effect{synth.Delay{Time: 1.5, Feedback: 0.35, Damp: 2600, PingPong: true}}},
	},
	Master: []synth.Effect{
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 30},
			{Kind: synth.HighShelf, Freq: 10000, Gain: -1},
		}},
		synth.Compressor{Threshold: -14, Ratio: 2, Attack: 0.02, Release: 0.2, Makeup: 1},
		synth.Limiter{},
	},
	Target: -14,
}

func Woodlands() synth.Song { return woodlands.Song() }

// A kick drum tuned up and shortened: a boot on a stage's boards.
func wlStomp() synth.Patch {
	p := patches.Kick()
	p.Oscs[0].Fixed = 58
	p.Amp.Decay = 0.28
	return p
}

func wlSnare() synth.Patch {
	p := patches.Snare()
	p.Amp.Decay = 0.12
	return p
}
