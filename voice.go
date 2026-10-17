package synth

import "math"

// Cutoff is recomputed this often rather than every sample; tan() dominates otherwise.
const filterEvery = 4

type oscState struct {
	ownEnv   bool
	osc      Osc
	env      Envelope
	ratio    float64
	level    float64
	phase    float64
	modPhase float64
	pluck    *pluck
	noise    *rng
}

type unisonVoice struct {
	detune float64 // cents
	pan    float64
	gain   float64
	seed   uint64
}

func unisonVoices(u Unison, pan float64, seed uint64) []unisonVoice {
	n := u.voices()
	out := make([]unisonVoice, n)
	g := 1 / math.Sqrt(float64(n))
	for i := range out {
		out[i] = unisonVoice{pan: pan, gain: g, seed: mixSeed(seed, uint64(i))}
		if n > 1 {
			x := float64(i)/float64(n-1) - 0.5
			out[i].detune = u.Detune * x
			out[i].pan = pan + 2*clamp(u.Spread, 0, 1)*x
		}
	}
	return out
}

// noteLength is how long a note sounds once every oscillator's envelope has released.
func noteLength(p Patch, n Note) float64 {
	length := 0.0
	for _, o := range p.Oscs {
		env := p.Amp
		if o.Env != nil {
			env = *o.Env
		}
		length = math.Max(length, env.length(n.Dur))
	}
	return length
}

// renderVoice adds one unison voice of one note into dst, starting at sample start.
func renderVoice(dst Stereo, start int, p Patch, n Note, v unisonVoice) {
	length := samples(noteLength(p, n))
	if start >= dst.Len() || length == 0 {
		return
	}
	length = min(length, dst.Len()-start)

	r := newRNG(v.seed)
	vel := n.velocity()
	gain := v.gain * (1 - p.Velocity + p.Velocity*vel)
	if p.Gain != 0 {
		gain *= p.Gain
	}

	bends := n.Slide != 0 || p.Sweep.Semis != 0
	for _, l := range p.LFOs {
		bends = bends || l.Target == Vibrato
	}

	oscs := make([]oscState, len(p.Oscs))
	for i, o := range p.Oscs {
		s := oscState{
			ownEnv: o.Env != nil,
			osc:    o,
			env:    p.Amp,
			ratio:  math.Pow(2, (o.Semi+(o.Cents+v.detune)/100)/12),
			level:  o.level(),
			phase:  o.Phase,
		}
		if o.Env != nil {
			s.env = *o.Env
		}
		// Unison voices in phase would sum to one loud voice instead of a chorus.
		if v.detune != 0 {
			s.phase += 0.5 + 0.5*r.float()
		}
		switch o.Wave {
		case Pluck:
			hz := Hz(n.Key) * s.ratio
			if o.Fixed > 0 {
				hz = o.Fixed * s.ratio
			}
			s.pluck = newPluck(hz, o.String, bends, newRNG(mixSeed(v.seed, uint64(i), 1)))
		case Noise:
			s.noise = newRNG(mixSeed(v.seed, uint64(i), 2))
		}
		oscs[i] = s
	}

	filtered := p.Filter.Cutoff > 0
	var f svf
	baseOctaves := p.Filter.KeyTrack*(n.Key-60)/12 + p.Filter.Velocity*(vel-1)

	driveNorm := 1.0
	if p.Drive > 0 {
		driveNorm = 1 / math.Tanh(1+p.Drive)
	}

	hasPan := false
	for _, l := range p.LFOs {
		hasPan = hasPan || l.Target == AutoPan
	}
	gl, gr := panGains(v.pan)

	lfoSeeds := make([]uint64, len(p.LFOs))
	for li := range lfoSeeds {
		lfoSeeds[li] = mixSeed(v.seed, uint64(li), 3)
	}
	base := Hz(n.Key)

	for i := range length {
		t := float64(i) / SampleRate

		var vib, wah, trem, pwm, pan float64
		for li, l := range p.LFOs {
			x := l.value(t, lfoSeeds[li])
			switch l.Target {
			case Vibrato:
				vib += x
			case Wah:
				wah += x
			case Tremolo:
				trem += x
			case PWM:
				pwm += x
			case AutoPan:
				pan += x
			}
		}

		if bends {
			key := n.Key + vib/100
			if n.Slide != 0 {
				key += n.Slide * math.Min(t/n.slideTime(), 1)
			}
			if p.Sweep.Time > 0 {
				key += p.Sweep.Semis * math.Exp(-t/p.Sweep.Time)
			}
			base = Hz(key)
		}

		amp := p.Amp.level(t, n.Dur)
		var x float64
		for oi := range oscs {
			s := &oscs[oi]
			e := amp
			if s.ownEnv {
				e = s.env.level(t, n.Dur)
			}
			if e == 0 && t > n.Dur {
				continue
			}
			x += s.next(base, t, pwm) * e * s.level
		}

		if filtered {
			if i%filterEvery == 0 {
				oct := baseOctaves + wah
				if p.Filter.EnvAmount != 0 {
					oct += p.Filter.EnvAmount * p.Filter.Env.level(t, n.Dur)
				}
				f.set(p.Filter.Cutoff*math.Exp2(oct), p.Filter.Resonance)
			}
			x = f.process(x, p.Filter.Mode)
		}

		if p.Drive > 0 {
			x = math.Tanh(x*(1+p.Drive)) * driveNorm
		}

		x *= gain * (1 - clamp(trem, 0, 1))

		if hasPan {
			gl, gr = panGains(v.pan + pan)
		}
		dst.L[start+i] += float32(x * gl)
		dst.R[start+i] += float32(x * gr)
	}
}

