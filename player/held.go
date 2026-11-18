package player

import (
	"time"

	"github.com/cbodonnell/synth"
)

// Held is a sound playing until it's let go, as a key held on the studio's keyboard is.
type Held struct {
	s *stream
	v *voice
}

// Hold plays buf among the effects until Release. It's for a note rendered longer than
// it's likely to be held, which a release then cuts short.
func (p *Player) Hold(buf synth.Stereo, gain float64) *Held { return p.effects.sustain(buf, gain) }

func (s *stream) sustain(buf synth.Stereo, gain float64) *Held {
	v := newVoice("", &memory{pcm: interleave(buf)}, 1, gain)
	v.priority = 1 << 30
	s.play(v, 0)
	return &Held{s, v}
}

// Release fades the sound out over d, at least the few milliseconds any stop takes.
func (h *Held) Release(d time.Duration) {
	frames := max(int(d.Seconds()*synth.SampleRate), rampFrames)
	h.s.mu.Lock()
	defer h.s.mu.Unlock()
	h.v.stopping = true
	h.v.gain.ramp(0, frames)
	if h.v.start < 0 {
		h.v.done = true
	}
}
