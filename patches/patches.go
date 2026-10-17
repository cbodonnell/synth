// Package patches is a starting library of instruments. Each is a function returning a
// fresh value, so a caller can tweak its copy without touching anyone else's.
package patches

import "github.com/cbodonnell/synth"

// Harpsichord registers two 8' choirs a few cents apart, which is where its shimmer
// comes from, and a 4' choir an octave up. Every string is plucked by a quill near its
// end, and the quill's click is its own short noise burst. Velocity does nothing, as on
// the real instrument; the damper falling on release is the Amp envelope.
func Harpsichord() synth.Patch {
	quill := synth.Envelope{Attack: 0.0005, Decay: 0.006, Release: 0.004}
	return synth.Patch{
		Name: "harpsichord",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 5, Brightness: 0.82, Pick: 0.12}},
			{Wave: synth.Pluck, Level: 0.8, Cents: 4, String: synth.String{Decay: 5.5, Brightness: 0.78, Pick: 0.17}},
			{Wave: synth.Pluck, Level: 0.42, Semi: 12, Cents: -2, String: synth.String{Decay: 3.5, Brightness: 0.88, Pick: 0.09}},
			{Wave: synth.Noise, Level: 0.12, Env: &quill},
		},
		Amp:    synth.Envelope{Attack: 0.0008, Release: 0.09},
		Filter: synth.Filter{Cutoff: 6500, KeyTrack: 0.35, Resonance: 0.08},
		Gain:   0.2,
	}
}

// OctaveSynth is a soft string machine in octaves, the "angel pad" that doubles a riff
// from above: detuned saws for body, a pulse breathing under PWM, a slow swell.
func OctaveSynth() synth.Patch {
	return synth.Patch{
		Name: "octave synth",
		Oscs: []synth.Osc{
			{Wave: synth.Saw, Level: 0.6},
			{Wave: synth.Saw, Level: 0.45, Cents: 7},
			{Wave: synth.Pulse, Level: 0.4, Semi: 12, Width: 0.35},
		},
		Unison: synth.Unison{Voices: 4, Detune: 16, Spread: 0.7},
		Amp:    synth.Envelope{Attack: 0.45, Decay: 0.6, Sustain: 0.8, Release: 1.1},
		Filter: synth.Filter{
			Cutoff: 1500, Resonance: 0.12, KeyTrack: 0.5,
			Env:       synth.Envelope{Attack: 0.6, Decay: 1.5, Sustain: 0.4, Release: 1},
			EnvAmount: 1.2,
		},
		LFOs: []synth.LFO{
			{Target: synth.PWM, Rate: 0.35, Depth: 0.12},
			{Target: synth.Vibrato, Rate: 5.2, Depth: 6, Delay: 0.6},
		},
		Velocity: 0.4,
		Gain:     0.09,
	}
}

// Bass is a picked electric bass: saw and sub through a filter that snaps shut.
func Bass() synth.Patch {
	return synth.Patch{
		Name: "bass",
		Oscs: []synth.Osc{
			{Wave: synth.Saw, Level: 0.7},
			{Wave: synth.Pulse, Level: 0.5, Semi: -12, Width: 0.5},
			{Wave: synth.Sine, Level: 0.6},
		},
		Amp: synth.Envelope{Attack: 0.003, Decay: 0.4, Sustain: 0.6, Release: 0.08},
		Filter: synth.Filter{
			Cutoff: 280, Resonance: 0.25,
			Env:       synth.Envelope{Decay: 0.18},
			EnvAmount: 3, Velocity: 1.5,
		},
		Drive:    0.8,
		Velocity: 0.6,
		Gain:     0.35,
	}
}

// EPiano is two-operator FM: a bright tine strike that mellows as the index falls.
func EPiano() synth.Patch {
	return synth.Patch{
		Name: "epiano",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, FM: synth.FM{Ratio: 1, Index: 2.2, Decay: 0.35, Sustain: 0.15}},
			{Wave: synth.Sine, Level: 0.3, FM: synth.FM{Ratio: 14, Index: 1.2, Decay: 0.02}},
		},
		Amp:      synth.Envelope{Attack: 0.002, Decay: 2.2, Sustain: 0, Release: 0.25},
		Unison:   synth.Unison{Voices: 2, Detune: 5, Spread: 0.4},
		LFOs:     []synth.LFO{{Target: synth.AutoPan, Rate: 3, Depth: 0.25}},
		Velocity: 0.8,
		Gain:     0.3,
	}
}

