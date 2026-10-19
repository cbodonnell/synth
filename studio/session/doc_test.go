package session

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
	"github.com/cbodonnell/synth/songs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const songsDir = "../../songs"

// copied loads the arrangement from a copy of its file, so a test can save it.
func copied(t *testing.T, a synth.Arrangement) (*Doc, []byte) {
	t.Helper()
	d, err := Load(a, []string{songsDir})
	require.NoError(t, err, a.ID)
	src, err := os.ReadFile(d.Path)
	require.NoError(t, err)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, filepath.Base(d.Path)), src, 0o644))
	d, err = Load(a, []string{dir})
	require.NoError(t, err)
	return d, src
}

func TestEveryArrangementLoadsAndWritesItselfBack(t *testing.T) {
	for _, a := range songs.Arrangements() {
		d, err := Load(a, []string{songsDir})
		require.NoError(t, err, a.ID)
		src, err := os.ReadFile(d.Path)
		require.NoError(t, err)
		w := &writer{file: d.f}
		w.diff(d.f.lit, reflect.ValueOf(a), reflect.ValueOf(a), true)
		out, err := w.apply()
		require.NoError(t, err, a.ID)
		assert.Equal(t, string(src), string(out), a.ID)
	}
}

// edits are the mixer's, each as it would apply to any arrangement.
var edits = map[string]func(a *synth.Arrangement){
	"fader": func(a *synth.Arrangement) { a.Tracks[0].Level = math.Round(a.Tracks[0].Level*10-15) / 10 },
	"pan": func(a *synth.Arrangement) {
		t := &a.Tracks[len(a.Tracks)-1]
		t.Pan = math.Round(t.Pan*100+math.Copysign(15, t.Pan)) / 100
	},
	"mute": func(a *synth.Arrangement) {
		for i := range a.Tracks {
			if !a.Tracks[i].Mute {
				a.Tracks[i].Mute = true
				return
			}
		}
	},
	"send": func(a *synth.Arrangement) {
		for i := range a.Tracks {
			for j := range a.Tracks[i].Sends {
				a.Tracks[i].Sends[j].Level = math.Round(a.Tracks[i].Sends[j].Level*10-35) / 10
				return
			}
		}
	},
	"new bus": func(a *synth.Arrangement) {
		a.Buses = append(a.Buses, synth.Bus{Name: "studio reverb", Effects: []synth.Effect{synth.Reverb{Size: 0.5, Mix: 1}}})
		for i := range a.Tracks {
			a.Tracks[i].Sends = append(a.Tracks[i].Sends, synth.Send{Bus: "studio reverb", Level: -12})
		}
	},
	"bus": func(a *synth.Arrangement) {
		if len(a.Buses) > 0 {
			a.Buses[0].Level = math.Round(a.Buses[0].Level*10+15) / 10
		}
	},
	"note drawn": gridEdit(false, func(g Grid) Grid { return g.Draw(g.Steps()-1, 1, 60) }),
	"note moved": gridEdit(false, func(g Grid) Grid {
		e := g.Events[0]
		g, _ = g.Move([]Ref{{e.Step, e.Keys[0].Key}}, 0, 1)
		return g
	}),
	"note stretched":   gridEdit(false, func(g Grid) Grid { return g.Stretch(g.Events[0].Step, 1) }),
	"note accented":    gridEdit(false, func(g Grid) Grid { return g.Accent(g.Events[0].Step, -1) }),
	"drum cell cycled": gridEdit(true, func(g Grid) Grid { return g.Cycle(0) }),
	"step halved": func(a *synth.Arrangement) {
		for i, t := range a.Tracks {
			for j, c := range t.Clips {
				if frac, _ := math.Frexp(c.Step); frac != 0.5 || len(c.Bars) == 0 {
					continue
				}
				g, _ := ReadGrid(c)
				sc, _ := ParseScale(a.Key)
				a.Tracks[i].Clips[j].Step /= 2
				a.Tracks[i].Clips[j].Bars = g.Halve().Write(sc)
				return
			}
		}
	},
	"out": func(a *synth.Arrangement) {
		for i := range a.Tracks {
			if a.Tracks[i].Out == "" && len(a.Buses) > 0 {
				a.Tracks[i].Out = a.Buses[0].Name
				return
			}
		}
	},
}

// gridEdit edits the first clip of the arrangement on a grid, a Hit clip or not.
func gridEdit(hit bool, f func(g Grid) Grid) func(a *synth.Arrangement) {
	return func(a *synth.Arrangement) {
		for i, t := range a.Tracks {
			for j, c := range t.Clips {
				g, err := ReadGrid(c)
				if err != nil || g.Hit != hit || len(g.Events) == 0 {
					continue
				}
				sc, _ := ParseScale(a.Key)
				a.Tracks[i].Clips[j].Bars = f(g).Write(sc)
				return
			}
		}
	}
}

