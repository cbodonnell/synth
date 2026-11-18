package patches

import "github.com/cbodonnell/synth"

// TapePad is two pulses breathing against each other under PWM over a triangle an
// octave down, dark and slow, with the gentle pitch wow of a warm cassette.
func TapePad() synth.Patch {
	return synth.Patch{
		Name: "tape pad",
		Oscs: []synth.Osc{
			{Wave: synth.Pulse, Width: 0.3},
			{Wave: synth.Pulse, Width: 0.6, Level: 0.7, Cents: 5},
			{Wave: synth.Triangle, Semi: -12, Level: 0.4},
		},
		Unison: synth.Unison{Voices: 3, Detune: 10, Spread: 0.8},
		Amp:    synth.Envelope{Attack: 1.2, Decay: 1.5, Sustain: 0.85, Release: 2},
		Filter: synth.Filter{
			Cutoff: 900, Resonance: 0.15, KeyTrack: 0.3,
			Env:       synth.Envelope{Attack: 1.5, Decay: 2, Sustain: 0.5, Release: 2},
			EnvAmount: 1,
		},
		LFOs: []synth.LFO{
			{Target: synth.PWM, Rate: 0.3, Depth: 0.15},
			{Target: synth.Vibrato, Shape: synth.TriangleShape, Rate: 0.55, Depth: 7},
		},
		Gain: 0.07,
	}
}

// GlassPluck is a short FM strike at a harmonic ratio, so it rings like a struck glass
// rather than a bell, over a triangle that keeps the note's body once the shine is gone.
func GlassPluck() synth.Patch {
	return synth.Patch{
		Name: "glass pluck",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, FM: synth.FM{Ratio: 2, Index: 2.5, Decay: 0.07, Sustain: 0.1}},
			{Wave: synth.Triangle, Level: 0.4, Cents: -4},
		},
		Amp:      synth.Envelope{Attack: 0.002, Decay: 0.5, Release: 0.15},
		Filter:   synth.Filter{Cutoff: 3500, KeyTrack: 0.3, Velocity: 1},
		Velocity: 0.6,
		Gain:     0.22,
	}
}

// GlideLead is a soft monosynth: saw and a hollow pulse an octave down, the filter half
// shut, vibrato arriving late. Its glide is the clip's; it wants a chorus and an echo.
func GlideLead() synth.Patch {
	return synth.Patch{
		Name: "glide lead",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Pulse, Semi: -12, Width: 0.35, Level: 0.5},
			{Wave: synth.Sine, Level: 0.4},
		},
		Amp: synth.Envelope{Attack: 0.03, Decay: 0.4, Sustain: 0.75, Release: 0.25},
		Filter: synth.Filter{
			Cutoff: 1600, Resonance: 0.2, KeyTrack: 0.4,
			Env:       synth.Envelope{Attack: 0.02, Decay: 0.3, Sustain: 0.3},
			EnvAmount: 0.8,
		},
		LFOs:     []synth.LFO{{Target: synth.Vibrato, Rate: 5, Depth: 12, Delay: 0.35}},
		Velocity: 0.4,
		Gain:     0.13,
	}
}

// PickBass is an electric bass picked near the bridge: a string with a little bite
// over a sine for the low end, so a melodic line high on the neck stays round.
func PickBass() synth.Patch {
	pick := synth.Envelope{Attack: 0.0005, Decay: 0.008, Release: 0.004}
	return synth.Patch{
		Name: "pick bass",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 3, Brightness: 0.55, Pick: 0.1}},
			{Wave: synth.Sine, Level: 0.5},
			{Wave: synth.Noise, Level: 0.05, Env: &pick},
		},
		Amp: synth.Envelope{Attack: 0.002, Decay: 1.2, Sustain: 0.5, Release: 0.07},
		Filter: synth.Filter{
			Cutoff: 700, Resonance: 0.15, KeyTrack: 0.3,
			Env:       synth.Envelope{Decay: 0.12},
			EnvAmount: 1.5, Velocity: 1,
		},
		Drive:    0.7,
		Velocity: 0.5,
		Gain:     0.45,
	}
}