func (s *oscState) next(base, t, pwm float64) float64 {
	hz := base * s.ratio
	if s.osc.Fixed > 0 {
		hz = s.osc.Fixed * s.ratio
	}
	dt := hz / SampleRate

	switch s.osc.Wave {
	case Pluck:
		return s.pluck.next(hz)
	case Noise:
		return s.noise.float()
	}

	p := s.phase - math.Floor(s.phase)
	var y float64
	switch s.osc.Wave {
	case Triangle:
		y = 4*math.Abs(p-0.5) - 1
	case Saw:
		y = 2*p - 1 - polyBLEP(p, dt)
	case Pulse:
		w := s.osc.Width
		if w == 0 {
			w = 0.5
		}
		w = clamp(w+pwm, 0.02, 0.98)
		y = -1
		if p < w {
			y = 1
		}
		y += polyBLEP(p, dt)
		q := p + 1 - w
		y -= polyBLEP(q-math.Floor(q), dt)
	default:
		fm := s.osc.FM
		if fm.Index != 0 {
			index := fm.Index
			if fm.Decay > 0 {
				sus := clamp(fm.Sustain, 0, 1)
				index *= sus + (1-sus)*math.Exp(-t/fm.Decay)
			}
			y = math.Sin(2*math.Pi*p + index*math.Sin(2*math.Pi*s.modPhase))
			s.modPhase += dt * fm.Ratio
			if s.modPhase >= 1 {
				s.modPhase -= math.Floor(s.modPhase)
			}
		} else {
			y = math.Sin(2 * math.Pi * p)
		}
	}
	s.phase = p + dt
	return y
}

// polyBLEP rounds off a waveform's hard edge so it doesn't alias into a fizz up top.
func polyBLEP(t, dt float64) float64 {
	switch {
	case dt <= 0:
		return 0
	case t < dt:
		t /= dt
		return t + t - t*t - 1
	case t > 1-dt:
		t = (t - 1) / dt
		return t*t + t + t + 1
	}
	return 0
}

func (l LFO) value(t float64, seed uint64) float64 {
	fade := 1.0
	if l.Delay > 0 {
		fade = clamp((t-l.Delay)/l.Delay, 0, 1)
		if fade == 0 {
			return 0
		}
	}
	cycles := l.Rate * t
	p := cycles - math.Floor(cycles)

	var y float64
	switch l.Shape {
	case TriangleShape:
		y = 1 - 4*math.Abs(p-0.5)
	case SquareShape:
		y = 1
		if p >= 0.5 {
			y = -1
		}
	case SawShape:
		y = 2*p - 1
	case RandomShape:
		y = newRNG(mixSeed(seed, uint64(cycles))).float()
	default:
		y = math.Sin(2 * math.Pi * p)
	}
	return y * l.Depth * fade
}
