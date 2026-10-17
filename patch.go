package synth

// Patch is an instrument: oscillators summed, filtered, saturated, and shaped by an
// amplitude envelope, all modulated per note. Every field is plain data, so a patch
// can be copied and tweaked like any other value.
type Patch struct {
	Name   string
	Oscs   []Osc
	Unison Unison
	Amp    Envelope
	Filter Filter
	LFOs   []LFO
	Sweep  Sweep

	// tanh saturation between the filter and the amp stage; 0 is clean.
	Drive float64

	// How much a note's velocity scales its level: 0 ignores it, 1 is fully proportional.
	Velocity float64

	Gain float64 // 0 means 1
}

type Wave int

const (
	Sine Wave = iota
	Triangle
	Saw
	Pulse
	Noise
	// Pluck is a Karplus-Strong string: a burst of noise ringing in a tuned delay line.
	Pluck
)

func (w Wave) String() string {
	switch w {
	case Sine:
		return "sine"
	case Triangle:
		return "triangle"
	case Saw:
		return "saw"
	case Pulse:
		return "pulse"
	case Noise:
		return "noise"
	case Pluck:
		return "pluck"
	}
	return "unknown"
}

type Osc struct {
	Wave  Wave
	Level float64 // 0 means 1
	Semi  float64 // transposition from the played key
	Cents float64

	// Nonzero pins the oscillator to this frequency regardless of the key: drum bodies,
	// and for Noise, nothing (noise has no pitch to pin).
	Fixed float64

	Width float64 // Pulse duty cycle; 0 means 0.5
	Phase float64 // starting phase, 0..1

	Pan float64 // offset from the part's pan, -1..1

	// Nil follows Patch.Amp. A transient (a quill's click, a hammer's thud) gets its own.
	Env *Envelope

	FM     FM     // Sine only
	String String // Pluck only
}

func (o Osc) level() float64 {
	if o.Level == 0 {
		return 1
	}
	return o.Level
}

// FM phase-modulates a Sine oscillator with a sine at Ratio times its frequency.
// The index falls from Index to Sustain*Index with time constant Decay, which is what
// gives FM bells and electric pianos a bright strike and a mellow ring.
type FM struct {
	Ratio   float64
	Index   float64 // 0 disables
	Decay   float64 // seconds; 0 holds Index
	Sustain float64 // fraction of Index left once Decay has run its course
}

// String tunes a Pluck oscillator.
type String struct {
	// Seconds to fall 60dB at A3. Shorter in the treble, by the square root of pitch,
	// as on a real string. 0 means 4.
	Decay float64

	// 0 is a dull felt-like thump, 1 a bright metallic ping.
	Brightness float64

	// Where along the string it is plucked, 0..0.5. Near the end (0.05-0.15) is
	// nasal and bright, as a harpsichord quill; near the middle is round and hollow.
	Pick float64
}

// Envelope is ADSR in seconds. Decay and release curves are exponential and reach their
// target exactly, so every note ends at a true zero and never clicks. Decay 0 with
// Sustain 0 holds full level while the key is down: the zero Envelope is an organ's gate.
type Envelope struct {
	Attack  float64
	Decay   float64
	Sustain float64 // level, 0..1
	Release float64
}

type Unison struct {
	Voices int     // 0 and 1 both mean a single voice
	Detune float64 // cents between the flattest and sharpest voice
	Spread float64 // stereo width, 0..1
}

func (u Unison) voices() int { return max(u.Voices, 1) }

type FilterMode int

const (
	LowPass FilterMode = iota
	HighPass
	BandPass
	Notch
)

// Filter is a resonant state-variable filter. Cutoff moves in octaves: the envelope,
// key tracking, velocity, and LFOs all add octaves to it.
type Filter struct {
	Mode      FilterMode
	Cutoff    float64 // Hz; 0 bypasses the filter
	Resonance float64 // 0..1; near 1 it whistles

	KeyTrack float64 // 1 moves the cutoff an octave per octave played, relative to C4

	Env       Envelope
	EnvAmount float64 // octaves added at the envelope's peak; may be negative

	// Octaves taken away from a note played at velocity 0. Full velocity is the nominal cutoff.
	Velocity float64
}

type LFOTarget int

const (
	Vibrato LFOTarget = iota // Depth in cents
	Wah                      // Depth in octaves of filter cutoff
	Tremolo                  // Depth as a fraction of level
	PWM                      // Depth as duty cycle, added to Width
	AutoPan                  // Depth as pan, -1..1
)

type Shape int

const (
	SineShape Shape = iota
	TriangleShape
	SquareShape
	SawShape
	RandomShape // sample and hold, a fresh value each cycle
)

// LFO restarts on every note, so its modulation is part of the note, not a free-running clock.
type LFO struct {
	Target LFOTarget
	Shape  Shape
	Rate   float64 // Hz
	Depth  float64
	Delay  float64 // seconds before it begins to fade in, over the same length again
}

// Sweep starts the pitch Semis away and settles exponentially with time constant Time:
// the pitch drop of a kick drum, the zap of a laser, the scoop of a brass stab.
type Sweep struct {
	Semis float64
	Time  float64
}
