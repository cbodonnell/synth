package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Brighter Days": festival progressive house, 128 BPM.
// One progression all the way through, F#m D A E, and one hook in the 3-3-2
// sixteenth syncopation that carries a whole drop.
//
// Piano alone for four bars, then the loop: an eight-bar breakdown (piano, pad, the
// hook on a pluck), an eight-bar build (snare roll, riser, the hook passing to the lead
// as its filter opens, a half-bar of air), and a sixteen-bar drop where everything pumps
// off the kick.
var brighterDays = synth.Arrangement{
	ID: "brighterdays", Name: "Brighter Days",
	Tempo: 128, Key: "F# minor",
	Sections: []synth.Section{
		{Name: "Piano", Bars: 4},
		{Name: "Breakdown", Bars: 8},
		{Name: "Build", Bars: 8},
		{Name: "Drop", Bars: 16},
	},
	Loop: "Breakdown",
	Tracks: []synth.Track{
		{
			Name: "kick", Patch: patches.HouseKick(), Level: 1.9, Out: "drums",
			Clips: []synth.Clip{{Name: "four", Step: 1, Hit: "C2", Bars: []string{"X x x x"}}},
			Play:  []synth.Play{{Clip: "four", Bar: 21, Times: 16}},
		},
		{
			// Skips the drum bus, whose compressor would only bring it back up: it holds
			// back so the drop's can land.
			Name: "build kick", Patch: patches.HouseKick(), Level: -8,
			Clips: []synth.Clip{{Name: "four", Step: 1, Hit: "C2", Bars: []string{"x x x x"}}},
			Play:  []synth.Play{{Clip: "four", Bar: 13, Times: 6}},
		},
		{
			Name: "clap", Patch: patches.Clap(), Level: 14.8, Out: "drums",
			Sends: []synth.Send{{Bus: "hall", Level: -12}},
			Clips: []synth.Clip{{Name: "two and four", Step: 0.5, Strum: 0.011, Hit: "D#3,D#3,D#3", Bars: []string{
				". . X . . . X .",
			}}},
			Play: []synth.Play{{Clip: "two and four", Bar: 21, Times: 16}},
		},
		{
			// Doubles its rate every two bars and grows louder as it goes.
			Name: "snare roll", Patch: patches.Snare(), Level: 5.1, Out: "drums",
			Sends: []synth.Send{{Bus: "hall", Level: -14}},
			Clips: []synth.Clip{
				{Name: "quarters", Step: 1, Vel: 0.4, Hit: "D2", Bars: []string{"x x x x"}},
				{Name: "eighths", Step: 0.5, Vel: 0.55, Hit: "D2", Bars: []string{"x x x x x x x x"}},
				{Name: "sixteenths", Step: 0.25, Vel: 0.7, Hit: "D2", Bars: []string{"x x x x x x x x x x x x x x x x"}},
				{Name: "thirty-seconds", Step: 0.125, Vel: 0.9, Hit: "D2", Bars: []string{
					"x x x x x x x x x x x x x x x x . . . . . . . . . . . . . . . .",
				}},
			},
			Play: []synth.Play{
				{Clip: "quarters", Bar: 13, Times: 4},
				{Clip: "eighths", Bar: 17, Times: 2},
				{Clip: "sixteenths", Bar: 19},
				{Clip: "thirty-seconds", Bar: 20},
			},
		},
		{
			Name: "hats", Patch: patches.HiHat(), Level: -1.9, Pan: 0.25, Out: "drums",
			Clips: []synth.Clip{{Name: "sixteenths", Step: 0.25, Vel: 0.6, Hit: "F#2", Bars: []string{
				"o o x o o o x o o o x o o o x o",
			}}},
			Play: []synth.Play{{Clip: "sixteenths", Bar: 17, Times: 3}, {Clip: "sixteenths", Bar: 21, Times: 16}},
		},
		{
			Name: "open hat", Patch: patches.OpenHat(), Level: -3.1, Pan: -0.2, Out: "drums",
			Clips: []synth.Clip{{Name: "offbeats", Step: 0.5, Hit: "A#2", Bars: []string{". x . x . x . x"}}},
			Play:  []synth.Play{{Clip: "offbeats", Bar: 21, Times: 16}},
		},
		{
			Name: "crash", Patch: patches.Cymbal(2.5), Level: 2.9, Pan: -0.1, Out: "drums",
			Sends: []synth.Send{{Bus: "hall", Level: -14}},
			Clips: []synth.Clip{
				{Name: "soft", Step: 4, Vel: 0.6, Bars: []string{"C#3"}},
				{Name: "drop", Step: 4, Bars: []string{
					"C#3!", "-", "-", "-", "-", "-", "-", "-",
					"C#3", "-", "-", "-", "-", "-", "-", "-",
				}},
			},
			Play: []synth.Play{{Clip: "soft", Bar: 5}, {Clip: "drop", Bar: 21}},
		},
		{
			Name: "boom", Patch: bdBoom(), Level: -1.9, Out: "drums",
			Sends: []synth.Send{{Bus: "hall", Level: -10.5}},
			Clips: []synth.Clip{{Name: "boom", Notes: []synth.Note{{Dur: 3.2, Key: 36}}}},
			Play:  []synth.Play{{Clip: "boom", Bar: 21}},
		},
		{
			Name: "bass", Patch: patches.OffbeatBass(),
			Inserts: []synth.Effect{synth.Sidechain{Key: "kick, build kick", Depth: 0.9, Release: 0.16}},
			Clips: []synth.Clip{{Name: "offbeats", Step: 0.5, Gate: 0.8, Bars: []string{
				". F#1 . F#2 . F#1 . F#2",
				". D1 . D2 . D1 . D2",
				". A1 . A2 . A1 . A2",
				". E1 . E2 . E1 . E2",
			}}},
			Play: []synth.Play{{Clip: "offbeats", Bar: 21, Times: 4}},
		},
		{
			Name: "piano", Patch: patches.HousePiano(), Level: 1.9,
			Inserts: []synth.Effect{synth.Sidechain{Key: "kick, build kick", Depth: 0.45, Release: 0.2}},
			Sends:   []synth.Send{{Bus: "hall", Level: -10.5}},
			Clips: []synth.Clip{
				{Name: "chords", Step: 0.5, Gate: 0.95, Strum: 0.008, Vel: 0.75, Bars: []string{
					"A3,C#4,F#4! - - A3,C#4,F#4 - - A3,C#4,F#4 -",
					"A3,D4,F#4! - - A3,D4,F#4 - - A3,D4,F#4 -",
					"A3,C#4,E4! - - A3,C#4,E4 - - A3,C#4,E4 -",
					"G#3,B3,E4! - - G#3,B3,E4 - - G#3,B3,E4 -",
				}},
				{Name: "air", Step: 0.5, Gate: 0.95, Strum: 0.008, Vel: 0.75, Bars: []string{
					"A3,C#4,F#4! - - A3,C#4,F#4 - - A3,C#4,F#4 -",
					"A3,D4,F#4! - - A3,D4,F#4 - - A3,D4,F#4 -",
					"A3,C#4,E4! - - A3,C#4,E4 - - A3,C#4,E4 -",
					"G#3,B3,E4! - - G#3,B3,E4 . . . .",
				}},
			},
			Play: []synth.Play{{Clip: "chords", Bar: 1, Times: 4}, {Clip: "air", Bar: 17}, {Clip: "chords", Bar: 21, Times: 4}},
		},
		{
			Name: "piano left", Patch: patches.HousePiano(), Level: 1.9,
			Inserts: []synth.Effect{synth.Sidechain{Key: "kick, build kick", Depth: 0.45, Release: 0.2}},
			Sends:   []synth.Send{{Bus: "hall", Level: -10.5}},
			Clips: []synth.Clip{
				{Name: "roots", Step: 0.5, Gate: 0.95, Strum: 0.008, Vel: 0.75, Bars: []string{
					"F#2,C#3 - - - - - - -",
					"D2,A2 - - - - - - -",
					"A2,E3 - - - - - - -",
					"E2,B2 - - - - - - -",
				}},
				{Name: "air", Step: 0.5, Gate: 0.95, Strum: 0.008, Vel: 0.75, Bars: []string{
					"F#2,C#3 - - - - - - -",
					"D2,A2 - - - - - - -",
					"A2,E3 - - - - - - -",
					"E2,B2 - - - . . . .",
				}},
			},
			Play: []synth.Play{{Clip: "roots", Bar: 1, Times: 4}, {Clip: "air", Bar: 17}, {Clip: "roots", Bar: 21, Times: 4}},
		},
		{
			Name: "pad", Patch: patches.SawPad(),
			Inserts: []synth.Effect{synth.Sidechain{Key: "kick, build kick", Depth: 0.75, Release: 0.25}},
			Sends:   []synth.Send{{Bus: "hall", Level: -9.1}},
			Clips: []synth.Clip{
				{Name: "chords", Step: 4, Gate: 0.98, Vel: 0.7, Bars: []string{"A3,C#4,F#4", "A3,D4,F#4", "A3,C#4,E4", "G#3,B3,E4"}},
				{Name: "air", Step: 4, Gate: 0.98, Vel: 0.7, Bars: []string{"A3,C#4,F#4", "A3,D4,F#4", "A3,C#4,E4", "."}},
			},
			Play: []synth.Play{{Clip: "chords", Bar: 5, Times: 3}, {Clip: "air", Bar: 17}, {Clip: "chords", Bar: 21, Times: 4}},
		},
		{
			Name: "pluck", Patch: patches.HousePluck(), Level: 9.5, Pan: 0.15,
			Inserts: []synth.Effect{synth.AutoFilter{Points: []synth.Point{{At: 16, Value: 1400}, {At: 48, Value: 6000}}}},
			Sends:   []synth.Send{{Bus: "echo", Level: -9.1}, {Bus: "hall", Level: -10.5}},
			Clips: []synth.Clip{{Name: "hook", Step: 0.25, Gate: 0.9, Bars: []string{
				"F#5 - - E5 - - C#5 - E5 - - F#5 - - A5 -",
				"F#5 - - E5 - - D5 - E5 - - F#5 - - D5 -",
				"E5 - - C#5 - - A4 - C#5 - - E5 - - A5 -",
				"G#5 - - F#5 - - E5 - B4 - - E5 - - G#5 -",
				"F#5 - - E5 - - C#5 - E5 - - F#5 - - A5 -",
				"F#5 - - E5 - - D5 - E5 - - F#5 - - D5 -",
				"E5 - - C#5 - - A4 - C#5 - - E5 - - A5 -",
				"G#5 - - F#5 - - E5 - F#5 - - G#5 - - B5 -",
			}}},
			Play: []synth.Play{{Clip: "hook", Bar: 5}},
		},
		{
			Name: "chord stabs", Patch: patches.HousePluck(), Level: 2.9, Pan: -0.15,
			Inserts: []synth.Effect{synth.Sidechain{Key: "kick, build kick", Depth: 0.6, Release: 0.18}},
			Sends:   []synth.Send{{Bus: "hall", Level: -12}},
			Clips: []synth.Clip{{Name: "stabs", Step: 0.5, Gate: 0.6, Bars: []string{
				". A3,C#4,F#4 . A3,C#4,F#4 . A3,C#4,F#4 . A3,C#4,F#4!",
				". A3,D4,F#4 . A3,D4,F#4 . A3,D4,F#4 . A3,D4,F#4!",
				". A3,C#4,E4 . A3,C#4,E4 . A3,C#4,E4 . A3,C#4,E4!",
				". G#3,B3,E4 . G#3,B3,E4 . G#3,B3,E4 . G#3,B3,E4!",
			}}},
			Play: []synth.Play{{Clip: "stabs", Bar: 21, Times: 4}},
		},
		{
			Name: "lead", Patch: patches.SuperSaw(), Level: 8.9,
			Inserts: []synth.Effect{
				synth.AutoFilter{Resonance: 0.2, Points: []synth.Point{
					{At: 48, Value: 180},
					{At: 78, Value: 2200},
					{At: 79.98, Value: 2200},
					{At: 80, Value: 20000},
				}},
				synth.Sidechain{Key: "kick, build kick", Depth: 0.5, Release: 0.2},
			},
			Sends: []synth.Send{{Bus: "echo", Level: -12}, {Bus: "hall", Level: -9.1}},
			Clips: []synth.Clip{
				{Name: "build", Step: 0.25, Gate: 0.9, Vel: 0.5, Bars: []string{
					"F#5 - - E5 - - C#5 - E5 - - F#5 - - A5 -",
					"F#5 - - E5 - - D5 - E5 - - F#5 - - D5 -",
					"E5 - - C#5 - - A4 - C#5 - - E5 - - A5 -",
					"G#5 - - F#5 - - E5 - B4 - - E5 - - G#5 -",
					"F#5 - - E5 - - C#5 - E5 - - F#5 - - A5 -",
					"F#5 - - E5 - - D5 - E5 - - F#5 - - D5 -",
					"E5 - - C#5 - - A4 - C#5 - - E5 - - A5 -",
					"G#5 - - F#5 - - E5 - . . . . . . . .",
				}},
				{Name: "hook", Step: 0.25, Gate: 0.9, Bars: []string{
					"F#5 - - E5 - - C#5 - E5 - - F#5 - - A5 -",
					"F#5 - - E5 - - D5 - E5 - - F#5 - - D5 -",
					"E5 - - C#5 - - A4 - C#5 - - E5 - - A5 -",
					"G#5 - - F#5 - - E5 - B4 - - E5 - - G#5 -",
					"F#5 - - E5 - - C#5 - E5 - - F#5 - - A5 -",
					"F#5 - - E5 - - D5 - E5 - - F#5 - - D5 -",
					"E5 - - C#5 - - A4 - C#5 - - E5 - - A5 -",
					"G#5 - - F#5 - - E5 - F#5 - - G#5 - - B5 -",
				}},
			},
			Play: []synth.Play{{Clip: "build", Bar: 13}, {Clip: "hook", Bar: 21, Times: 2}},
		},
		{
			Name: "riser", Patch: patches.Riser(14.0625), Level: 1.6,
			Sends: []synth.Send{{Bus: "hall", Level: -8}},
			Clips: []synth.Clip{{Name: "rise", Notes: []synth.Note{{Dur: 30, Key: 54, Slide: 12}}, Length: 8}},
			Play:  []synth.Play{{Clip: "rise", Bar: 13}},
		},
	},
	Buses: []synth.Bus{
		{Name: "drums", Effects: []synth.Effect{
			synth.EQ{Bands: []synth.Band{{Kind: synth.LowShelf, Freq: 60, Gain: 2}}},
			synth.Compressor{Threshold: -16, Ratio: 3, Attack: 0.008, Release: 0.1, Makeup: 2},
		}},
		// The tails pump with everything else, which is half of what makes it breathe.
		{Name: "hall", Effects: []synth.Effect{
			synth.Reverb{Size: 0.88, Damp: 0.3, PreDelay: 0.03},
			synth.Sidechain{Key: "kick, build kick", Depth: 0.7, Release: 0.25},
		}},
		{Name: "echo", Effects: []synth.Effect{
			synth.Delay{Time: 0.75, Feedback: 0.4, Damp: 4500, PingPong: true},
			synth.Sidechain{Key: "kick, build kick", Depth: 0.6, Release: 0.25},
		}},
	},
	Master: []synth.Effect{
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 28},
			{Kind: synth.HighShelf, Freq: 9000, Gain: 1.5},
		}},
		synth.Compressor{Threshold: -12, Ratio: 2, Attack: 0.015, Release: 0.15},
		synth.Limiter{Ceiling: -0.8},
	},
	Target: -14,
}

func BrighterDays() synth.Song { return brighterDays.Song() }

// A kick tuned down and left to ring, for the first beat of the drop.
func bdBoom() synth.Patch {
	p := patches.HouseKick()
	p.Amp.Decay = 1.8
	p.Sweep = synth.Sweep{Semis: 36, Time: 0.08}
	return p
}
