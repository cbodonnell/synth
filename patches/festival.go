package patches

import "github.com/cbodonnell/synth"

// Sitar is a bright string over a flat bridge, the jawari, that lets it buzz. The buzz
// is a resonant filter thrown open by the pluck and closing as the note rings; a thin
// string an octave up stands in for the shimmer of the sympathetic strings.
func Sitar() synth.Patch {
	return synth.Patch{
		Name: "sitar",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 6, Brightness: 0.92, Pick: 0.05}},
			{Wave: synth.Pluck, Level: 0.25, Semi: 12, Cents: 3, String: synth.String{Decay: 8, Brightness: 0.85, Pick: 0.07}},
		},
		Amp: synth.Envelope{Attack: 0.002, Release: 0.3},
		Filter: synth.Filter{
			Cutoff: 1300, Resonance: 0.55, KeyTrack: 0.6,
			Env:       synth.Envelope{Decay: 0.7, Sustain: 0.2},
			EnvAmount: 2.2,
		},
		Drive: 1.5,
		Gain:  0.25,
	}
}

// Tanpura is the drone behind a raga: long open strings whose buzz blooms a moment
// after the pluck, as the string settles onto threads laid under it on the bridge.
func Tanpura() synth.Patch {
	return synth.Patch{
		Name: "tanpura",
		Oscs: []synth.Osc{
			{Wave: synth.Pluck, String: synth.String{Decay: 10, Brightness: 0.75, Pick: 0.12}},
			{Wave: synth.Pluck, Level: 0.6, Cents: 2, String: synth.String{Decay: 9, Brightness: 0.7, Pick: 0.2}},
		},
		Amp: synth.Envelope{Attack: 0.003, Release: 1.5},
		Filter: synth.Filter{
			Cutoff: 800, Resonance: 0.5, KeyTrack: 0.5,
			Env:       synth.Envelope{Attack: 0.4, Decay: 1.6, Sustain: 0.3, Release: 1},
			EnvAmount: 1.6,
		},
		Gain: 0.2,
	}
}

// Handpan is a steel shell hammered into tuned dimples and struck with the fingers:
// each note tuned with an octave and a compound fifth above it, which die sooner than
// the fundamental, and a soft knock of skin on steel.
func Handpan() synth.Patch {
	knock := synth.Envelope{Attack: 0.001, Decay: 0.02, Release: 0.01}
	return synth.Patch{
		Name: "handpan",
		Oscs: []synth.Osc{
			{Wave: synth.Sine},
			{Wave: synth.Sine, Semi: 12, Cents: 1.5, Level: 0.3, Env: &synth.Envelope{Attack: 0.003, Decay: 1.2, Release: 0.4}},
			{Wave: synth.Sine, Semi: 19, Level: 0.12, Env: &synth.Envelope{Attack: 0.003, Decay: 0.6, Release: 0.3}},
			{Wave: synth.Noise, Level: 0.05, Env: &knock},
		},
		Amp:      synth.Envelope{Attack: 0.004, Decay: 3, Release: 0.8},
		Filter:   synth.Filter{Cutoff: 4000},
		Velocity: 0.7,
		Gain:     0.3,
	}
}

// Didgeridoo is lips buzzing into a long tube: saw and pulse through a resonant
// lowpass, the vowel the mouth shapes as the filter's envelope, and the tongue's
// flutter as a quick wah. A note well above the drone is the overblown toot.
func Didgeridoo() synth.Patch {
	return synth.Patch{
		Name: "didgeridoo",
		Oscs: []synth.Osc{
			{Wave: synth.Saw},
			{Wave: synth.Pulse, Width: 0.3, Level: 0.6, Cents: -4},
			{Wave: synth.Sine, Level: 0.5},
		},
		Amp: synth.Envelope{Attack: 0.05, Release: 0.2},
		Filter: synth.Filter{
			Cutoff: 320, Resonance: 0.7, KeyTrack: 0.8,
			Env:       synth.Envelope{Attack: 0.08, Decay: 0.4, Sustain: 0.3, Release: 0.2},
			EnvAmount: 1.5,
		},
		LFOs:     []synth.LFO{{Target: synth.Wah, Rate: 4.5, Depth: 0.3, Delay: 0.2}},
		Drive:    2,
		Velocity: 0.5,
		Gain:     0.15,
	}
}

// HandDrum is a djembe whose pitch follows the played key, so a low key is the bass
// tone struck in the middle of the skin and a high one the open tone at the rim.
// The slap is its own patch.
func HandDrum() synth.Patch {
	skin := synth.Envelope{Attack: 0.0005, Decay: 0.015, Release: 0.01}
	return synth.Patch{
		Name: "hand drum",
		Oscs: []synth.Osc{
			{Wave: synth.Sine},
			{Wave: synth.Triangle, Semi: 12, Level: 0.2},
			{Wave: synth.Noise, Level: 0.35, Env: &skin},
		},
		Sweep:    synth.Sweep{Semis: 4, Time: 0.03},
		Amp:      synth.Envelope{Attack: 0.001, Decay: 0.22, Release: 0.05},
		Filter:   synth.Filter{Cutoff: 3000},
		Drive:    0.8,
		Velocity: 0.7,
		Gain:     0.6,
	}
}

// Slap is the fingers cracking against a djembe's rim: nearly all skin, hardly a pitch.
func Slap() synth.Patch {
	body := synth.Envelope{Attack: 0.0005, Decay: 0.03, Release: 0.01}
	return synth.Patch{
		Name: "slap",
		Oscs: []synth.Osc{
			{Wave: synth.Noise},
			{Wave: synth.Sine, Fixed: 420, Level: 0.4, Env: &body},
		},
		Amp:      synth.Envelope{Attack: 0.0005, Decay: 0.07, Release: 0.02},
		Filter:   synth.Filter{Mode: synth.BandPass, Cutoff: 2200, Resonance: 0.35},
		Velocity: 0.7,
		Gain:     0.5,
	}
}

// Shaker is seeds in a gourd: high noise with a soft attack, since the seeds land a
// little apart.
func Shaker() synth.Patch {
	return synth.Patch{
		Name:     "shaker",
		Oscs:     []synth.Osc{{Wave: synth.Noise}},
		Amp:      synth.Envelope{Attack: 0.012, Decay: 0.05, Release: 0.03},
		Filter:   synth.Filter{Mode: synth.HighPass, Cutoff: 5500, Resonance: 0.2},
		Velocity: 0.8,
		Gain:     0.25,
	}
}

// Wind is noise through a resonant bandpass that drifts slowly and gusts.
func Wind() synth.Patch {
	return synth.Patch{
		Name:   "wind",
		Oscs:   []synth.Osc{{Wave: synth.Noise}},
		Amp:    synth.Envelope{Attack: 1.5, Release: 2.5},
		Filter: synth.Filter{Mode: synth.BandPass, Cutoff: 600, Resonance: 0.55},
		LFOs: []synth.LFO{
			{Target: synth.Wah, Rate: 0.12, Depth: 1.3},
			{Target: synth.Tremolo, Rate: 0.3, Depth: 0.3},
		},
		Gain: 0.15,
	}
}
