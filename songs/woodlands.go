package songs

import (
	"fmt"
	"strings"

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
const (
	wlBeat = 60.0 / 116
	wlStep = wlBeat / 2
	wlBar  = 4 * wlBeat
	wlPass = 4 * wlBar
)

var (
	wlForm  = form{"IIAABBCCBB", 32}
	wlForm6 = form{wlForm.sections, 64}
)

var wlChords = map[byte][4]string{
	'I': {"G", "Em", "C", "D"},
	'A': {"G", "Em", "C", "D"},
	'B': {"C", "G", "Em", "D"},
	'C': {"Em", "C", "G", "D"},
}

var (
	wlGuitar = map[string]string{
		"G":  "G2,B2,D3,G3,B3,G4",
		"Em": "E2,B2,E3,G3,B3,E4",
		"C":  "C3,E3,G3,C4,E4",
		"D":  "D3,A3,D4,F#4",
	}
	// Forward rolls, thumb, index, middle; the second half of each bar puts the thumb
	// on the short fifth string's high G, the drone every banjo tune leans on.
	wlRoll = map[string][3]string{
		"G":  {"G3", "B3", "D4"},
		"Em": {"G3", "B3", "E4"},
		"C":  {"G3", "C4", "E4"},
		"D":  {"F#3", "A3", "D4"},
	}
	wlRoot = map[string][2]string{
		"G":  {"G1", "D2"},
		"Em": {"E1", "B1"},
		"C":  {"C2", "G1"},
		"D":  {"D2", "A1"},
	}
	wlArp = map[string][5]string{
		"G":  {"G4", "B4", "D5", "F#5", "A5"},
		"Em": {"E4", "G4", "B4", "D5", "F#5"},
		"C":  {"C4", "E4", "G4", "B4", "D5"},
		"D":  {"D4", "F#4", "A4", "C#5", "E5"},
	}
	wlPad = map[string]string{
		"G":  "G3,B3,D4,F#4",
		"Em": "E3,G3,B3,D4",
		"C":  "C3,G3,B3,E4",
		"D":  "D3,F#3,A3,E4",
	}
)

var (
	wlWhistle = []string{
		"D5 - G5 - B5 - A5 G5 | E5 - - - G5 - E5 D5 | E5 - G5 - C6 - B5 A5 | A5 - - - F#5/G5 - - . |",
		"D5 - G5 - B5 - A5 G5 | E5 - - - G5 - B5 D6 | C6 - B5 - A5 - G5 E5 | F#5 - A5 - G5 - - . |",
	}
	wlChorus = map[byte]string{
		'B': "G5 - - - E5 - G5 - | B5 - - - A5 - G5 - | B5 - - - D6 - B5 - | A5 - - - - - . . |",
	}
	wlFireflies = []string{
		". . E7 . . . . . | . . . . . B6 . . | . D7 . . . . . . | . . . . . . A6 . |",
		". . . . . G6 . . | . . D7 . . . . . | . . . . . . . E7 | . . . B6 . . . . |",
	}
)

// wlEach writes one bar for each of a section's four chords.
func wlEach(s byte, bar func(chord string) string) string {
	var b strings.Builder
	for _, c := range wlChords[s] {
		b.WriteString(bar(c) + " ")
	}
	return b.String()
}

// tremolo turns each held note into a run of picks across its length, four to an eighth.
func tremolo(line string) string {
	tokens := strings.Fields(strings.ReplaceAll(line, "|", " "))
	for i, t := range tokens {
		if t == "." || t == "-" {
			continue
		}
		n := 1
		for i+n < len(tokens) && tokens[i+n] == "-" {
			n++
		}
		tokens[i] = fmt.Sprintf("%s*%d", t, 4*n)
	}
	return strings.Join(tokens, " ")
}

func Woodlands() synth.Song {
	eighths := func(gate, humanize float64, seed uint64) synth.Phrase {
		return synth.Phrase{Step: wlStep, Gate: gate, Humanize: humanize, Seed: seed}
	}
	band := func(p int) bool { return wlForm.sections[p] == 'A' || wlForm.sections[p] == 'B' }
	night := func(p int) bool { return wlForm.sections[p] != 'A' && wlForm.sections[p] != 'B' || p >= 8 }
	build := 7

	strummer := eighths(0.95, 0.006, 1)
	strummer.Strum = 0.012
	guitar := strummer.Notes(wlForm.score(func(p int, s byte) string {
		rhythm := "X . x x . x x x"
		switch {
		case p == 0 || p == 6:
			rhythm = "X - - - - - - -"
		case p == build:
			rhythm = "X x x x X x x x"
		}
		return wlEach(s, func(c string) string { return synth.Pattern(wlGuitar[c], rhythm) + "|" })
	}))

	banjo := synth.Phrase{Step: wlStep / 2, Gate: 1, Humanize: 0.004, Seed: 3}.Notes(wlForm6.score(func(p int, s byte) string {
		if p == 0 || !band(p) && p != 1 {
			return ""
		}
		return wlEach(s, func(c string) string {
			r := wlRoll[c]
			t, i, m := r[0], r[1], r[2]
			return strings.Join([]string{t + "!", i, m, t, i, m, t, m, "G4!", i, m, "G4", i, m, "G4", m, "|"}, " ")
		})
	}))

	bass := eighths(0.85, 0.004, 5).Notes(wlForm.score(func(p int, s byte) string {
		if p < 2 {
			return ""
		}
		return wlEach(s, func(c string) string {
			r := wlRoot[c]
			switch {
			case p == 6:
				return r[0] + " - - - - - - - |"
			case p == build:
				return fmt.Sprintf("%s %s %s %s %s %s %s %s |", r[0], r[0], r[0], r[0], r[0], r[0], r[0], r[0])
			}
			return fmt.Sprintf("%s - . %s %s - . %s |", r[0], r[0], r[1], r[1])
		})
	}))

	whistler := eighths(0.92, 0.006, 7)
	whistler.Glide = 0.08
	whistle := whistler.Notes(wlForm.score(func(p int, s byte) string {
		if s != 'A' {
			return ""
		}
		return wlWhistle[p%2]
	}))

	glock := eighths(1, 0.003, 9).Notes(wlForm.plays(wlChorus, from(4)))
	mandolin := synth.Transpose(eighths(1, 0, 0).Notes(wlForm.plays(map[byte]string{'B': tremolo(wlChorus['B'])}, from(4))), -12)
	sung := eighths(0.95, 0.01, 11).Notes(wlForm.plays(wlChorus, from(4)))
	gang := synth.Merge(synth.Transpose(sung, -12), synth.Transpose(sung, -24))
	shout := eighths(1, 0.01, 13).Notes(wlForm.plays(map[byte]string{'B': silence(30) + "D4,A4! ."}, from(4)))

	arp := synth.Phrase{Step: wlStep / 2, Gate: 0.7, Vel: 0.75}.Notes(wlForm6.score(func(p int, s byte) string {
		if !night(p) || p < 2 {
			return ""
		}
		return wlEach(s, func(c string) string { return arp(wlArp[c][0], wlArp[c][1], wlArp[c][2], wlArp[c][3], wlArp[c][4]) })
	}))
	pad := eighths(1, 0, 0).Notes(wlForm.score(func(p int, s byte) string {
		if !night(p) || p < 2 {
			return ""
		}
		return wlEach(s, func(c string) string { return wlPad[c] + " - - - - - - - |" })
	}))

	fireflies := eighths(1, 0.12, 15).Notes(wlForm.score(func(p int, _ byte) string {
		if !night(p) {
			return ""
		}
		return wlFireflies[p%2]
	}))
	crickets := eighths(1, 0.05, 17).Notes(wlForm.score(func(p int, _ byte) string {
		if !night(p) || p >= 8 {
			return ""
		}
		return strings.Repeat("C4 . . . C4 . C4 . | . . C4 . . . . . | ", 2)
	}))

	kit := eighths(1, 0.004, 19)
	drums := func(sound string, rhythm func(p int, s byte) string) []synth.Note {
		return kit.Notes(wlForm.score(func(p int, s byte) string {
			r := rhythm(p, s)
			if r == "" {
				return ""
			}
			return synth.Pattern(sound, r)
		}))
	}
	stomp := drums("C2", func(p int, s byte) string {
		switch {
		case p == 1:
			return strings.Repeat("X . . . x . . . | ", 4)
		case band(p):
			return strings.Repeat("X . x . X . x . | ", 4)
		case p == build:
			return strings.Repeat("X . x . X . x . | ", 3) + "X x X x X x X x |"
		}
		return ""
	})
	clapper := kit
	clapper.Strum = 0.009
	claps := clapper.Notes(wlForm.score(func(p int, s byte) string {
		hands := "D#3,D#3,D#3"
		switch {
		case band(p) || p == 1:
			return synth.Pattern(hands, strings.Repeat(". . X . . . X . | ", 4))
		case p == build:
			return synth.Pattern(hands, strings.Repeat(". . . . . . . . | ", 2)+". . X . . . X . | X . X . X x X x |")
		}
		return ""
	}))
	tambourine := drums("C5", func(p int, s byte) string {
		if s == 'B' || p == 3 {
			return strings.Repeat("o x X x o x X x | ", 4)
		}
		return ""
	})
	roll := drums("D2", func(p int, _ byte) string {
		if p != build {
			return ""
		}
		return silence(16) + "o o x o x x x x | x*2 x*2 x*2 x*2 X*4 - X*4 - |"
	})
	crash := drums("C#3", func(p int, _ byte) string {
		if p == 2 || p == 4 || p == 8 {
			return "X " + silence(31)
		}
		return ""
	})

	riserLength := 2 * wlBar
	riser := []synth.Note{{At: float64(build)*wlPass + 2*wlBar, Dur: riserLength, Key: synth.MustKey("D3"), Slide: 12}}

	stompPatch := patches.Kick()
	stompPatch.Oscs[0].Fixed = 58
	stompPatch.Amp.Decay = 0.28
	snare := patches.Snare()
	snare.Amp.Decay = 0.12

	barn := func(l float64) synth.Send { return synth.Send{Bus: "barn", Level: l} }
	woods := func(l float64) synth.Send { return synth.Send{Bus: "woods", Level: l} }
	echo := func(l float64) synth.Send { return synth.Send{Bus: "echo", Level: l} }
	drum := func(name string, p synth.Patch, notes []synth.Note, pan, gain float64, sends ...synth.Send) synth.Part {
		return synth.Part{Name: name, Patch: p, Notes: notes, Pan: pan, Gain: gain, Out: "drums", Sends: sends}
	}

	return synth.Song{
		ID:   "woodlands",
		Name: "The Woodlands",
		Parts: []synth.Part{
			{Name: "crickets", Patch: patches.Cricket(), Notes: crickets, Pan: -0.6, Sends: []synth.Send{woods(0.5)}},
			{Name: "fireflies", Patch: patches.Firefly(), Notes: fireflies, Gain: 0.8, Sends: []synth.Send{woods(0.6), echo(0.3)}},
			{
				Name: "acoustic guitar", Patch: patches.AcousticGuitar(), Notes: guitar, Pan: -0.35,
				Inserts: []synth.Effect{synth.EQ{Bands: []synth.Band{
					{Kind: synth.LowCut, Freq: 90},
					{Kind: synth.Peak, Freq: 250, Gain: -2},
					{Kind: synth.HighShelf, Freq: 6000, Gain: 2},
				}}},
				Sends: []synth.Send{barn(0.15)},
			},
			{Name: "banjo", Patch: patches.Banjo(), Notes: banjo, Pan: 0.4, Gain: 1.7, Sends: []synth.Send{barn(0.12)}},
			{
				Name: "upright bass", Patch: patches.UprightBass(), Notes: bass,
				Inserts: []synth.Effect{synth.Compressor{Threshold: -20, Ratio: 3, Attack: 0.01, Makeup: 3}},
				Sends:   []synth.Send{barn(0.04)},
			},
			{Name: "whistle", Patch: patches.Whistle(), Notes: whistle, Pan: 0.1, Gain: 0.84, Sends: []synth.Send{woods(0.3), echo(0.2)}},
			{Name: "gang vocals", Patch: patches.GangVocals(), Notes: gang, Gain: 1.6, Sends: []synth.Send{barn(0.3), woods(0.2)}},
			{Name: "shout", Patch: patches.Shout(), Notes: shout, Gain: 2, Sends: []synth.Send{barn(0.35)}},
			{Name: "mandolin", Patch: patches.Mandolin(), Notes: mandolin, Pan: -0.5, Gain: 1.4, Sends: []synth.Send{barn(0.15)}},
			{Name: "glockenspiel", Patch: patches.Glockenspiel(), Notes: glock, Pan: 0.3, Gain: 0.8, Sends: []synth.Send{woods(0.3)}},
			{
				Name: "arp", Patch: patches.SynthArp(), Notes: arp, Pan: -0.2, Gain: 1.6,
				Inserts: []synth.Effect{synth.Phaser{Rate: 0.12, Depth: 1.6, Feedback: 0.5}},
				Sends:   []synth.Send{echo(0.3), woods(0.25)},
			},
			{Name: "pad", Patch: patches.SawPad(), Notes: pad, Gain: 0.63, Sends: []synth.Send{woods(0.5)}},
			{Name: "riser", Patch: patches.Riser(riserLength), Notes: riser, Sends: []synth.Send{woods(0.3)}},
			{Name: "roll", Patch: snare, Notes: roll, Pan: 0.05, Sends: []synth.Send{barn(0.2)}},
			drum("stomp", stompPatch, stomp, 0, 0.84, barn(0.1)),
			drum("claps", patches.Clap(), claps, 0, 2, barn(0.25)),
			drum("tambourine", patches.Tambourine(), tambourine, 0.45, 1.2, barn(0.1)),
			drum("crash", patches.Cymbal(2.4), crash, -0.3, 1, barn(0.15)),
		},
		Buses: []synth.Bus{
			{Name: "drums", Effects: []synth.Effect{
				synth.EQ{Bands: []synth.Band{{Kind: synth.LowShelf, Freq: 100, Gain: 2}}},
				synth.Compressor{Threshold: -18, Ratio: 3, Attack: 0.006, Release: 0.12, Makeup: 2},
			}},
			// A wooden room for the band; the woods are wider, a little darker, and longer.
			{Name: "barn", Effects: []synth.Effect{synth.Reverb{Size: 0.55, Damp: 0.5, PreDelay: 0.015}}},
			{Name: "woods", Effects: []synth.Effect{synth.Reverb{Size: 0.9, Damp: 0.55, PreDelay: 0.04}}},
			{Name: "echo", Effects: []synth.Effect{
				synth.Delay{Time: 3 * wlStep, Feedback: 0.35, Damp: 2600, PingPong: true},
			}},
		},
		Master: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowCut, Freq: 30},
				{Kind: synth.HighShelf, Freq: 10000, Gain: -1},
			}},
			synth.Compressor{Threshold: -14, Ratio: 2, Attack: 0.02, Release: 0.2, Makeup: 1},
			synth.Limiter{},
		},
		Length:    float64(len(wlForm.sections)) * wlPass,
		Loop:      true,
		LoopStart: 2 * wlPass,
		Meter:     synth.Meter{Beat: wlBeat},
		Target:    -14,
	}
}
