package patches

import "github.com/cbodonnell/synth"

// SuperSaw is the big festival lead: seven detuned saws spread wide, an octave above
// for shine and a square below for weight, with a filter that bites on each note.
func SuperSaw() synth.Patch {
	return synth.Patch{
		Name: "supersaw",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Saw, Semi: 12, Level: 0.45},
			{Wave: synth.Pulse, Semi: -12, Level: 0.3},
		},
		Unison: synth.Unison{Voices: 7, Detune: 38, Spread: 0.9},
		Amp:    synth.Envelope{Attack: 0.003, Decay: 0.3, Sustain: 0.65, Release: 0.15},
		Filter: synth.Filter{
			Cutoff: 3200, Resonance: 0.15, KeyTrack: 0.3,
			Env:       synth.Envelope{Decay: 0.25},
			EnvAmount: 1.5,
		},
		Velocity: 0.4,
		Gain:     0.09,
	}
}

// SawPad is the supersaw slowed down: no bite, a long swell and a longer tail.
func SawPad() synth.Patch {
	return synth.Patch{
		Name: "saw pad",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Saw, Semi: 12, Level: 0.3},
		},
		Unison: synth.Unison{Voices: 6, Detune: 30, Spread: 1},
		Amp:    synth.Envelope{Attack: 0.5, Decay: 1, Sustain: 0.8, Release: 1.2},
		Filter: synth.Filter{Cutoff: 1800, Resonance: 0.1},
		LFOs:   []synth.LFO{{Target: synth.Wah, Rate: 0.2, Depth: 0.4}},
		Gain:   0.05,
	}
}

// HousePluck is all filter envelope: a saw stack that opens for an instant and shuts.
func HousePluck() synth.Patch {
	return synth.Patch{
		Name: "house pluck",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Pulse, Semi: 12, Level: 0.3, Width: 0.3},
		},
		Unison: synth.Unison{Voices: 3, Detune: 18, Spread: 0.6},
		Amp:    synth.Envelope{Attack: 0.002, Decay: 0.35, Release: 0.1},
		Filter: synth.Filter{
			Cutoff: 350, Resonance: 0.3, KeyTrack: 0.5,
			Env:       synth.Envelope{Decay: 0.13},
			EnvAmount: 4,
			Velocity:  1,
		},
		Velocity: 0.5,
		Gain:     0.14,
	}
}

// HousePiano is a hammered pair of strings a few cents apart, the way a piano doubles
// and triples its strings, with a felt-softened hammer that brightens with velocity.
func HousePiano() synth.Patch {
	hammer := synth.Envelope{Attack: 0.0005, Decay: 0.01, Release: 0.005}
	return synth.Patch{
		Name: "house piano",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 7, Brightness: 0.7, Pick: 0.12}},
			{Wave: synth.Pluck, Level: 0.8, Cents: 2.5, String: synth.String{Decay: 6, Brightness: 0.65, Pick: 0.14}},
			{Wave: synth.Sine, Level: 0.25, Env: &synth.Envelope{Attack: 0.002, Decay: 1.5, Release: 0.2}},
			{Wave: synth.Noise, Level: 0.06, Env: &hammer},
		},
		Amp:      synth.Envelope{Attack: 0.001, Release: 0.25},
		Filter:   synth.Filter{Cutoff: 3500, KeyTrack: 0.5, Velocity: 1.5},
		Velocity: 0.7,
		Gain:     0.2,
	}
}

// OffbeatBass is the pumping house bass: a saw stack kept dark, with a sine for the sub.
func OffbeatBass() synth.Patch {
	return synth.Patch{
		Name: "offbeat bass",
		Oscs: []synth.Osc{
			{Wave: synth.Saw, Level: 0.6},
			{Wave: synth.Sine, Level: 0.8},
		},
		Unison: synth.Unison{Voices: 2, Detune: 10, Spread: 0.2},
		Amp:    synth.Envelope{Attack: 0.003, Decay: 0.15, Sustain: 0.7, Release: 0.06},
		Filter: synth.Filter{
			Cutoff: 380, Resonance: 0.2,
			Env:       synth.Envelope{Decay: 0.09},
			EnvAmount: 1.5,
		},
		Drive: 0.5,
		Gain:  0.35,
	}
}

// HouseKick is a 909-style kick: a short click and a sine that drops into a long body.
func HouseKick() synth.Patch {
	click := synth.Envelope{Attack: 0.0003, Decay: 0.003, Release: 0.002}
	return synth.Patch{
		Name: "house kick",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, Fixed: 48},
			{Wave: synth.Triangle, Fixed: 48, Level: 0.2},
			{Wave: synth.Noise, Level: 0.35, Env: &click},
		},
		Sweep: synth.Sweep{Semis: 28, Time: 0.028},
		Amp:   synth.Envelope{Attack: 0.0008, Decay: 0.45, Release: 0.04},
		Drive: 1.5,
		Gain:  0.9,
	}
}

// Clap is filtered noise. Play it as a strummed chord of one note, three or four times
// over, so it arrives as a flam of hands rather than one burst.
func Clap() synth.Patch {
	return synth.Patch{
		Name:     "clap",
		Oscs:     []synth.Osc{{Wave: synth.Noise}},
		Amp:      synth.Envelope{Attack: 0.0005, Decay: 0.11, Release: 0.03},
		Filter:   synth.Filter{Mode: synth.BandPass, Cutoff: 1400, Resonance: 0.35},
		Velocity: 0.6,
		Gain:     0.45,
	}
}

func OpenHat() synth.Patch {
	return synth.Patch{
		Name:     "open hat",
		Oscs:     []synth.Osc{{Wave: synth.Noise}},
		Amp:      synth.Envelope{Attack: 0.001, Decay: 0.22, Release: 0.04},
		Filter:   synth.Filter{Mode: synth.HighPass, Cutoff: 8000, Resonance: 0.2},
		Velocity: 0.6,
		Gain:     0.3,
	}
}

// Riser swells for build seconds: noise and a saw rising in level and opening through a
// resonant filter together. Give its note a Slide upward for the pitch to climb too.
func Riser(build float64) synth.Patch {
	return synth.Patch{
		Name: "riser",
		Oscs: []synth.Osc{
			{Wave: synth.Noise},
			{Wave: synth.Saw, Level: 0.25},
		},
		Unison: synth.Unison{Voices: 3, Detune: 40, Spread: 1},
		Amp:    synth.Envelope{Attack: build, Sustain: 1, Release: 0.05},
		Filter: synth.Filter{
			Cutoff: 250, Resonance: 0.45,
			Env:       synth.Envelope{Attack: build, Sustain: 1},
			EnvAmount: 5.5,
		},
		Gain: 0.12,
	}
}
