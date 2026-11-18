package songs

import (
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
// lifts at every bar line, and sings a slow line an octave above.
//
// The first pass is the ostinato alone with the swell box shut, the second adds the
// pedal and the piano's arpeggios; the loop brings in the chords, then the piano's
// tune, doubled in octaves as the box opens fully, and an octave above the ostinato,
// before the box shuts again. A beat is a dotted quarter, and each token an eighth.
var interstellar = synth.Arrangement{
	ID: "interstellar", Name: "Interstellar (intro)",
	Tempo: 64, Meter: []int{2}, Key: "A minor",
	Sections: []synth.Section{
		{Name: "Ostinato", Bars: 8},
		{Name: "Pedal", Bars: 8},
		{Name: "Chords", Bars: 8},
		{Name: "Tune", Bars: 8},
		{Name: "Octaves", Bars: 8},
		{Name: "Shut", Bars: 8},
	},
	Loop: "Chords",
	Tracks: []synth.Track{
		{
			// The swell box opens across the intro, then breathes once per two passes of
			// the loop, shut again at the seam so the body joins its own start.
			Name: "ostinato", Patch: patches.PipeOrgan(), Level: -1.9, Pan: -0.15,
			Inserts: []synth.Effect{
				synth.AutoFilter{
					Points: []synth.Point{{Value: 700}, {At: 32, Value: 1400}, {At: 64, Value: 8000}, {At: 96, Value: 1400}},
				},
			},
			Sends: []synth.Send{{Bus: "nave", Level: -3.1}},
			Clips: []synth.Clip{{Name: "A", Step: 1.0 / 3, Gate: 0.8, Vel: 0.7, Bars: []string{
				"A4 E5 C5 E5 A4 E5",
				"G4 E5 B4 E5 G4 E5",
				"A4 E5 C5 E5 A4 E5",
				"G4 E5 B4 E5 G4 E5",
				"A4 E5 C5 E5 A4 E5",
				"G4 E5 B4 E5 G4 E5",
				"A4 E5 C5 E5 A4 E5",
				"B4 E5 D5 E5 B4 E5",
			}}},
			Play: []synth.Play{{Clip: "A", Bar: 1, Times: 6}},
		},
		{
			Name: "high", Patch: patches.PipeOrgan(), Level: -6.9, Pan: 0.25,
			Inserts: []synth.Effect{
				synth.AutoFilter{
					Points: []synth.Point{{Value: 700}, {At: 32, Value: 1400}, {At: 64, Value: 8000}, {At: 96, Value: 1400}},
				},
			},
			Sends: []synth.Send{{Bus: "nave", Level: -1.9}},
			Clips: []synth.Clip{{Name: "E", Step: 1.0 / 3, Gate: 0.6, Vel: 0.5, Bars: []string{
				". E6 . E6 . E6",
			}}},
			Play: []synth.Play{{Clip: "E", Bar: 25, Times: 16}},
		},
		{
			Name: "chords", Patch: patches.PipeOrgan(), Level: -9.1, Pan: 0.1,
			Inserts: []synth.Effect{
				synth.AutoFilter{
					Points: []synth.Point{{Value: 700}, {At: 32, Value: 1400}, {At: 64, Value: 8000}, {At: 96, Value: 1400}},
				},
			},
			Sends: []synth.Send{{Bus: "nave", Level: -4.4}},
			Clips: []synth.Clip{{Name: "A", Step: 1.0 / 3, Gate: 0.98, Bars: []string{
				"A2,E3,A3,C4 - - - - -",
				"G2,E3,G3,B3 - - - - -",
				"F2,C3,A3,E4 - - - - -",
				"E2,B2,G3,B3 - - - - -",
				"A2,E3,A3,C4 - - - - -",
				"G2,E3,G3,B3 - - - - -",
				"F2,C3,A3,E4 - - - - -",
				"G2,D3,B3,E4 - - - - -",
			}}},
			Play: []synth.Play{{Clip: "A", Bar: 17, Times: 4}},
		},
		{
			Name: "piano", Patch: patches.GrandPiano(), Level: 4.1, Pan: -0.25,
			Sends: []synth.Send{{Bus: "nave", Level: -9.1}},
			Clips: []synth.Clip{{Name: "A", Step: 1.0 / 3, Vel: 0.45, Humanize: 0.006, Seed: 1, Pedal: true, Bars: []string{
				"A2 E3 A3 C4 A3 E3",
				"G2 E3 G3 B3 G3 E3",
				"F2 C3 E3 A3 E3 C3",
				"E2 B2 E3 G3 E3 B2",
				"A2 E3 A3 C4 A3 E3",
				"G2 E3 G3 B3 G3 E3",
				"F2 C3 E3 A3 E3 C3",
				"G2 D3 G3 B3 E4 B3",
			}}},
			Play: []synth.Play{{Clip: "A", Bar: 9, Times: 5}},
		},
		{
			Name: "piano tune", Patch: patches.GrandPiano(), Level: 2.9, Pan: 0.2,
			Sends: []synth.Send{{Bus: "nave", Level: -8}},
			Clips: []synth.Clip{{Name: "A", Step: 1.0 / 3, Gate: 0.95, Vel: 0.6, Humanize: 0.006, Seed: 3, Bars: []string{
				"E6 - - - D6 C6",
				"B5 - - - - -",
				"C6 - - - B5 A5",
				"G5 - - B5 - -",
				"E6 - - - D6 C6",
				"B5 - - - C6 D6",
				"E6 - - - F6 E6",
				"D6 - - B5 - -",
			}}},
			Play: []synth.Play{{Clip: "A", Bar: 25, Times: 2}},
		},
		{
			Name: "piano octave", Patch: patches.GrandPiano(), Level: 2.9, Pan: 0.2,
			Sends: []synth.Send{{Bus: "nave", Level: -8}},
			Clips: []synth.Clip{{Name: "A", Step: 1.0 / 3, Gate: 0.95, Vel: 0.42, Humanize: 0.006, Seed: 3, Bars: []string{
				"E5 - - - D5 C5",
				"B4 - - - - -",
				"C5 - - - B4 A4",
				"G4 - - B4 - -",
				"E5 - - - D5 C5",
				"B4 - - - C5 D5",
				"E5 - - - F5 E5",
				"D5 - - B4 - -",
			}}},
			Play: []synth.Play{{Clip: "A", Bar: 33}},
		},
		{
			Name: "pedal", Patch: patches.PipeOrgan(), Level: -1.4,
			Sends: []synth.Send{{Bus: "nave", Level: -6}},
			Clips: []synth.Clip{{Name: "A", Step: 1.0 / 3, Gate: 0.98, Bars: []string{
				"A1 - - - - -",
				"G1 - - - - -",
				"F1 - - - - -",
				"E1 - - - - -",
				"A1 - - - - -",
				"G1 - - - - -",
				"F1 - - - - -",
				"G1 - - - - -",
			}}},
			Play: []synth.Play{{Clip: "A", Bar: 9, Times: 5}},
		},
	},
	Buses: []synth.Bus{
		{Name: "nave", Effects: []synth.Effect{synth.Reverb{Size: 0.97, Damp: 0.3, PreDelay: 0.045}}},
	},
	Master: []synth.Effect{
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 25},
			{Kind: synth.HighShelf, Freq: 10000, Gain: -2},
		}},
		synth.Compressor{Threshold: -16, Ratio: 2, Attack: 0.03, Release: 0.3, Makeup: 5},
		synth.Limiter{},
	},
	Target: -14,
}

func Interstellar() synth.Song { return interstellar.Song() }
