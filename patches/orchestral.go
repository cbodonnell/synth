package patches

import "github.com/cbodonnell/synth"

// Strings is a section playing legato: detuned saws over a triangle, a slow bow, the
// filter opening a little as the note speaks, and vibrato arriving late.
func Strings() synth.Patch {
	return synth.Patch{
		Name: "strings",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Saw, Cents: 7, Level: 0.7},
			{Wave: synth.Triangle, Level: 0.5},
		},
		Unison: synth.Unison{Voices: 5, Detune: 18, Spread: 0.9},
		Amp:    synth.Envelope{Attack: 0.35, Decay: 1, Sustain: 0.85, Release: 0.9},
		Filter: synth.Filter{
			Cutoff: 1500, KeyTrack: 0.5,
			Env:       synth.Envelope{Attack: 0.4, Decay: 1.2, Sustain: 0.6, Release: 0.9},
			EnvAmount: 0.7,
			Velocity:  1.2,
		},
		LFOs:     []synth.LFO{{Target: synth.Vibrato, Rate: 5.2, Depth: 9, Delay: 0.5}},
		Velocity: 0.6,
		Gain:     0.12,
	}
}

// Tremolo is Strings bowed fast on each note, for tension under a storm.
func Tremolo() synth.Patch {
	p := Strings()
	p.Name = "tremolo strings"
	p.Amp.Attack = 0.15
	p.LFOs = append(p.LFOs, synth.LFO{Target: synth.Tremolo, Shape: synth.TriangleShape, Rate: 9, Depth: 0.7})
	return p
}

// Spiccato is strings bounced off the string: a short bowed note with a scratch at its
// start, for ostinatos.
func Spiccato() synth.Patch {
	bow := synth.Envelope{Attack: 0.001, Decay: 0.03, Release: 0.01}
	return synth.Patch{
		Name: "spiccato",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Saw, Cents: 9, Level: 0.7},
			{Wave: synth.Noise, Level: 0.15, Env: &bow},
		},
		Unison: synth.Unison{Voices: 4, Detune: 16, Spread: 0.8},
		Amp:    synth.Envelope{Attack: 0.006, Decay: 0.2, Release: 0.12},
		Filter: synth.Filter{
			Cutoff: 1100, KeyTrack: 0.5,
			Env:       synth.Envelope{Attack: 0.004, Decay: 0.09},
			EnvAmount: 1.4,
			Velocity:  1.5,
		},
		Velocity: 0.8,
		Gain:     0.25,
	}
}

// Horns is a French horn section: round and dark when soft, the filter opening with
// velocity until it brays.
func Horns() synth.Patch {
	return synth.Patch{
		Name: "horns",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Pulse, Width: 0.45, Cents: -5, Level: 0.6},
			{Wave: synth.Sine, Level: 0.6},
		},
		Unison: synth.Unison{Voices: 3, Detune: 10, Spread: 0.5},
		Sweep:  synth.Sweep{Semis: -0.6, Time: 0.05},
		Amp:    synth.Envelope{Attack: 0.09, Decay: 0.6, Sustain: 0.8, Release: 0.45},
		Filter: synth.Filter{
			Cutoff: 420, KeyTrack: 0.7,
			Env:       synth.Envelope{Attack: 0.1, Decay: 0.7, Sustain: 0.55, Release: 0.4},
			EnvAmount: 1.5,
			Velocity:  2,
		},
		LFOs:     []synth.LFO{{Target: synth.Vibrato, Rate: 4.8, Depth: 5, Delay: 0.6}},
		Drive:    0.6,
		Velocity: 0.7,
		Gain:     0.08,
	}
}

// Timpani is a tuned kettle drum: a fundamental that swells a touch sharp as it is
// struck, two inharmonic partials that die first, and the stick's thud.
func Timpani() synth.Patch {
	partial := synth.Envelope{Attack: 0.002, Decay: 0.5, Release: 0.3}
	stick := synth.Envelope{Attack: 0.0005, Decay: 0.03, Release: 0.01}
	return synth.Patch{
		Name: "timpani",
		Oscs: []synth.Osc{
			{Wave: synth.Sine},
			{Wave: synth.Sine, Semi: 7, Cents: 2, Level: 0.35, Env: &partial},
			{Wave: synth.Sine, Semi: 12, Cents: -14, Level: 0.2, Env: &partial},
			{Wave: synth.Noise, Level: 0.25, Env: &stick},
		},
		Sweep:    synth.Sweep{Semis: 0.5, Time: 0.06},
		Amp:      synth.Envelope{Attack: 0.002, Decay: 2.2, Release: 1.2},
		Filter:   synth.Filter{Cutoff: 1000, Velocity: 1.5},
		Velocity: 0.85,
		Gain:     0.5,
	}
}

// Taiko is a big drum hit hard: a deep body that drops in pitch and the skin's slap.
// Play it around A1.
func Taiko() synth.Patch {
	skin := synth.Envelope{Attack: 0.0005, Decay: 0.06, Release: 0.02}
	return synth.Patch{
		Name: "taiko",
		Oscs: []synth.Osc{
			{Wave: synth.Sine},
			{Wave: synth.Triangle, Semi: 12, Level: 0.25},
			{Wave: synth.Noise, Level: 0.4, Env: &skin},
		},
		Sweep:    synth.Sweep{Semis: 9, Time: 0.035},
		Amp:      synth.Envelope{Attack: 0.001, Decay: 0.9, Release: 0.5},
		Filter:   synth.Filter{Cutoff: 900, Velocity: 1.5},
		Velocity: 0.9,
		Gain:     0.65,
	}
}

// Choir is voices on "ah": a wide stack of saws through a resonant lowpass that sits
// where a vowel's formant would, slow to speak, with a singer's vibrato.
func Choir() synth.Patch {
	return synth.Patch{
		Name: "choir",
		Oscs: []synth.Osc{
			{Wave: synth.Saw, Level: 0.6},
			{Wave: synth.Pulse, Width: 0.3, Cents: 6, Level: 0.4},
			{Wave: synth.Sine},
		},
		Unison: synth.Unison{Voices: 6, Detune: 22, Spread: 1},
		Amp:    synth.Envelope{Attack: 0.8, Decay: 1, Sustain: 0.9, Release: 1.4},
		Filter: synth.Filter{Cutoff: 900, Resonance: 0.35, KeyTrack: 0.3},
		LFOs: []synth.LFO{
			{Target: synth.Vibrato, Rate: 5, Depth: 11, Delay: 0.7},
			{Target: synth.Wah, Rate: 0.2, Depth: 0.15},
		},
		Velocity: 0.5,
		Gain:     0.06,
	}
}
