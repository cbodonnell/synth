package synth

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func riff(tempo float64) Arrangement {
	return Arrangement{
		ID: "riff", Tempo: tempo, Meter: []int{3, 4},
		Sections: []Section{{"Intro", 1}, {"A", 2}, {"B", 1}},
		Loop:     "A",
		Tracks: []Track{
			{
				Name: "kick", Patch: tone(Sine),
				Clips: []Clip{{Name: "beat", Step: 1, Hit: "C2", Bars: []string{"X . .", "x . x ."}}},
				Play:  []Play{{Clip: "beat", Bar: 1, Times: 2}},
			},
			{
				Name: "pad", Patch: tone(Saw), Out: "verb",
				Inserts: []Effect{
					Sidechain{Key: "kick", Depth: 0.5},
					AutoFilter{Points: []Point{{At: 3, Value: 500}, {At: 7, Value: 5000}}},
				},
				Clips: []Clip{{Name: "hold", Notes: []Note{{At: 0.5, Dur: 2, Key: 60}}, Length: 2}},
				Play:  []Play{{Clip: "hold", Bar: 2}},
			},
		},
		Buses: []Bus{{Name: "verb", Effects: []Effect{Delay{Time: 0.5, Feedback: 0.3}}}},
	}
}

func TestArrangementTimesFollowTheMeter(t *testing.T) {
	s := riff(120).Song()
	require.NoError(t, Validate(s))
	assert.InDelta(t, 3*0.5, s.LoopStart, 1e-9)
	assert.InDelta(t, (3+4+3+4)*0.5, s.Length, 1e-9)
	assert.True(t, s.Loop)

	var kicks []float64
	for _, n := range s.Parts[0].Notes {
		kicks = append(kicks, n.At)
	}
	assert.InDeltaSlice(t, []float64{0, 1.5, 2.5, 3.5, 5, 6}, kicks, 1e-9)

	pad := s.Parts[1].Notes[0]
	assert.InDelta(t, 1.5+0.25, pad.At, 1e-9)
	assert.InDelta(t, 1, pad.Dur, 1e-9)
	assert.Equal(t, kicks, s.Parts[1].Inserts[0].(Sidechain).Hits)
}

func TestArrangementBeatsInEffectsFollowTheTempo(t *testing.T) {
	slow, fast := riff(60).Song(), riff(120).Song()
	assert.InDelta(t, 2*fast.Buses[0].Effects[0].(Delay).Time, slow.Buses[0].Effects[0].(Delay).Time, 1e-9)
	assert.InDelta(t, 2*fast.Parts[1].Inserts[1].(AutoFilter).Points[1].At, slow.Parts[1].Inserts[1].(AutoFilter).Points[1].At, 1e-9)
}

func TestArrangementLeavesMutedTracksOut(t *testing.T) {
	a := riff(120)
	a.Tracks[0].Mute = true
	s := a.Song()
	require.Len(t, s.Parts, 1)
	assert.Equal(t, "pad", s.Parts[0].Name)
	assert.NotEmpty(t, s.Parts[0].Inserts[0].(Sidechain).Hits)
}

func TestArrangementPedalHoldsToTheBarLine(t *testing.T) {
	a := riff(120)
	a.Tracks[0].Clips[0].Pedal = true
	notes := a.Song().Parts[0].Notes
	assert.InDelta(t, 1.5-pedalLift, notes[0].Dur, 1e-9)
	assert.InDelta(t, 3.5-2.5-pedalLift, notes[2].Dur, 1e-9)
}

