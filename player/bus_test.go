package player

import (
	"math"
	"testing"
	"time"

	"github.com/cbodonnell/synth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func constant(n int, x float32) []float32 {
	out := make([]float32, n*2)
	for i := range out {
		out[i] = x
	}
	return out
}

func peakOf(x []float32) float32 {
	var p float32
	for _, f := range x {
		p = max(p, abs(f))
	}
	return p
}

func TestBusSumsMusicAndEffects(t *testing.T) {
	music, effects := newStream(1, 0), newStream(1, 0)
	b := newBus(music, effects, defaultCeiling)
	music.play(newVoice("song", &memory{pcm: clip(0.25, 0.25)}, 1, 1), 0)
	effects.play(newVoice("fx", &memory{pcm: clip(0.5)}, 1, 1), 0)
	out := pull(t, b, 3)
	assert.Equal(t, clip(0.75, 0.25, 0), out, "under the ceiling the mix is untouched")
}

func TestAPileUpStaysUnderTheCeiling(t *testing.T) {
	music, effects := newStream(1, 0), newStream(1, 16)
	b := newBus(music, effects, defaultCeiling)
	music.play(newVoice("song", &memory{pcm: sine(48000, 220)}, 1, 0.6), 0)
	for i := range 4 {
		effects.play(newVoice("fanfare", &memory{pcm: sine(48000, 100+float64(i)*7)}, 1, 0.7), 0)
	}
	out := pull(t, b, 24000)
	ceiling := float32(math.Pow(10, defaultCeiling/20))
	assert.LessOrEqual(t, peakOf(out), ceiling+1e-6)
	assert.Greater(t, peakOf(out), ceiling*0.9, "limited, not silenced")
}

func TestTheLimiterLetsGo(t *testing.T) {
	l := newLimiter(defaultCeiling, limiterRelease)
	l.process(constant(10, 4))
	require.Less(t, l.gain, float32(0.25))

	quiet := constant(synth.SampleRate, 0.1)
	l.process(quiet)
	assert.InDelta(t, 1, l.gain, 0.01, "recovered within a second")
	assert.InDelta(t, 0.1, quiet[len(quiet)-1], 0.001)
}

func TestHeldMusicStandsStill(t *testing.T) {
	music := newStream(0, 0)
	music.hold = true
	b := newBus(music, newStream(1, 0), defaultCeiling)
	pcm := sine(1000, 50)
	for i := range pcm {
		pcm[i] *= 0.5
	}
	v := newVoice("song", &memory{pcm: pcm}, 1, 1)
	music.play(v, 0)

	pull(t, b, 500)
	assert.True(t, music.held())
	assert.Zero(t, music.frames(), "nothing mixed at gain 0")
	_, started := music.started(v)
	assert.False(t, started)

	music.setGain(1)
	out := pull(t, b, 500)
	assert.Equal(t, pcm[rampFrames*2:500*2], out[rampFrames*2:], "picks up from the start once audible")

	m, mixed := b.musicAt()
	assert.Equal(t, int64(500), m)
	assert.Equal(t, int64(1000), mixed)
}

func TestAheadReadsItsSourceExactly(t *testing.T) {
	pcm := sine(aheadFrames*3+123, 41)
	a := newAhead(&memory{pcm: pcm})
	var got []float32
	for chunk := 0; ; chunk++ {
		buf := make([]float32, (97+chunk%3*700)*2)
		n := a.read(buf)
		got = append(got, buf[:n*2]...)
		if n < len(buf)/2 {
			break
		}
	}
	assert.Equal(t, pcm, got)
}

func TestStoppedMusicLetsItsDecoderGo(t *testing.T) {
	const intro, body = 100, 1000
	pcm := sine(intro+body+blendFrames, 50)
	a := newAhead(&looped{src: &memory{pcm: pcm}, start: intro, end: intro + body})
	s := newStream(1, 0)
	v := newVoice("song", a, 1, 1)
	s.play(v, 0)
	pull(t, s, 5000)
	s.stop(v)
	pull(t, s, 2*rampFrames)
	require.Empty(t, s.voices)

	deadline := time.After(time.Second)
	for {
		select {
		case _, ok := <-a.full:
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("the decoder is still running")
		}
	}
}
