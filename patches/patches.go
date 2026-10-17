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

// Recorder is a near-sine flute with a breath of noise, the chiff of each tongued note,
// and vibrato that only arrives once a note is held.
func Recorder() synth.Patch {
	chiff := synth.Envelope{Attack: 0.002, Decay: 0.03, Release: 0.01}
	return synth.Patch{
		Name: "recorder",
		Oscs: []synth.Osc{
			{Wave: synth.Sine},
			{Wave: synth.Triangle, Level: 0.25, Semi: 12},
			{Wave: synth.Noise, Level: 0.03},
			{Wave: synth.Noise, Level: 0.15, Env: &chiff},
		},
		Amp:      synth.Envelope{Attack: 0.03, Decay: 0.2, Sustain: 0.85, Release: 0.1},
		Filter:   synth.Filter{Cutoff: 3500, KeyTrack: 0.5},
		LFOs:     []synth.LFO{{Target: synth.Vibrato, Rate: 5, Depth: 14, Delay: 0.25}},
		Velocity: 0.5,
		Gain:     0.25,
	}
}

// PipeOrgan is a church organ's principal chorus: open flues at 16', 8', 4', and 2', a
// stopped flute and a quint, each rank a cent or two off the next so they beat slowly,
// and the wind's chiff as a pipe speaks. It has no Leslie; a pipe organ moves in the room.
func PipeOrgan() synth.Patch {
	chiff := synth.Envelope{Attack: 0.005, Decay: 0.05, Release: 0.02}
	return synth.Patch{
		Name: "pipe organ",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, Semi: -12, Level: 0.5},
			{Wave: synth.Saw, Level: 0.6},
			{Wave: synth.Pulse, Level: 0.4, Cents: 1.5},
			{Wave: synth.Saw, Semi: 12, Level: 0.35, Cents: -1},
			{Wave: synth.Sine, Semi: 19, Level: 0.15},
			{Wave: synth.Saw, Semi: 24, Level: 0.12, Cents: 2},
			{Wave: synth.Noise, Level: 0.06, Env: &chiff},
		},
		Amp:    synth.Envelope{Attack: 0.06, Release: 0.25},
		Filter: synth.Filter{Cutoff: 2800, KeyTrack: 0.5},
		Gain:   0.06,
	}
}

// Tick is a clock's escapement: a click of band-passed noise and a ping too short to
// have a pitch.
func Tick() synth.Patch {
	return synth.Patch{
		Name:     "tick",
		Oscs:     []synth.Osc{{Wave: synth.Noise}, {Wave: synth.Sine, Fixed: 3200, Level: 0.5}},
		Amp:      synth.Envelope{Attack: 0.0005, Decay: 0.025, Release: 0.01},
		Filter:   synth.Filter{Mode: synth.BandPass, Cutoff: 4000, Resonance: 0.4},
		Velocity: 0.8,
		Gain:     0.3,
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

// GrandPiano is a concert grand played softly: three strings a key, struck near the
// middle by a felt hammer so the tone is round, with a long bloom and a slow release
// for notes the sustain pedal holds past their written length.
func GrandPiano() synth.Patch {
	hammer := synth.Envelope{Attack: 0.0005, Decay: 0.012, Release: 0.005}
	return synth.Patch{
		Name: "grand piano",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 9, Brightness: 0.55, Pick: 0.2}},
			{Wave: synth.Pluck, Level: 0.8, Cents: 1.8, String: synth.String{Decay: 8, Brightness: 0.5, Pick: 0.22}},
			{Wave: synth.Pluck, Level: 0.6, Cents: -1.4, String: synth.String{Decay: 10, Brightness: 0.45, Pick: 0.18}},
			{Wave: synth.Sine, Level: 0.2, Env: &synth.Envelope{Attack: 0.002, Decay: 2.5, Release: 0.3}},
			{Wave: synth.Noise, Level: 0.04, Env: &hammer},
		},
		Amp:      synth.Envelope{Attack: 0.001, Release: 0.6},
		Filter:   synth.Filter{Cutoff: 2600, KeyTrack: 0.6, Velocity: 1.5},
		Velocity: 0.75,
		Gain:     0.16,
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
