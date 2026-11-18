package player

import (
	"testing"
	"time"

	"github.com/cbodonnell/synth"
	"github.com/stretchr/testify/assert"
)

func TestAHeldSoundFadesOverItsRelease(t *testing.T) {
	s := newStream(1, 0)
	buf := synth.Stereo{L: make([]float32, 48000), R: make([]float32, 48000)}
	for i := range buf.L {
		buf.L[i], buf.R[i] = 1, 1
	}
	h := s.sustain(buf, 1)
	pull(t, s, 1000)
	h.Release(100 * time.Millisecond)
	out := pull(t, s, 6000)
	assert.InDelta(t, 4800, sounding(out), 1, "silent once the release is over")
	assert.Greater(t, out[2*2400], float32(0.4), "halfway down halfway through")
	assert.Empty(t, s.voices)
}
