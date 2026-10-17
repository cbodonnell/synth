package synth

import "math"

// pluck is a Karplus-Strong string. A string that holds its pitch is tuned with an
// allpass for the fractional part of the delay, without which chords drift audibly out
// of tune in the treble. A string that bends reads its delay line with 4-point Hermite
// interpolation instead, which can move smoothly where the allpass would click.
type pluck struct {
	buf  []float64
	pos  int
	gain float64 // loss per trip around the loop
	s    float64 // loop lowpass weight: 0.5 is the classic average, lower is brighter
	prev float64

	apC      float64
	apX, apY float64

	bends bool
	mask  int
}

func newPluck(hz float64, p String, bends bool, r *rng) *pluck {
	hz = clamp(hz, 20, SampleRate/4)

	brightness := clamp(p.Brightness, 0, 1)
	s := 0.5 - 0.45*brightness

	total := SampleRate / hz
	d := int(total - s - 0.1)
	frac := total - s - float64(d)

	decay := p.Decay
	if decay <= 0 {
		decay = 4
	}
	t60 := decay * math.Sqrt(220/hz)

	k := &pluck{
		gain:  math.Pow(10, -3/(hz*t60)),
		s:     s,
		apC:   (1 - frac) / (1 + frac),
		bends: bends,
	}
	burst := excitation(max(d, 2), p, brightness, r)
	if !bends {
		k.buf = burst
		return k
	}

	// Room to bend down an octave.
	size := 1
	for size < 2*len(burst)+8 {
		size <<= 1
	}
	k.buf = make([]float64, size)
	k.mask = size - 1
	copy(k.buf, burst)
	k.pos = len(burst)
	return k
}

func excitation(n int, p String, brightness float64, r *rng) []float64 {
	// A darker pick is a softer burst: lowpass the noise before it enters the string.
	lp := onePole{a: 0.15 + 0.85*brightness}
	burst := make([]float64, n)
	for i := range burst {
		burst[i] = lp.process(r.float())
	}

	// Plucking at a fraction p of the length cancels every harmonic with a node there.
	if pick := int(math.Round(clamp(p.Pick, 0, 0.5) * float64(n))); pick > 0 {
		for i := n - 1; i >= pick; i-- {
			burst[i] -= burst[i-pick]
		}
	}

	var mean, peak float64
	for _, x := range burst {
		mean += x
	}
	mean /= float64(n)
	for i := range burst {
		burst[i] -= mean
		peak = math.Max(peak, math.Abs(burst[i]))
	}
	if peak > 0 {
		for i := range burst {
			burst[i] /= peak
		}
	}
	return burst
}

func (k *pluck) next(hz float64) float64 {
	if k.bends {
		return k.nextBent(hz)
	}
	y := k.buf[k.pos]
	lowpassed := k.gain * ((1-k.s)*y + k.s*k.prev)
	k.prev = y
	ap := k.apC*lowpassed + k.apX - k.apC*k.apY
	k.apX, k.apY = lowpassed, ap
	k.buf[k.pos] = ap
	k.pos++
	if k.pos == len(k.buf) {
		k.pos = 0
	}
	return y
}

func (k *pluck) nextBent(hz float64) float64 {
	d := clamp(SampleRate/math.Max(hz, 1)-k.s, 4, float64(len(k.buf)-4))
	rp := float64(k.pos) - d
	i := int(math.Floor(rp))
	f := rp - float64(i)
	xm1, x0 := k.buf[(i-1)&k.mask], k.buf[i&k.mask]
	x1, x2 := k.buf[(i+1)&k.mask], k.buf[(i+2)&k.mask]
	c1 := 0.5 * (x1 - xm1)
	c2 := xm1 - 2.5*x0 + 2*x1 - 0.5*x2
	c3 := 0.5*(x2-xm1) + 1.5*(x0-x1)
	y := ((c3*f+c2)*f+c1)*f + x0

	k.buf[k.pos] = k.gain * ((1-k.s)*y + k.s*k.prev)
	k.prev = y
	k.pos = (k.pos + 1) & k.mask
	return y
}
