package songs

import (
	"strings"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Brighter Days": festival progressive house, 128 BPM.
// One progression all the way through, F#m D A E, and one hook in the 3-3-2
// sixteenth syncopation that carries a whole drop.
//
// Piano alone for four bars, then the loop: an eight-bar breakdown (piano, pad, the
// hook on a pluck), an eight-bar build (snare roll, riser, the hook passing to the lead
// as its filter opens, a half-bar of air), and a sixteen-bar drop where everything pumps off the kick.
const (
	bdBeat = 60.0 / 128
	bdBar  = 4 * bdBeat
	bd8    = bdBeat / 2
	bd16   = bdBeat / 4

	bdBreakdown = 4 * bdBar
	bdBuild     = 12 * bdBar
	bdDrop      = 20 * bdBar
	bdEnd       = 36 * bdBar
)

var (
	bdChords = []string{"A3,C#4,F#4", "A3,D4,F#4", "A3,C#4,E4", "G#3,B3,E4"}
	bdRoots  = []string{"F#2,C#3", "D2,A2", "A2,E3", "E2,B2"}
	bdBass   = []string{"F#1", "D1", "A1", "E1"}
)

// bars repeats one bar per chord of the progression, n bars in all.
func bars(n int, bar func(chord int) string) string {
	var b strings.Builder
	for i := range n {
		b.WriteString(bar(i % 4))
		b.WriteString(" | ")
	}
	return b.String()
}

func silence(steps int) string { return strings.Repeat(". ", steps) }

func BrighterDays() synth.Song {
	hook := []string{
		"F#5 - - E5 - - C#5 - E5 - - F#5 - - A5 - |",
		"F#5 - - E5 - - D5 - E5 - - F#5 - - D5 - |",
		"E5 - - C#5 - - A4 - C#5 - - E5 - - A5 - |",
		"G#5 - - F#5 - - E5 - B4 - - E5 - - G#5 - |",
	}
	lift := "G#5 - - F#5 - - E5 - F#5 - - G#5 - - B5 - |"
	gap := "G#5 - - F#5 - - E5 - . . . . . . . . |"
	eight := strings.Join(hook, " ") + strings.Join(hook[:3], " ") + lift
	sixteenths := synth.Phrase{Step: bd16, Gate: 0.9}

	pluckNotes := sixteenths.Notes(silence(4*16), eight)
	leadNotes := synth.Merge(
		synth.Phrase{At: bdBuild, Step: bd16, Gate: 0.9, Vel: 0.5}.Notes(
			strings.Join(hook, " "), strings.Join(hook[:3], " "), gap,
		),
		synth.Phrase{At: bdDrop, Step: bd16, Gate: 0.9}.Notes(eight, eight),
	)
	stabNotes := synth.Phrase{Step: bd8, Gate: 0.6}.Notes(
		silence(20*8),
		bars(16, func(c int) string { return synth.Pattern(bdChords[c], ". x . x . x . X") }),
	)

	piano := synth.Phrase{Step: bd8, Gate: 0.95, Strum: 0.008, Vel: 0.75}
	pianoNotes := synth.Merge(
		piano.Notes(
			bars(19, func(c int) string { return synth.Pattern(bdChords[c], "X - - x - - x -") }),
			synth.Pattern(bdChords[3], "X - - x . . . ."),
			bars(16, func(c int) string { return synth.Pattern(bdChords[c], "X - - x - - x -") }),
		),
		piano.Notes(
			bars(19, func(c int) string { return bdRoots[c] + " - - - - - - -" }),
			bdRoots[3]+" - - - . . . .",
			bars(16, func(c int) string { return bdRoots[c] + " - - - - - - -" }),
		),
	)

	padNotes := synth.Phrase{Step: bdBar, Gate: 0.98, Vel: 0.7}.Notes(
		". . . .",
		bars(15, func(c int) string { return bdChords[c] }),
		". |",
		bars(16, func(c int) string { return bdChords[c] }),
	)

	bassNotes := synth.Phrase{Step: bd8, Gate: 0.8}.Notes(
		silence(20*8),
		bars(16, func(c int) string {
			r := bdBass[c]
			hi := strings.Replace(r, "1", "2", 1)
			return ". " + r + " . " + hi + " . " + r + " . " + hi
		}),
	)

	// The build's kick skips the drum bus, whose compressor would only bring it back up:
	// it holds back so the drop's can land.
	buildKickNotes := synth.Phrase{At: bdBuild, Step: bdBeat}.Notes(synth.Pattern("C2", strings.Repeat("x x x x | ", 6)))
	kickNotes := synth.Phrase{At: bdDrop, Step: bdBeat}.Notes(synth.Pattern("C2", strings.Repeat("X x x x | ", 16)))
	hits := synth.Onsets(synth.Merge(buildKickNotes, kickNotes))

	eighths := synth.Phrase{Step: bd8}
	clapNotes := synth.Phrase{Step: bd8, Strum: 0.011}.Notes(
		silence(20*8),
		synth.Pattern("D#3,D#3,D#3", strings.Repeat(". . X . . . X . | ", 16)),
	)
	openHatNotes := eighths.Notes(
		silence(20*8),
		synth.Pattern("A#2", strings.Repeat(". x . x . x . x | ", 16)),
	)
	hatNotes := synth.Phrase{Step: bd16, Vel: 0.6}.Notes(
		silence(16*16),
		synth.Pattern("F#2", strings.Repeat("o o x o ", 3*4)), silence(16),
		synth.Pattern("F#2", strings.Repeat("o o x o ", 16*4)),
	)

	// The roll doubles its rate every two bars and grows louder as it goes.
	roll := func(at, step float64, vel float64, rhythm string) []synth.Note {
		return synth.Phrase{At: at, Step: step, Vel: vel}.Notes(synth.Pattern("D2", rhythm))
	}
	snareNotes := synth.Merge(
		roll(bdBuild, bdBeat, 0.4, strings.Repeat("x x x x ", 4)),
		roll(bdBuild+4*bdBar, bd8, 0.55, strings.Repeat("x ", 16)),
		roll(bdBuild+6*bdBar, bd16, 0.7, strings.Repeat("x ", 16)),
		roll(bdBuild+7*bdBar, bd16/2, 0.9, strings.Repeat("x ", 16)),
	)

	crashNotes := synth.Merge(
		synth.Phrase{At: bdBreakdown, Step: bdBar, Vel: 0.6}.Notes("C#3"),
		synth.Phrase{At: bdDrop, Step: 8 * bdBar}.Notes("C#3! C#3"),
	)

	riserLength := 7.5 * bdBar
	riserNotes := []synth.Note{{At: bdBuild, Dur: riserLength, Key: synth.MustKey("F#3"), Slide: 12}}

	boom := patches.HouseKick()
	boom.Amp.Decay = 1.8
	boom.Sweep = synth.Sweep{Semis: 36, Time: 0.08}
	boomNotes := []synth.Note{{At: bdDrop, Dur: 1.5, Key: synth.MustKey("C2")}}

	pump := func(depth, release float64) synth.Sidechain {
		return synth.Sidechain{Hits: hits, Depth: depth, Release: release}
	}
	reverb := func(l float64) synth.Send { return synth.Send{Bus: "hall", Level: l} }
	delay := func(l float64) synth.Send { return synth.Send{Bus: "echo", Level: l} }
	drum := func(name string, p synth.Patch, notes []synth.Note, pan, gain float64, sends ...synth.Send) synth.Part {
		return synth.Part{Name: name, Patch: p, Notes: notes, Pan: pan, Gain: gain, Out: "drums", Sends: sends}
	}

	return synth.Song{
		ID:   "brighterdays",
		Name: "Brighter Days",
		Parts: []synth.Part{
			drum("kick", patches.HouseKick(), kickNotes, 0, 1.25),
			{Name: "build kick", Patch: patches.HouseKick(), Notes: buildKickNotes, Gain: 0.4},
			drum("clap", patches.Clap(), clapNotes, 0, 5.5, reverb(0.25)),
			drum("snare roll", patches.Snare(), snareNotes, 0, 1.8, reverb(0.2)),
			drum("hats", patches.HiHat(), hatNotes, 0.25, 0.8),
			drum("open hat", patches.OpenHat(), openHatNotes, -0.2, 0.7),
			drum("crash", patches.Cymbal(2.5), crashNotes, -0.1, 1.4, reverb(0.2)),
			drum("boom", boom, boomNotes, 0, 0.8, reverb(0.3)),
			{
				Name: "bass", Patch: patches.OffbeatBass(), Notes: bassNotes,
				Inserts: []synth.Effect{pump(0.9, 0.16)},
			},
			{
				Name: "piano", Patch: patches.HousePiano(), Notes: pianoNotes, Gain: 1.25,
				Inserts: []synth.Effect{pump(0.45, 0.2)},
				Sends:   []synth.Send{reverb(0.3)},
			},
			{
				Name: "pad", Patch: patches.SawPad(), Notes: padNotes,
				Inserts: []synth.Effect{pump(0.75, 0.25)},
				Sends:   []synth.Send{reverb(0.35)},
			},
			{
				Name: "pluck", Patch: patches.HousePluck(), Notes: pluckNotes, Pan: 0.15, Gain: 3,
				Inserts: []synth.Effect{synth.AutoFilter{Points: []synth.Point{
					{At: bdBreakdown, Value: 1400}, {At: bdBuild, Value: 6000},
				}}},
				Sends: []synth.Send{delay(0.35), reverb(0.3)},
			},
			{
				Name: "chord stabs", Patch: patches.HousePluck(), Notes: stabNotes, Pan: -0.15, Gain: 1.4,
				Inserts: []synth.Effect{pump(0.6, 0.18)},
				Sends:   []synth.Send{reverb(0.25)},
			},
			{
				Name: "lead", Patch: patches.SuperSaw(), Notes: leadNotes, Gain: 2.8,
				Inserts: []synth.Effect{
					synth.AutoFilter{Resonance: 0.2, Points: []synth.Point{
						{At: bdBuild, Value: 180},
						{At: bdDrop - bdBar/2, Value: 2200},
						{At: bdDrop - 0.01, Value: 2200},
						{At: bdDrop, Value: 20000},
					}},
					pump(0.5, 0.2),
				},
				Sends: []synth.Send{delay(0.25), reverb(0.35)},
			},
			{
				Name: "riser", Patch: patches.Riser(riserLength), Notes: riserNotes, Gain: 1.2,
				Sends: []synth.Send{reverb(0.4)},
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
				pump(0.7, 0.25),
			}},
			{Name: "echo", Effects: []synth.Effect{
				synth.Delay{Time: 0.75 * bdBeat, Feedback: 0.4, Damp: 4500, PingPong: true},
				pump(0.6, 0.25),
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
		Length:    bdEnd,
		Loop:      true,
		LoopStart: bdBreakdown,
		Meter:     synth.Meter{Beat: bdBeat},
	}
}
