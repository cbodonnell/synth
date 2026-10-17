package patches

import "github.com/cbodonnell/synth"

// Guitars are plucked strings with long sustain; the amp is the part's insert chain,
// so a chord distorts as a chord and its notes intermodulate the way a real amp's do.
// GuitarAmp is that chain: fuzz, then a 4x12 cabinet's band-limit.
func GuitarAmp(fuzz float64) []synth.Effect {
	return []synth.Effect{
		synth.Drive{Amount: fuzz, Tone: 7000},
		synth.EQ{Bands: []synth.Band{
			{Kind: synth.LowCut, Freq: 90},
			{Kind: synth.Peak, Freq: 140, Gain: 2},
			{Kind: synth.Peak, Freq: 700, Gain: -2.5, Q: 0.8},
			{Kind: synth.Peak, Freq: 2200, Gain: 3, Q: 1},
			{Kind: synth.HighCut, Freq: 5200, Q: 0.6},
			{Kind: synth.HighCut, Freq: 7000},
		}},
	}
}

func RhythmGuitar() synth.Patch {
	return synth.Patch{
		Name: "rhythm guitar",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 7, Brightness: 0.55, Pick: 0.15}},
			{Wave: synth.Pluck, Level: 0.5, Cents: 3, String: synth.String{Decay: 6, Brightness: 0.6, Pick: 0.22}},
		},
		Amp:  synth.Envelope{Attack: 0.002, Release: 0.06},
		Gain: 0.3,
	}
}

// LeadGuitar is a wah into a fuzz, Hendrix's order. The wah is the filter, opened by
// each pick and settling back as the note sustains; vibrato waits until a note is held.
func LeadGuitar() synth.Patch {
	return synth.Patch{
		Name: "lead guitar",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 9, Brightness: 0.6, Pick: 0.18}},
		},
		Amp: synth.Envelope{Attack: 0.002, Release: 0.08},
		Filter: synth.Filter{
			Cutoff: 650, Resonance: 0.72, KeyTrack: 0.6,
			Env:       synth.Envelope{Attack: 0.04, Decay: 0.5, Sustain: 0.35},
			EnvAmount: 2,
		},
		LFOs:  []synth.LFO{{Target: synth.Vibrato, Rate: 5.6, Depth: 28, Delay: 0.22}},
		Drive: 9,
		Gain:  0.22,
	}
}

// Organ is a tonewheel organ: sine drawbars at the 16', 8', 5 1/3', 4', 2 2/3', and 2'
// footages, a key click, and a Leslie, which is three LFOs at one rate since the rotor
// moves pitch, level, and position all at once.
func Organ() synth.Patch {
	click := synth.Envelope{Attack: 0.0005, Decay: 0.004, Release: 0.003}
	const leslie = 6.4
	return synth.Patch{
		Name: "organ",
		Oscs: []synth.Osc{
			{Wave: synth.Sine, Semi: -12, Level: 0.7},
			{Wave: synth.Sine, Level: 0.8},
			{Wave: synth.Sine, Semi: 7, Level: 0.5},
			{Wave: synth.Sine, Semi: 12, Level: 0.45},
			{Wave: synth.Sine, Semi: 19, Level: 0.25},
			{Wave: synth.Sine, Semi: 24, Level: 0.2},
			{Wave: synth.Noise, Level: 0.05, Env: &click},
		},
		Amp: synth.Envelope{Attack: 0.006, Release: 0.06},
		LFOs: []synth.LFO{
			{Target: synth.Vibrato, Rate: leslie, Depth: 7},
			{Target: synth.Tremolo, Rate: leslie, Depth: 0.12},
			{Target: synth.AutoPan, Rate: leslie, Depth: 0.35},
		},
		Drive: 0.8,
		Gain:  0.07,
	}
}

// HofnerBass is a short-scale bass played with a pick near the bridge, warm and woody,
// with a sine under it to keep the low end round.
func HofnerBass() synth.Patch {
	return synth.Patch{
		Name: "hofner bass",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 3.5, Brightness: 0.45, Pick: 0.2}},
			{Wave: synth.Sine, Level: 0.5},
		},
		Amp:      synth.Envelope{Attack: 0.003, Decay: 1.2, Sustain: 0.6, Release: 0.07},
		Filter:   synth.Filter{Cutoff: 1100, Resonance: 0.15},
		Drive:    0.7,
		Velocity: 0.5,
		Gain:     0.45,
	}
}

// SynthArp is a Juno-style pluck: pulse and saw through a filter that snaps shut.
func SynthArp() synth.Patch {
	return synth.Patch{
		Name: "synth arp",
		Oscs: []synth.Osc{
			{Wave: synth.Pulse, Width: 0.4},
			{Wave: synth.Saw, Level: 0.6, Cents: -6},
		},
		Amp: synth.Envelope{Attack: 0.002, Decay: 0.25, Sustain: 0.2, Release: 0.12},
		Filter: synth.Filter{
			Cutoff: 500, Resonance: 0.35, KeyTrack: 0.4,
			Env:       synth.Envelope{Decay: 0.14},
			EnvAmount: 3.2,
			Velocity:  1,
		},
		LFOs:     []synth.LFO{{Target: synth.PWM, Rate: 0.5, Depth: 0.1}},
		Velocity: 0.5,
		Gain:     0.09,
	}
}

// Tom follows the played key, so one patch is a whole rack of toms.
func Tom() synth.Patch {
	skin := synth.Envelope{Attack: 0.0005, Decay: 0.03, Release: 0.01}
	return synth.Patch{
		Name: "tom",
		Oscs: []synth.Osc{
			{Wave: synth.Sine},
			{Wave: synth.Triangle, Semi: 7, Level: 0.25},
			{Wave: synth.Noise, Level: 0.25, Env: &skin},
		},
		Sweep:    synth.Sweep{Semis: 7, Time: 0.06},
		Amp:      synth.Envelope{Attack: 0.001, Decay: 0.45, Release: 0.05},
		Filter:   synth.Filter{Cutoff: 2500},
		Drive:    0.6,
		Velocity: 0.6,
		Gain:     0.55,
	}
}

// Cymbal is the TR-808's answer to metal: six square waves at inharmonic, fixed
// frequencies, high-passed, with noise for the wash. Decay picks crash or ride.
func Cymbal(decay float64) synth.Patch {
	sq := func(hz float64) synth.Osc { return synth.Osc{Wave: synth.Pulse, Fixed: hz, Level: 0.35} }
	return synth.Patch{
		Name: "cymbal",
		Oscs: []synth.Osc{
			sq(205.3), sq(304.4), sq(369.6), sq(522.7), sq(540), sq(800),
			{Wave: synth.Noise, Level: 0.8},
		},
		Amp:      synth.Envelope{Attack: 0.001, Decay: decay, Release: 0.1},
		Filter:   synth.Filter{Mode: synth.HighPass, Cutoff: 6000, Resonance: 0.15},
		Velocity: 0.7,
		Gain:     0.18,
	}
}
