package synth

import (
	"math"
	"slices"
)

// Effect processes a whole buffer in place, starting from silence each call, so an
// effect value is reusable data rather than a stateful object.
type Effect interface {
	Process(s Stereo)
	// Seconds the effect keeps sounding after its input falls silent.
	Tail() float64
}

func wetDry(mix float64) (wet, dry float32) {
	if mix <= 0 {
		return 1, 0
	}
	m := clamp(mix, 0, 1)
	return float32(m), float32(1 - m)
}

type BandKind int

const (
	Peak BandKind = iota
	LowCut
	HighCut
	LowShelf
	HighShelf
)

type Band struct {
	Kind BandKind
	Freq float64
	Gain float64 // dB; ignored by the cuts
	Q    float64 // 0 means 0.707
}

type EQ struct{ Bands []Band }

func (e EQ) Tail() float64 { return 0.05 }

func (e EQ) Process(s Stereo) {
	for _, b := range e.Bands {
		l, r := newBiquad(b), newBiquad(b)
		for i := range s.L {
			s.L[i] = float32(l.process(float64(s.L[i])))
			s.R[i] = float32(r.process(float64(s.R[i])))
		}
	}
}

// Drive is tanh saturation with a tone control after it: the cheap amp a 1981 record went through.
type Drive struct {
	Amount float64 // 0 is nearly clean, 10 is fuzz
	Tone   float64 // lowpass Hz after the clipper; 0 leaves it open
	Mix    float64 // wet fraction; 0 means 1
}

func (d Drive) Tail() float64 { return 0 }

func (d Drive) Process(s Stereo) {
	g := 1 + d.Amount
	norm := 1 / math.Tanh(g)
	wet, dry := wetDry(d.Mix)
	tl, tr := newOnePole(d.Tone), newOnePole(d.Tone)
	for i := range s.L {
		l := math.Tanh(float64(s.L[i])*g) * norm
		r := math.Tanh(float64(s.R[i])*g) * norm
		if d.Tone > 0 {
			l, r = tl.process(l), tr.process(r)
		}
		s.L[i] = s.L[i]*dry + float32(l)*wet
		s.R[i] = s.R[i]*dry + float32(r)*wet
	}
}

// Chorus sweeps two short delays in opposite phase, one per side, which widens as it thickens.
type Chorus struct {
	Rate     float64 // Hz; 0 means 0.8
	Depth    float64 // seconds of sweep; 0 means 0.003
	Delay    float64 // seconds at the center of the sweep; 0 means 0.012
	Feedback float64
	Mix      float64 // wet fraction; 0 means 1
}

func (c Chorus) Tail() float64 { return 0.05 }

func (c Chorus) Process(s Stereo) {
	rate, depth, base := c.Rate, c.Depth, c.Delay
	if rate == 0 {
		rate = 0.8
	}
	if depth == 0 {
		depth = 0.003
	}
	if base == 0 {
		base = 0.012
	}
	wet, dry := wetDry(c.Mix)
	size := samples(base+depth) + 4
	bl, br := make([]float32, size), make([]float32, size)
	fb := float32(clamp(c.Feedback, 0, 0.95))

	for i := range s.L {
		t := float64(i) / SampleRate
		ph := 2 * math.Pi * rate * t
		dl := (base + depth*math.Sin(ph)) * SampleRate
		dr := (base + depth*math.Sin(ph+math.Pi)) * SampleRate
		yl := readFrac(bl, i, dl)
		yr := readFrac(br, i, dr)
		bl[i%size] = s.L[i] + yl*fb
		br[i%size] = s.R[i] + yr*fb
		s.L[i] = s.L[i]*dry + yl*wet
		s.R[i] = s.R[i]*dry + yr*wet
	}
}

