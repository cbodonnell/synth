package bake

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tone(id string, loop bool) synth.Song {
	click := synth.Envelope{Attack: 0.001, Decay: 0.01}
	p := synth.Patch{
		Oscs: []synth.Osc{{Wave: synth.Saw}, {Wave: synth.Noise, Level: 0.1, Env: &click}},
		Amp:  synth.Envelope{Attack: 0.01, Decay: 0.2, Sustain: 0.5, Release: 0.1},
	}
	s := synth.Song{
		ID: id,
		Parts: []synth.Part{{Name: "a", Patch: p,
			Notes: synth.Phrase{Step: 0.25}.Notes("C4 E4 G4 C5")}},
		Buses:  []synth.Bus{{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.3}}}},
		Master: []synth.Effect{synth.Limiter{}},
		Length: 1,
	}
	if loop {
		s.Loop, s.LoopStart = true, 0.25
	}
	return s
}

func TestFingerprintFollowsTheSongNotItsPointers(t *testing.T) {
	a, b := tone("a", true), tone("a", true)
	assert.Equal(t, Fingerprint(a), Fingerprint(b))

	b.Parts[0].Patch.Oscs[1].Env.Decay = 0.02
	assert.NotEqual(t, Fingerprint(a), Fingerprint(b))

	c := tone("a", true)
	c.Master = []synth.Effect{synth.Limiter{Ceiling: -3}}
	assert.NotEqual(t, Fingerprint(a), Fingerprint(c))
}

func decodedLen(t *testing.T, path string) int {
	f, err := os.ReadFile(path)
	require.NoError(t, err)
	d, err := vorbis.DecodeF32(bytes.NewReader(f))
	require.NoError(t, err)
	pcm, err := io.ReadAll(d)
	require.NoError(t, err)
	return len(pcm) / 8
}

func TestBakeIsSampleExactAndTracksStaleness(t *testing.T) {
	if _, err := exec.LookPath("oggenc"); err != nil {
		t.Skip("oggenc not installed")
	}
	dir := t.TempDir()
	loop, shot := tone("loop", true), tone("shot", false)
	require.NoError(t, Bake(dir, loop, shot))

	m, err := Load(os.DirFS(dir))
	require.NoError(t, err)
	l := synth.RenderLoop(loop)
	e := m["loop"]
	start := l.Intro.Len() - l.Body.Len()
	assert.GreaterOrEqual(t, e.Intro, start)
	assert.Less(t, e.Intro, l.Intro.Len(), "the intro stops before a second full pass")
	assert.Equal(t, l.Body.Len(), e.Body)
	assert.Equal(t, e.Intro+e.Body+min(overlap, e.Body), decodedLen(t, filepath.Join(dir, "loop.ogg")))
	assert.Equal(t, synth.Render(shot).Len(), decodedLen(t, filepath.Join(dir, "shot.ogg")))

	stale, err := Stale(os.DirFS(dir), loop, shot)
	require.NoError(t, err)
	assert.Empty(t, stale)

	shot.Parts[0].Gain = 0.5
	stale, err = Stale(os.DirFS(dir), loop, shot, tone("new", false))
	require.NoError(t, err)
	assert.Equal(t, []string{"new", "shot"}, stale)

	info, err := os.Stat(filepath.Join(dir, "loop.ogg"))
	require.NoError(t, err)
	require.NoError(t, Bake(dir, loop))
	again, err := os.Stat(filepath.Join(dir, "loop.ogg"))
	require.NoError(t, err)
	assert.Equal(t, info.ModTime(), again.ModTime(), "a fresh bake is left alone")
	assert.NoFileExists(t, filepath.Join(dir, "shot.ogg"))
}

func TestSettleFindsWhereTwoBuffersMeet(t *testing.T) {
	a, b := synth.NewStereo(10), synth.NewStereo(10)
	assert.Equal(t, 0, settle(a, b))
	a.L[3] = 0.5
	assert.Equal(t, 4, settle(a, b))
	b.R[9] = 0.5
	assert.Equal(t, 10, settle(a, b))
}

func TestBakedLoopCarriesOnIntoItsStart(t *testing.T) {
	if _, err := exec.LookPath("oggenc"); err != nil {
		t.Skip("oggenc not installed")
	}
	dir := t.TempDir()
	loop := tone("loop", true)
	require.NoError(t, Bake(dir, loop))
	m, err := Load(os.DirFS(dir))
	require.NoError(t, err)
	e := m["loop"]

	f, err := os.ReadFile(filepath.Join(dir, e.Files[0]))
	require.NoError(t, err)
	d, err := vorbis.DecodeF32(bytes.NewReader(f))
	require.NoError(t, err)
	pcm, err := io.ReadAll(d)
	require.NoError(t, err)
	at := func(i int) float32 { return math.Float32frombits(binary.LittleEndian.Uint32(pcm[i*4:])) }

	var worst float64
	for k := range 2 * 256 {
		worst = max(worst, math.Abs(float64(at(2*(e.Intro+e.Body)+k)-at(2*e.Intro+k))))
	}
	assert.Less(t, worst, 0.02, "past the end, the file is the loop's start again")
}

func TestBakeWritesEveryTake(t *testing.T) {
	if _, err := exec.LookPath("oggenc"); err != nil {
		t.Skip("oggenc not installed")
	}
	dir := t.TempDir()
	shot := tone("shot", false)
	shot.Vary.Takes = 3
	require.NoError(t, Bake(dir, shot))
	m, err := Load(os.DirFS(dir))
	require.NoError(t, err)
	require.Len(t, m["shot"].Files, 3)
	for _, f := range m["shot"].Files {
		assert.FileExists(t, filepath.Join(dir, f))
	}

	shot.Vary.Takes = 1
	require.NoError(t, Bake(dir, shot))
	assert.NoFileExists(t, filepath.Join(dir, "shot.2.ogg"), "a take no longer wanted is removed")
}
