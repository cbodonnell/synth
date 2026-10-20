package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Long Light", flywheel's theme: the Turning at sundown, heard from the hill. Downtempo
// in G major over open chords, warm and tape-soft: a chorused Juno pad and arp breathing
// on the kick, a picked bass, a clean guitar tune lost in dotted-eighth echoes.
//
// Four bars of intro (pad, and the arp opening up), then a 32-bar loop: A is the groove
// on Gmaj9 D/F# Em9 C6/9; A' brings in the guitar and Rhodes; B drops the kick for
// Em9 C6/9 Gmaj9 Am11 under a string-machine tune and builds back; A” is everything,
// a glockenspiel answering the guitar.
var longLight = synth.Arrangement{
	ID: "theme", Name: "Long Light",
	Tempo: 100, Key: "G major",
	Sections: []synth.Section{
		{Name: "Intro", Bars: 4},
		{Name: "A", Bars: 8},
		{Name: "A'", Bars: 8},
		{Name: "B", Bars: 8},
		{Name: "A''", Bars: 8},
	},
	Loop: "A",
	Tracks: []synth.Track{
		{
			Name: "kick", Patch: longLightKick(), Level: -9.8970004, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -26.0206}},
			Clips: []synth.Clip{
				{Name: "groove", Step: 0.25, Swing: 0.08, Humanize: 0.003, Seed: 8, Hit: "C2", Bars: []string{
					"X . . . . . x . X . . . . . . .",
					"X . . . . . x . X . . x . . . .",
				}},
				{Name: "four", Step: 0.25, Swing: 0.08, Humanize: 0.003, Seed: 8, Hit: "C2", Bars: []string{
					"x . . . x . . . x . . . x . . .",
				}},
			},
			Play: []synth.Play{
				{Clip: "groove", Bar: 5, Times: 8},
				{Clip: "four", Bar: 27, Times: 2},
				{Clip: "groove", Bar: 29, Times: 4},
			},
		},
		{
			Name: "snare", Patch: longLightSnare(), Level: 2.9225607, Pan: 0.05, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -13.9794}, {Bus: "sky", Level: -16.4781748}},
			Clips: []synth.Clip{
				{Name: "backbeat", Step: 0.25, Swing: 0.08, Humanize: 0.003, Seed: 8, Hit: "C3", Bars: []string{
					". . . . X . . . . . . . X . . o",
				}},
				{Name: "fill", Step: 0.25, Swing: 0.08, Humanize: 0.003, Seed: 8, Hit: "C3", Bars: []string{
					". . . . . . . . o . o . x . x x",
				}},
			},
			Play: []synth.Play{
				{Clip: "backbeat", Bar: 5, Times: 16},
				{Clip: "fill", Bar: 28},
				{Clip: "backbeat", Bar: 29, Times: 8},
			},
		},
		{
			Name: "hihat", Patch: patches.HiHat(), Level: -0.91515, Pan: 0.3, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -20}},
			Clips: []synth.Clip{
				{Name: "eighths", Step: 0.25, Swing: 0.08, Humanize: 0.003, Seed: 8, Hit: "C5", Bars: []string{
					"x . x . x . x . x . x . x . x .",
				}},
				{Name: "sixteenths", Step: 0.25, Swing: 0.08, Humanize: 0.003, Seed: 8, Hit: "C5", Bars: []string{
					"x o x o x o x o x o x o x o x o",
				}},
				{Name: "offbeats", Step: 0.25, Swing: 0.08, Humanize: 0.003, Seed: 8, Hit: "C5", Bars: []string{
					". . x . . . x . . . x . . . x .",
				}},
			},
			Play: []synth.Play{
				{Clip: "eighths", Bar: 5, Times: 8},
				{Clip: "sixteenths", Bar: 13, Times: 8},
				{Clip: "offbeats", Bar: 21, Times: 8},
				{Clip: "sixteenths", Bar: 29, Times: 8},
			},
		},
		{
			Name: "shaker", Patch: patches.Shaker(), Pan: -0.35, Out: "drums",
			Sends: []synth.Send{{Bus: "room", Level: -20}},
			Clips: []synth.Clip{{Name: "shake", Step: 0.25, Swing: 0.08, Humanize: 0.003, Seed: 8, Hit: "C4", Bars: []string{
				"o o x o o o x o o o x o o o x o",
			}}},
			Play: []synth.Play{{Clip: "shake", Bar: 21, Times: 8}},
		},
		{
			Name: "pad", Patch: longLightPad(), Level: 3.521825,
			Inserts: []synth.Effect{
				synth.Chorus{Rate: 0.5, Depth: 0.004, Mix: 0.5},
				synth.Sidechain{Key: "kick", Depth: 0.35, Release: 0.3},
			},
			Sends: []synth.Send{{Bus: "sky", Level: -9.118639}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.5, Gate: 0.98, Vel: 0.7, Bars: []string{
					"B3,D4,F#4,A4 - - - - - - -",
					"A3,D4,E4,F#4 - - - - - - -",
					"G3,B3,D4,F#4 - - - - - - -",
					"G3,A3,D4,E4 - - - - - - -",
				}},
				{Name: "B", Step: 0.5, Gate: 0.98, Vel: 0.7, Bars: []string{
					"G3,B3,D4,F#4 - - - - - - -",
					"G3,A3,D4,E4 - - - - - - -",
					"B3,D4,F#4,A4 - - - - - - -",
					"G3,C4,D4,E4 - - - - - - -",
				}},
			},
			Play: []synth.Play{
				{Clip: "A", Bar: 1, Times: 5},
				{Clip: "B", Bar: 21, Times: 2},
				{Clip: "A", Bar: 29, Times: 2},
			},
		},
		{
			Name: "arp", Patch: longLightArp(), Level: 6.848454, Pan: -0.2,
			Inserts: []synth.Effect{
				synth.AutoFilter{Resonance: 0.15, Points: []synth.Point{{At: 0, Value: 500}, {At: 16, Value: 5000}}},
				synth.Chorus{Rate: 0.5, Depth: 0.004, Mix: 0.5},
				synth.Sidechain{Key: "kick", Depth: 0.25, Release: 0.3},
			},
			Sends: []synth.Send{{Bus: "echo", Level: -9.118639}, {Bus: "sky", Level: -13.9794}},
			Clips: []synth.Clip{
				{Name: "A", Step: 0.25, Gate: 0.6, Vel: 0.65, Swing: 0.08, Humanize: 0.003, Seed: 2, Bars: []string{
					"B4 F#5 A5 D5 F#5 A5 B4 A5 B4 F#5 A5 D5 F#5 A5 B4 A5",
					"A4 E5 F#5 D5 E5 F#5 A4 F#5 A4 E5 F#5 D5 E5 F#5 A4 F#5",
					"G4 D5 F#5 B4 D5 F#5 G4 F#5 G4 D5 F#5 B4 D5 F#5 G4 F#5",
					"G4 D5 E5 A4 D5 E5 G4 E5 G4 D5 E5 A4 D5 E5 G4 E5",
				}},
				{Name: "B", Step: 0.25, Gate: 0.6, Vel: 0.65, Swing: 0.08, Humanize: 0.003, Seed: 2, Bars: []string{
					"G4 D5 F#5 B4 D5 F#5 G4 F#5 G4 D5 F#5 B4 D5 F#5 G4 F#5",
					"G4 D5 E5 A4 D5 E5 G4 E5 G4 D5 E5 A4 D5 E5 G4 E5",
					"B4 F#5 A5 D5 F#5 A5 B4 A5 B4 F#5 A5 D5 F#5 A5 B4 A5",
					"G4 D5 E5 C5 D5 E5 G4 E5 G4 D5 E5 C5 D5 E5 G4 E5",
				}},
			},
			Play: []synth.Play{
				{Clip: "A", Bar: 1, Times: 5},
				{Clip: "B", Bar: 21},
				{Clip: "A", Bar: 29, Times: 2},
			},
		},
		{
			Name: "bass", Patch: longLightBass(), Level: -4.436975,
			Inserts: []synth.Effect{
				synth.Compressor{Threshold: -20, Ratio: 4, Attack: 0.01, Makeup: 4},
				synth.Sidechain{Key: "kick", Depth: 0.3, Release: 0.3},
			},
			Clips: []synth.Clip{
				{Name: "pickup", Step: 0.5, Gate: 0.85, Humanize: 0.006, Seed: 3, Bars: []string{
					". . . . . . G1 A1",
				}},
				{Name: "A", Step: 0.5, Gate: 0.85, Humanize: 0.006, Seed: 3, Bars: []string{
					"G1 - - G1 . G1 G2 D2",
					"F#1 - - F#1 . F#1 F#2 A1",
					"E1 - - E1 . E1 E2 B1",
					"C2 - - C2 . C2 C3 G2",
				}},
				{Name: "B long", Step: 0.5, Gate: 0.85, Humanize: 0.006, Seed: 3, Bars: []string{
					"E1 - - - - - E2 -",
					"C2 - - - - - C3 -",
					"G1 - - - - - G2 -",
					"A1 - - - - - A2 -",
				}},
				{Name: "B", Step: 0.5, Gate: 0.85, Humanize: 0.006, Seed: 3, Bars: []string{
					"E1 - - E1 . E1 E2 B1",
					"C2 - - C2 . C2 C3 G2",
					"G1 - - G1 . G1 G2 D2",
					"A1 - - A1 . A1 A2 E2",
				}},
			},
			Play: []synth.Play{
				{Clip: "pickup", Bar: 4},
				{Clip: "A", Bar: 5, Times: 4},
				{Clip: "B long", Bar: 21},
				{Clip: "B", Bar: 25},
				{Clip: "A", Bar: 29, Times: 2},
			},
		},
		{
			Name: "guitar", Patch: patches.RhythmGuitar(), Level: 1.583625, Pan: 0.25,
			Inserts: append(patches.GuitarAmp(0.6), synth.Chorus{Rate: 0.7, Mix: 0.35}),
			Sends:   []synth.Send{{Bus: "echo", Level: -6.93575}, {Bus: "sky", Level: -10.4575749}},
			Clips: []synth.Clip{{Name: "tune", Step: 0.5, Gate: 0.9, Glide: 0.05, Humanize: 0.006, Seed: 4, Bars: []string{
				"B4 - - A4 B4 - D5 -",
				"E5 - - D5 - - A4 -",
				"B4 - - A4 G4 - E4 -",
				"G4 - - - A4 - - -",
				"B4 - - A4 B4 - D5 E5",
				"F#5 - - E5 - - D5 -",
				"E5 - - D5 B4 - A4 -",
				"A4 - - - - - - -",
			}}},
			Play: []synth.Play{{Clip: "tune", Bar: 13}, {Clip: "tune", Bar: 29}},
		},
		{
			Name: "rhodes", Patch: patches.EPiano(), Level: -7.9588, Pan: -0.3,
			Inserts: []synth.Effect{synth.EQ{Bands: []synth.Band{{Kind: synth.HighCut, Freq: 4000}}}},
			Sends:   []synth.Send{{Bus: "room", Level: -16.4781748}, {Bus: "echo", Level: -16.4781748}},
			Clips: []synth.Clip{{Name: "comp", Step: 0.5, Vel: 0.55, Strum: 0.01, Humanize: 0.006, Seed: 5, Bars: []string{
				". . B3,D4,F#4,A4 - . . B3,D4,F#4,A4 -",
				". . A3,D4,E4,F#4 - . . A3,D4,E4,F#4 -",
				". . G3,B3,D4,F#4 - . . G3,B3,D4,F#4 -",
				". . G3,A3,D4,E4 - . . G3,A3,D4,E4 -",
			}}},
			Play: []synth.Play{{Clip: "comp", Bar: 13, Times: 2}, {Clip: "comp", Bar: 29, Times: 2}},
		},
		{
			// The lift, then the guitar's tune an octave down beneath it.
			Name: "strings", Patch: patches.OctaveSynth(), Level: 2.278867, Pan: 0.1,
			Inserts: []synth.Effect{synth.Chorus{Rate: 0.5, Depth: 0.004, Mix: 0.5}},
			Sends:   []synth.Send{{Bus: "sky", Level: -7.9588}, {Bus: "echo", Level: -13.9794}},
			Clips: []synth.Clip{
				{Name: "lift", Step: 0.5, Vel: 0.7, Glide: 0.12, Humanize: 0.006, Seed: 6, Bars: []string{
					"B4 - - - - - - -",
					"C5 - - - E5 - - -",
					"D5 - - - - - B4 -",
					"A4 - - - - - - -",
					"G4 - - - B4 - - -",
					"E5 - - - D5 - - -",
					"F#5 - - - D5 - - -",
					"E5 - - - - - D5/E5 -",
				}},
				{Name: "tune", Step: 0.5, Gate: 0.9, Glide: 0.05, Humanize: 0.006, Seed: 4, Bars: []string{
					"B3 - - A3 B3 - D4 -",
					"E4 - - D4 - - A3 -",
					"B3 - - A3 G3 - E3 -",
					"G3 - - - A3 - - -",
					"B3 - - A3 B3 - D4 E4",
					"F#4 - - E4 - - D4 -",
					"E4 - - D4 B3 - A3 -",
					"A3 - - - - - - -",
				}},
			},
			Play: []synth.Play{{Clip: "lift", Bar: 21}, {Clip: "tune", Bar: 29}},
		},
		{
			Name: "glockenspiel", Patch: patches.Glockenspiel(), Level: -6.0206, Pan: 0.45,
			Sends: []synth.Send{{Bus: "sky", Level: -7.9588}, {Bus: "echo", Level: -9.118639}},
			Clips: []synth.Clip{{Name: "answer", Step: 0.5, Vel: 0.5, Humanize: 0.006, Seed: 7, Bars: []string{
				". . . . . . D6 A5",
				". . . . . . . .",
				". . . . . . B5 .",
				"A5 . G5 . . . . .",
			}}},
			Play: []synth.Play{{Clip: "answer", Bar: 29, Times: 2}},
		},
	},
	Buses: []synth.Bus{
		{Name: "drums", Level: -3.098039, Effects: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowShelf, Freq: 90, Gain: 2},
				{Kind: synth.Peak, Freq: 450, Gain: -3, Q: 0.9},
				{Kind: synth.HighCut, Freq: 9000},
			}},
			synth.Drive{Amount: 1.5, Mix: 0.3},
			synth.Compressor{Threshold: -18, Ratio: 4, Attack: 0.008, Release: 0.12, Makeup: 4},
		}},
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.35, Damp: 0.6, PreDelay: 0.01}}},
		{Name: "sky", Effects: []synth.Effect{synth.Reverb{Size: 0.9, Damp: 0.45, PreDelay: 0.04}}},
		{Name: "echo", Effects: []synth.Effect{
			synth.Delay{Time: 0.75, Feedback: 0.45, Damp: 2400, PingPong: true},
			synth.Reverb{Size: 0.6, Damp: 0.5, Mix: 0.3},
		}},
	},
	Master: []synth.Effect{
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 30},
			{Kind: synth.LowShelf, Freq: 150, Gain: 1},
			{Kind: synth.HighShelf, Freq: 8000, Gain: -2.5},
		}},
		synth.Drive{Amount: 1, Tone: 12000, Mix: 0.25},
		synth.Compressor{Threshold: -16, Ratio: 2, Attack: 0.02, Release: 0.25},
		synth.Limiter{},
	},
}

func LongLight() synth.Song { return longLight.Song() }

func longLightKick() synth.Patch {
	p := patches.Kick()
	p.Oscs[0].Fixed = 48
	p.Amp.Decay = 0.28
	return p
}

func longLightSnare() synth.Patch {
	p := patches.Snare()
	p.Filter.Cutoff = 2200
	return p
}

func longLightPad() synth.Patch {
	p := patches.SawPad()
	p.Filter.Cutoff = 1300
	p.Amp.Attack = 0.8
	return p
}

func longLightArp() synth.Patch {
	p := patches.SynthArp()
	p.Filter.Cutoff = 420
	p.Filter.EnvAmount = 2.6
	return p
}

func longLightBass() synth.Patch {
	p := patches.Bass()
	p.Filter.Cutoff = 380
	return p
}