func TestEditsSaveAndUndoToTheSameBytes(t *testing.T) {
	for _, a := range songs.Arrangements() {
		for name, edit := range edits {
			d, src := copied(t, a)
			d.Edit(edit)
			if !d.Dirty() {
				continue
			}
			require.NoError(t, d.Save(), "%s %s", a.ID, name)
			saved, _ := os.ReadFile(d.Path)
			assert.NotEqual(t, string(src), string(saved), "%s %s", a.ID, name)

			again, err := Load(d.Now(), []string{filepath.Dir(d.Path)})
			require.NoError(t, err, "%s %s: the saved file holds the edit", a.ID, name)
			assert.Equal(t, d.Now(), again.Now())

			require.True(t, d.Undo())
			require.NoError(t, d.Save(), "%s %s undone", a.ID, name)
			back, _ := os.ReadFile(d.Path)
			assert.Equal(t, string(src), string(back), "%s %s undone", a.ID, name)
		}
	}
}

// fixture is a small arrangement's source, with code in it, and its value.
func fixture(t *testing.T) (*Doc, string) {
	t.Helper()
	src := `package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

var quiet = -6.0

var waltz = synth.Arrangement{
	ID: "waltz", Tempo: 120, Meter: []int{3},
	Sections: []synth.Section{{Name: "A", Bars: 2}, {Name: "B", Bars: 4}},
	Loop:     "B",
	Tracks: []synth.Track{
		{
			Name: "bell", Patch: patches.Bell(), Level: quiet, // the tune
			Clips: []synth.Clip{{Name: "up", Step: 1.0 / 1, Bars: []string{"C5 E5 G5"}}},
			Play:  []synth.Play{{Clip: "up", Bar: 1, Times: 6}},
		},
		// The bass waits for B.
		{
			Name: "bass", Patch: patches.Bass(), Level: -3, Pan: -0.2,
			Sends: []synth.Send{{Bus: "room", Level: -20}}, // a little
			Clips: []synth.Clip{{Name: "root", Step: 1, Bars: []string{"C2 - -", "G1 - -"}}},
			Play:  []synth.Play{{Clip: "root", Bar: 3, Times: 2}},
		},
	},
	Buses: []synth.Bus{
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.4}}},
	},
}
`
	a := synth.Arrangement{
		ID: "waltz", Tempo: 120, Meter: []int{3},
		Sections: []synth.Section{{Name: "A", Bars: 2}, {Name: "B", Bars: 4}},
		Loop:     "B",
		Tracks: []synth.Track{
			{
				Name: "bell", Patch: patches.Bell(), Level: -6,
				Clips: []synth.Clip{{Name: "up", Step: 1, Bars: []string{"C5 E5 G5"}}},
				Play:  []synth.Play{{Clip: "up", Bar: 1, Times: 6}},
			},
			{
				Name: "bass", Patch: patches.Bass(), Level: -3, Pan: -0.2,
				Sends: []synth.Send{{Bus: "room", Level: -20}},
				Clips: []synth.Clip{{Name: "root", Step: 1, Bars: []string{"C2 - -", "G1 - -"}}},
				Play:  []synth.Play{{Clip: "root", Bar: 3, Times: 2}},
			},
		},
		Buses: []synth.Bus{{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.4}}}},
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "waltz.go")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))
	d, err := Load(a, []string{dir})
	require.NoError(t, err)
	return d, src
}

func TestLoadRefusesSourceThatIsntTheBuild(t *testing.T) {
	d, _ := fixture(t)
	a := d.Now()
	a.Tracks = slices.Clone(a.Tracks)
	a.Tracks[1].Pan = 0.5
	_, err := Load(a, []string{filepath.Dir(d.Path)})
	assert.Error(t, err)
}

func TestCodeIsCustomAndLiteralsAreNot(t *testing.T) {
	d, _ := fixture(t)
	assert.True(t, d.Custom("Tracks", 0, "Patch"), "a call")
	assert.True(t, d.Custom("Tracks", 0, "Level"), "a var")
	assert.True(t, d.Custom("Tracks", 0, "Clips", 0, "Step"), "arithmetic")
	assert.False(t, d.Custom("Tracks", 1, "Level"))
	assert.False(t, d.Custom("Tracks", 0, "Pan"), "absent, so a control can add it")
	assert.False(t, d.Custom("Tracks", 1, "Sends", 0, "Level"))
	assert.False(t, d.Custom("Buses", 0, "Effects", 0, "Size"))

	d.Edit(func(a *synth.Arrangement) {
		a.Tracks = append([]synth.Track{{Name: "new"}}, a.Tracks...)
	})
	assert.True(t, d.Custom("Tracks", 1, "Patch"), "found by name")
	assert.False(t, d.Custom("Tracks", 0, "Patch"), "the studio's own")
}