func readFrac(buf []float32, i int, delay float64) float32 {
	n := len(buf)
	pos := float64(i) - delay
	j := int(math.Floor(pos))
	frac := float32(pos - float64(j))
	a := buf[((j%n)+n)%n]
	b := buf[(((j+1)%n)+n)%n]
	return a + (b-a)*frac
}

// Phaser sweeps a chain of allpasses and mixes it back with the dry signal, so notches
// comb through the spectrum. Unlike the other effects, Mix 0 means an even blend:
// fully wet, a phaser is only a phase shift and can't be heard.
type Phaser struct {
	Rate     float64 // Hz
	Depth    float64 // octaves swept either side of Center
	Center   float64 // Hz; 0 means 700
	Stages   int     // 0 means 6
	Feedback float64 // 0..0.9; more is more vocal
	Mix      float64 // wet fraction; 0 means 0.5
}

func (p Phaser) Tail() float64 { return 0.01 }

func (p Phaser) Process(s Stereo) {
	center, stages, mix := p.Center, p.Stages, p.Mix
	if center == 0 {
		center = 700
	}
	if stages == 0 {
		stages = 6
	}
	if mix == 0 {
		mix = 0.5
	}
	wet, dry := float32(clamp(mix, 0, 1)), float32(1-clamp(mix, 0, 1))
	fb := clamp(p.Feedback, 0, 0.9)

	type side struct {
		x, y []float64
		last float64
	}
	sides := [2]side{
		{x: make([]float64, stages), y: make([]float64, stages)},
		{x: make([]float64, stages), y: make([]float64, stages)},
	}
	chans := [2][]float32{s.L, s.R}

	for i := range s.L {
		t := float64(i) / SampleRate
		for c := range 2 {
			// The right side runs a quarter cycle behind, which spreads the sweep across the field.
			lfo := math.Sin(2*math.Pi*p.Rate*t + float64(c)*math.Pi/2)
			f := clamp(center*math.Exp2(p.Depth*lfo), 20, 0.45*SampleRate)
			tn := math.Tan(math.Pi * f / SampleRate)
			a := (tn - 1) / (tn + 1)

			sd := &sides[c]
			in := float64(chans[c][i])
			x := in + fb*sd.last
			for k := range stages {
				y := a*x + sd.x[k] - a*sd.y[k]
				sd.x[k], sd.y[k] = x, y
				x = y
			}
			sd.last = x
			chans[c][i] = float32(in)*dry + float32(x)*wet
		}
	}
}

type Delay struct {
	Time     float64 // seconds
	Feedback float64 // 0..0.95
	Damp     float64 // lowpass Hz on each repeat; 0 leaves them bright
	PingPong bool    // repeats alternate sides
	Mix      float64 // wet fraction; 0 means 1
}

func (d Delay) Tail() float64 {
	fb := clamp(d.Feedback, 0, 0.95)
	if fb <= 0.001 {
		return d.Time
	}
	return d.Time * (1 + math.Log(0.001)/math.Log(fb))
}

func (d Delay) Process(s Stereo) {
	n := samples(d.Time)
	if n <= 0 {
		return
	}
	wet, dry := wetDry(d.Mix)
	fb := float32(clamp(d.Feedback, 0, 0.95))
	bl, br := make([]float32, n), make([]float32, n)
	pl, pr := newOnePole(d.Damp), newOnePole(d.Damp)
	damp := func(p *onePole, x float32) float32 {
		if d.Damp <= 0 {
			return x
		}
		return float32(p.process(float64(x)))
	}

	for i := range s.L {
		j := i % n
		yl, yr := bl[j], br[j]
		if d.PingPong {
			bl[j] = damp(&pl, (s.L[i]+s.R[i])*0.5+yr*fb)
			br[j] = damp(&pr, yl)
		} else {
			bl[j] = damp(&pl, s.L[i]+yl*fb)
			br[j] = damp(&pr, s.R[i]+yr*fb)
		}
		s.L[i] = s.L[i]*dry + yl*wet
		s.R[i] = s.R[i]*dry + yr*wet
	}
}

