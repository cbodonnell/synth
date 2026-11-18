package session

import (
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/stretchr/testify/assert"
)

func TestNotesMoveWithinTheClip(t *testing.T) {
	notes := []synth.Note{{At: 0.1, Dur: 0.1, Key: 60}, {At: 0.3, Dur: 0.1, Key: 125, Slide: 2}}
	both := []int{0, 1}

	early := MoveNotes(notes, both, -1, 0, 0.5)
	assert.Equal(t, []float64{0, 0.2}, []float64{early[0].At, early[1].At}, "the first note stops at the start, the rest keep their spacing")
	late := MoveNotes(notes, both, 1, 0, 0.5)
	assert.Less(t, late[1].At, 0.5)
	assert.InDelta(t, 0.2, late[1].At-late[0].At, 1e-9)
	assert.Equal(t, notes[1].Key, MoveNotes(notes, both, 0, 12, 0.5)[1].Key, "the slide is already at the keyboard's top")

	assert.Len(t, DeleteNotes(notes, []int{0}), 1)
	assert.Equal(t, 0.05, Snap(0.061, 0.05))
	assert.Equal(t, 0.0612, Snap(0.06123, 0))
}
