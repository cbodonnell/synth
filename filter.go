package synth

import "math"

// svf is the trapezoidal state-variable filter (Zavalishin/Simper): stable under fast
// cutoff modulation, which a biquad is not.
type svf struct {
	ic1, ic2   float64
	a1, a2, a3 float64
	k          float64
}

func (f *svf) set(cutoff, resonance float64) {
	cutoff = clamp(cutoff, 20, 0.45*SampleRate)
	g := math.Tan(math.Pi * cutoff / SampleRate)
	f.k = 2 - 1.96*clamp(resonance, 0, 1)
	f.a1 = 1 / (1 + g*(g+f.k))
	f.a2 = g * f.a1
	f.a3 = g * f.a2
}

func (f *svf) process(x float64, mode FilterMode) float64 {
	v3 := x - f.ic2
	v1 := f.a1*f.ic1 + f.a2*v3
	v2 := f.ic2 + f.a2*f.ic1 + f.a3*v3
	f.ic1 = 2*v1 - f.ic1
	f.ic2 = 2*v2 - f.ic2
	switch mode {
	case HighPass:
		return x - f.k*v1 - v2
	case BandPass:
		return v1
	case Notch:
		return x - f.k*v1
	default:
		return v2
	}
}

// biquad is a fixed RBJ-cookbook section, for EQ where nothing moves.
type biquad struct {
	b0, b1, b2, a1, a2 float64
	x1, x2, y1, y2     float64
}

func (q *biquad) process(x float64) float64 {
	y := q.b0*x + q.b1*q.x1 + q.b2*q.x2 - q.a1*q.y1 - q.a2*q.y2
	q.x2, q.x1 = q.x1, x
	q.y2, q.y1 = q.y1, y
	return y
}

func newBiquad(b Band) biquad {
	freq := clamp(b.Freq, 10, 0.45*SampleRate)
	q := b.Q
	if q <= 0 {
		q = math.Sqrt2 / 2
	}
	w := 2 * math.Pi * freq / SampleRate
	cw, sw := math.Cos(w), math.Sin(w)
	alpha := sw / (2 * q)
	a := math.Pow(10, b.Gain/40)

	var b0, b1, b2, a0, a1, a2 float64
	switch b.Kind {
	case LowCut:
		b0, b1, b2 = (1+cw)/2, -(1 + cw), (1+cw)/2
		a0, a1, a2 = 1+alpha, -2*cw, 1-alpha
	case HighCut:
		b0, b1, b2 = (1-cw)/2, 1-cw, (1-cw)/2
		a0, a1, a2 = 1+alpha, -2*cw, 1-alpha
	case LowShelf:
		sa := 2 * math.Sqrt(a) * alpha
		b0 = a * ((a + 1) - (a-1)*cw + sa)
		b1 = 2 * a * ((a - 1) - (a+1)*cw)
		b2 = a * ((a + 1) - (a-1)*cw - sa)
		a0 = (a + 1) + (a-1)*cw + sa
		a1 = -2 * ((a - 1) + (a+1)*cw)
		a2 = (a + 1) + (a-1)*cw - sa
	case HighShelf:
		sa := 2 * math.Sqrt(a) * alpha
		b0 = a * ((a + 1) + (a-1)*cw + sa)
		b1 = -2 * a * ((a - 1) + (a+1)*cw)
		b2 = a * ((a + 1) + (a-1)*cw - sa)
		a0 = (a + 1) - (a-1)*cw + sa
		a1 = 2 * ((a - 1) - (a+1)*cw)
		a2 = (a + 1) - (a-1)*cw - sa
	default: // Peak
		b0, b1, b2 = 1+alpha*a, -2*cw, 1-alpha*a
		a0, a1, a2 = 1+alpha/a, -2*cw, 1-alpha/a
	}
	return biquad{b0: b0 / a0, b1: b1 / a0, b2: b2 / a0, a1: a1 / a0, a2: a2 / a0}
}

// onePole is a gentle lowpass for damping inside feedback loops.
type onePole struct{ a, y float64 }

func newOnePole(cutoff float64) onePole {
	return onePole{a: 1 - math.Exp(-2*math.Pi*clamp(cutoff, 10, 0.45*SampleRate)/SampleRate)}
}

func (p *onePole) process(x float64) float64 {
	p.y += p.a * (x - p.y)
	return p.y
}
