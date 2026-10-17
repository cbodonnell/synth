package player

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pull reads n frames from s, the way Ebiten would, in uneven chunks.
func pull(t *testing.T, s *stream, n int) []float32 {
	out := make([]float32, 0, n*2)
	for chunk := 0; len(out) < n*2; chunk++ {
		b := make([]byte, min(97+chunk%3*500, n-len(out)/2)*bytesPerFrame)
		got, err := s.Read(b)
		require.NoError(t, err)
		require.Equal(t, len(b), got, "a stream never runs dry")
		f := make([]float32, got/4)
		floats(f, b[:got])
		out = append(out, f...)
	}
	return out
}

func clip(frames ...float32) []float32 {
	out := make([]float32, 0, len(frames)*2)
	for _, f := range frames {
		out = append(out, f, f)
	}
	return out
}

func sine(n int, period float64) []float32 {
	out := make([]float32, n*2)
	for i := range n {
		x := float32(math.Sin(2 * math.Pi * float64(i) / period))
		out[i*2], out[i*2+1] = x, x
	}
	return out
}

func sounding(x []float32) int {
	n := 0
	for i := 0; i < len(x); i += 2 {
		if x[i] != 0 {
			n = i/2 + 1
		}
	}
	return n
}

func TestVoicesStartedTogetherStayAligned(t *testing.T) {
	s := newStream(1, 0)
	pull(t, s, 333)
	s.play(newVoice("a", &memory{pcm: clip(1, 0, 0)}, 1, 0.5), 0)
	s.play(newVoice("b", &memory{pcm: clip(1, 0, 0)}, 1, 0.25), 0)
	out := pull(t, s, 10)
	assert.Equal(t, float32(0.75), out[0], "both land on the first frame of the next read")
	assert.Zero(t, out[2])
}

func TestGainChangesRamp(t *testing.T) {
	s := newStream(1, 0)
	v := newVoice("a", &memory{pcm: clip(make([]float32, 4000)...)}, 1, 1)
	for i := range v.src.(*memory).pcm {
		v.src.(*memory).pcm[i] = 1
	}
	s.play(v, 0)
	pull(t, s, 100)
	s.setGain(0)
	out := pull(t, s, 1000)
	limit := float32(1.0/rampFrames) + 1e-6
	for i := 2; i < len(out); i += 2 {
		require.LessOrEqualf(t, abs(out[i]-out[i-2]), limit, "frame %d", i/2)
	}
	assert.Zero(t, out[len(out)-1])
}

func TestStoppedVoiceFadesOut(t *testing.T) {
	s := newStream(1, 0)
	v := newVoice("a", &memory{pcm: sine(48000, 100)}, 1, 1)
	s.play(v, 0)
	pull(t, s, 1000)
	s.stop(v)
	out := pull(t, s, 2*rampFrames)
	assert.InDelta(t, rampFrames, sounding(out), 1, "silent within the ramp")
	assert.Empty(t, s.voices)
}

func TestLoopedSeamIsContinuous(t *testing.T) {
	const intro, body, period = 1000, 2000, 100
	pcm := sine(intro+body+blendFrames, period)
	src := &looped{src: &memory{pcm: pcm}, start: intro, end: intro + body}
	s := newStream(1, 0)
	s.play(newVoice("loop", src, 1, 1), 0)
	out := pull(t, s, intro+5*body)
	for i := range len(out) / 2 {
		j := i
		if i >= intro {
			j = intro + (i-intro)%body
		}
		require.InDeltaf(t, pcm[j*2], out[i*2], 1e-5, "frame %d", i)
	}
}

func TestRateTwoHalvesTheLength(t *testing.T) {
	s := newStream(1, 0)
	s.play(newVoice("a", &memory{pcm: sine(1000, 50)}, 2, 1), 0)
	n := sounding(pull(t, s, 1000))
	assert.InDelta(t, 500, n, 3)
}

func TestRateOneIsExact(t *testing.T) {
	s := newStream(1, 0)
	pcm := sine(500, 37)
	s.play(newVoice("a", &memory{pcm: pcm}, 1, 1), 0)
	assert.Equal(t, pcm, pull(t, s, 500))
}

func TestASoundPastItsLimitReplacesItsOldest(t *testing.T) {
	s := newStream(1, 0)
	var vs []*voice
	for range 3 {
		v := newVoice("step", &memory{pcm: sine(48000, 100)}, 1, 1)
		vs = append(vs, v)
		require.True(t, s.play(v, 2))
		pull(t, s, 10)
	}
	assert.True(t, vs[0].stopping)
	assert.False(t, vs[1].stopping)
	assert.False(t, vs[2].stopping)
}

func TestHighPrioritySurvivesAFlood(t *testing.T) {
	s := newStream(1, 4)
	fanfare := newVoice("fanfare", &memory{pcm: sine(48000, 100)}, 1, 1)
	fanfare.priority = 2
	s.play(fanfare, 0)
	for range 20 {
		v := newVoice("coin", &memory{pcm: sine(48000, 100)}, 1, 1)
		s.play(v, 0)
		pull(t, s, 10)
	}
	assert.False(t, fanfare.stopping)
	assert.Len(t, s.live(), 4)

	low := newVoice("dust", &memory{pcm: sine(48000, 100)}, 1, 1)
	low.priority = -1
	assert.False(t, s.play(low, 0), "a sound lower than everything playing is dropped")
}

func TestVaryNeverRepeatsATakeAndStaysInRange(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	v := synth.Variation{Pitch: 2, Gain: 3, Takes: 3}
	last := -1
	for range 500 {
		take, rate, gain := vary(v, rng, last)
		require.NotEqual(t, last, take)
		require.True(t, take >= 0 && take < 3)
		require.InDelta(t, 0, 12*math.Log2(rate), 2)
		require.InDelta(t, 0, 20*math.Log10(gain), 3)
		last = take
	}

	take, rate, gain := vary(synth.Variation{}, rng, -1)
	assert.Equal(t, 0, take)
	assert.Equal(t, 1.0, rate)
	assert.Equal(t, 1.0, gain)
}

func abs(x float32) float32 {
	return float32(math.Abs(float64(x)))
}
