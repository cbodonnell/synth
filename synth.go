// Package synth renders music and sound effects for Ebiten games from plain data.
// A Patch describes an instrument, a Part plays Notes through it, and a Song mixes
// Parts through effect buses. Render is pure: the same Song always renders the same
// samples, because nothing in it reads a clock or math/rand.
package synth

import "math"

const SampleRate = 48000

// Revision changes whenever rendering changes what a Song sounds like, so a baked song
// knows it must be baked again.
const Revision = 1

// Stereo is a rendered buffer. L and R are always the same length.
type Stereo struct {
	L, R []float32
}

func NewStereo(n int) Stereo {
	return Stereo{L: make([]float32, n), R: make([]float32, n)}
}

func (s Stereo) Len() int { return len(s.L) }

func (s Stereo) Seconds() float64 { return float64(s.Len()) / SampleRate }

func (s Stereo) Peak() float32 {
	var peak float32
	for i := range s.L {
		peak = max(peak, abs32(s.L[i]), abs32(s.R[i]))
	}
	return peak
}

func (s Stereo) add(src Stereo, gain float32) {
	n := min(s.Len(), src.Len())
	for i := range n {
		s.L[i] += src.L[i] * gain
		s.R[i] += src.R[i] * gain
	}
}

func samples(seconds float64) int {
	if seconds <= 0 {
		return 0
	}
	return int(math.Round(seconds * SampleRate))
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func clamp(x, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, x))
}

// Equal-power, so a sound swept across the field keeps its loudness.
func panGains(pan float64) (l, r float64) {
	a := (clamp(pan, -1, 1) + 1) * math.Pi / 4
	return math.Cos(a), math.Sin(a)
}

func dbToGain(db float64) float64 { return math.Pow(10, db/20) }

// splitmix64: a fixed, seedable source so noise and humanizing are reproducible.
type rng struct{ s uint64 }

func newRNG(seed uint64) *rng { return &rng{s: seed} }

func (r *rng) next() uint64 {
	r.s += 0x9e3779b97f4a7c15
	z := r.s
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// float returns a value in [-1, 1).
func (r *rng) float() float64 {
	return float64(r.next()>>11)/(1<<52) - 1
}

func mixSeed(parts ...uint64) uint64 {
	h := uint64(0xcbf29ce484222325)
	for _, p := range parts {
		h ^= p
		h *= 0x100000001b3
		h ^= h >> 29
	}
	return h
}