// Bell is FM at an inharmonic ratio, which is what makes metal sound like metal.
func Bell() synth.Patch {
	return synth.Patch{
		Name: "bell",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, FM: synth.FM{Ratio: 3.5, Index: 4, Decay: 1.2, Sustain: 0.2}},
		},
		Amp:  synth.Envelope{Attack: 0.001, Decay: 3, Release: 0.5},
		Gain: 0.35,
	}
}

// Lead is a square wave with delayed vibrato: the chiptune voice, band-limited.
func Lead() synth.Patch {
	return synth.Patch{
		Name:   "lead",
		Oscs:   []synth.Osc{{Wave: synth.Pulse, Width: 0.25}},
		Amp:    synth.Envelope{Attack: 0.005, Decay: 0.1, Sustain: 0.7, Release: 0.05},
		LFOs:   []synth.LFO{{Target: synth.Vibrato, Rate: 5.5, Depth: 25, Delay: 0.15}},
		Filter: synth.Filter{Cutoff: 5000},
		Gain:   0.2,
	}
}

func Kick() synth.Patch {
	click := synth.Envelope{Attack: 0.0005, Decay: 0.004, Release: 0.002}
	return synth.Patch{
		Name: "kick",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, Fixed: 50},
			{Wave: synth.Noise, Level: 0.3, Env: &click},
		},
		Sweep: synth.Sweep{Semis: 30, Time: 0.035},
		Amp:   synth.Envelope{Attack: 0.001, Decay: 0.35, Release: 0.05},
		Drive: 1.2,
		Gain:  0.8,
	}
}

func Snare() synth.Patch {
	body := synth.Envelope{Attack: 0.001, Decay: 0.08, Release: 0.02}
	return synth.Patch{
		Name: "snare",
		Oscs: []synth.Osc{
			{Wave: synth.Noise},
			{Wave: synth.Triangle, Fixed: 185, Level: 0.8, Env: &body},
		},
		Sweep:  synth.Sweep{Semis: 5, Time: 0.02},
		Amp:    synth.Envelope{Attack: 0.001, Decay: 0.18, Release: 0.04},
		Filter: synth.Filter{Mode: synth.BandPass, Cutoff: 2800, Resonance: 0.2},
		Gain:   0.7,
	}
}

func HiHat() synth.Patch {
	return synth.Patch{
		Name:     "hihat",
		Oscs:     []synth.Osc{{Wave: synth.Noise}},
		Amp:      synth.Envelope{Attack: 0.0005, Decay: 0.045, Release: 0.01},
		Filter:   synth.Filter{Mode: synth.HighPass, Cutoff: 7500, Resonance: 0.3},
		Velocity: 0.7,
		Gain:     0.4,
	}
}

// Laser is a sweep and nothing else; the shape of the sweep is the sound.
func Laser() synth.Patch {
	return synth.Patch{
		Name:   "laser",
		Oscs:   []synth.Osc{{Wave: synth.Saw}, {Wave: synth.Pulse, Semi: 7, Level: 0.5, Width: 0.125}},
		Sweep:  synth.Sweep{Semis: 36, Time: 0.06},
		Amp:    synth.Envelope{Attack: 0.001, Decay: 0.25, Release: 0.03},
		Filter: synth.Filter{Cutoff: 3000, Resonance: 0.6, Env: synth.Envelope{Decay: 0.2}, EnvAmount: 2},
		Gain:   0.3,
	}
}

// Coin is a plain square blip; the rising two-note figure belongs to the score.
func Coin() synth.Patch {
	return synth.Patch{
		Name: "coin",
		Oscs: []synth.Osc{{Wave: synth.Pulse, Width: 0.5}},
		Amp:  synth.Envelope{Attack: 0.001, Decay: 0.25, Release: 0.02},
		Gain: 0.25,
	}
}
