// Package bake renders songs ahead of time to Ogg Vorbis, so a game decodes them instead of
// synthesizing them: a full band costs seconds to render, and far more on one WebAssembly
// thread, while decoding costs almost nothing. The songs stay the source; a Manifest
// fingerprints each one, so a bake that no longer matches its song is caught by Stale in a
// test and passed over by the player, which renders that song live instead.
//
// Encoding shells out to oggenc (vorbis-tools). Only baking needs it; playing doesn't.
package bake

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cbodonnell/synth"
)

const ManifestFile = "bake.json"

// Past a loop's end the file carries the start of the body again: the player blends its
// first 256 samples into the loop start, which hides the codec's seam.
const overlap = 2048

const quality = "6"

type Entry struct {
	Files       []string `json:"files"` // one per take of a sound with Vary.Takes
	Fingerprint string   `json:"fingerprint"`
	// Samples played once before the body, and the body's length; Body 0 is a one-shot.
	// Intro runs past the song's LoopStart by however long the first pass takes to settle
	// into the loop, so the body in the file starts that far into the song's loop.
	Intro int `json:"intro,omitempty"`
	Body  int `json:"body,omitempty"`
}

type Manifest map[string]Entry

func Load(fsys fs.FS) (Manifest, error) {
	b, err := fs.ReadFile(fsys, ManifestFile)
	if err != nil {
		return nil, err
	}
	m := Manifest{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("bake: %s: %w", ManifestFile, err)
	}
	return m, nil
}

// Fresh reports the entry baked from exactly this song.
func (m Manifest) Fresh(s synth.Song) (Entry, bool) {
	e, ok := m[s.ID]
	return e, ok && e.Fingerprint == Fingerprint(s)
}

// Bake writes every song to dir and removes what no song claims any more. A song whose
// bake is already fresh is left alone, so rebaking only touches what changed.
func Bake(dir string, songs ...synth.Song) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	old, err := Load(os.DirFS(dir))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	m := Manifest{}
	for _, s := range songs {
		if s.ID == "" {
			return errors.New("bake: a song needs an ID")
		}
		if _, dup := m[s.ID]; dup {
			return fmt.Errorf("bake: %q twice", s.ID)
		}
		if err := synth.Validate(s); err != nil {
			return fmt.Errorf("bake: %s: %w", s.ID, err)
		}
		if e, ok := old.Fresh(s); ok && present(os.DirFS(dir), e) {
			m[s.ID] = e
			continue
		}
		e, err := bake(dir, s)
		if err != nil {
			return fmt.Errorf("bake: %s: %w", s.ID, err)
		}
		m[s.ID] = e
	}
	kept := map[string]bool{}
	for _, e := range m {
		for _, f := range e.Files {
			kept[f] = true
		}
	}
	for _, e := range old {
		for _, f := range e.Files {
			if !kept[f] {
				os.Remove(filepath.Join(dir, f))
			}
		}
	}
	b, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ManifestFile), append(b, '\n'), 0o644)
}

func bake(dir string, s synth.Song) (Entry, error) {
	e := Entry{Fingerprint: Fingerprint(s)}
	if s.Loop {
		l := synth.RenderLoop(s)
		start := l.Intro.Len() - l.Body.Len()
		first := slice(l.Intro, start, l.Intro.Len())
		// The first pass and the body differ only while the intro's tail rings into one
		// and the loop's own tail wraps onto the other. Past that the file can turn into
		// the body, so it's the intro and one pass, not two; the body is rotated to begin
		// where they meet, and the loop still lands on the same music.
		t := settle(first, l.Body)
		body := synth.Concat(slice(l.Body, t, l.Body.Len()), slice(l.Body, 0, t))
		e.Intro, e.Body = start+t, body.Len()
		buf := synth.Concat(slice(l.Intro, 0, start+t), body, slice(body, 0, min(overlap, body.Len())))
		e.Files = []string{s.ID + ".ogg"}
		return e, encode(filepath.Join(dir, e.Files[0]), buf)
	}
	for take := range max(s.Vary.Takes, 1) {
		file := s.ID + ".ogg"
		if take > 0 {
			file = fmt.Sprintf("%s.%d.ogg", s.ID, take)
		}
		e.Files = append(e.Files, file)
		if err := encode(filepath.Join(dir, file), synth.Render(synth.Take(s, take))); err != nil {
			return e, err
		}
	}
	return e, nil
}

func encode(path string, buf synth.Stereo) error {
	wav, err := os.CreateTemp("", "bake-*.wav")
	if err != nil {
		return err
	}
	defer os.Remove(wav.Name())
	if _, err := wav.Write(synth.WAV(buf)); err != nil {
		wav.Close()
		return err
	}
	if err := wav.Close(); err != nil {
		return err
	}
	out, err := exec.Command("oggenc", "-Q", "-q", quality, "-o", path, wav.Name()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("oggenc (brew install vorbis-tools): %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func slice(s synth.Stereo, from, to int) synth.Stereo {
	return synth.Stereo{L: s.L[from:to], R: s.R[from:to]}
}

// Under 16-bit resolution: past here, the first pass and the body are the same music.
const agree = 1.0 / 32768

// settle is the first sample from which a and b agree to the end.
func settle(a, b synth.Stereo) int {
	for i := min(a.Len(), b.Len()) - 1; i >= 0; i-- {
		if math.Abs(float64(a.L[i]-b.L[i])) > agree || math.Abs(float64(a.R[i]-b.R[i])) > agree {
			return i + 1
		}
	}
	return 0
}

// Stale lists the songs whose bake is missing or no longer matches them, and the baked
// IDs no song claims: everything a rebake would change.
func Stale(fsys fs.FS, songs ...synth.Song) ([]string, error) {
	m, err := Load(fsys)
	if err != nil {
		return nil, err
	}
	var stale []string
	seen := map[string]bool{}
	for _, s := range songs {
		seen[s.ID] = true
		if e, ok := m.Fresh(s); !ok || !present(fsys, e) {
			stale = append(stale, s.ID)
		}
	}
	for id := range m {
		if !seen[id] {
			stale = append(stale, id)
		}
	}
	slices.Sort(stale)
	return stale, nil
}

func present(fsys fs.FS, e Entry) bool {
	for _, f := range e.Files {
		if _, err := fs.Stat(fsys, f); err != nil {
			return false
		}
	}
	return len(e.Files) > 0
}
