package session

import (
	"slices"
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/songs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryClipsBarsReadAndWriteBack(t *testing.T) {
	for _, a := range songs.Arrangements() {
		sc, _ := ParseScale(a.Key)
		for _, tr := range a.Tracks {
			for _, c := range tr.Clips {
				g, err := ReadGrid(c)
				if len(c.Bars) == 0 {
					assert.Error(t, err)
					continue
				}
				require.NoError(t, err, "%s %s %s", a.ID, tr.Name, c.Name)
				assert.Equal(t, c.Bars, g.Write(sc), "%s %s %s", a.ID, tr.Name, c.Name)
			}
		}
	}
}

// notes is what the clip plays from the song's start, at a beat a second.
func notes(c synth.Clip) []synth.Note {
	a := synth.Arrangement{
		Tempo:    60,
		Sections: []synth.Section{{Name: "A", Bars: len(c.Bars)}},
		Tracks:   []synth.Track{{Name: "t", Clips: []synth.Clip{c}, Play: []synth.Play{{Clip: c.Name, Bar: 1}}}},
	}
	return a.Song().Parts[0].Notes
}

func TestDrawnNotesReadBackAsDrawn(t *testing.T) {
	sc, _ := ParseScale("D minor")
	g, err := ReadGrid(synth.Clip{Step: 0.5, Bars: []string{". . . . . . . .", ". . . . . . . ."}})
	require.NoError(t, err)

	g = g.Draw(0, 3, 62)
	g = g.Draw(0, 1, 65, 69)
	g = g.Draw(6, 4, 70)
	g = g.Draw(2, 1, 74) // cuts the chord short
	g = g.Accent(6, 1).Ratchet(2, 3).Slide(Ref{2, 74}, 76)
	g, moved := g.Move([]Ref{{6, 70}}, 1, 12)
	assert.Equal(t, []Ref{{7, 82}}, moved)
	assert.Equal(t, 2, g.Events[0].Len, "a note drawn under a chord cuts it")

	bars := g.Write(sc)
	back, err := ReadGrid(synth.Clip{Step: 0.5, Bars: bars})
	require.NoError(t, err)
	require.Len(t, back.Events, len(g.Events))
	for i, e := range g.Events {
		assert.True(t, e.same(back.Events[i]), "%v reads back as %v", e, back.Events[i])
		assert.Equal(t, e.Step, back.Events[i].Step)
		assert.Equal(t, e.Len, back.Events[i].Len)
	}
	assert.Contains(t, bars[0], "Bb5!", "spelled in the key")

	c := synth.Clip{Name: "c", Step: 0.5, Bars: bars}
	assert.NotPanics(t, func() { notes(c) })
}

func TestStepsHalveAndDoubleLosingNothing(t *testing.T) {
	sc, _ := ParseScale("C")
	c := synth.Clip{Name: "c", Step: 0.5, Bars: []string{"C4 - E4*3 . G4,B4! - - D4/E4"}}
	g, err := ReadGrid(c)
	require.NoError(t, err)

	half := c
	half.Step /= 2
	half.Bars = g.Halve().Write(sc)
	assert.Equal(t, notes(c), notes(half))

	h, err := ReadGrid(half)
	require.NoError(t, err)
	d, ok := h.Double()
	require.True(t, ok)
	assert.Equal(t, c.Bars, d.Write(sc))

	_, ok = h.Draw(1, 1, 60).Double()
	assert.False(t, ok, "a note between the longer steps")
}

func TestDrumCellsCycle(t *testing.T) {
	g, err := ReadGrid(synth.Clip{Step: 1, Hit: "C2", Bars: []string{". . . ."}})
	require.NoError(t, err)
	var seen []string
	for range 4 {
		g = g.Cycle(1)
		seen = append(seen, g.Write(Scale{})[0])
	}
	assert.Equal(t, []string{". x . .", ". X . .", ". o . .", ". . . ."}, seen)
}

func TestStepInputRestsAndHolds(t *testing.T) {
	g, err := ReadGrid(synth.Clip{Step: 1, Bars: []string{"C4 D4 E4 F4"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"C4 - . F4"}, g.Hold(1).Rest(2).Write(Scale{}))
}

func TestScalesSpellAndStack(t *testing.T) {
	for _, a := range songs.Arrangements() {
		_, ok := ParseScale(a.Key)
		assert.True(t, ok, a.Key)
	}
	bb, _ := ParseScale("Bb minor")
	var names []string
	for _, k := range []int{70, 72, 73, 75, 77, 78, 80, 71} {
		names = append(names, bb.Spell(k))
	}
	assert.Equal(t, []string{"Bb4", "C5", "Db5", "Eb5", "F5", "Gb5", "Ab5", "B4"}, names)
	for _, k := range []int{58, 59, 60, 61} {
		key, err := synth.ParseKey(bb.Spell(k))
		require.NoError(t, err)
		assert.Equal(t, float64(k), key)
	}
	e, _ := ParseScale("E mixolydian")
	assert.Equal(t, "G#4", e.Spell(68))
	assert.Equal(t, []int{64, 68, 71}, e.Chord(64, 3))
	assert.Equal(t, []int{69, 73, 76, 80}, e.Chord(70, 4), "from the degree below")
	assert.True(t, slices.Equal(Scale{}.Chord(60, 3), []int{60, 64, 67}))
}
