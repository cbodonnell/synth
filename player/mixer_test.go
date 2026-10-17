package player

import (
	"testing"
	"time"

	"github.com/cbodonnell/synth"
	"github.com/stretchr/testify/assert"
)

type recorder struct {
	played  []string
	gains   []float64
	effects int
	stopped int
	pos     time.Duration
	playing bool
}

func (r *recorder) Play(synth.Song, float64) { r.effects++ }
func (r *recorder) PlayMusic(s synth.Song, g float64) {
	r.played = append(r.played, s.ID)
	r.gains = append(r.gains, g)
}
func (r *recorder) StopMusic()             { r.stopped++ }
func (r *recorder) SetMusicGain(g float64) { r.gains = append(r.gains, g) }
func (r *recorder) MusicPosition() (time.Duration, bool) {
	return r.pos, r.playing
}
func (r *recorder) lastGain() float64 { return r.gains[len(r.gains)-1] }

func song(id string) synth.Song {
	return synth.Song{
		ID: id, Length: 10, Loop: true, LoopStart: 4,
		Meter: synth.Meter{Beat: 1, Bars: []int{3}},
	}
}

func TestPlayMusicIgnoresTheSongAlreadyPlaying(t *testing.T) {
	r := &recorder{}
	m := NewMixer(r)
	m.PlayMusic(song("a"))
	m.PlayMusic(song("a"))
	m.PlayMusic(song("b"))
	assert.Equal(t, []string{"a", "b"}, r.played)
}

func TestPlayMusicWithNoIDStops(t *testing.T) {
	r := &recorder{}
	m := NewMixer(r)
	m.PlayMusic(song("a"))
	m.PlayMusic(synth.Song{})
	assert.Equal(t, 1, r.stopped)
	assert.Empty(t, m.Current())
}

func TestMuteSilencesAndRestores(t *testing.T) {
	r := &recorder{}
	m := NewMixer(r)
	m.SetLevels(Levels{Master: 0.8, Music: 1, Effects: 1})
	m.PlayMusic(song("a"))
	before := r.lastGain()

	m.ToggleMute()
	assert.True(t, m.Muted())
	assert.Zero(t, r.lastGain())
	m.Play(song("fx"))
	assert.Zero(t, r.effects, "a muted mixer plays no effects")

	m.ToggleMute()
	assert.Equal(t, before, r.lastGain())
}

func TestLevelsClamp(t *testing.T) {
	m := NewMixer(Silent{})
	m.SetLevels(Levels{Master: 2, Music: -1, Effects: 0.5})
	assert.Equal(t, Levels{Master: 1, Music: 0, Effects: 0.5}, m.Levels())
}

func TestBeatFollowsTheLoopedPlayhead(t *testing.T) {
	r := &recorder{}
	m := NewMixer(r)

	_, ok := m.Beat()
	assert.False(t, ok, "no music")

	m.PlayMusic(song("a"))
	_, ok = m.Beat()
	assert.False(t, ok, "not playing yet")

	r.playing, r.pos = true, 11500*time.Millisecond
	pos, ok := m.Beat()
	assert.True(t, ok)
	assert.Equal(t, synth.Position{Bar: 1, Beat: 2, Phase: 0.5}, pos, "11.5s played is 5.5s into the score")
}

func TestPlayGainSkipsSilentEffects(t *testing.T) {
	r := &recorder{}
	m := NewMixer(r)
	m.PlayGain(song("a"), 0)
	m.PlayGain(song("a"), -1)
	assert.Zero(t, r.effects)
	m.PlayGain(song("a"), 0.5)
	assert.Equal(t, 1, r.effects)
}
