package player

import "github.com/cbodonnell/synth"

// No separate mute flag: a level of 0 is off.
type Levels struct {
	Master  float64
	Music   float64
	Effects float64
}

var DefaultLevels = Levels{Master: 0.5, Music: 0.5, Effects: 0.5}

// Mix trim under the Music level, so music always sits under effects.
const musicBalance = 0.6

// Mixer is the one place a game touches audio: what's playing, how loud, and where the
// beat is. Use it from the game loop only; it isn't safe for concurrent use.
type Mixer struct {
	sink    Sink
	levels  Levels
	current synth.Song
	restore float64
}

func NewMixer(sink Sink) *Mixer {
	return &Mixer{sink: sink, levels: DefaultLevels, restore: DefaultLevels.Master}
}

// PlayMusic is a no-op if s is already playing; a Song with no ID stops the music.
func (m *Mixer) PlayMusic(s synth.Song) {
	if s.ID == "" {
		m.StopMusic()
		return
	}
	if s.ID == m.current.ID {
		return
	}
	m.current = s
	m.sink.PlayMusic(s, m.musicGain())
}

func (m *Mixer) StopMusic() {
	if m.current.ID == "" {
		return
	}
	m.current = synth.Song{}
	m.sink.StopMusic()
}

func (m *Mixer) Current() string { return m.current.ID }

func (m *Mixer) Play(s synth.Song) {
	if g := m.levels.Master * m.levels.Effects; g > 0 {
		m.sink.Play(s, g)
	}
}

func (m *Mixer) Levels() Levels { return m.levels }

func (m *Mixer) SetLevels(l Levels) {
	m.levels = Levels{Master: clampLevel(l.Master), Music: clampLevel(l.Music), Effects: clampLevel(l.Effects)}
	m.sink.SetMusicGain(m.musicGain())
}

func (m *Mixer) ToggleMute() {
	l := m.levels
	if !m.Muted() {
		m.restore = l.Master
		l.Master = 0
	} else {
		l.Master = m.restore
		if l.Master <= 0 {
			l.Master = DefaultLevels.Master
		}
	}
	m.SetLevels(l)
}

func (m *Mixer) Muted() bool { return m.levels.Master <= 0 }

func (m *Mixer) musicGain() float64 { return m.levels.Master * m.levels.Music * musicBalance }

// SongTime is where the music is in its score, loops unwound.
func (m *Mixer) SongTime() (float64, bool) {
	if m.current.ID == "" {
		return 0, false
	}
	pos, ok := m.sink.MusicPosition()
	if !ok || pos < 0 {
		return 0, false
	}
	return m.current.SongTime(pos.Seconds()), true
}

// Beat reads the current song's meter at its playhead, for animating on the beat.
// False means silence or no position yet; a game should keep its own clock then.
func (m *Mixer) Beat() (synth.Position, bool) {
	t, ok := m.SongTime()
	if !ok || m.current.Meter.Beat <= 0 {
		return synth.Position{}, false
	}
	return m.current.Meter.At(t), true
}

func clampLevel(v float64) float64 {
	return max(0, min(1, v))
}