// Reverb is Freeverb: eight damped combs in parallel into four allpasses, per side,
// with the right side's lengths offset so the two decorrelate into width.
type Reverb struct {
	Size     float64 // 0..1, room to hall
	Damp     float64 // 0..1, how quickly the highs die
	Width    float64 // 0..1; 0 means 1
	PreDelay float64 // seconds
	Mix      float64 // wet fraction; 0 means 1
}

var (
	combTuning    = [8]int{1116, 1188, 1277, 1356, 1422, 1491, 1557, 1617}
	allpassTuning = [4]int{556, 441, 341, 225}
)

const stereoSpread = 23

func (r Reverb) feedback() float64 { return clamp(r.Size, 0, 1)*0.28 + 0.7 }

func (r Reverb) Tail() float64 {
	avg := 1378.0 / 44100
	return r.PreDelay + 3*avg/-math.Log10(r.feedback())
}

func (r Reverb) Process(s Stereo) {
	scale := float64(SampleRate) / 44100
	fb := r.feedback()
	damp := clamp(r.Damp, 0, 1) * 0.4
	width := r.Width
	if width == 0 {
		width = 1
	}
	wet, dry := wetDry(r.Mix)
	wet1 := float64(wet) * 3 * (width/2 + 0.5)
	wet2 := float64(wet) * 3 * ((1 - width) / 2)

	type comb struct {
		buf   []float64
		i     int
		store float64
	}
	type allpass struct {
		buf []float64
		i   int
	}
	newSide := func(offset int) ([8]comb, [4]allpass) {
		var cs [8]comb
		var as [4]allpass
		for k, n := range combTuning {
			cs[k].buf = make([]float64, int(float64(n+offset)*scale))
		}
		for k, n := range allpassTuning {
			as[k].buf = make([]float64, int(float64(n+offset)*scale))
		}
		return cs, as
	}
	cl, al := newSide(0)
	cr, ar := newSide(stereoSpread)

	run := func(cs *[8]comb, as *[4]allpass, in float64) float64 {
		var out float64
		for k := range cs {
			c := &cs[k]
			y := c.buf[c.i]
			c.store = y*(1-damp) + c.store*damp
			c.buf[c.i] = in + c.store*fb
			c.i = (c.i + 1) % len(c.buf)
			out += y
		}
		for k := range as {
			a := &as[k]
			y := a.buf[a.i]
			a.buf[a.i] = out + y*0.5
			a.i = (a.i + 1) % len(a.buf)
			out = y - out
		}
		return out
	}

	pre := samples(r.PreDelay)
	var preBuf []float64
	if pre > 0 {
		preBuf = make([]float64, pre)
	}

	for i := range s.L {
		in := float64(s.L[i]+s.R[i]) * 0.015
		if pre > 0 {
			j := i % pre
			in, preBuf[j] = preBuf[j], in
		}
		ol := run(&cl, &al, in)
		or := run(&cr, &ar, in)
		s.L[i] = s.L[i]*dry + float32(ol*wet1+or*wet2)
		s.R[i] = s.R[i]*dry + float32(or*wet1+ol*wet2)
	}
}

// Compressor is feed-forward and stereo-linked, so it never shifts the image.
type Compressor struct {
	Threshold float64 // dB
	Ratio     float64 // 0 means 4
	Attack    float64 // seconds; 0 means 0.005
	Release   float64 // seconds; 0 means 0.1
	Makeup    float64 // dB
}

func (c Compressor) Tail() float64 { return 0 }

