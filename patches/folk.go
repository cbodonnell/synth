package patches

import "github.com/cbodonnell/synth"

// Banjo is a bright string over a drum head: a short, twangy decay and the head's
// honk as a resonant peak thrown open by each pick.
func Banjo() synth.Patch {
	return synth.Patch{
		Name: "banjo",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 1.6, Brightness: 0.9, Pick: 0.08}},
		},
		Amp: synth.Envelope{Attack: 0.002, Release: 0.08},
		Filter: synth.Filter{
			Cutoff: 1800, Resonance: 0.45, KeyTrack: 0.4,
			Env:       synth.Envelope{Decay: 0.12},
			EnvAmount: 1.5,
		},
		Velocity: 0.6,
		Gain:     0.3,
	}
}

// Mandolin is paired steel strings tuned a few cents apart, played tremolo.
func Mandolin() synth.Patch {
	return synth.Patch{
		Name: "mandolin",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 2.5, Brightness: 0.8, Pick: 0.12}},
			{Wave: synth.Pluck, Cents: 6, String: synth.String{Decay: 2.5, Brightness: 0.8, Pick: 0.14}},
		},
		Amp:      synth.Envelope{Attack: 0.002, Release: 0.1},
		Filter:   synth.Filter{Cutoff: 5000},
		Velocity: 0.7,
		Gain:     0.2,
	}
}

// AcousticGuitar is a steel-string flat-top: the rhythm guitar's strings with no amp,
// played into a microphone, so the body's boom and the strings' sparkle both survive.
func AcousticGuitar() synth.Patch {
	return synth.Patch{
		Name: "acoustic guitar",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 4, Brightness: 0.75, Pick: 0.13}},
			{Wave: synth.Pluck, Level: 0.5, Cents: 4, String: synth.String{Decay: 3.5, Brightness: 0.8, Pick: 0.2}},
		},
		Amp:      synth.Envelope{Attack: 0.002, Release: 0.12},
		Filter:   synth.Filter{Cutoff: 7000},
		Velocity: 0.6,
		Gain:     0.25,
	}
}

// UprightBass is a plucked gut string: dark and thumpy, gone soon after the pluck.
func UprightBass() synth.Patch {
	return synth.Patch{
		Name: "upright bass",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 1.8, Brightness: 0.3, Pick: 0.3}},
			{Wave: synth.Sine, Level: 0.6, Env: &synth.Envelope{Attack: 0.004, Decay: 0.7, Release: 0.1}},
		},
		Amp:      synth.Envelope{Attack: 0.003, Release: 0.08},
		Filter:   synth.Filter{Cutoff: 900, Resonance: 0.1},
		Velocity: 0.5,
		Gain:     0.5,
	}
}

// Whistle is a person whistling: nearly a pure sine, a little breath, and a lazy vibrato.
func Whistle() synth.Patch {
	return synth.Patch{
		Name: "whistle",
		Oscs: []synth.Osc{
			{Wave: synth.Sine},
			{Wave: synth.Noise, Level: 0.04},
		},
		Amp:      synth.Envelope{Attack: 0.04, Decay: 0.3, Sustain: 0.8, Release: 0.08},
		Filter:   synth.Filter{Cutoff: 4000, KeyTrack: 0.6},
		LFOs:     []synth.LFO{{Target: synth.Vibrato, Rate: 5.5, Depth: 18, Delay: 0.2}},
		Velocity: 0.4,
		Gain:     0.2,
	}
}

// Glockenspiel is steel bars struck with brass mallets: FM nearer harmonic than a
// bell's, so it rings in tune, and a quick decay.
func Glockenspiel() synth.Patch {
	return synth.Patch{
		Name: "glockenspiel",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, FM: synth.FM{Ratio: 2.76, Index: 1.5, Decay: 0.15, Sustain: 0.1}},
			{Wave: synth.Sine, Semi: 12, Level: 0.2, Env: &synth.Envelope{Attack: 0.001, Decay: 0.3, Release: 0.1}},
		},
		Amp:      synth.Envelope{Attack: 0.001, Decay: 1.4, Release: 0.4},
		Velocity: 0.6,
		Gain:     0.25,
	}
}

