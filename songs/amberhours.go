package songs

import (
	"strings"

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
const (
	ahBeat  = 60.0 / 168
	ahStep  = ahBeat / 2
	ahCycle = 13 * ahBeat
)

var ahForm = form{"AAAABBAA", 26}

var (
	ahLeft = map[byte]string{
		'A': "D2 - F3,A3 . F3,A3 . | Bb1 - F3,Bb3 . F3,Bb3 . | C2 - E3,G3 . E3,G3 . | A1 - E3,G3 . G2 F2 E2 C#2 |",
		'B': "Bb1 - F3,Bb3 . F3,Bb3 . | A1 - F3,A3 . F3,A3 . | G1 - G3,Bb3 . G3,Bb3 . | A1 - E3,G3 . G2 F2 E2 C#2 |",
	}
	ahRight = map[byte]string{
		'A': ". A4 D5 - E5 F5 | - D5 . Bb4 C5 D5 | E5 - C5 . G4 Bb4 | A4 - . E4 A4 C#5 E5 C#5 |",
		'B': "D4 F4 Bb4 F4 D4 F4 | C4 F4 A4 F4 C4 F4 | D4 G4 Bb4 G4 D4 G4 | C#4 E4 A4 E4 C#4 E4 G4 E4 |",
	}
	ahMelody = map[byte]string{
		'A': ahRight['A'],
		'B': "D5 - - C5 Bb4 C5 | A4 - - - F4 A4 | Bb4 - D5 - G5 F5 | E5 - - C#5 - E5 D5 C#5 |",
	}
	ahBass = map[byte]string{
		'A': "D2 - - A1 - - | Bb1 - - F2 - - | C2 - - G1 - - | A1 - - E2 G1 F1 E1 C#2 |",
		'B': "Bb1 - - F2 - - | A1 - - F2 - - | G1 - - D2 - - | A1 - - E2 G1 F1 E1 C#2 |",
	}
	ahOrgan = map[byte]string{
		'A': "D3,F3,A3,D4 - - - - - | D3,F3,Bb3,D4 - - - - - | C3,E3,G3,C4 - - - - - | C#3,E3,A3,E4 - - - - - - - |",
		'B': "D3,F3,Bb3,D4 - - - - - | C3,F3,A3,C4 - - - - - | D3,G3,Bb3,D4 - - - - - | C#3,E3,A3,E4 - - - - - - - |",
	}
	ahPedal = map[byte]string{
		'A': "D2 - - - - - | Bb1 - - - - - | C2 - - - - - | A1 - - - - - - - |",
		'B': "Bb1 - - - - - | A1 - - - - - | G1 - - - - - | A1 - - - - - - - |",
	}
	// Two notes held against the harmony as it moves: A and E are the ninth over Dm,
	// the major seventh and sharp eleventh over Bb.
	ahOstinato = map[byte]string{
		'A': strings.Repeat("A4 E5 ", 13),
		'B': strings.Repeat("D5 A5 ", 9) + strings.Repeat("C#5 A5 ", 4),
	}
)

// The last pass of each section ends on a tom fill instead of a snare.
func ahFill(pass int) bool { return pass == 5 || pass == len(ahForm.sections)-1 }

func AmberHours() synth.Song {
	take := func(first int, seed uint64) []synth.Note {
		return synth.Merge(
			synth.Phrase{Step: ahStep, Gate: 0.95, Humanize: 0.005, Seed: seed}.Notes(ahForm.plays(ahLeft, from(first))),
			synth.Phrase{Step: ahStep, Gate: 0.8, Strum: 0.005, Humanize: 0.005, Seed: seed + 1}.Notes(ahForm.plays(ahRight, from(first))),
		)
	}

	melody := synth.Phrase{Step: ahStep, Gate: 0.92, Humanize: 0.006, Seed: 11}.Notes(ahForm.score(func(pass int, s byte) string {
		if s == 'B' || pass == 6 {
			return ahMelody[s]
		}
		return ""
	}))
	bass := synth.Phrase{Step: ahStep, Gate: 0.9, Humanize: 0.004, Seed: 13}.Notes(ahForm.plays(ahBass, from(2)))

	held := synth.Phrase{Step: ahStep, Gate: 0.98}
	organ := held.Notes(ahForm.plays(ahOrgan, from(4)))
	pedal := held.Notes(ahForm.plays(ahPedal, func(pass int) bool { return pass == 1 || pass >= 4 }))
	ostinato := synth.Phrase{Step: ahStep, Gate: 0.85, Vel: 0.7}.Notes(ahForm.plays(ahOstinato, from(4)))

	// The swell box opens across B and stays open for the return of the hook.
	swell := synth.AutoFilter{Points: []synth.Point{
		{At: 4 * ahCycle, Value: 500},
		{At: 6 * ahCycle, Value: 7000},
	}}

	tick := synth.Phrase{Step: ahStep}.Notes(ahForm.score(func(int, byte) string {
		return synth.Pattern("C5", "X . x . x . | X . x . x . | X . x . x . | X . x . x . x . |")
	}))

	kit := synth.Phrase{Step: ahStep, Humanize: 0.004, Seed: 17}
	drums := func(sound string, rhythm func(pass int) string) []synth.Note {
		return kit.Notes(ahForm.score(func(pass int, _ byte) string {
			if pass < 2 {
				return ""
			}
			return synth.Pattern(sound, rhythm(pass))
		}))
	}
	kick := drums("C2", func(int) string {
		return strings.Repeat("x . . . . . | ", 3) + "x . . . x . . . |"
	})
	snare := drums("D2", func(pass int) string {
		last := ". . o . x . . o |"
		if ahFill(pass) {
			last = ". . o . . . . . |"
		}
		return strings.Repeat(". . o . x . | ", 3) + last
	})
	hats := drums("F#2", func(int) string {
		return strings.Repeat("x o x o x o | ", 3) + "x o x o x o x o |"
	})
	toms := kit.Notes(ahForm.score(func(pass int, _ byte) string {
		if !ahFill(pass) {
			return ""
		}
		return silence(22) + "D3 A2 F2 D2!"
	}))

	snarePatch := patches.Snare()
	snarePatch.Amp.Decay = 0.12
	snarePatch.Filter.Cutoff = 1800

	room := func(l float64) synth.Send { return synth.Send{Bus: "room", Level: l} }
	echo := func(l float64) synth.Send { return synth.Send{Bus: "echo", Level: l} }
	nave := func(l float64) synth.Send { return synth.Send{Bus: "nave", Level: l} }
	drum := func(name string, p synth.Patch, notes []synth.Note, pan, gain float64) synth.Part {
		return synth.Part{Name: name, Patch: p, Notes: notes, Pan: pan, Gain: gain, Out: "drums", Sends: []synth.Send{room(0.15)}}
	}

	return synth.Song{
		ID:   "amberhours",
		Name: "Amber Hours",
		Parts: []synth.Part{
			{Name: "harpsichord", Patch: patches.Harpsichord(), Notes: take(0, 1), Pan: -0.3, Out: "harpsichords", Sends: []synth.Send{room(0.2)}},
			{Name: "harpsichord double", Patch: patches.Harpsichord(), Notes: take(1, 3), Pan: 0.4, Gain: 0.8, Out: "harpsichords", Sends: []synth.Send{room(0.2)}},
			{Name: "recorder", Patch: patches.Recorder(), Notes: melody, Pan: 0.1, Gain: 0.8, Sends: []synth.Send{echo(0.2), room(0.3)}},
			{
				Name: "bass", Patch: patches.HofnerBass(), Notes: bass, Gain: 0.42,
				Inserts: []synth.Effect{synth.Compressor{Threshold: -20, Ratio: 4, Attack: 0.01, Makeup: 4}},
				Sends:   []synth.Send{room(0.04)},
			},
			{Name: "organ", Patch: patches.PipeOrgan(), Notes: organ, Gain: 0.6, Inserts: []synth.Effect{swell}, Sends: []synth.Send{nave(0.6)}},
			{Name: "pedal", Patch: patches.PipeOrgan(), Notes: pedal, Gain: 0.85, Sends: []synth.Send{nave(0.5)}},
			{Name: "ostinato", Patch: patches.PipeOrgan(), Notes: ostinato, Pan: -0.2, Gain: 0.75, Inserts: []synth.Effect{swell}, Sends: []synth.Send{nave(0.7)}},
			{Name: "clock", Patch: patches.Tick(), Notes: tick, Pan: 0.5, Gain: 1.4, Sends: []synth.Send{room(0.1)}},
			drum("kick", patches.Kick(), kick, 0, 0.25),
			drum("snare", snarePatch, snare, 0.1, 1.1),
			drum("hats", patches.HiHat(), hats, 0.3, 0.75),
			drum("toms", patches.Tom(), toms, -0.15, 0.22),
		},
		Buses: []synth.Bus{
			{Name: "harpsichords", Effects: harpsichordDesk()},
			{Name: "drums", Effects: []synth.Effect{
				synth.EQ{Bands: []synth.Band{{Kind: synth.HighCut, Freq: 8000}}},
				synth.Compressor{Threshold: -20, Ratio: 3, Attack: 0.006, Release: 0.12, Makeup: 3},
			}},
			{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.55, Damp: 0.5, PreDelay: 0.012, Width: 0.9}}},
			{Name: "nave", Effects: []synth.Effect{synth.Reverb{Size: 0.97, Damp: 0.3, PreDelay: 0.045}}},
			{Name: "echo", Effects: []synth.Effect{
				synth.Delay{Time: 3 * ahStep, Feedback: 0.3, Damp: 2500, PingPong: true},
			}},
		},
		Master: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowCut, Freq: 28},
				{Kind: synth.HighShelf, Freq: 11000, Gain: -1.5},
			}},
			synth.Compressor{Threshold: -14, Ratio: 2, Attack: 0.015, Release: 0.2, Makeup: 1},
			synth.Limiter{},
		},
		Length:    float64(len(ahForm.sections)) * ahCycle,
		Loop:      true,
		LoopStart: 2 * ahCycle,
		Meter:     synth.Meter{Beat: ahBeat, Bars: []int{3, 3, 3, 4}},
	}
}