func (c Compressor) Process(s Stereo) {
	ratio, attack, release := c.Ratio, c.Attack, c.Release
	if ratio == 0 {
		ratio = 4
	}
	if attack == 0 {
		attack = 0.005
	}
	if release == 0 {
		release = 0.1
	}
	ca := math.Exp(-1 / (attack * SampleRate))
	cr := math.Exp(-1 / (release * SampleRate))
	makeup := dbToGain(c.Makeup)

	env := 0.0
	for i := range s.L {
		peak := float64(max(abs32(s.L[i]), abs32(s.R[i])))
		if peak > env {
			env = ca*env + (1-ca)*peak
		} else {
			env = cr*env + (1-cr)*peak
		}
		g := makeup
		if env > 1e-9 {
			over := 20*math.Log10(env) - c.Threshold
			if over > 0 {
				g *= dbToGain(-over * (1 - 1/ratio))
			}
		}
		s.L[i] *= float32(g)
		s.R[i] *= float32(g)
	}
}

// Crush quantizes and sample-holds: the grain of an 8-bit console.
type Crush struct {
	Bits int // 0 leaves depth alone
	Hold int // samples each value is held; 0 and 1 leave the rate alone
}

func (c Crush) Tail() float64 { return 0 }

func (c Crush) Process(s Stereo) {
	levels := float32(0)
	if c.Bits > 0 {
		levels = float32(math.Pow(2, float64(c.Bits)) / 2)
	}
	var hl, hr float32
	for i := range s.L {
		if c.Hold <= 1 || i%c.Hold == 0 {
			hl, hr = s.L[i], s.R[i]
			if levels > 0 {
				hl = float32(math.Round(float64(hl*levels))) / levels
				hr = float32(math.Round(float64(hr*levels))) / levels
			}
		}
		s.L[i], s.R[i] = hl, hr
	}
}

// Limiter holds peaks under Ceiling without clipping them. Rendering is offline, so it
// sees each peak coming and eases the gain down to meet it rather than reacting late.
type Limiter struct {
	Ceiling float64 // dB; 0 means -1
	Release float64 // seconds; 0 means 0.08
}

func (l Limiter) Tail() float64 { return 0 }

func (l Limiter) Process(s Stereo) {
	ceiling, release := l.Ceiling, l.Release
	if ceiling == 0 {
		ceiling = -1
	}
	if release == 0 {
		release = 0.08
	}
	c := dbToGain(ceiling)
	n := s.Len()
	if n == 0 {
		return
	}

	req := make([]float64, n)
	for i := range req {
		peak := float64(max(abs32(s.L[i]), abs32(s.R[i])))
		req[i] = 1
		if peak > c {
			req[i] = c / peak
		}
	}

	ahead := samples(0.005)
	floor := slidingMin(req, ahead)

	// Every window that averages over a peak's sample also took its minimum over that
	// peak, so the eased gain is never above what the peak needs.
	prefix := make([]float64, n+1)
	for i, x := range floor {
		prefix[i+1] = prefix[i] + x
	}
	half := ahead / 2
	eased := make([]float64, n)
	for i := range eased {
		lo, hi := max(i-half, 0), min(i+half, n-1)
		eased[i] = (prefix[hi+1] - prefix[lo]) / float64(hi-lo+1)
	}

	rc := math.Exp(-1 / (release * SampleRate))
	g := 1.0
	for i := range n {
		g = math.Min(eased[i], g*rc+(1-rc)*eased[i])
		s.L[i] *= float32(g)
		s.R[i] *= float32(g)
	}
}

// slidingMin is the minimum of x over [i-r, i+r] for every i, in linear time.
func slidingMin(x []float64, r int) []float64 {
	n := len(x)
	out := make([]float64, n)
	var dq []int
	next := 0
	for i := range n {
		for ; next < n && next <= i+r; next++ {
			for len(dq) > 0 && x[dq[len(dq)-1]] >= x[next] {
				dq = dq[:len(dq)-1]
			}
			dq = append(dq, next)
		}
		for dq[0] < i-r {
			dq = dq[1:]
		}
		out[i] = x[dq[0]]
	}
	return out
}