// Vibraphone is aluminium bars over resonator tubes whose fans turn: a nearly pure
// tone with a mallet's brief brightness, and the fans' slow tremolo.
func Vibraphone() synth.Patch {
	return synth.Patch{
		Name: "vibraphone",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, FM: synth.FM{Ratio: 4, Index: 0.8, Decay: 0.04}},
			{Wave: synth.Sine, Semi: 24, Level: 0.08, Env: &synth.Envelope{Attack: 0.001, Decay: 0.25, Release: 0.1}},
		},
		Amp:      synth.Envelope{Attack: 0.002, Decay: 3.5, Release: 0.6},
		LFOs:     []synth.LFO{{Target: synth.Tremolo, Rate: 4.5, Depth: 0.35}},
		Velocity: 0.6,
		Gain:     0.25,
	}
}

// Bird is one chirp: a tone falling into place as it warbles. Play it high (C7 is
// about 2kHz) and ratchet it for a trill.
func Bird() synth.Patch {
	return synth.Patch{
		Name:     "bird",
		Oscs:     []synth.Osc{{Wave: synth.Sine}, {Wave: synth.Sine, Semi: 12, Level: 0.1}},
		Sweep:    synth.Sweep{Semis: 5, Time: 0.025},
		Amp:      synth.Envelope{Attack: 0.006, Decay: 0.08, Release: 0.03},
		LFOs:     []synth.LFO{{Target: synth.Vibrato, Rate: 24, Depth: 70}},
		Velocity: 0.7,
		Gain:     0.1,
	}
}

// Surf is waves on a shore: noise kept low and swelling slowly, opening as each breaks.
// Hold one note for as long as the sea should last.
func Surf() synth.Patch {
	return synth.Patch{
		Name:   "surf",
		Oscs:   []synth.Osc{{Wave: synth.Noise}},
		Amp:    synth.Envelope{Attack: 2, Release: 3},
		Filter: synth.Filter{Cutoff: 700, Resonance: 0.1},
		LFOs: []synth.LFO{
			{Target: synth.Tremolo, Rate: 0.11, Depth: 0.6},
			{Target: synth.Wah, Rate: 0.11, Depth: 1.2},
		},
		Gain: 0.12,
	}
}

// SunLead is a warm, round lead: a sine modulated at unison ratio, held, for a reed's
// buzz without a saw's edge, with a pulse an octave up for air and a soft filter bloom
// on each note. It wants glide, a chorus, and a long echo.
func SunLead() synth.Patch {
	return synth.Patch{
		Name: "sun lead",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, FM: synth.FM{Ratio: 1, Index: 1.6, Decay: 0.25, Sustain: 0.55}},
			{Wave: synth.Pulse, Semi: 12, Width: 0.2, Level: 0.12, Cents: 6},
			{Wave: synth.Triangle, Semi: -12, Level: 0.25},
		},
		Amp: synth.Envelope{Attack: 0.01, Decay: 0.5, Sustain: 0.7, Release: 0.3},
		Filter: synth.Filter{
			Cutoff: 2200, Resonance: 0.1, KeyTrack: 0.4,
			Env:       synth.Envelope{Attack: 0.03, Decay: 0.4, Sustain: 0.4},
			EnvAmount: 0.7,
		},
		LFOs:     []synth.LFO{{Target: synth.Vibrato, Rate: 5.2, Depth: 14, Delay: 0.3}},
		Velocity: 0.4,
		Gain:     0.22,
	}
}

// CleanGuitar is a clean electric picked with a light touch: two strings a few cents
// apart and panned apart, which is the chorus pedal before any chorus is added.
func CleanGuitar() synth.Patch {
	return synth.Patch{
		Name: "clean guitar",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, Pan: -0.3, String: synth.String{Decay: 2.5, Brightness: 0.7, Pick: 0.25}},
			{Wave: synth.Pluck, Pan: 0.3, Level: 0.8, Cents: 7, String: synth.String{Decay: 2.2, Brightness: 0.6, Pick: 0.3}},
		},
		Amp:      synth.Envelope{Attack: 0.001, Decay: 1.5, Sustain: 0.3, Release: 0.2},
		Filter:   synth.Filter{Cutoff: 3200, KeyTrack: 0.4, Velocity: 1},
		Velocity: 0.5,
		Gain:     0.3,
	}
}

