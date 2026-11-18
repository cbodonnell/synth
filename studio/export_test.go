package studio

import (
	"cmp"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/songs"
	"github.com/cbodonnell/synth/studio/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sameNotes asserts got plays want's notes, each within drift seconds.
func sameNotes(t *testing.T, want, got []synth.Note, drift float64, msg string) {
	t.Helper()
	require.Lenf(t, got, len(want), "%s: notes", msg)
	byAt := func(a, b synth.Note) int { return cmp.Compare(a.At, b.At) }
	want, got = slices.Clone(want), slices.Clone(got)
	slices.SortStableFunc(want, byAt)
	slices.SortStableFunc(got, byAt)
	used := make([]bool, len(got))
	for i, w := range want {
		k := -1
		for j, g := range got {
			if !used[j] && g.Key == w.Key && math.Abs(g.At-w.At) <= drift+1e-6 {
				k = j
				break
			}
		}
		require.GreaterOrEqualf(t, k, 0, "%s: note %d (%v) missing", msg, i, w)
		used[k] = true
		g := got[k]
		require.InDeltaf(t, w.Dur, g.Dur, 0.02, "%s: note %d dur", msg, i)
		require.InDeltaf(t, w.Slide, g.Slide, 1e-9, "%s: note %d slide", msg, i)
	}
}

func TestArrangeKeepsEachSongsNotes(t *testing.T) {
	for _, s := range songs.Songs() {
		a := Arrange(s)
		require.NoError(t, a.Validate(), s.ID)
		got := a.Song()
		assert.InDelta(t, s.Length, got.Length, 1e-6, s.ID)
		assert.InDelta(t, s.LoopStart, got.LoopStart, 1e-6, s.ID)
		require.Len(t, got.Parts, len(s.Parts), s.ID)
		for i, p := range s.Parts {
			sameNotes(t, p.Notes, got.Parts[i].Notes, humanize(a.Tracks[i]), s.ID+" "+p.Name)
		}
	}
}

func humanize(t synth.Track) float64 {
	h := 0.0
	for _, c := range t.Clips {
		h = math.Max(h, c.Humanize)
	}
	return h
}

// The exported source compiles, and the song it builds plays the original's notes.
func TestExportCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a program")
	}
	require.NoError(t, os.MkdirAll("testdata", 0o755))
	dir, err := os.MkdirTemp("testdata", "export-")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
		os.Remove("testdata")
	})

	var main strings.Builder
	main.WriteString("package main\n\nimport (\n\t\"encoding/json\"\n\t\"os\"\n\n\t\"github.com/cbodonnell/synth\"\n)\n\n")
	main.WriteString("func main() {\n\tjson.NewEncoder(os.Stdout).Encode([]synth.Song{\n")
	for _, s := range songs.Songs() {
		src, err := session.Source(Arrange(s), "main")
		require.NoError(t, err, string(src))
		require.NoError(t, os.WriteFile(filepath.Join(dir, s.ID+".go"), src, 0o644))
		main.WriteString("\t\t" + session.Ident(s.ID, true) + "(),\n")
	}
	main.WriteString("\t})\n}\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte(main.String()), 0o644))

	out, err := exec.Command("go", "run", "./"+dir).Output()
	if ee, ok := err.(*exec.ExitError); ok {
		t.Fatalf("%v\n%s", err, ee.Stderr)
	}
	require.NoError(t, err)
	var built []json.RawMessage
	require.NoError(t, json.Unmarshal(out, &built))
	for i, s := range songs.Songs() {
		var got struct {
			Parts []struct{ Notes []synth.Note }
		}
		require.NoError(t, json.Unmarshal(built[i], &got))
		a := Arrange(s)
		for k, p := range s.Parts {
			sameNotes(t, p.Notes, got.Parts[k].Notes, humanize(a.Tracks[k]), s.ID+" "+p.Name)
		}
	}
}
