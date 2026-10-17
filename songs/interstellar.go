package songs

import (
	"math"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// Hans Zimmer, "Interstellar" main theme (2014), the organ intro, in A minor, 6/8.
//
// Everything hangs on E. The ostinato ticks like a clock, a chord tone and E above it
// in eighths, while the harmony moves underneath: Am, Em over G, Fmaj7, Em, and on the
// second half a G6 that turns back home. Over F the ostinato keeps Am's notes, which
// are the chord's third and seventh, so the hands never move and the floor does.
//
// The piano rolls each chord upward in the left hand under the sustain pedal, which
// lifts at every bar line, and sings a slow line an octave above the ostinato.
//
// The first pass is the ostinato alone with the swell box shut, the second adds the
// pedal and the piano's arpeggios; the loop brings in the chords, then the piano's
// tune, doubled in octaves as the box opens fully, and an octave above the ostinato,
// before the box shuts again. Each token is an eighth.
const (
	isStep  = 60.0 / 192
	isCycle = 48 * isStep
)

var isForm = form{"AAAAAA", 48}

var (
	isOstinato = map[byte]string{
		'A': "A4 E5 C5 E5 A4 E5 | G4 E5 B4 E5 G4 E5 | A4 E5 C5 E5 A4 E5 | G4 E5 B4 E5 G4 E5 |" +
			"A4 E5 C5 E5 A4 E5 | G4 E5 B4 E5 G4 E5 | A4 E5 C5 E5 A4 E5 | B4 E5 D5 E5 B4 E5 |",
	}
	isHigh = map[byte]string{
		'A': ". E6 . E6 . E6 | . E6 . E6 . E6 | . E6 . E6 . E6 | . E6 . E6 . E6 |" +
			". E6 . E6 . E6 | . E6 . E6 . E6 | . E6 . E6 . E6 | . E6 . E6 . E6 |",
	}
	isChords = map[byte]string{
		'A': "A2,E3,A3,C4 - - - - - | G2,E3,G3,B3 - - - - - | F2,C3,A3,E4 - - - - - | E2,B2,G3,B3 - - - - - |" +
			"A2,E3,A3,C4 - - - - - | G2,E3,G3,B3 - - - - - | F2,C3,A3,E4 - - - - - | G2,D3,B3,E4 - - - - - |",
	}
	isArpeggio = map[byte]string{
		'A': "A2 E3 A3 C4 A3 E3 | G2 E3 G3 B3 G3 E3 | F2 C3 E3 A3 E3 C3 | E2 B2 E3 G3 E3 B2 |" +
			"A2 E3 A3 C4 A3 E3 | G2 E3 G3 B3 G3 E3 | F2 C3 E3 A3 E3 C3 | G2 D3 G3 B3 E4 B3 |",
	}
	isTune = map[byte]string{
		'A': "E6 - - - D6 C6 | B5 - - - - - | C6 - - - B5 A5 | G5 - - B5 - - |" +
			"E6 - - - D6 C6 | B5 - - - C6 D6 | E6 - - - F6 E6 | D6 - - B5 - - |",
	}
	isPedal = map[byte]string{
		'A': "A1 - - - - - | G1 - - - - - | F1 - - - - - | E1 - - - - - |" +
			"A1 - - - - - | G1 - - - - - | F1 - - - - - | G1 - - - - - |",
	}
)

func Interstellar() synth.Song {
	ostinato := synth.Phrase{Step: isStep, Gate: 0.8, Vel: 0.7}.Notes(isForm.plays(isOstinato, from(0)))
	high := synth.Phrase{Step: isStep, Gate: 0.6, Vel: 0.5}.Notes(isForm.plays(isHigh, only(3, 4)))
	held := synth.Phrase{Step: isStep, Gate: 0.98}
	chords := held.Notes(isForm.plays(isChords, from(2)))
	pedal := held.Notes(isForm.plays(isPedal, from(1)))

	arpeggio := sustained(synth.Phrase{Step: isStep, Vel: 0.45, Humanize: 0.006, Seed: 1}.Notes(isForm.plays(isArpeggio, from(1))))
	tune := synth.Phrase{Step: isStep, Gate: 0.95, Vel: 0.6, Humanize: 0.006, Seed: 3}
	melody := tune.Notes(isForm.plays(isTune, only(3, 4)))
	for _, n := range tune.Notes(isForm.plays(isTune, only(4))) {
		n.Key -= 12
		n.Vel *= 0.7
		melody = append(melody, n)
	}

	// Opens across the intro, then breathes once per two passes of the loop, shut again
	// at the seam so the body joins its own start.
	swell := synth.AutoFilter{Points: []synth.Point{
		{At: 0, Value: 700},
		{At: 2 * isCycle, Value: 1400},
		{At: 4 * isCycle, Value: 8000},
		{At: 6 * isCycle, Value: 1400},
	}}

	nave := func(l float64) synth.Send { return synth.Send{Bus: "nave", Level: l} }

	return synth.Song{
		ID:   "interstellar",
		Name: "Interstellar (intro)",
		Parts: []synth.Part{
			{Name: "ostinato", Patch: patches.PipeOrgan(), Notes: ostinato, Pan: -0.15, Gain: 0.8, Inserts: []synth.Effect{swell}, Sends: []synth.Send{nave(0.7)}},
			{Name: "high", Patch: patches.PipeOrgan(), Notes: high, Pan: 0.25, Gain: 0.45, Inserts: []synth.Effect{swell}, Sends: []synth.Send{nave(0.8)}},
			{Name: "chords", Patch: patches.PipeOrgan(), Notes: chords, Pan: 0.1, Gain: 0.35, Inserts: []synth.Effect{swell}, Sends: []synth.Send{nave(0.6)}},
			{Name: "piano", Patch: patches.GrandPiano(), Notes: arpeggio, Pan: -0.25, Gain: 1.6, Sends: []synth.Send{nave(0.35)}},
			{Name: "piano tune", Patch: patches.GrandPiano(), Notes: melody, Pan: 0.2, Gain: 1.4, Sends: []synth.Send{nave(0.4)}},
			{Name: "pedal", Patch: patches.PipeOrgan(), Notes: pedal, Gain: 0.85, Sends: []synth.Send{nave(0.5)}},
		},
		Buses: []synth.Bus{
			{Name: "nave", Effects: []synth.Effect{synth.Reverb{Size: 0.97, Damp: 0.3, PreDelay: 0.045}}},
		},
		Master: []synth.Effect{
			synth.EQ{Bands: []synth.Band{
				{Kind: synth.LowCut, Freq: 25},
				{Kind: synth.HighShelf, Freq: 10000, Gain: -2},
			}},
			synth.Compressor{Threshold: -16, Ratio: 2, Attack: 0.03, Release: 0.3, Makeup: 1},
			synth.Limiter{},
		},
		Length:    float64(len(isForm.sections)) * isCycle,
		Loop:      true,
		LoopStart: 2 * isCycle,
		Meter:     synth.Meter{Beat: 3 * isStep, Bars: []int{2}},
	}
}

// sustained holds each note to the next bar line, where the pedal lifts.
func sustained(notes []synth.Note) []synth.Note {
	bar := 6 * isStep
	for i, n := range notes {
		notes[i].Dur = max(n.Dur, (math.Floor(n.At/bar+1e-6)+1)*bar-n.At-0.02)
	}
	return notes
}
