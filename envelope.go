package synth

import "math"

// A literal zero-sample jump from silence clicks at any volume, on attack and release alike.
const (
	minAttack  = 0.002
	minRelease = 0.005
)

// Steepness of the exponential segments; normalized so each still lands exactly on target.
const curve = 5.0

var curveFloor = math.Exp(-curve)

func fall(x float64) float64 {
	if x >= 1 {
		return 0
	}
	return (math.Exp(-curve*x) - curveFloor) / (1 - curveFloor)
}

func (e Envelope) attack() float64  { return math.Max(e.Attack, minAttack) }
func (e Envelope) release() float64 { return math.Max(e.Release, minRelease) }

// level is the envelope t seconds after note-on, with the key released at off.
func (e Envelope) level(t, off float64) float64 {
	if t < 0 {
		return 0
	}
	if t < off {
		return e.held(t)
	}
	rt := t - off
	r := e.release()
	if rt >= r {
		return 0
	}
	return e.held(off) * fall(rt/r)
}

func (e Envelope) held(t float64) float64 {
	a := e.attack()
	if t < a {
		return t / a
	}
	s := clamp(e.Sustain, 0, 1)
	d := t - a
	if e.Decay <= 0 || d >= e.Decay {
		if e.Decay <= 0 && s == 0 {
			return 1
		}
		return s
	}
	return s + (1-s)*fall(d/e.Decay)
}

// length is how long a note held for dur keeps sounding.
func (e Envelope) length(dur float64) float64 {
	held := math.Max(dur, 0)
	// A note with no sustain is silent once its decay finishes, release or not.
	if e.Sustain <= 0 && e.Decay > 0 {
		held = math.Min(held, e.attack()+e.Decay)
	}
	return held + e.release()
}
