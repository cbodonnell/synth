package songs

import (
	"strings"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Lava Lamp Sun": a four-piece and a keyboard player, somewhere between Electric
// Ladyland and Currents. E minor, a swung 98.
//
// Four bars of intro (organ and arp over a building kit), then a sixteen-bar loop:
// A is a unison riff, bass and two guitars on the minor pentatonic with the blue
// fifth; B lifts to phased major-seventh chords under a wah-fuzz solo and lands on
// E7#9 to turn back into the riff. Tokens are eighths, except the arp's sixteenths.
const (
	llBeat  = 60.0 / 98
	llStep  = llBeat / 2
	llBar   = 4 * llBeat
	llSwing = 0.18
)

const (
	em9    = "E3,G3,B3,D4,F#4"
	cmaj7  = "C3,G3,B3,E4"
	b7     = "B2,D#3,A3,F#4"
	bm7    = "B2,F#3,A3,D4"
	am7    = "A2,E3,G3,C4"
	d6     = "D3,A3,B3,F#4"
	e7sh9  = "E3,G#3,D4,G4"
	strum  = "X . x x . x . x |"
	strum2 = "X . x x"
)

func rest(bars int) string { return strings.Repeat(". . . . . . . . | ", bars) }

// arp runs five tones up and back down twice: one bar of sixteenths.
func arp(a, b, c, d, e string) string {
	half := strings.Join([]string{a, b, c, d, e, d, c, b}, " ")
	return half + " " + half + " |"
}

func arpRest(bars int) string { return strings.Repeat(". ", 16*bars) }

func eighths(seed uint64) synth.Phrase {
	return synth.Phrase{Step: llStep, Swing: llSwing, Humanize: 0.006, Seed: seed}
}

func LavaLamp() synth.Song {
	bass := eighths(10)
	bass.Gate = 0.9
	bassNotes := bass.Notes(
		rest(2),
		"C2 - - - - - G1 - |",
		"B1 - - - A1 - F#1 - |",

		"E1 - E2 E1 G1 . A1 Bb1 | B1 - . D2 . E2 D2 B1 |",
		"E1 - E2 E1 G1 . A1 Bb1 | B1 - . D2 . E2 D2 B1 |",
		"A1 - A2 A1 C2 . D2 Eb2 | E2 - . G2 . A2 G2 E2 |",
		"E1 - E2 E1 G1 . A1 Bb1 | B1 - . D2 . E2 D2 B1 |",

		"C2 - . C2 G1 - B1 C2 |",
		"B1 - . B1 F#1 - A1 B1 |",
		"A1 - . A1 E2 - G2 E2 |",
		"D2 - . D2 A1 - D2 C#2 |",
		"C2 - . C2 G1 - B1 C2 |",
		"B1 - . B1 F#1 - A1 B1 |",
		"A1 - E2 A1 D2 - A1 D2 |",
		"E1 - - - . . B1 D2 |",
	)

	riff := []string{
		"E2,B2,E3 - E3 E2 G2 . A2 Bb2 | B2,F#3,B3 - . D3 . E3 D3 B2 |",
		"E2,B2,E3 - E3 E2 G2 . A2 Bb2 | B2,F#3,B3 - . D3 . E3 D3 B2 |",
		"A2,E3,A3 - A3 A2 C3 . D3 Eb3 | E3,B3,E4 - . G3 . A3 G3 E3 |",
		"E2,B2,E3 - E3 E2 G2 . A2 Bb2 | B2,F#3,B3 - . D3 . E3 D3 B2 |",
	}
	chords := []string{
		synth.Pattern(cmaj7, strum), synth.Pattern(bm7, strum),
		synth.Pattern(am7, strum), synth.Pattern(d6, strum),
		synth.Pattern(cmaj7, strum), synth.Pattern(bm7, strum),
		synth.Pattern(am7, strum2) + synth.Pattern(d6, strum2) + " |",
		synth.Pattern(e7sh9, "X - - - . . . . |"),
	}
	guitar := func(seed uint64) []synth.Note {
		p := eighths(seed)
		p.Gate, p.Strum = 0.85, 0.011
		lines := append([]string{rest(3), synth.Pattern(b7, "X - - - - - . . |")}, riff...)
		return p.Notes(append(lines, chords...)...)
	}

	lead := eighths(30)
	lead.Glide = 0.11
	leadNotes := lead.Notes(
		rest(4),
		rest(6),
		". . . . E5 G5 A5/B5 - |",
		"- G5 E5 D5 E5 - . . |",

		"G5 - - E5 G5/A5 - - G5 |",
		"F#5 - D5 - B4 - D5/E5 - |",
		"E5 - - - . C5 D5 E5 |",
		"F#5/G5 - - - F#5 - D5 - |",
		"B5/C6 - - - B5 G5 E5 G5 |",
		"A5 - F#5 - D5 - E5/F#5 - |",
		"E5 - D5 - C5 - A4 - |",
		"G4/A4 - - - G4 E4 . . |",
	)

	organ := synth.Phrase{Step: llStep, Gate: 0.97, Vel: 0.7}.Notes(
		em9+" - - - - - - - | - - - - - - - - |",
		"C3,E3,G3,B3,E4 - - - - - - - |",
		"B2,D#3,F#3,A3 - - - - - . . |",
		rest(8),
		"C3,E3,G3,B3 - - - - - - - |",
		"B2,D3,F#3,A3 - - - - - - - |",
		"A2,C3,E3,G3 - - - - - - - |",
		"D3,F#3,A3,B3 - - - - - - - |",
		"C3,E3,G3,B3 - - - - - - - |",
		"B2,D3,F#3,A3 - - - - - - - |",
		"A2,C3,E3,G3 - - - D3,F#3,A3,B3 - - - |",
		e7sh9+" - - - . . . . |",
	)

	arpNotes := synth.Phrase{Step: llStep / 2, Swing: 0.12, Gate: 0.7, Vel: 0.75}.Notes(
		arp("E4", "G4", "B4", "D5", "F#5"),
		arp("E4", "G4", "B4", "D5", "F#5"),
		arp("C4", "E4", "G4", "B4", "E5"),
		arp("B3", "D#4", "F#4", "A4", "B4"),
		arpRest(8),
		arp("C4", "E4", "G4", "B4", "E5"),
		arp("B3", "D4", "F#4", "A4", "D5"),
		arp("A3", "C4", "E4", "G4", "C5"),
		arp("D4", "F#4", "A4", "B4", "D5"),
		arp("C4", "E4", "G4", "B4", "E5"),
		arp("B3", "D4", "F#4", "A4", "D5"),
		arp("A3", "C4", "E4", "G4", "C5"),
		arpRest(1),
	)

	kit := eighths(40)
	kit.Humanize = 0.003
	hit := func(sound string, bars ...string) []synth.Note {
		for i, b := range bars {
			bars[i] = synth.Pattern(sound, b)
		}
		return kit.Notes(bars...)
	}
	stomp := "X . . x x . . . | X . . x x . x . |"
	open := "X . . . x . . . | X . . x . . x . |"
	backbeat := ". . X . . . X . |"
	ghosted := ". o X . . o X . |"
	hats := "x o x o x o x o |"

	kick := hit("C2",
		rest(2), "x . . . x . . . |", "x . . . . . . . |",
		stomp, stomp, stomp, stomp,
		open, open, open, "X . . . x . . . |", "X . . . . . . . |",
	)
	snare := hit("D2",
		rest(3), ". . X . . . . . |",
		strings.Repeat(backbeat, 7), ". . X . X x X X |",
		strings.Repeat(ghosted, 7), rest(1),
	)
	hihat := hit("F#2",
		rest(2), hats, hats,
		strings.Repeat(hats, 8),
	)
	ride := hit("D#3",
		rest(12),
		strings.Repeat("X x x x X x x x |", 7), "X . . . . . . . |",
	)
	crash := hit("C#3",
		rest(4), "X . . . . . . . |", rest(7), "X . . . . . . . |",
	)
	toms := kit.Notes(
		rest(3), ". . . . A2 A2 E2 B1! |",
		rest(15), ". . . . A2 E2 B1 B1! |",
	)

	snarePatch := patches.Snare()
	snarePatch.Amp.Decay = 0.24
	kickPatch := patches.Kick()
	kickPatch.Oscs[0].Fixed = 55

	room := func(l float64) synth.Send { return synth.Send{Bus: "room", Level: l} }
	plate := func(l float64) synth.Send { return synth.Send{Bus: "plate", Level: l} }
	echo := func(l float64) synth.Send { return synth.Send{Bus: "echo", Level: l} }

	drum := func(name string, p synth.Patch, notes []synth.Note, pan, gain float64, sends ...synth.Send) synth.Part {
		return synth.Part{Name: name, Patch: p, Notes: notes, Pan: pan, Gain: gain, Out: "drums", Sends: sends}
	}

	rhythm := func(name string, pan, rate float64, seed uint64) synth.Part {
		return synth.Part{
			Name: name, Patch: patches.RhythmGuitar(), Notes: guitar(seed), Pan: pan, Gain: 0.29,
			Inserts: append(patches.GuitarAmp(6), synth.Phaser{Rate: rate, Depth: 1.3, Feedback: 0.45}),
			Sends:   []synth.Send{room(0.12)},
		}
	}

	return synth.Song{
		ID:   "lavalamp",
		Name: "Lava Lamp Sun",
		Parts: []synth.Part{
			drum("kick", kickPatch, kick, 0, 0.66, room(0.04)),
			drum("snare", snarePatch, snare, 0.05, 1.3, room(0.2), plate(0.08)),
			drum("hihat", patches.HiHat(), hihat, 0.3, 0.4),
			drum("ride", patches.Cymbal(0.9), ride, -0.3, 0.48, room(0.1)),
			drum("crash", patches.Cymbal(2.2), crash, -0.2, 0.65, room(0.2)),
			drum("toms", patches.Tom(), toms, 0.1, 0.25, room(0.2)),
			{
				Name: "bass", Patch: patches.HofnerBass(), Notes: bassNotes, Gain: 0.63,
				Inserts: []synth.Effect{synth.Compressor{Threshold: -20, Ratio: 4, Attack: 0.01, Makeup: 5}},
				Sends:   []synth.Send{room(0.03)},
			},
			rhythm("guitar left", -0.75, 0.23, 20),
			rhythm("guitar right", 0.75, 0.31, 21),
			{
				Name: "lead guitar", Patch: patches.LeadGuitar(), Notes: leadNotes, Pan: 0.1, Gain: 0.25,
				Inserts: patches.GuitarAmp(2),
				Sends:   []synth.Send{echo(0.28), plate(0.22)},
			},
			{
				Name: "organ", Patch: patches.Organ(), Notes: organ, Gain: 0.31,
				Sends: []synth.Send{plate(0.3)},
			},
			{
				Name: "arp", Patch: patches.SynthArp(), Notes: arpNotes, Pan: -0.2, Gain: 1.75,
				Inserts: []synth.Effect{synth.Phaser{Rate: 0.15, Depth: 1.8, Feedback: 0.6}},
				Sends:   []synth.Send{echo(0.35), plate(0.2)},
			},
		},
		Buses: []synth.Bus{
			// Squashed and saturated as a kit, the way a 70s record or Kevin Parker's drums are.
			{Name: "drums", Effects: []synth.Effect{
				synth.EQ{Bands: []synth.Band{
					{Kind: synth.LowShelf, Freq: 90, Gain: 3},
					{Kind: synth.Peak, Freq: 450, Gain: -4, Q: 0.9},
					{Kind: synth.HighCut, Freq: 9000},
				}},
				synth.Compressor{Threshold: -22, Ratio: 5, Attack: 0.004, Release: 0.09, Makeup: 6},
				synth.Drive{Amount: 2, Tone: 8000, Mix: 0.35},
			}, Gain: 0.6},
			{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.45, Damp: 0.6, PreDelay: 0.008}}},
			{Name: "plate", Effects: []synth.Effect{synth.Reverb{Size: 0.78, Damp: 0.3, PreDelay: 0.025}}},
			{Name: "echo", Effects: []synth.Effect{
				synth.Delay{Time: 0.75 * llBeat, Feedback: 0.38, Damp: 2800, PingPong: true},
				synth.Reverb{Size: 0.5, Damp: 0.5, Mix: 0.3},
			}},
		},
		Master: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowCut, Freq: 30},
				{Kind: synth.HighShelf, Freq: 10000, Gain: -1.5},
			}},
			synth.Drive{Amount: 0.6, Mix: 0.25},
			synth.Compressor{Threshold: -14, Ratio: 2.5, Attack: 0.02, Release: 0.25, Makeup: 0.8},
			synth.Limiter{},
		},
		Length:    20 * llBar,
		Loop:      true,
		LoopStart: 4 * llBar,
		Meter:     synth.Meter{Beat: llBeat},
		Target:    -14,
	}
}
