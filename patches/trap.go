package patches

import "github.com/cbodonnell/synth"

// EightOhEight is the long, tuned sub of trap: a sine with a punch of pitch at the
// front, driven until it growls on small speakers. Glide between notes with a slide
// token (C2/F1); keep the line monophonic by letting each note hold to the next.
func EightOhEight() synth.Patch {
	return synth.Patch{
		Name: "808",
		Oscs: []synth.Osc{
			{Wave: synth.Sine},
			{Wave: synth.Triangle, Level: 0.15},
		},
		Sweep:  synth.Sweep{Semis: 10, Time: 0.015},
		Amp:    synth.Envelope{Attack: 0.002, Decay: 1.6, Sustain: 0.45, Release: 0.1},
		Filter: synth.Filter{Cutoff: 1200},
		Drive:  2.8,
		Gain:   0.5,
	}
}

// TrapKick is short and hard, all click and knock, leaving the low end to the 808.
func TrapKick() synth.Patch {
	click := synth.Envelope{Attack: 0.0003, Decay: 0.004, Release: 0.002}
	return synth.Patch{
		Name: "trap kick",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, Fixed: 58},
			{Wave: synth.Noise, Level: 0.5, Env: &click},
		},
		Sweep: synth.Sweep{Semis: 32, Time: 0.02},
		Amp:   synth.Envelope{Attack: 0.0005, Decay: 0.18, Release: 0.03},
		Drive: 2.5,
		Gain:  0.85,
	}
}

// TrapSnare is cranked higher and longer than a rock snare, built to sit in a big reverb.
func TrapSnare() synth.Patch {
	body := synth.Envelope{Attack: 0.0005, Decay: 0.06, Release: 0.02}
	return synth.Patch{
		Name: "trap snare",
		Oscs: []synth.Osc{
			{Wave: synth.Noise},
			{Wave: synth.Triangle, Fixed: 240, Level: 0.9, Env: &body},
			{Wave: synth.Sine, Fixed: 340, Level: 0.4, Env: &body},
		},
		Sweep:  synth.Sweep{Semis: 7, Time: 0.015},
		Amp:    synth.Envelope{Attack: 0.0005, Decay: 0.26, Release: 0.05},
		Filter: synth.Filter{Mode: synth.HighPass, Cutoff: 600, Resonance: 0.1},
		Drive:  1,
		Gain:   0.6,
	}
}

// Brass is the festival-trap horn section: saws scooping up into pitch, a filter that
// swells open like breath behind a mouthpiece, and enough unison to sound like a section.
func Brass() synth.Patch {
	return synth.Patch{
		Name: "brass",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Saw, Cents: 9, Level: 0.8},
			{Wave: synth.Pulse, Semi: -12, Width: 0.4, Level: 0.35},
		},
		Unison: synth.Unison{Voices: 4, Detune: 14, Spread: 0.7},
		Sweep:  synth.Sweep{Semis: -1.5, Time: 0.035},
		Amp:    synth.Envelope{Attack: 0.012, Decay: 0.2, Sustain: 0.75, Release: 0.09},
		Filter: synth.Filter{
			Cutoff: 600, Resonance: 0.2, KeyTrack: 0.6,
			Env:       synth.Envelope{Attack: 0.03, Decay: 0.25, Sustain: 0.45},
			EnvAmount: 2.6,
			Velocity:  1,
		},
		Drive:    1.2,
		Velocity: 0.5,
		Gain:     0.1,
	}
}

// AirHorn is two horns a quarter-tone apart, beating against each other, honked through
// a narrow band and overdriven: the sound of a DJ who has something to announce.
func AirHorn() synth.Patch {
	return synth.Patch{
		Name: "air horn",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Saw, Cents: 55, Level: 0.9},
			{Wave: synth.Pulse, Semi: 12, Cents: 20, Width: 0.3, Level: 0.4},
		},
		Sweep:  synth.Sweep{Semis: -2, Time: 0.03},
		Amp:    synth.Envelope{Attack: 0.008, Release: 0.06},
		Filter: synth.Filter{Mode: synth.BandPass, Cutoff: 1600, Resonance: 0.35},
		LFOs:   []synth.LFO{{Target: synth.Vibrato, Rate: 7, Depth: 12}},
		Drive:  4,
		Gain:   0.25,
	}
}
