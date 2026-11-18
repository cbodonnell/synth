package session

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/songs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// saveAndUndo saves the doc's edit, checks the file holds it, then undoes and saves it
// back to src.
func saveAndUndo(t *testing.T, d *Doc, src []byte, what string) {
	t.Helper()
	require.NoError(t, d.Save(), what)
	saved, _ := os.ReadFile(d.Path)
	assert.NotEqual(t, string(src), string(saved), what)
	again, err := Load(d.Now(), []string{filepath.Dir(d.Path)})
	require.NoError(t, err, "%s: the saved file holds the edit", what)
	assert.Equal(t, d.Now(), again.Now(), what)

	require.True(t, d.Undo(), what)
	require.NoError(t, d.Save(), "%s undone", what)
	back, _ := os.ReadFile(d.Path)
	assert.Equal(t, string(src), string(back), "%s undone", what)
}

func TestCodeMadeLiteralSavesAndUndoes(t *testing.T) {
	patched, chained := false, false
	for _, a := range songs.Arrangements() {
		for i := range a.Tracks {
			d, src := copied(t, a)
			if !patched && d.Custom("Tracks", i, "Patch") {
				patched = true
				require.True(t, d.MakeLiteral("Tracks", i, "Patch"))
				assert.False(t, d.Custom("Tracks", i, "Patch"))
				assert.True(t, d.Dirty())
				saveAndUndo(t, d, src, a.ID+" patch made literal")
				assert.True(t, d.Custom("Tracks", i, "Patch"), "the call is back")
			}
			c := Chain{TrackChain, i}
			if d, src := copied(t, a); !chained && d.Custom(c.Path()...) {
				chained = true
				require.True(t, d.MakeLiteral(c.Path()...))
				saveAndUndo(t, d, src, a.ID+" chain made literal")
			}
		}
	}
	assert.True(t, patched && chained, "a patch and a chain written as code")
}

func TestPatchFromTheBrowser(t *testing.T) {
	d, src := fixture(t)
	require.True(t, d.SetPatch(1, "Kick"))
	assert.True(t, d.Custom("Tracks", 1, "Patch"), "written as the call")
	out, err := d.Source()
	require.NoError(t, err)
	assert.Contains(t, string(out), "Patch: patches.Kick(),")
	saveAndUndo(t, d, []byte(src), "patch set")

	src = `package songs

import "github.com/cbodonnell/synth"

var solo = synth.Arrangement{
	ID: "solo", Tempo: 120,
	Sections: []synth.Section{{Name: "A", Bars: 1}},
	Tracks:   []synth.Track{{Name: "sine", Patch: synth.Patch{Oscs: []synth.Osc{{}}}}},
}
`
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "solo.go"), []byte(src), 0o644))
	d, err = Load(synth.Arrangement{
		ID: "solo", Tempo: 120,
		Sections: []synth.Section{{Name: "A", Bars: 1}},
		Tracks:   []synth.Track{{Name: "sine", Patch: synth.Patch{Oscs: []synth.Osc{{}}}}},
	}, []string{dir})
	require.NoError(t, err)
	name, ok := d.AddTrack("HouseKick")
	require.True(t, ok)
	assert.Equal(t, name, d.Now().Tracks[1].Name)
	out, err = d.Source()
	require.NoError(t, err)
	assert.Contains(t, string(out), `"github.com/cbodonnell/synth/patches"`, "imported")
	saveAndUndo(t, d, []byte(src), "track added")
}

type wobble struct {
	Rate   float64
	Steps  int
	On     bool
	Shape  synth.Shape
	Points []synth.Point
	Inner  struct{ Depth float64 }
}

func (wobble) Tail() float64          { return 0 }
func (wobble) Process(s synth.Stereo) {}

func TestAGamesEffectHasControls(t *testing.T) {
	a := synth.Arrangement{Master: []synth.Effect{wobble{Rate: -30, Steps: 3}}}
	cs := Controls(a.Master[0], []any{"Master", 0})
	kinds := map[string]ControlKind{}
	for _, c := range cs {
		kinds[c.Label] = c.Kind
	}
	assert.Equal(t, map[string]ControlKind{"Rate": Knob, "Steps": Knob, "On": Toggle, "Shape": Cycle, "Points": Points, "Inner Depth": Knob}, kinds)
	assert.Less(t, cs[0].Param.Min, -30.0, "a range holding its value")
	assert.Equal(t, 1.0, cs[1].Param.Step, "whole numbers")

	require.True(t, Set(&a, cs[1].Path, 4.0))
	require.True(t, Set(&a, cs[5].Path, 0.5))
	assert.Equal(t, 4, a.Master[0].(wobble).Steps)
	assert.Equal(t, 0.5, a.Master[0].(wobble).Inner.Depth)
}

func TestCurvesPlot(t *testing.T) {
	env, release := EnvelopeShape(synth.Envelope{Attack: 0.01, Decay: 0.2, Sustain: 0.5, Release: 0.3}, 64)
	assert.Zero(t, env[0])
	assert.InDelta(t, 0, env[len(env)-1], 1e-9)
	assert.Greater(t, release, 0.0)
	assert.Less(t, release, 1.0)
	lp := FilterResponse(synth.Filter{Cutoff: 1000}, 64)
	assert.Greater(t, lp[0], lp[len(lp)-1], "a lowpass falls")
	hp := FilterResponse(synth.Filter{Mode: synth.HighPass, Cutoff: 1000}, 64)
	assert.Less(t, hp[0], hp[len(hp)-1], "a highpass rises")

	ink := color.RGBA{255, 255, 255, 255}
	img := Plot([]float64{0, 1}, -1, 40, 20, 2, ink, color.RGBA{})
	assert.Equal(t, ink, img.RGBAAt(0, 19), "starts at the bottom")
	assert.Equal(t, ink, img.RGBAAt(39, 0), "ends at the top")
}

func TestTypedParameters(t *testing.T) {
	hz := synth.Param{Unit: synth.UnitHz, Min: 20, Max: 20000, Log: true}
	secs := synth.Param{Unit: synth.UnitSeconds, Max: 5}
	for _, c := range []struct {
		p    synth.Param
		s    string
		want float64
	}{
		{hz, "3k", 3000}, {hz, "1234.5", 1230}, {hz, "99 kHz", 20000}, {secs, "250 ms", 0.25}, {secs, "1/3", 0.333},
	} {
		got, ok := ParseParam(c.p, c.s)
		assert.True(t, ok, c.s)
		assert.InDelta(t, c.want, got, 1e-9, c.s)
	}
	_, ok := ParseParam(hz, "loud")
	assert.False(t, ok)
	assert.InDelta(t, 1000, FromPosition(hz, Position(hz, 1000, 0), 0), 1e-9, "a log knob's travel")
}