func TestArrangementValidate(t *testing.T) {
	require.NoError(t, riff(120).Validate())

	for name, edit := range map[string]func(*Arrangement){
		"unknown clip":   func(a *Arrangement) { a.Tracks[0].Play[0].Clip = "nope" },
		"short bar":      func(a *Arrangement) { a.Tracks[0].Clips[0].Bars[1] = "x . x" },
		"misplaced bar":  func(a *Arrangement) { a.Tracks[0].Play[0].Bar = 2 },
		"overlap":        func(a *Arrangement) { a.Tracks[1].Play = append(a.Tracks[1].Play, Play{Clip: "hold", Bar: 3}) },
		"past the end":   func(a *Arrangement) { a.Tracks[1].Play[0].Bar = 4 },
		"unknown loop":   func(a *Arrangement) { a.Loop = "C" },
		"unknown bus":    func(a *Arrangement) { a.Tracks[1].Out = "nope" },
		"unknown key":    func(a *Arrangement) { a.Tracks[1].Inserts[0] = Sidechain{Key: "snare"} },
		"bad note":       func(a *Arrangement) { a.Tracks[0].Clips[0].Hit = "H2" },
		"duplicate clip": func(a *Arrangement) { a.Tracks[0].Clips = append(a.Tracks[0].Clips, a.Tracks[0].Clips[0]) },
	} {
		a := riff(120)
		edit(&a)
		assert.Error(t, a.Validate(), name)
	}
}

func blip() Arrangement {
	return Arrangement{
		ID: "blip", Vary: Variation{Pitch: 1, Takes: 2}, Voices: 3, Priority: 2,
		Tracks: []Track{
			{
				Name: "tone", Patch: tone(Sine),
				Clips: []Clip{{Name: "a", Notes: []Note{{At: 0.05, Dur: 0.1, Key: 69}, {At: 0.25, Dur: 0.2, Key: 72}}}},
				Play:  []Play{{Clip: "a", Bar: 1}},
			},
		},
		Master: []Effect{Normalize{Peak: -6}, Fader{Points: []Point{{At: 0.3, Value: 0}}}},
	}
}

func TestSoundArrangementIsInSeconds(t *testing.T) {
	a := blip()
	require.NoError(t, a.Validate())
	s := a.Song()
	assert.Zero(t, s.Length)
	assert.Equal(t, Meter{}, s.Meter)
	assert.Equal(t, a.Vary, s.Vary)
	assert.Equal(t, 3, s.Voices)
	assert.Equal(t, 2, s.Priority)
	var at []float64
	for _, n := range s.Parts[0].Notes {
		at = append(at, n.At)
	}
	assert.Equal(t, []float64{0.05, 0.25}, at)
	assert.Equal(t, 0.2, s.Parts[0].Notes[1].Dur)
	assert.Equal(t, 0.3, s.Master[1].(Fader).Points[0].At)
	assert.Greater(t, Render(s).Seconds(), 0.45)

	for name, edit := range map[string]func(*Arrangement){
		"sections":  func(a *Arrangement) { a.Sections = []Section{{"A", 1}} },
		"bars":      func(a *Arrangement) { a.Tracks[0].Clips[0] = Clip{Name: "a", Step: 0.1, Bars: []string{"C4"}} },
		"later bar": func(a *Arrangement) { a.Tracks[0].Play[0].Bar = 2 },
		"repeated":  func(a *Arrangement) { a.Tracks[0].Play[0].Times = 2 },
	} {
		a := blip()
		edit(&a)
		assert.Error(t, a.Validate(), name)
	}
}

func TestNormalizeLandsThePeak(t *testing.T) {
	for _, peak := range []float64{-12, -1} {
		a := blip()
		a.Master = []Effect{Normalize{Peak: peak}}
		got := 20 * math.Log10(float64(Render(a.Song()).Peak()))
		assert.InDelta(t, peak, got, 0.01)
	}
	silent := Stereo{L: make([]float32, 8), R: make([]float32, 8)}
	Normalize{Peak: -6}.Process(silent)
	assert.Zero(t, silent.Peak())
}

func TestPhraseWrittenInPiecesDriftsAsOne(t *testing.T) {
	p := Phrase{Step: 0.25, Swing: 0.2, Humanize: 0.01, Seed: 3}
	whole := p.Notes("C4 D4 E4 F4", "G4 A4 B4 C5")
	second := p
	second.At = 1
	pieces := append(p.Notes("C4 D4 E4 F4"), second.Notes("G4 A4 B4 C5")...)
	require.Len(t, pieces, len(whole))
	for i := range whole {
		assert.InDelta(t, whole[i].At, pieces[i].At, 1e-12)
		assert.Equal(t, whole[i].Vel, pieces[i].Vel)
	}
}