// Kalimba is a thumb piano: a tine whose overtone sits at an inharmonic ratio and dies
// at once, leaving a hollow, wooden sine.
func Kalimba() synth.Patch {
	return synth.Patch{
		Name: "kalimba",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, FM: synth.FM{Ratio: 5.4, Index: 1.4, Decay: 0.03}},
			{Wave: synth.Sine, Semi: 12, Level: 0.12, Env: &synth.Envelope{Attack: 0.001, Decay: 0.12, Release: 0.05}},
		},
		Amp:      synth.Envelope{Attack: 0.001, Decay: 0.9, Release: 0.2},
		Velocity: 0.6,
		Gain:     0.3,
	}
}

// SwellPad opens on every chord: detuned saws behind a filter whose envelope takes
// two seconds to arrive, so each change blooms like the sun coming through.
func SwellPad() synth.Patch {
	return synth.Patch{
		Name: "swell pad",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Triangle, Semi: 12, Level: 0.35},
		},
		Unison: synth.Unison{Voices: 4, Detune: 16, Spread: 0.9},
		Amp:    synth.Envelope{Attack: 0.6, Decay: 1, Sustain: 0.85, Release: 1.5},
		Filter: synth.Filter{
			Cutoff: 450, Resonance: 0.2, KeyTrack: 0.2,
			Env:       synth.Envelope{Attack: 2, Decay: 2, Sustain: 0.6, Release: 1.5},
			EnvAmount: 1.8,
		},
		LFOs: []synth.LFO{{Target: synth.Vibrato, Shape: synth.TriangleShape, Rate: 0.4, Depth: 5}},
		Gain: 0.06,
	}
}

// RoundBass is a synth bass for a live band: a pulse and a sine an octave apart through
// a filter that plucks and settles, round enough to play melodies on.
func RoundBass() synth.Patch {
	return synth.Patch{
		Name: "round bass",
		Oscs: []synth.Osc{
			{Wave: synth.Pulse, Width: 0.4, Level: 0.6},
			{Wave: synth.Sine, Semi: -12, Level: 0.5},
			{Wave: synth.Sine, Level: 0.5},
		},
		Amp: synth.Envelope{Attack: 0.003, Decay: 0.5, Sustain: 0.65, Release: 0.08},
		Filter: synth.Filter{
			Cutoff: 320, Resonance: 0.2, KeyTrack: 0.3,
			Env:       synth.Envelope{Decay: 0.15},
			EnvAmount: 2, Velocity: 1,
		},
		Drive:    0.6,
		Velocity: 0.5,
		Gain:     0.35,
	}
}

// TightSnare is a snare tuned up with the wires snug: a short crack of noise over a
// tonal body that pops rather than booms.
func TightSnare() synth.Patch {
	body := synth.Envelope{Attack: 0.001, Decay: 0.05, Release: 0.02}
	return synth.Patch{
		Name: "tight snare",
		Oscs: []synth.Osc{
			{Wave: synth.Noise},
			{Wave: synth.Sine, Fixed: 240, Level: 0.9, Env: &body},
			{Wave: synth.Sine, Fixed: 380, Level: 0.4, Env: &body},
		},
		Sweep:    synth.Sweep{Semis: 4, Time: 0.015},
		Amp:      synth.Envelope{Attack: 0.0005, Decay: 0.13, Release: 0.03},
		Filter:   synth.Filter{Mode: synth.HighPass, Cutoff: 900, Resonance: 0.1},
		Velocity: 0.7,
		Gain:     0.6,
	}
}

// Crowd is a field of people talking at a distance: noise in a voice's band, its
// formant jumping at random as a syllable might. Hold a note as long as they stay.
func Crowd() synth.Patch {
	return synth.Patch{
		Name:   "crowd",
		Oscs:   []synth.Osc{{Wave: synth.Noise}},
		Amp:    synth.Envelope{Attack: 2.5, Release: 3},
		Filter: synth.Filter{Mode: synth.BandPass, Cutoff: 800, Resonance: 0.35},
		LFOs: []synth.LFO{
			{Target: synth.Wah, Shape: synth.RandomShape, Rate: 6, Depth: 0.6},
			{Target: synth.Tremolo, Shape: synth.RandomShape, Rate: 4.3, Depth: 0.35},
			{Target: synth.Wah, Rate: 0.07, Depth: 0.4},
		},
		Gain: 0.1,
	}
}
