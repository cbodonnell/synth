package synth

import (
	"cmp"
	"math"
	"slices"
)

// Point is one corner of an automation curve, in the unit of whatever it automates.
type Point struct {
	At    float64 // seconds into the song
	Value float64
}

// automated is an effect that follows a curve through song time.
type automated interface{ points() []Point }

// follow is the line through p at t, holding its first and last values beyond its ends.
// lerp moves between neighbors, so each automation interpolates in its own domain.
func follow(p []Point, t float64, lerp func(a, b, x float64) float64) float64 {
	switch {
	case t <= p[0].At:
		return p[0].Value
	case t >= p[len(p)-1].At:
		return p[len(p)-1].Value
	}
	k := 1
	for p[k].At < t {
		k++
	}
	x := (t - p[k-1].At) / math.Max(p[k].At-p[k-1].At, 1e-9)
	return lerp(p[k-1].Value, p[k].Value, x)
}

func linear(a, b, x float64) float64 { return a + (b-a)*x }

func exponential(a, b, x float64) float64 {
	a, b = math.Max(a, 1), math.Max(b, 1)
	return a * math.Pow(b/a, x)
}

func pointsInOrder(p []Point) bool {
	return slices.IsSortedFunc(p, func(a, b Point) int { return cmp.Compare(a.At, b.At) })
}

// AutoFilter is a filter whose cutoff follows Points (Hz) over song time, moving
// exponentially between them: the slow sweep that opens a build.
type AutoFilter struct {
	Mode      FilterMode
	Resonance float64
	Points    []Point // Hz, ascending in At
}

func (a AutoFilter) points() []Point { return a.Points }

func (a AutoFilter) Tail() float64 { return 0.05 }

func (a AutoFilter) cutoff(t float64) float64 {
	if len(a.Points) == 0 {
		return 0.45 * SampleRate
	}
	return follow(a.Points, t, exponential)
}

func (a AutoFilter) Process(s Stereo) {
	var l, r svf
	for i := range s.L {
		if i%filterEvery == 0 {
			c := a.cutoff(float64(i) / SampleRate)
			l.set(c, a.Resonance)
			r.set(c, a.Resonance)
		}
		s.L[i] = float32(l.process(float64(s.L[i]), a.Mode))
		s.R[i] = float32(r.process(float64(s.R[i]), a.Mode))
	}
}

// faderFloor is where a Fader falls silent. A fade from -Inf moves from here, since
// no line in dB can leave -Inf.
const faderFloor = -80

// Fader moves the level along Points (dB) over song time, linearly in dB: fades,
// swells, and, as a bus's first effect, a send opening into a drop.
type Fader struct {
	Points []Point // dB, ascending in At; -Inf is silence
}

func (f Fader) points() []Point { return f.Points }

func (f Fader) Tail() float64 { return 0 }

func (f Fader) gain(t float64) float64 {
	db := follow(f.Points, t, func(a, b, x float64) float64 {
		return linear(math.Max(a, faderFloor), math.Max(b, faderFloor), x)
	})
	if db <= faderFloor {
		return 0
	}
	return dbToGain(db)
}

func (f Fader) Process(s Stereo) {
	if len(f.Points) == 0 {
		return
	}
	for i := range s.L {
		g := float32(f.gain(float64(i) / SampleRate))
		s.L[i] *= g
		s.R[i] *= g
	}
}

// Panner moves the stereo image along Points (-1 left to 1 right) over song time. It
// balances rather than folds: centered is unity, and a side's own content fades as
// the image leaves it, with equal power so a drift never dips in level.
type Panner struct {
	Points []Point // -1..1, ascending in At
}

func (p Panner) points() []Point { return p.Points }

func (p Panner) Tail() float64 { return 0 }

func (p Panner) Process(s Stereo) {
	if len(p.Points) == 0 {
		return
	}
	for i := range s.L {
		l, r := panGains(follow(p.Points, float64(i)/SampleRate, linear))
		s.L[i] *= float32(l * math.Sqrt2)
		s.R[i] *= float32(r * math.Sqrt2)
	}
}
