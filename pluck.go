package synth

import (
	"math"
	"math/cmplx"
	"sync"
)

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

	// A struck string's stiffness: allpasses that delay the low partials more than the
	// high, so the overtones ring sharp of whole multiples as a piano's do.
	disp         int
	dispA        float64
	dispX, dispY [strikeStages]float64

	// A struck string starts at rest and the hammer's push, heard through the
	// soundboard, is fed in here a sample at a time.
	drive []float64
	at    int

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
	k.load(excitation(max(d, 2), p, brightness, r))
	return k
}

func (k *pluck) load(burst []float64) {
	if !k.bends {
		k.buf = burst
		return
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
}

func excitation(n int, p String, brightness float64, r *rng) []float64 {
	// A darker pick is a softer burst: lowpass the noise before it enters the string.
	lp := onePole{a: 0.15 + 0.85*brightness}
	burst := make([]float64, n)
	for i := range burst {
		burst[i] = lp.process(r.float())
	}

	return shaped(burst, p.Pick)
}

// shaped cancels the harmonics with a node where the string is plucked or struck, a
// fraction pick of its length, then centers and normalizes the burst.
func shaped(burst []float64, pick float64) []float64 {
	n := len(burst)
	if pick := int(math.Round(clamp(pick, 0, 0.5) * float64(n))); pick > 0 {
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
	if k.at < len(k.drive) {
		y += k.drive[k.at]
		k.at++
	}
	lowpassed := k.gain * ((1-k.s)*y + k.s*k.prev)
	k.prev = y
	for i := range k.disp {
		x := lowpassed
		lowpassed = k.dispA*x + k.dispX[i] - k.dispA*k.dispY[i]
		k.dispX[i], k.dispY[i] = x, lowpassed
	}
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
	if k.at < len(k.drive) {
		y += k.drive[k.at]
		k.at++
	}

	k.buf[k.pos] = k.gain * ((1-k.s)*y + k.s*k.prev)
	k.prev = y
	k.pos = (k.pos + 1) & k.mask
	return y
}

const (
	strikeStages = 4
	// How loud a struck string rings over its first second, as an RMS.
	strikeLevel = 0.3
	// A struck string loses its highs to the loop at this weight, whatever the hammer.
	strikeDamp = 0.2
)

// newStrike is a stiff string struck by a felt hammer. The string starts at rest and
// the hammer pushes it over its contact time, longer when soft, so a soft or gentle
// strike has fewer highs, and the push reaches the string through a soundboard. Bass
// strings ring far longer than treble ones, and the overtones are stretched by the
// stiffness a piano string has at each pitch, with the fundamental kept in tune.
func newStrike(hz float64, p String, vel float64, bends bool) *pluck {
	hz = clamp(hz, 20, SampleRate/4)
	decay := p.Decay
	if decay <= 0 {
		decay = 4
	}
	t60 := decay * math.Pow(220/hz, 0.8)
	total := SampleRate / hz

	// A treble string goes round its loop thousands of times a second, so the loop's
	// lowpass is lighter there, and the loss it does take is given back to the
	// fundamental, which then rings for t60.
	damp := strikeDamp * math.Min(1, 523/hz)
	w0 := 2 * math.Pi * hz / SampleRate
	lp := math.Sqrt((1-damp)*(1-damp) + damp*damp + 2*damp*(1-damp)*math.Cos(w0))
	k := &pluck{gain: math.Min(0.99999, math.Pow(10, -3/(hz*t60))/lp), s: damp, bends: bends}

	delay := total - damp
	if !bends {
		k.dispA = stiffness(hz, w0, delay)
		if k.dispA != 0 {
			k.disp = strikeStages
			delay -= strikeStages * allpassDelay(k.dispA, w0)
		}
	}
	d := max(int(delay-0.1), 2)
	frac := delay - float64(d)
	k.apC = (1 - frac) / (1 + frac)

	pick := p.Pick
	if pick == 0 {
		pick = 0.125
	}
	k.load(make([]float64, d))
	k.drive = hammer(total, p.Brightness, vel, pick)
	voice(k.drive, total, k.gain, damp)
	return k
}

// hammer is what a hammer feeds a string of period samples: its push, a pulse that
// is shorter and more peaked the harder the felt is driven, with the harmonics that
// have a node at the strike point cancelled, and the soundboard's ringing of it after (commuted: string and board are
// linear, so the board can come first).
func hammer(period, hardness, vel, pick float64) []float64 {
	h := clamp(hardness, 0, 1) * (0.4 + 0.6*clamp(vel, 0, 1))
	contact := 0.003 * math.Pow(0.1, h)
	w := int(clamp(contact*SampleRate, 2, period))
	sharp := 1 + 2*h
	push := make([]float64, w)
	for i := range push {
		push[i] = math.Pow(0.5-0.5*math.Cos(2*math.Pi*(float64(i)+0.5)/float64(w)), sharp)
	}
	moved := struck(push, int(math.Round(clamp(pick, 0, 0.5)*period)))

	// The push is measured by the fundamental it gives, so the board can be set
	// against it.
	var re, im float64
	for i, x := range moved {
		re += x * math.Cos(2*math.Pi*float64(i)/period)
		im += x * math.Sin(2*math.Pi*float64(i)/period)
	}
	scale := strikeLevel * period / (2 * math.Hypot(re, im))

	// The board rings at a level of its own beside the string's, so a treble string,
	// moved hard to sound its fundamental, doesn't knock the board as hard.
	board := soundboard()
	ring := make([]float64, len(board)+len(moved)-1)
	for i, m := range moved {
		if m == 0 {
			continue
		}
		for j, b := range board {
			ring[i+j] += m * b
		}
	}
	var energy float64
	knock := min(len(ring), SampleRate/30)
	for _, x := range ring[:knock] {
		energy += x * x
	}
	ringScale := boardLevel * strikeLevel / math.Sqrt(energy/float64(knock))

	out := ring
	for i := range out {
		out[i] *= ringScale
		if i < len(moved) {
			out[i] += moved[i] * scale
		}
	}

	// A loop rings at its offset as at its harmonics, which a string fixed at both ends
	// can't, so what is fed below half the fundamental is taken out.
	highpass(out, SampleRate/period/2)
	return out
}

// voice scales a string's drive so it sounds as loud over its first second at any
// pitch, as a piano is voiced even: the energy the loop will ring at each harmonic,
// for as long as its loss per trip lets it, weighted as the ear hears it: the lows
// down and the highs up.
func voice(drive []float64, period, gain, damp float64) {
	f0 := SampleRate / period
	var e float64
	for k := 1; float64(k)*f0 < 5000; k++ {
		w := 2 * math.Pi * float64(k) / period
		rot := complex(math.Cos(w), -math.Sin(w))
		ph, sum := complex(1, 0), complex(0, 0)
		for _, x := range drive {
			sum += complex(x, 0) * ph
			ph *= rot
		}
		amp := 2 * cmplx.Abs(sum) / period
		trip := gain * math.Sqrt((1-damp)*(1-damp)+damp*damp+2*damp*(1-damp)*math.Cos(w))
		tau := -1 / (f0 * math.Log(trip))
		f := float64(k) * f0
		heard := f * f / (f*f + 150*150) * (1 + 2*f*f/(f*f+2500*2500))
		e += amp * amp / 2 * heard * tau / 2 * (1 - math.Exp(-2/tau))
	}
	if e <= 0 {
		return
	}
	scale := strikeLevel / math.Sqrt(e)
	for i := range drive {
		drive[i] *= scale
	}
}

// highpass is a two-pole Butterworth highpass at hz, in place.
func highpass(x []float64, hz float64) {
	w := 2 * math.Pi * hz / SampleRate
	alpha := math.Sin(w) / math.Sqrt2
	cw := math.Cos(w)
	a0 := 1 + alpha
	b0, b1, b2 := (1+cw)/2/a0, -(1+cw)/a0, (1+cw)/2/a0
	a1, a2 := -2*cw/a0, (1-alpha)/a0
	var x1, x2, y1, y2 float64
	for i, v := range x {
		y := b0*v + b1*x1 + b2*x2 - a1*y1 - a2*y2
		x2, x1, y2, y1 = x1, v, y1, y
		x[i] = y
	}
}

// struck is a push as the string carries it: what is heard of a string follows the
// force it puts on the bridge, which follows its velocity, so the push goes in as it
// is, with the harmonics that have a node at the strike point, pk samples along,
// cancelled.
func struck(push []float64, pk int) []float64 {
	out := make([]float64, len(push)+pk)
	copy(out, push)
	if pk > 0 {
		for i := len(out) - 1; i >= pk; i-- {
			out[i] -= out[i-pk]
		}
	}
	return out
}

// boardModes are a soundboard's resonances: this many, spread from boardLow to
// boardHigh Hz, the higher ones dying sooner and weaker. Over its first 30ms the board
// rings at boardLevel of the string's fundamental.
const (
	boardModes = 40
	boardLow   = 80.0
	boardHigh  = 2000.0
	boardLevel = 0.25
)

// soundboard is the board's impulse response, the same for every string and render.
var soundboard = sync.OnceValue(func() []float64 {
	n := int(0.15 * SampleRate)
	ir := make([]float64, n)
	r := newRNG(0x50a4db0a4d)
	for m := range boardModes {
		f := boardLow * math.Pow(boardHigh/boardLow, (float64(m)+0.5+0.4*r.float())/boardModes)
		tau := clamp(0.02*math.Sqrt(200/f), 0.004, 0.025)
		amp := r.float() / (1 + f/2000) * f / (f + 250)
		phase := math.Pi * r.float()
		for i := 1; i < n; i++ {
			t := float64(i) / SampleRate
			v := amp * math.Exp(-t/tau) * math.Sin(2*math.Pi*f*t+phase)
			ir[i] += v
		}
	}
	return ir
})

// allpassDelay is the phase delay in samples of (a + z^-1) / (1 + a z^-1) at w.
func allpassDelay(a, w float64) float64 {
	phase := -math.Atan2(math.Sin(w), a+math.Cos(w)) + math.Atan2(a*math.Sin(w), 1+a*math.Cos(w))
	return -phase / w
}

// stiffness is the dispersion coefficient that stretches the string's sixth partial (or
// the highest below a third of the sample rate) as far as a piano string's at this
// pitch, f_n = n f_0 sqrt(1 + B n^2), within the room the delay line leaves.
func stiffness(hz, w0, room float64) float64 {
	b := math.Min(1.5e-4*math.Pow(hz/110, 1.2), 0.02)
	n := min(6, int(SampleRate/3/hz))
	if n < 2 {
		return 0
	}
	want := math.Sqrt(1 + b*float64(n*n))
	stretch := func(a float64) float64 {
		loop := func(w float64) float64 { return room + strikeStages*(allpassDelay(a, w)-allpassDelay(a, w0)) }
		w := float64(n) * w0
		for range 8 {
			w = 2 * math.Pi * float64(n) / loop(w)
		}
		return w / (float64(n) * w0)
	}
	lo, hi := -0.95, 0.0
	for lo < -0.01 && strikeStages*allpassDelay(lo, w0) > room-3 {
		lo /= 2
	}
	if stretch(lo) <= want {
		return lo
	}
	for range 40 {
		mid := (lo + hi) / 2
		if stretch(mid) > want {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}
