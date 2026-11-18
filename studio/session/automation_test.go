package session

import (
	"slices"
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sweep struct {
	Curve []synth.Point
}

func (sweep) Tail() float64          { return 0 }
func (sweep) Process(s synth.Stereo) {}

func automated() synth.Arrangement {
	a := waltz()
	a.Tracks = slices.Clone(a.Tracks)
	a.Tracks[1].Inserts = []synth.Effect{synth.Fader{Points: []synth.Point{{At: 0, Value: -6}, {At: 4, Value: 0}, {At: 8, Value: -3}}}}
	a.Master = []synth.Effect{sweep{Curve: []synth.Point{{At: 3, Value: 1}}}}
	return a
}

func TestAPointDraggedPastItsNeighboursStaysInOrder(t *testing.T) {
	a := automated()
	l := Automations(a)[0]
	i := 0
	for _, at := range []float64{6, 10, 4, 0, 2} {
		var err error
		i, err = MovePoint(&a, l, i, synth.Point{At: at, Value: -1})
		require.NoError(t, err)
		pts := Automations(a)[0].Points
		assert.Equal(t, synth.Point{At: at, Value: -1}, pts[i], "the index follows the point")
		assert.True(t, slices.IsSortedFunc(pts, func(p, q synth.Point) int { return int(p.At - q.At) }), "%v", pts)
		assert.Len(t, pts, 3)
	}
	assert.NoError(t, a.Validate())
}

func TestAGamesOwnAutomationIsALaneInBeats(t *testing.T) {
	a := automated()
	lanes := Automations(a)
	require.Len(t, lanes, 2)
	game := lanes[1]
	assert.Equal(t, "Curve", game.Field)
	assert.Less(t, game.Param.Min, game.Param.Max)

	_, err := AddPoint(&a, game, synth.Point{At: 1, Value: 0.5})
	require.NoError(t, err)
	assert.Len(t, Automations(a)[1].Points, 2)
	beat := 60 / a.Tempo
	assert.InDelta(t, 3*beat, a.Song().Master[0].(sweep).Curve[1].At, 1e-12)

	InsertSection(&a, 0, 1)
	assert.Equal(t, 6.0, Automations(a)[1].Points[1].At, "a bar of 3/4 inserted before it")
}