// Sidechain ducks its input at each of Hits, the times a kick lands, and lets it swell
// back: the pump of house music. Keyed by the kick's notes rather than its audio, so it
// pumps exactly where the kick plays and nowhere else.
type Sidechain struct {
	Hits    []float64 // seconds, ascending; see Onsets
	Depth   float64   // 0..1, how far the level drops
	Attack  float64   // seconds to the bottom of the duck; 0 means 0.004
	Release float64   // seconds back to full; 0 means 0.2
}

func (sc Sidechain) Tail() float64 { return 0 }

func (sc Sidechain) Process(s Stereo) {
	attack, release := sc.Attack, sc.Release
	if attack == 0 {
		attack = 0.004
	}
	if release == 0 {
		release = 0.2
	}
	depth := clamp(sc.Depth, 0, 1)
	next := 0
	last := math.Inf(-1)
	for i := range s.L {
		t := float64(i) / SampleRate
		for next < len(sc.Hits) && sc.Hits[next] <= t+attack {
			last = sc.Hits[next]
			next++
		}
		// The duck begins attack early, so the level is already down when the kick lands.
		dt := t - (last - attack)
		var duck float64
		switch {
		case dt < 0 || dt > attack+release:
			duck = 0
		case dt < attack:
			duck = dt / attack
		default:
			x := (dt - attack) / release
			duck = fall(x)
		}
		g := float32(1 - depth*duck)
		s.L[i] *= g
		s.R[i] *= g
	}
}

// AutoFilter is a filter whose cutoff follows a line through Points over song time,
// moving exponentially between them: the slow sweep that opens a build.
type AutoFilter struct {
	Mode      FilterMode
	Resonance float64
	Points    []Point // ascending in At
}

type Point struct {
	At     float64 // seconds into the song
	Cutoff float64 // Hz
}

func (a AutoFilter) Tail() float64 { return 0.05 }

func (a AutoFilter) cutoff(t float64) float64 {
	p := a.Points
	switch {
	case len(p) == 0:
		return 0.45 * SampleRate
	case t <= p[0].At:
		return p[0].Cutoff
	case t >= p[len(p)-1].At:
		return p[len(p)-1].Cutoff
	}
	k := 1
	for p[k].At < t {
		k++
	}
	x := (t - p[k-1].At) / math.Max(p[k].At-p[k-1].At, 1e-9)
	lo, hi := math.Max(p[k-1].Cutoff, 1), math.Max(p[k].Cutoff, 1)
	return lo * math.Pow(hi/lo, x)
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

// TapeStop spins the audio down to a halt, as if the reels lost power, then holds
// silence until Until: the breath before a drop. Playback slows linearly over Length,
// so pitch falls with it.
type TapeStop struct {
	At     float64 // seconds into the song
	Length float64
	Until  float64 // silence lasts until here; 0 means it ends with the stop
}

func (ts TapeStop) Tail() float64 { return 0 }

func (ts TapeStop) Process(s Stereo) {
	a, n := samples(ts.At), samples(ts.Length)
	if a >= s.Len() || n <= 0 {
		return
	}
	n = min(n, s.Len()-a)
	src := Stereo{L: slices.Clone(s.L[a : a+n]), R: slices.Clone(s.R[a : a+n])}
	for i := range n {
		tau := float64(i)
		pos := tau - tau*tau/(2*float64(n))
		j := int(pos)
		f := float32(pos - float64(j))
		k := min(j+1, n-1)
		// The level sinks with the speed, so the stop settles on silence instead of a step.
		g := float32(math.Sqrt(1 - tau/float64(n)))
		s.L[a+i] = (src.L[j] + (src.L[k]-src.L[j])*f) * g
		s.R[a+i] = (src.R[j] + (src.R[k]-src.R[j])*f) * g
	}
	for i := a + n; i < min(samples(ts.Until), s.Len()); i++ {
		s.L[i], s.R[i] = 0, 0
	}
}
