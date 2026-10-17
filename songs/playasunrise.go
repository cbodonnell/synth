package songs

import (
	"strings"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// "Playa Sunrise": Golden Brown's lilt carried out to the desert at dawn. E mixolydian
// at 148, swung a little, over a tanpura and a didgeridoo that never leave E.
//
// The phrase is thirteen beats, three bars of three and one of four. A is E D A, the
// handpan keeping Golden Brown's interlock of root, off-beat, and answer; its four-beat
// bar falls E D C#m B in parallel chords, a sitar singing tenths above the bass. B
// borrows C for C D E, the same fall with C natural, and a bansuri takes the tune.
// The didgeridoo breathes in threes against the waltz, 3+3+3+3+3+3+3+3+2 to the pass.
//
// The intro is wind, tanpura, and handpan, then the didgeridoo and a shaker; the loop
// is A A B B A A, adding djembes, bass, and sitar, then a pad that opens through B and
// wind chimes, the flute joining the sitar on the hook for the last pass.
// Each token is an eighth.
const (
	psBeat  = 60.0 / 148
	psStep  = psBeat / 2
	psCycle = 13 * psBeat
	psSwing = 0.12
)

var psForm = form{"AAAABBAA", 26}

var (
	psHandpan = map[byte]string{
		'A': "E3 G#4 . B3,B4 . E4 | D3 F#4 . A3,A4 . D4 | A2 C#4 . E3,E4 . A3 | E3 G#4,B4 D3 F#4,A4 C#3 E4,G#4 B2 D#4,F#4 |",
		'B': "C3 E4 . G3,G4 . C4 | D3 F#4 . A3,A4 . D4 | E3 G#4 . B3,B4 . E4 | E3 G#4,B4 D3 F#4,A4 C3 E4,G4 B2 D#4,F#4 |",
	}
	psBass = map[byte]string{
		'A': "E2 - - . B1 E2 | D2 - - . A1 D2 | A1 - - . E2 A1 | E2 - D2 - C#2 - B1 - |",
		'B': "C2 - - . G1 C2 | D2 - - . A1 D2 | E2 - - . B1 E2 | E2 - D2 - C2 - B1 - |",
	}
	psHook = map[byte]string{
		'A': ". B4 E5 - F#5 G#5 | A5 - F#5 . E5 D5 | E5 - C#5 . A4 B4 | F#5/G#5 - F#5 - E5 - D#5 - |",
		'B': "G5 - - E5 - G5 | A5 - - F#5 - A5 | B5 - - - A5 G#5 | G#5 - F#5 - E5 - D#5 - |",
	}
	psPad = map[byte]string{
		'A': "E3,B3,E4,G#4 - - - - - | D3,A3,D4,F#4 - - - - - | E3,A3,C#4,E4 - - - - - | E3,B3,E4,G#4 - - - B2,F#3,B3,D#4 - - - |",
		'B': "C3,G3,C4,E4 - - - - - | D3,A3,D4,F#4 - - - - - | E3,B3,E4,G#4 - - - - - | E3,B3,E4,G#4 - - - B2,F#3,B3,D#4 - - - |",
	}
	psChimes = map[byte]string{
		'A': ". . . . . . | . . B6 . . . | . . . . . . | E7 . . . . . C#7 . |",
		'B': ". . . . G#6 . | . . . . . . | . B6 . . . . | . . E7 . . . . . |",
	}
)

// Pa and three Sa, the tanpura's cycle, laid on the bars so each one starts on the fifth.
var psTanpura = strings.Repeat("B2 - E3 - E3 - | ", 3) + "B2 - E3 - E3 - E2 - |"

// The last pass of each section ends on a djembe roll instead of the slap.
func psFill(pass int) bool { return pass == 5 || pass == len(psForm.sections)-1 }

func PlayaSunrise() synth.Song {
	swung := func(gate, humanize float64, seed uint64) synth.Phrase {
		return synth.Phrase{Step: psStep, Gate: gate, Swing: psSwing, Humanize: humanize, Seed: seed}
	}
	every := func(int) bool { return true }
	each := func(line string) map[byte]string { return map[byte]string{'A': line, 'B': line} }

	handpan := swung(1, 0.005, 1).Notes(psForm.plays(psHandpan, every))
	tanpura := swung(1, 0.008, 3).Notes(psForm.plays(each(psTanpura), every))
	bass := swung(0.9, 0.004, 5).Notes(psForm.plays(psBass, from(2)))
	sitar := swung(0.95, 0.006, 7)
	sitar.Glide = 0.12
	hook := sitar.Notes(psForm.score(func(pass int, s byte) string {
		if s == 'A' && pass >= 2 {
			return psHook[s]
		}
		return ""
	}))
	flute := swung(0.92, 0.006, 9)
	flute.Glide = 0.1
	melody := flute.Notes(psForm.plays(psHook, func(pass int) bool { return pass == 4 || pass == 5 || pass == 7 }))
	pad := swung(1, 0, 0).Notes(psForm.plays(psPad, from(4)))
	chimes := swung(1, 0.03, 11).Notes(psForm.plays(psChimes, only(4, 5, 7)))
	wind := synth.Phrase{Step: psStep}.Notes(psForm.plays(each("E3"+strings.Repeat(" -", 25)), every))
	didge := swung(0.92, 0.006, 13).Notes(psForm.plays(each(
		strings.Repeat("E2! - - E2 - - | ", 3)+"E2! - - E2 - - G#3 - |",
	), from(1)))

	// The swell opens across B and stays open for the return of the hook.
	swell := synth.AutoFilter{Points: []synth.Point{
		{At: 4 * psCycle, Cutoff: 500},
		{At: 6 * psCycle, Cutoff: 6000},
	}}

	kit := swung(1, 0.004, 17)
	drums := func(first int, rhythm func(pass int) string) []synth.Note {
		return kit.Notes(psForm.score(func(pass int, _ byte) string {
			if pass < first {
				return ""
			}
			return rhythm(pass)
		}))
	}
	shaker := drums(1, func(int) string {
		return synth.Pattern("C5", strings.Repeat("x o x o x o | ", 3)+"x o x o x o x o |")
	})
	djembe := drums(2, func(pass int) string {
		last := "D2! . D4 D2 . D4 D4? D4 |"
		if psFill(pass) {
			last = "D2! D4*2 D4 F3 D3*3 D3 A2? D2! |"
		}
		return strings.Repeat("D2! . D4 D4? . D4 | ", 3) + last
	})
	slap := drums(2, func(pass int) string {
		last := ". . . . x . . . |"
		if psFill(pass) {
			last = silence(8)
		}
		return synth.Pattern("C4", strings.Repeat(". . . . x . | ", 3)+last)
	})

	bansuri := patches.Recorder()
	bansuri.Name = "bansuri"
	bansuri.Oscs[2].Level = 0.07
	bansuri.Amp.Attack = 0.05
	bansuri.Filter.Cutoff = 2800
	bansuri.LFOs = []synth.LFO{{Target: synth.Vibrato, Rate: 5.2, Depth: 22, Delay: 0.3}}

	sky := func(l float64) synth.Send { return synth.Send{Bus: "sky", Level: l} }
	dub := func(l float64) synth.Send { return synth.Send{Bus: "dub", Level: l} }
	drum := func(name string, p synth.Patch, notes []synth.Note, pan, gain float64) synth.Part {
		return synth.Part{Name: name, Patch: p, Notes: notes, Pan: pan, Gain: gain, Out: "drums", Sends: []synth.Send{sky(0.08)}}
	}

	return synth.Song{
		ID:   "playasunrise",
		Name: "Playa Sunrise",
		Parts: []synth.Part{
			{Name: "wind", Patch: patches.Wind(), Notes: wind, Sends: []synth.Send{sky(0.3)}},
			{Name: "tanpura", Patch: patches.Tanpura(), Notes: tanpura, Pan: -0.35, Gain: 1.25, Sends: []synth.Send{sky(0.3)}},
			{Name: "handpan", Patch: patches.Handpan(), Notes: handpan, Pan: 0.2, Gain: 0.45, Sends: []synth.Send{sky(0.35), dub(0.12)}},
			{Name: "didgeridoo", Patch: patches.Didgeridoo(), Notes: didge, Gain: 0.8, Sends: []synth.Send{sky(0.15)}},
			{
				Name: "bass", Patch: patches.HofnerBass(), Notes: bass, Gain: 0.42,
				Inserts: []synth.Effect{synth.Compressor{Threshold: -20, Ratio: 4, Attack: 0.01, Makeup: 4}},
				Sends:   []synth.Send{sky(0.04)},
			},
			{Name: "sitar", Patch: patches.Sitar(), Notes: hook, Pan: -0.15, Gain: 1.5, Sends: []synth.Send{sky(0.3), dub(0.25)}},
			{Name: "bansuri", Patch: bansuri, Notes: melody, Pan: 0.25, Gain: 0.65, Sends: []synth.Send{sky(0.4), dub(0.3)}},
			{
				Name: "pad", Patch: patches.SawPad(), Notes: pad,
				Inserts: []synth.Effect{swell, synth.Phaser{Rate: 0.15, Depth: 2, Feedback: 0.5}},
				Sends:   []synth.Send{sky(0.5)},
			},
			{Name: "chimes", Patch: patches.Bell(), Notes: chimes, Pan: 0.5, Gain: 0.33, Sends: []synth.Send{sky(0.6)}},
			drum("shaker", patches.Shaker(), shaker, 0.35, 1.1),
			drum("djembe", patches.HandDrum(), djembe, -0.1, 0.7),
			drum("slap", patches.Slap(), slap, -0.1, 2.2),
		},
		Buses: []synth.Bus{
			{Name: "drums", Effects: []synth.Effect{
				synth.Compressor{Threshold: -18, Ratio: 2.5, Attack: 0.008, Release: 0.12, Makeup: 2},
			}},
			{Name: "sky", Effects: []synth.Effect{synth.Reverb{Size: 0.85, Damp: 0.4, PreDelay: 0.03}}},
			{Name: "dub", Effects: []synth.Effect{
				synth.Delay{Time: 3 * psStep, Feedback: 0.4, Damp: 2200, PingPong: true},
			}},
		},
		Master: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowCut, Freq: 30},
				{Kind: synth.HighShelf, Freq: 11000, Gain: -1.5},
			}},
			synth.Compressor{Threshold: -14, Ratio: 2, Attack: 0.015, Release: 0.2, Makeup: 1},
			synth.Limiter{},
		},
		Length:    float64(len(psForm.sections)) * psCycle,
		Loop:      true,
		LoopStart: 2 * psCycle,
		Meter:     synth.Meter{Beat: psBeat, Bars: []int{3, 3, 3, 4}},
	}
}