func TestSavesTouchOnlyWhatChanged(t *testing.T) {
	d, src := fixture(t)
	d.Edit(func(a *synth.Arrangement) {
		a.Tracks[1].Level = 0
		a.Tracks[1].Sends = nil
		a.Tracks[0].Mute = true
		a.Tracks[1].Clips[0].Bars = []string{"C2 - -", "A1 - -"}
	})
	out, err := d.Source()
	require.NoError(t, err)
	got := string(out)
	assert.Contains(t, got, "Level: quiet, // the tune", "code and comments survive")
	assert.Contains(t, got, "// The bass waits for B.", "a comment on its own line stays")
	assert.Contains(t, got, `Name: "bass", Patch: patches.Bass(), Pan: -0.2,`, "a field gone to zero is removed")
	assert.NotContains(t, got, "a little", "a deleted item takes its end-of-line comment")
	assert.Regexp(t, `Mute: +true,\n\t\t},`, got, "a new field goes at the end")
	assert.Contains(t, got, `Bars: []string{"C2 - -", "A1 - -"}}`, "a bar is rewritten where it is")
	assert.Contains(t, got, "Step: 1.0 / 1,", "untouched")
	assert.Equal(t, strings.Count(src, "\n"), strings.Count(got, "\n"))
}

func TestSaveRefusesAFileChangedOnDisk(t *testing.T) {
	d, src := fixture(t)
	d.Edit(func(a *synth.Arrangement) { a.Tracks[1].Pan = 0.3 })
	require.NoError(t, os.WriteFile(d.Path, []byte(src+"\n// edited elsewhere\n"), 0o644))
	assert.ErrorIs(t, d.Save(), ErrChanged)
	assert.True(t, d.Dirty())
}

func TestUndoStepsAndHolds(t *testing.T) {
	d, _ := fixture(t)
	d.Hold()
	for _, pan := range []float64{0.1, 0.2, 0.3} {
		d.Edit(func(a *synth.Arrangement) { a.Tracks[1].Pan = pan })
	}
	d.Release()
	d.Edit(func(a *synth.Arrangement) { a.Tracks[1].Level = -9 })
	assert.True(t, d.Dirty())

	require.True(t, d.Undo())
	assert.Equal(t, 0.3, d.Now().Tracks[1].Pan, "the level was one step")
	require.True(t, d.Undo())
	assert.Equal(t, -0.2, d.Now().Tracks[1].Pan, "the drag was one step")
	assert.False(t, d.Dirty())
	assert.False(t, d.Undo())

	require.True(t, d.Redo())
	assert.Equal(t, 0.3, d.Now().Tracks[1].Pan)
	d.Edit(func(a *synth.Arrangement) { a.Tracks[1].Pan = 0 })
	assert.False(t, d.Redo(), "an edit drops what was undone")

	for i := range UndoDepth + 10 {
		d.Edit(func(a *synth.Arrangement) { a.Tracks[1].Level = float64(i) })
	}
	n := 0
	for d.Undo() {
		n++
	}
	assert.Equal(t, UndoDepth, n)
}

func TestSavingOneDocKeepsItsNeighboursSavable(t *testing.T) {
	d, src := fixture(t)
	lit := src[strings.Index(src, "var waltz"):]
	lit = strings.Replace(strings.Replace(lit, "var waltz", "var other", 1), `ID: "waltz"`, `ID: "other"`, 1)
	require.NoError(t, os.WriteFile(d.Path, []byte(src+"\n"+lit), 0o644))

	docs := NewDocs([]string{filepath.Dir(d.Path)})
	waltz := d.Now()
	other := waltz
	other.ID = "other"
	w, err := docs.Open(waltz)
	require.NoError(t, err)
	o, err := docs.Open(other)
	require.NoError(t, err)

	w.Edit(func(a *synth.Arrangement) { a.Tracks[1].Pan = 0.5 })
	o.Edit(func(a *synth.Arrangement) { a.Tracks[1].Pan = -0.5 })
	require.NoError(t, docs.Save(w))
	assert.Equal(t, []string{"other"}, docs.Dirty())
	require.NoError(t, docs.Save(o))
	assert.Empty(t, docs.Dirty())
}
