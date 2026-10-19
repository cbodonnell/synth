package player

import (
	"math"
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stereo(x []float32) synth.Stereo {
	s := synth.NewStereo(len(x) / 2)
	for i := range s.L {
		s.L[i], s.R[i] = x[i*2], x[i*2+1]
	}
	return s
}

// mixPreview reads n frames of v in uneven chunks, as the bus does.
func mixPreview(v *Preview, n int) []float32 {
	out := make([]float32, 0, n*2)
	for chunk := 0; len(out) < n*2; chunk++ {
		buf := make([]float32, min(97+chunk%3*500, n-len(out)/2)*2)
		v.mix(buf)
		out = append(out, buf...)
	}
	return out
}

func biggestStep(x []float32) float32 {
	var step float32
	for i := 2; i < len(x); i += 2 {
		step = max(step, abs(x[i]-x[i-2]))
	}
	return step
}

func TestPreviewSwapsWithoutAStep(t *testing.T) {
	v := newPreview()
	v.Play(stereo(constant(10000, 1)), 0, 0, 0)
	mixPreview(v, 1000)
	v.Swap(stereo(constant(10000, -1)), 0, 0, nil)
	out := mixPreview(v, 1000)
	assert.Equal(t, float32(-1), out[len(out)-1], "the new buffer takes over")
	assert.LessOrEqual(t, biggestStep(out), float32(2.0/rampFrames)+1e-6)

	v.Stop()
	out = mixPreview(v, 1000)
	assert.Zero(t, out[len(out)-1])
	assert.LessOrEqual(t, biggestStep(out), float32(1.0/rampFrames)+1e-6)
}

func TestPreviewLoopsItsRegionSeamlessly(t *testing.T) {
	const period, start, end = 100.0, 1000, 3000
	v := newPreview()
	v.Play(stereo(sine(4000, period)), start-50, start, end)
	out := mixPreview(v, 10000)[rampFrames*2:]
	for i := 0; i < len(out); i += 2 {
		want := float32(math.Sin(2 * math.Pi * float64(start-50+rampFrames+i/2) / period))
		require.InDeltaf(t, want, out[i], 1e-4, "frame %d", i/2)
	}
}

func TestPreviewPositionFollowsTheLoop(t *testing.T) {
	v := newPreview()
	heard := int64(0)
	v.heard = func() int64 { return heard }
	v.Play(stereo(constant(4000, 0.5)), 2500, 1000, 3000)
	pos, ok := v.Position()
	require.True(t, ok)
	assert.Equal(t, 2500, pos, "where it starts, until it's heard")
	mixPreview(v, 2000)
	for at, want := range map[int64]int{0: 2500, 400: 2900, 500: 1000, 1999: 2499} {
		heard = at
		pos, ok := v.Position()
		require.True(t, ok)
		assert.Equal(t, want, pos, "heard %d", at)
	}

	v.Play(stereo(constant(100, 0.5)), 0, 0, 0)
	mixPreview(v, 200)
	heard = 2200
	_, ok = v.Position()
	assert.False(t, ok, "a buffer that doesn't loop plays out")
}
