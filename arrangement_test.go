package synth

import (
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