// Tambourine is jingles: inharmonic squares high-passed into a shimmer, with a rattle
// of noise that lasts a moment longer.
func Tambourine() synth.Patch {
	sq := func(hz float64) synth.Osc {
		return synth.Osc{Wave: synth.Pulse, Fixed: hz, Level: 0.25, Env: &synth.Envelope{Attack: 0.001, Decay: 0.06, Release: 0.02}}
	}
	return synth.Patch{
		Name:     "tambourine",
		Oscs:     []synth.Osc{sq(3150), sq(4210), sq(5480), {Wave: synth.Noise}},
		Amp:      synth.Envelope{Attack: 0.003, Decay: 0.14, Release: 0.04},
		Filter:   synth.Filter{Mode: synth.HighPass, Cutoff: 6000, Resonance: 0.3},
		Velocity: 0.8,
		Gain:     0.2,
	}
}

// Firefly glows rather than strikes: a soft sine that swells and fades, drifting
// from side to side.
func Firefly() synth.Patch {
	return synth.Patch{
		Name: "firefly",
		Oscs: []synth.Osc{
			{Wave: synth.Sine},
			{Wave: synth.Sine, Semi: 12, Cents: 4, Level: 0.15},
		},
		Amp:  synth.Envelope{Attack: 0.12, Decay: 0.9, Release: 0.6},
		LFOs: []synth.LFO{{Target: synth.AutoPan, Shape: synth.RandomShape, Rate: 0.7, Depth: 0.7}},
		Gain: 0.12,
	}
}

// Cricket is a field cricket's chirp: a fixed tone near 4.5kHz, its wings' strokes
// chopping it into pulses. Hold a note for the length of one chirp.
func Cricket() synth.Patch {
	return synth.Patch{
		Name: "cricket",
		Oscs: []synth.Osc{{Wave: synth.Sine, Fixed: 4400}, {Wave: synth.Sine, Fixed: 8800, Level: 0.1}},
		Amp:  synth.Envelope{Attack: 0.005, Sustain: 1, Release: 0.01},
		LFOs: []synth.LFO{{Target: synth.Tremolo, Shape: synth.SquareShape, Rate: 32, Depth: 1}},
		Gain: 0.05,
	}
}

// GangVocals is a crowd singing along, open-throated: detuned saws through a vowel's
// worth of resonant lowpass, each voice wobbling a little.
func GangVocals() synth.Patch {
	return synth.Patch{
		Name: "gang vocals",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Pulse, Width: 0.3, Level: 0.4},
		},
		Unison: synth.Unison{Voices: 6, Detune: 28, Spread: 0.8},
		Amp:    synth.Envelope{Attack: 0.09, Decay: 0.3, Sustain: 0.8, Release: 0.25},
		Filter: synth.Filter{Cutoff: 900, Resonance: 0.45, KeyTrack: 0.3},
		LFOs:   []synth.LFO{{Target: synth.Vibrato, Rate: 5, Depth: 15, Delay: 0.2}},
		Gain:   0.05,
	}
}

// Shout is the gang's "Hey!": a crowd's vowel barked and dropping in pitch.
func Shout() synth.Patch {
	return synth.Patch{
		Name: "shout",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Noise, Level: 0.3},
		},
		Unison: synth.Unison{Voices: 8, Detune: 70, Spread: 0.9},
		Sweep:  synth.Sweep{Semis: 3, Time: 0.12},
		Amp:    synth.Envelope{Attack: 0.01, Decay: 0.2, Release: 0.06},
		Filter: synth.Filter{Mode: synth.BandPass, Cutoff: 1100, Resonance: 0.4},
		Drive:  1.5,
		Gain:   0.1,
	}
}
