package songs

import (
	"strings"

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
const (
	ssBeat = 60.0 / 140
	ssBar  = 4 * ssBeat
	ss8    = ssBeat / 2
	ss16   = ssBeat / 4

	ssBuild = 4 * ssBar
	ssStop  = 11.25 * ssBar
	ssDrop  = 12 * ssBar
	ssEnd   = 28 * ssBar
)

var (
	ssPad   = []string{"F3,Ab3,C4", "F3,Ab3,Db4", "F3,Bb3,Db4", "E3,G3,C4"}
	ssBrass = []string{"F3,C4,F4,Ab4", "Db3,Ab3,Db4,F4", "Bb2,F3,Bb3,Db4", "C3,G3,C4,E4"}
	ssBells = []string{
		"F5 C5 Ab5 C5 G5 C5 Ab5 C5",
		"F5 Db5 Ab5 Db5 G5 Db5 Ab5 Db5",
		"F5 Db5 Bb5 Db5 Ab5 Db5 F5 Db5",
		"E5 C5 G5 C5 E5 G5 C6 G5",
	}
	ss808 = []string{
		"F1 - - - - - - F1 - - F1 - - - - -",
		"Db1 - - - - - - - - - Db1 - - Db1 - -",
		"Bb1 - - - - - - Bb1 - - Bb1 - - - - -",
		"C2 - - - - - - - - - C2 - - C2/F1 - -",
	}
	ssKicks = []string{
		"X . . . . . . x . . X . . . . .",
		"X . . . . . . . . . X . . x . .",
	}
	ssHats = []string{
		"x x x x x x x*3 x",
		"x x x*2 x x x x*4 x*3",
		"x x x x x*3 x x x",
		"x x x*2 x x*3 x*3 x*4 x*6",
	}
)

func SmokeSignal() synth.Song {
	tape := synth.TapeStop{At: ssStop, Length: 0.25 * ssBar, Until: ssDrop}

	bells := synth.Phrase{Step: ss8, Gate: 0.9, Vel: 0.7}.Notes(
		bars(11, func(c int) string { return ssBells[c] }),
		strings.Join(strings.Fields(ssBells[3])[:4], " "), silence(4),
	)
	pad := synth.Phrase{Step: ssBar, Gate: 0.98, Vel: 0.6}.Notes(
		bars(12, func(c int) string { return ssPad[c] }),
		bars(16, func(c int) string { return ssPad[c] }),
	)

	// The build: four bars of quarters, two of eighths, one of sixteenths, half a bar of
	// thirty-seconds, each louder than the last.
	roll := func(bar, step, vel float64, n int) []synth.Note {
		return synth.Phrase{At: ssBuild + bar*ssBar, Step: step, Vel: vel}.Notes(
			synth.Pattern("D2", strings.Repeat("x ", n)))
	}
	snareRoll := synth.Merge(
		roll(0, ssBeat, 0.35, 16),
		roll(4, ss8, 0.5, 16),
		roll(6, ss16, 0.65, 16),
		roll(7, ss16/2, 0.85, 16),
	)
	riser := []synth.Note{{At: ssBuild, Dur: ssStop - ssBuild, Key: synth.MustKey("F3"), Slide: 12}}
	horns := synth.Phrase{At: ssStop + 0.25*ssBar/2, Step: ss16, Gate: 0.75}.Notes("F4 . F4 . F4! - - - - -")

	sixteenths := synth.Phrase{Step: ss16}
	kick := sixteenths.Notes(
		silence(4*16),
		strings.Repeat("C2 "+silence(15), 6), silence(2*16),
		bars(16, func(c int) string { return synth.Pattern("C2", ssKicks[c%2]) }),
	)
	snare := sixteenths.Notes(
		silence(12*16),
		bars(16, func(c int) string {
			if c == 3 {
				return synth.Pattern("D2", ". . . . . . . . X . . . . . o x")
			}
			return synth.Pattern("D2", ". . . . . . . . X . . . . . . .")
		}),
	)
	clap := synth.Phrase{Step: ss16, Strum: 0.009}.Notes(
		silence(12*16),
		bars(16, func(int) string { return synth.Pattern("D#3,D#3,D#3", ". . . . . . . . X . . . . . . .") }),
	)
	hats := synth.Phrase{Step: ss8, Vel: 0.7}.Notes(
		silence(6*8),
		synth.Pattern("F#2", strings.Repeat("x x x x x x x x | ", 5)), synth.Pattern("F#2", "x x"), silence(6),
		bars(16, func(c int) string { return synth.Pattern("F#2", ssHats[c]) }),
	)
	openHat := synth.Phrase{Step: ss8, Vel: 0.6}.Notes(
		silence(12*8),
		bars(16, func(c int) string {
			if c%2 == 1 {
				return synth.Pattern("A#2", ". . . x . . . .")
			}
			return silence(8)
		}),
	)
	crash := synth.Phrase{At: ssDrop, Step: 8 * ssBar}.Notes("C#3! C#3")
	boom := []synth.Note{{At: ssDrop, Dur: 1.2, Key: synth.MustKey("C2")}}

	sub := synth.Phrase{Step: ss16, Gate: 0.97, Glide: 0.09}.Notes(
		silence(12*16),
		bars(16, func(c int) string { return ss808[c] }),
	)
	brass := synth.Phrase{Step: ss16, Gate: 0.85, Strum: 0.004}.Notes(
		silence(12*16),
		bars(16, func(c int) string { return synth.Pattern(ssBrass[c], "X . . x . . x . . . x . X - - -") }),
	)
	melody := []string{
		"C5 - - C5 - - Ab4 - C5 - - Eb5 - C5 - - |",
		"Db5 - - C5 - - Ab4 - F4 - - Ab4 - - - |",
		"Bb4 - - Db5 - - F5 - Eb5 - - Db5 - C5 - |",
		"E5/F5 - - - C5 - - - G4 - - Bb4 - E5 - |",
	}
	lead := synth.Phrase{At: ssDrop + 8*ssBar, Step: ss16, Gate: 0.9, Glide: 0.07}.Notes(
		strings.Join(melody, " "), strings.Join(melody, " "),
	)

	boomPatch := patches.EightOhEight()
	boomPatch.Sweep = synth.Sweep{Semis: 24, Time: 0.12}
	boomPatch.Amp.Decay = 1.4

	leadPatch := patches.SuperSaw()
	leadPatch.Unison.Voices = 5
	leadPatch.LFOs = []synth.LFO{{Target: synth.Vibrato, Rate: 6, Depth: 30, Delay: 0.15}}

	hits := synth.Onsets(kick)
	duck := func(depth float64) synth.Sidechain {
		return synth.Sidechain{Hits: hits, Depth: depth, Release: 0.22}
	}
	hall := func(l float64) synth.Send { return synth.Send{Bus: "hall", Level: l} }
	echo := func(l float64) synth.Send { return synth.Send{Bus: "echo", Level: l} }
	drum := func(name string, p synth.Patch, notes []synth.Note, pan, gain float64, sends ...synth.Send) synth.Part {
		return synth.Part{Name: name, Patch: p, Notes: notes, Pan: pan, Gain: gain, Out: "drums", Sends: sends}
	}
	stopped := func(p synth.Part) synth.Part {
		p.Inserts = append(p.Inserts, tape)
		return p
	}

	return synth.Song{
		ID:   "smokesignal",
		Name: "Smoke Signal",
		Parts: []synth.Part{
			stopped(drum("kick", patches.TrapKick(), kick, 0, 0.86)),
			drum("snare", patches.TrapSnare(), snare, 0, 1.4, hall(0.3)),
			drum("clap", patches.Clap(), clap, 0, 4.2, hall(0.3)),
			// The roll skips the drum bus, whose compressor would undo its crescendo.
			stopped(synth.Part{Name: "snare roll", Patch: patches.TrapSnare(), Notes: snareRoll, Gain: 0.4, Sends: []synth.Send{hall(0.2)}}),
			stopped(drum("hats", patches.HiHat(), hats, 0.2, 0.77)),
			drum("open hat", patches.OpenHat(), openHat, -0.2, 0.7),
			drum("crash", patches.Cymbal(2.5), crash, -0.1, 1.2, hall(0.2)),
			{Name: "808", Patch: patches.EightOhEight(), Notes: sub, Gain: 0.9},
			{Name: "boom", Patch: boomPatch, Notes: boom, Gain: 0.93, Sends: []synth.Send{hall(0.2)}},
			stopped(synth.Part{
				Name: "bells", Patch: patches.Bell(), Notes: bells, Pan: 0.15, Gain: 0.5,
				Sends: []synth.Send{echo(0.3), hall(0.25)},
			}),
			stopped(synth.Part{
				Name: "pad", Patch: patches.SawPad(), Notes: pad,
				Inserts: []synth.Effect{duck(0.6)},
				Sends:   []synth.Send{hall(0.3)},
			}),
			stopped(synth.Part{
				Name: "riser", Patch: patches.Riser(ssStop - ssBuild), Notes: riser, Gain: 2.1,
				Sends: []synth.Send{hall(0.3)},
			}),
			{Name: "air horn", Patch: patches.AirHorn(), Notes: horns, Gain: 0.9, Sends: []synth.Send{hall(0.35)}},
			{
				Name: "brass", Patch: patches.Brass(), Notes: brass, Gain: 1.3,
				Inserts: []synth.Effect{duck(0.35)},
				Sends:   []synth.Send{hall(0.2)},
			},
			{
				Name: "lead", Patch: leadPatch, Notes: lead, Gain: 2.5,
				Inserts: []synth.Effect{duck(0.3)},
				Sends:   []synth.Send{echo(0.25), hall(0.25)},
			},
		},
		Buses: []synth.Bus{
			{Name: "drums", Effects: []synth.Effect{
				synth.Compressor{Threshold: -14, Ratio: 3, Attack: 0.005, Release: 0.08, Makeup: 2},
				synth.Drive{Amount: 1.5, Mix: 0.3},
			}},
			{Name: "hall", Effects: []synth.Effect{synth.Reverb{Size: 0.85, Damp: 0.4, PreDelay: 0.02}}},
			{Name: "echo", Effects: []synth.Effect{
				synth.Delay{Time: 0.75 * ssBeat, Feedback: 0.4, Damp: 3500, PingPong: true},
			}},
		},
		Master: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowCut, Freq: 25},
				{Kind: synth.HighShelf, Freq: 9000, Gain: 1},
			}},
			synth.Compressor{Threshold: -12, Ratio: 2, Attack: 0.01, Release: 0.15},
			synth.Limiter{Ceiling: -0.8},
		},
		Length:    ssEnd,
		Loop:      true,
		LoopStart: ssBuild,
		Meter:     synth.Meter{Beat: ssBeat},
	}
}
