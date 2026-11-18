package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Smoke Signal": 2015 festival trap, 140 BPM played half-time, in F minor over the
// epic i VI iv V (Fm Db Bbm C).
//
// Four bars of music-box bells over a dark pad, then the loop: an eight-bar build
// (riser, a snare roll that keeps doubling, the whole track tape-stopping into a half
// bar of silence that the air horns fill) and a sixteen-bar drop of 808, brass, rolling
// hats, and a snare on three. Drum and 808 lines are in sixteenths, hats in eighths.
// The tape stop is an insert on each part that should spin down, so the horns survive it.
var smokeSignal = synth.Arrangement{
	ID: "smokesignal", Name: "Smoke Signal",
	Tempo: 140, Key: "F minor",
	Sections: []synth.Section{{Name: "Bells", Bars: 4}, {Name: "Build", Bars: 8}, {Name: "Drop", Bars: 16}},
	Loop:     "Build",
	Tracks: []synth.Track{
		{
			Name: "kick", Patch: patches.TrapKick(), Level: -1.3, Out: "drums",
			Inserts: []synth.Effect{synth.TapeStop{At: 45, Length: 1, Until: 48}},
			Clips: []synth.Clip{
				{Name: "build", Step: 0.25, Hit: "C2", Bars: []string{"x . . . . . . . . . . . . . . ."}},
				{Name: "drop", Step: 0.25, Hit: "C2", Bars: []string{
					"X . . . . . . x . . X . . . . .",
					"X . . . . . . . . . X . . x . .",
				}},
			},
			Play: []synth.Play{{Clip: "build", Bar: 5, Times: 6}, {Clip: "drop", Bar: 13, Times: 8}},
		},
		{
			Name: "snare", Patch: patches.TrapSnare(), Level: 2.9, Out: "drums",
			Sends: []synth.Send{{Bus: "hall", Level: -10.5}},
			Clips: []synth.Clip{{Name: "three", Step: 0.25, Hit: "D2", Bars: []string{
				". . . . . . . . X . . . . . . .",
				". . . . . . . . X . . . . . . .",
				". . . . . . . . X . . . . . . .",
				". . . . . . . . X . . . . . o x",
			}}},
			Play: []synth.Play{{Clip: "three", Bar: 13, Times: 4}},
		},
		{
			Name: "clap", Patch: patches.Clap(), Level: 12.5, Out: "drums",
			Sends: []synth.Send{{Bus: "hall", Level: -10.5}},
			Clips: []synth.Clip{{Name: "three", Step: 0.25, Strum: 0.009, Hit: "D#3,D#3,D#3", Bars: []string{
				". . . . . . . . X . . . . . . .",
			}}},
			Play: []synth.Play{{Clip: "three", Bar: 13, Times: 16}},
		},
		{
			// Four bars of quarters, two of eighths, one of sixteenths, half a bar of
			// thirty-seconds, each louder than the last. It skips the drum bus, whose
			// compressor would undo its crescendo.
			Name: "snare roll", Patch: patches.TrapSnare(), Level: -8,
			Inserts: []synth.Effect{synth.TapeStop{At: 45, Length: 1, Until: 48}},
			Sends:   []synth.Send{{Bus: "hall", Level: -14}},
			Clips: []synth.Clip{
				{Name: "quarters", Step: 1, Vel: 0.35, Hit: "D2", Bars: []string{"x x x x"}},
				{Name: "eighths", Step: 0.5, Vel: 0.5, Hit: "D2", Bars: []string{"x x x x x x x x"}},
				{Name: "sixteenths", Step: 0.25, Vel: 0.65, Hit: "D2", Bars: []string{"x x x x x x x x x x x x x x x x"}},
				{Name: "thirty-seconds", Step: 0.125, Vel: 0.85, Hit: "D2", Bars: []string{
					"x x x x x x x x x x x x x x x x . . . . . . . . . . . . . . . .",
				}},
			},
			Play: []synth.Play{
				{Clip: "quarters", Bar: 5, Times: 4},
				{Clip: "eighths", Bar: 9, Times: 2},
				{Clip: "sixteenths", Bar: 11},
				{Clip: "thirty-seconds", Bar: 12},
			},
		},
		{
			Name: "hats", Patch: patches.HiHat(), Level: -2.3, Pan: 0.2, Out: "drums",
			Inserts: []synth.Effect{synth.TapeStop{At: 45, Length: 1, Until: 48}},
			Clips: []synth.Clip{
				{Name: "eighths", Step: 0.5, Vel: 0.7, Hit: "F#2", Bars: []string{"x x x x x x x x"}},
				{Name: "stop", Step: 0.5, Vel: 0.7, Hit: "F#2", Bars: []string{"x x . . . . . ."}},
				{Name: "rolls", Step: 0.5, Vel: 0.7, Hit: "F#2", Bars: []string{
					"x x x x x x x*3 x",
					"x x x*2 x x x x*4 x*3",
					"x x x x x*3 x x x",
					"x x x*2 x x*3 x*3 x*4 x*6",
				}},
			},
			Play: []synth.Play{{Clip: "eighths", Bar: 7, Times: 5}, {Clip: "stop", Bar: 12}, {Clip: "rolls", Bar: 13, Times: 4}},
		},
		{
			Name: "open hat", Patch: patches.OpenHat(), Level: -3.1, Pan: -0.2, Out: "drums",
			Clips: []synth.Clip{{Name: "every other", Step: 0.5, Vel: 0.6, Hit: "A#2", Bars: []string{
				". . . . . . . .",
				". . . x . . . .",
			}}},
			Play: []synth.Play{{Clip: "every other", Bar: 13, Times: 8}},
		},
		{
			Name: "crash", Patch: patches.Cymbal(2.5), Level: 1.6, Pan: -0.1, Out: "drums",
			Sends: []synth.Send{{Bus: "hall", Level: -14}},
			Clips: []synth.Clip{{Name: "drop", Step: 4, Bars: []string{
				"C#3!", "-", "-", "-", "-", "-", "-", "-",
				"C#3", "-", "-", "-", "-", "-", "-", "-",
			}}},
			Play: []synth.Play{{Clip: "drop", Bar: 13}},
		},
		{
			Name: "808", Patch: patches.EightOhEight(), Level: -0.9,
			Clips: []synth.Clip{{Name: "drop", Step: 0.25, Gate: 0.97, Glide: 0.09, Bars: []string{
				"F1 - - - - - - F1 - - F1 - - - - -",
				"Db1 - - - - - - - - - Db1 - - Db1 - -",
				"Bb1 - - - - - - Bb1 - - Bb1 - - - - -",
				"C2 - - - - - - - - - C2 - - C2/F1 - -",
			}}},
			Play: []synth.Play{{Clip: "drop", Bar: 13, Times: 4}},
		},
		{
			Name: "boom", Patch: ssBoom(), Level: -0.6,
			Sends: []synth.Send{{Bus: "hall", Level: -14}},
			Clips: []synth.Clip{{Name: "boom", Notes: []synth.Note{{Dur: 2.8, Key: 36}}}},
			Play:  []synth.Play{{Clip: "boom", Bar: 13}},
		},
		{
			Name: "bells", Patch: patches.Bell(), Level: -6, Pan: 0.15,
			Inserts: []synth.Effect{synth.TapeStop{At: 45, Length: 1, Until: 48}},
			Sends:   []synth.Send{{Bus: "echo", Level: -10.5}, {Bus: "hall", Level: -12}},
			Clips: []synth.Clip{
				{Name: "music box", Step: 0.5, Gate: 0.9, Vel: 0.7, Bars: []string{
					"F5 C5 Ab5 C5 G5 C5 Ab5 C5",
					"F5 Db5 Ab5 Db5 G5 Db5 Ab5 Db5",
					"F5 Db5 Bb5 Db5 Ab5 Db5 F5 Db5",
					"E5 C5 G5 C5 E5 G5 C6 G5",
				}},
				{Name: "stop", Step: 0.5, Gate: 0.9, Vel: 0.7, Bars: []string{
					"F5 C5 Ab5 C5 G5 C5 Ab5 C5",
					"F5 Db5 Ab5 Db5 G5 Db5 Ab5 Db5",
					"F5 Db5 Bb5 Db5 Ab5 Db5 F5 Db5",
					"E5 C5 G5 C5 . . . .",
				}},
			},
			Play: []synth.Play{{Clip: "music box", Bar: 1, Times: 2}, {Clip: "stop", Bar: 9}},
		},
		{
			Name: "pad", Patch: patches.SawPad(),
			Inserts: []synth.Effect{
				synth.Sidechain{Key: "kick", Depth: 0.6, Release: 0.22},
				synth.TapeStop{At: 45, Length: 1, Until: 48},
			},
			Sends: []synth.Send{{Bus: "hall", Level: -10.5}},
			Clips: []synth.Clip{{Name: "chords", Step: 4, Gate: 0.98, Vel: 0.6, Bars: []string{
				"F3,Ab3,C4", "F3,Ab3,Db4", "F3,Bb3,Db4", "E3,G3,C4",
			}}},
			Play: []synth.Play{{Clip: "chords", Bar: 1, Times: 7}},
		},
		{
			Name: "riser", Patch: patches.Riser(29 * 60.0 / 140), Level: 6.4,
			Inserts: []synth.Effect{synth.TapeStop{At: 45, Length: 1, Until: 48}},
			Sends:   []synth.Send{{Bus: "hall", Level: -10.5}},
			Clips:   []synth.Clip{{Name: "rise", Notes: []synth.Note{{Dur: 29, Key: 53, Slide: 12}}, Length: 8}},
			Play:    []synth.Play{{Clip: "rise", Bar: 5}},
		},
		{
			Name: "air horn", Patch: patches.AirHorn(), Level: -0.9,
			Sends: []synth.Send{{Bus: "hall", Level: -9.1}},
			Clips: []synth.Clip{{Name: "horns", Step: 0.25, Gate: 0.75, Bars: []string{
				". . . . . . F4 . F4 . F4! - - - - -",
			}}},
			Play: []synth.Play{{Clip: "horns", Bar: 12}},
		},
		{
			Name: "brass", Patch: patches.Brass(), Level: 2.3,
			Inserts: []synth.Effect{synth.Sidechain{Key: "kick", Depth: 0.35, Release: 0.22}},
			Sends:   []synth.Send{{Bus: "hall", Level: -14}},
			Clips: []synth.Clip{{Name: "stabs", Step: 0.25, Gate: 0.85, Strum: 0.004, Bars: []string{
				"F3,C4,F4,Ab4! . . F3,C4,F4,Ab4 . . F3,C4,F4,Ab4 . . . F3,C4,F4,Ab4 . F3,C4,F4,Ab4! - - -",
				"Db3,Ab3,Db4,F4! . . Db3,Ab3,Db4,F4 . . Db3,Ab3,Db4,F4 . . . Db3,Ab3,Db4,F4 . Db3,Ab3,Db4,F4! - - -",
				"Bb2,F3,Bb3,Db4! . . Bb2,F3,Bb3,Db4 . . Bb2,F3,Bb3,Db4 . . . Bb2,F3,Bb3,Db4 . Bb2,F3,Bb3,Db4! - - -",
				"C3,G3,C4,E4! . . C3,G3,C4,E4 . . C3,G3,C4,E4 . . . C3,G3,C4,E4 . C3,G3,C4,E4! - - -",
			}}},
			Play: []synth.Play{{Clip: "stabs", Bar: 13, Times: 4}},
		},
		{
			Name: "lead", Patch: ssLead(), Level: 8,
			Inserts: []synth.Effect{synth.Sidechain{Key: "kick", Depth: 0.3, Release: 0.22}},
			Sends:   []synth.Send{{Bus: "echo", Level: -12}, {Bus: "hall", Level: -12}},
			Clips: []synth.Clip{{Name: "melody", Step: 0.25, Gate: 0.9, Glide: 0.07, Bars: []string{
				"C5 - - C5 - - Ab4 - C5 - - Eb5 - C5 - -",
				"Db5 - - C5 - - Ab4 - F4 - - Ab4 - - - -",
				"Bb4 - - Db5 - - F5 - Eb5 - - Db5 - C5 - -",
				"E5/F5 - - - C5 - - - G4 - - Bb4 - E5 - -",
			}}},
			Play: []synth.Play{{Clip: "melody", Bar: 21, Times: 2}},
		},
	},
	Buses: []synth.Bus{
		{Name: "drums", Effects: []synth.Effect{
			synth.Compressor{Threshold: -14, Ratio: 3, Attack: 0.005, Release: 0.08, Makeup: 2},
			synth.Drive{Amount: 1.5, Mix: 0.3},
		}},
		{Name: "hall", Effects: []synth.Effect{synth.Reverb{Size: 0.85, Damp: 0.4, PreDelay: 0.02}}},
		{Name: "echo", Effects: []synth.Effect{synth.Delay{Time: 0.75, Feedback: 0.4, Damp: 3500, PingPong: true}}},
	},
	Master: []synth.Effect{
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 25},
			{Kind: synth.HighShelf, Freq: 9000, Gain: 1},
		}},
		synth.Compressor{Threshold: -12, Ratio: 2, Attack: 0.01, Release: 0.15, Makeup: -1.7},
		synth.Limiter{Ceiling: -0.8},
	},
	Target: -14,
}

func SmokeSignal() synth.Song { return smokeSignal.Song() }

// The 808 struck from two octaves up and left to ring, for the first beat of the drop.
func ssBoom() synth.Patch {
	p := patches.EightOhEight()
	p.Sweep = synth.Sweep{Semis: 24, Time: 0.12}
	p.Amp.Decay = 1.4
	return p
}

func ssLead() synth.Patch {
	p := patches.SuperSaw()
	p.Unison.Voices = 5
	p.LFOs = []synth.LFO{{Target: synth.Vibrato, Rate: 6, Depth: 30, Delay: 0.15}}
	return p
}
