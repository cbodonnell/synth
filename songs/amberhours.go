package songs

import (
	"strings"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Amber Hours": a harpsichord waltz that keeps tripping over an extra bar, in the
// spirit of Golden Brown. D minor at 168.
//
// The phrase is fourteen beats, three bars of three and one of five. A is the hook,
// Dm Bb C A, its five-beat bar walking the bass down to C# to fall back into the tonic.
// B moves to the relative major, Bb F/A Gm A, with a recorder singing over broken
// chords. Two passes of A on harpsichord alone make the intro; the loop is A A B B A A
// with bass and a brushed kit, the recorder taking up the hook on the way home.
// Each token is an eighth.
const (
	ahBeat  = 60.0 / 168
	ahStep  = ahBeat / 2
	ahCycle = 14 * ahBeat
	ahPass  = 28
	ahForm  = "AAAABBAA"
)

var (
	ahLeft = map[byte]string{
		'A': "D2 - F3,A3 . F3,A3 . | Bb1 - F3,Bb3 . F3,Bb3 . | C2 - E3,G3 . E3,G3 . | A1 - E3,G3 . G2 - F2 - E2 C#2 |",
		'B': "Bb1 - F3,Bb3 . F3,Bb3 . | A1 - F3,A3 . F3,A3 . | G1 - G3,Bb3 . G3,Bb3 . | A1 - E3,G3 . G2 - F2 - E2 C#2 |",
	}
	ahRight = map[byte]string{
		'A': ". A4 D5 - E5 F5 | - D5 . Bb4 C5 D5 | E5 - C5 . G4 Bb4 | A4 - . E4 A4 C#5 E5 - D5 C#5 |",
		'B': "D4 F4 Bb4 F4 D4 F4 | C4 F4 A4 F4 C4 F4 | D4 G4 Bb4 G4 D4 G4 | C#4 E4 A4 E4 C#4 E4 G4 E4 C#4 E4 |",
	}
	ahMelody = map[byte]string{
		'A': ahRight['A'],
		'B': "D5 - - C5 Bb4 C5 | A4 - - - F4 A4 | Bb4 - D5 - G5 F5 | E5 - - - C#5 - E5 D5 C#5 - |",
	}
	ahBass = map[byte]string{
		'A': "D2 - - A1 - - | Bb1 - - F2 - - | C2 - - G1 - - | A1 - - E2 G1 - F1 - E1 C#2 |",
		'B': "Bb1 - - F2 - - | A1 - - F2 - - | G1 - - D2 - - | A1 - - E2 G1 - F1 - E1 C#2 |",
	}
)

// ahScore writes a line through the form, one pass at a time; play returns "" to rest.
func ahScore(play func(pass int, section byte) string) string {
	var b strings.Builder
	for pass := range len(ahForm) {
		bar := play(pass, ahForm[pass])
		if bar == "" {
			bar = silence(ahPass)
		}
		b.WriteString(bar + " ")
	}
	return b.String()
}

// The last pass of each section ends on a tom fill instead of a snare.
func ahFill(pass int) bool { return pass == 5 || pass == len(ahForm)-1 }

func AmberHours() synth.Song {
	from := func(first int, line map[byte]string) func(int, byte) string {
		return func(pass int, s byte) string {
			if pass < first {
				return ""
			}
			return line[s]
		}
	}
	take := func(first int, seed uint64) []synth.Note {
		return synth.Merge(
			synth.Phrase{Step: ahStep, Gate: 0.95, Humanize: 0.005, Seed: seed}.Notes(ahScore(from(first, ahLeft))),
			synth.Phrase{Step: ahStep, Gate: 0.8, Strum: 0.005, Humanize: 0.005, Seed: seed + 1}.Notes(ahScore(from(first, ahRight))),
		)
	}

	melody := synth.Phrase{Step: ahStep, Gate: 0.92, Humanize: 0.006, Seed: 11}.Notes(ahScore(func(pass int, s byte) string {
		if s == 'B' || pass == 6 {
			return ahMelody[s]
		}
		return ""
	}))
	bass := synth.Phrase{Step: ahStep, Gate: 0.9, Humanize: 0.004, Seed: 13}.Notes(ahScore(from(2, ahBass)))

	kit := synth.Phrase{Step: ahStep, Humanize: 0.004, Seed: 17}
	drums := func(sound string, rhythm func(pass int) string) []synth.Note {
		return kit.Notes(ahScore(func(pass int, _ byte) string {
			if pass < 2 {
				return ""
			}
			return synth.Pattern(sound, rhythm(pass))
		}))
	}
	kick := drums("C2", func(int) string {
		return strings.Repeat("x . . . . . | ", 3) + "x . . . x . . . . . |"
	})
	snare := drums("D2", func(pass int) string {
		last := ". . o . x . . . o . |"
		if ahFill(pass) {
			last = ". . o . x . . . . . |"
		}
		return strings.Repeat(". . o . x . | ", 3) + last
	})
	hats := drums("F#2", func(int) string {
		return strings.Repeat("x o x o x o | ", 3) + "x o x o x o x o x o |"
	})
	toms := kit.Notes(ahScore(func(pass int, _ byte) string {
		if !ahFill(pass) {
			return ""
		}
		return silence(24) + "D3 A2 F2 D2!"
	}))

	snarePatch := patches.Snare()
	snarePatch.Amp.Decay = 0.12
	snarePatch.Filter.Cutoff = 1800

	room := func(l float64) synth.Send { return synth.Send{Bus: "room", Level: l} }
	echo := func(l float64) synth.Send { return synth.Send{Bus: "echo", Level: l} }
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
			{Name: "echo", Effects: []synth.Effect{
				synth.Delay{Time: 3 * ahStep, Feedback: 0.3, Damp: 2500, PingPong: true},
			}},
		},
		Master: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowCut, Freq: 35},
				{Kind: synth.HighShelf, Freq: 11000, Gain: -1.5},
			}},
			synth.Compressor{Threshold: -14, Ratio: 2, Attack: 0.015, Release: 0.2, Makeup: 1},
			synth.Limiter{},
		},
		Length:    float64(len(ahForm)) * ahCycle,
		Loop:      true,
		LoopStart: 2 * ahCycle,
		Meter:     synth.Meter{Beat: ahBeat, Bars: []int{3, 3, 3, 5}},
	}
}
