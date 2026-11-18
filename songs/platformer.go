package songs

import (
	"github.com/cbodonnell/synth"
)

// Platformer is a side-scrolling MMO's effects: movement, combat, pickups, UI, and the
// little fanfares. Each is a sound effect of layered tracks, the way a sound designer
// stacks a transient, a body, and a tail. Occasional sounds share one small room;
// anything a game fires many times a second stays dry, since even a faint tail keeps a
// voice alive and the player only holds so many. Noise layers play at C4, where a
// filter's key tracking leaves its cutoff alone.
//
// A master Normalize sets each one's level by tier: UI at -12 dB, movement and pickups
// -10 (the coins' bright pings -13), combat -2, fanfares -3. Priority follows the game's
// tiers, so a flood of footsteps or coins gives way before a blow, and a blow before a
// fanfare. Whatever a game fires over and over varies each play and is capped, so it
// can't crowd out the rest.
func Platformer() []synth.Arrangement {
	return []synth.Arrangement{
		jumpSound, doublejumpSound, landSound, climbSound, pickupSound, mesoSound, dropSound,
		potionSound, buffSound, portalSound, swingSound, hitSound, critSound, magicSound,
		hurtSound, monsterdieSound, clickSound, openSound, closeSound, popSound, errorSound,
		deathSound, levelupSound, questSound,
	}
}

// A springy blip that rises into its note, over a puff of air from the feet.
var jumpSound = synth.Arrangement{
	ID:   "jump",
	Name: "jump",
	Tracks: []synth.Track{
		{
			Name: "blip",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Pulse, Width: 0.3}, {Wave: synth.Triangle, Level: 0.6, Semi: -12}},
				Amp:    synth.Envelope{Attack: 0.002, Decay: 0.13, Release: 0.03},
				Filter: synth.Filter{Cutoff: 3000},
				Sweep:  synth.Sweep{Semis: -9, Time: 0.03},
				Gain:   0.3,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.08, Key: 69, Slide: 4}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "puff",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Noise}},
				Amp:    synth.Envelope{Attack: 0.001, Decay: 0.05, Release: 0.01},
				Filter: synth.Filter{Mode: synth.BandPass, Cutoff: 1400, Resonance: 0.3},
				Gain:   0.25,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.03, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master: []synth.Effect{synth.Normalize{Peak: -10}},
	Vary:   synth.Variation{Pitch: 0.5, Gain: 1, Takes: 2},
	Voices: 2,
}

// The second jump in the air: a whoosh that swells and passes, under a rising whistle.
var doublejumpSound = synth.Arrangement{
	ID:   "doublejump",
	Name: "doublejump",
	Tracks: []synth.Track{
		{
			Name: "whoosh",
			Patch: synth.Patch{
				Oscs: []synth.Osc{{Wave: synth.Noise}},
				Amp:  synth.Envelope{Attack: 0.04, Decay: 0.28, Release: 0.01},
				Filter: synth.Filter{
					Mode:      synth.BandPass,
					Cutoff:    500,
					Resonance: 0.55,
					Env:       synth.Envelope{Attack: 0.08, Decay: 0.2},
					EnvAmount: 2.5,
				},
				LFOs: []synth.LFO{{Target: synth.AutoPan, Shape: synth.SawShape, Rate: 2.5, Depth: 0.5}},
				Gain: 0.9,
			},
			Sends: []synth.Send{{Bus: "room", Level: -16.5}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.2, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "whistle",
			Patch: synth.Patch{
				Oscs: []synth.Osc{{}, {Wave: synth.Triangle, Level: 0.3, Semi: 12}},
				Amp:  synth.Envelope{Attack: 0.01, Decay: 0.22, Release: 0.05},
				Gain: 0.2,
			},
			Sends: []synth.Send{{Bus: "room", Level: -16.5}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.15, Key: 76, Slide: 12, SlideTime: 0.15}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Buses: []synth.Bus{
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.25, Damp: 0.6, PreDelay: 0.01}}},
	},
	Master: []synth.Effect{synth.Normalize{Peak: -10}},
}

// Feet on the ground: a dull thud and a scuff.
var landSound = synth.Arrangement{
	ID:   "land",
	Name: "land",
	Tracks: []synth.Track{
		{
			Name: "thud",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{Fixed: 80}},
				Amp:   synth.Envelope{Attack: 0.001, Decay: 0.09, Release: 0.02},
				Sweep: synth.Sweep{Semis: 10, Time: 0.02},
				Drive: 0.8,
				Gain:  0.3,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.05, Key: 36}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "scuff",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Noise}},
				Amp:    synth.Envelope{Attack: 0.001, Decay: 0.05, Release: 0.01},
				Filter: synth.Filter{Cutoff: 900, Resonance: 0.1},
				Gain:   0.2,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{At: 0.004, Dur: 0.03, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master: []synth.Effect{synth.Normalize{Peak: -10}},
	Vary:   synth.Variation{Pitch: 1, Gain: 1.5, Takes: 3},
	Voices: 2,
}

// One hand on a rope or rung: a wooden tock. A game plays it every few frames of climbing.
var climbSound = synth.Arrangement{
	ID:   "climb",
	Name: "climb",
	Tracks: []synth.Track{
		{
			Name: "tock",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{Wave: synth.Triangle, Fixed: 520}, {Level: 0.3, Fixed: 1240}},
				Amp:   synth.Envelope{Attack: 0.0005, Decay: 0.035, Release: 0.01},
				Sweep: synth.Sweep{Semis: 5, Time: 0.005},
				Gain:  0.3,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.02, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "grain",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Noise}},
				Amp:    synth.Envelope{Attack: 0.001, Decay: 0.015, Release: 0.01},
				Filter: synth.Filter{Mode: synth.BandPass, Cutoff: 2500, Resonance: 0.4},
				Gain:   0.25,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.01, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master: []synth.Effect{synth.Normalize{Peak: -10}},
	Vary:   synth.Variation{Pitch: 1.5, Gain: 2, Takes: 3},
	Voices: 2,
}

// Picking an item up: a quick, soft bloop upward.
var pickupSound = synth.Arrangement{
	ID:   "pickup",
	Name: "pickup",
	Tracks: []synth.Track{
		{
			Name: "bloop",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{FM: synth.FM{Ratio: 2, Index: 1, Decay: 0.04}}, {Wave: synth.Triangle, Level: 0.4}},
				Amp:   synth.Envelope{Attack: 0.002, Decay: 0.12, Release: 0.03},
				Sweep: synth.Sweep{Semis: -5, Time: 0.015},
				Gain:  0.35,
			},
			Clips: []synth.Clip{
				{
					Name:  "a",
					Notes: []synth.Note{{Dur: 0.044, Key: 76, Vel: 0.8}, {At: 0.055, Dur: 0.044, Key: 83, Vel: 0.8}},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -10}},
	Vary:     synth.Variation{Pitch: 0.25, Gain: 1},
	Voices:   3,
	Priority: 1,
}

// Coins into the purse: two bright pings and the jingle of a few more behind them.
var mesoSound = synth.Arrangement{
	ID:   "meso",
	Name: "meso",
	Tracks: []synth.Track{
		{
			Name: "ping",
			Patch: synth.Patch{
				Oscs:     []synth.Osc{{FM: synth.FM{Ratio: 2.76, Index: 2.2, Decay: 0.05, Sustain: 0.25}}},
				Amp:      synth.Envelope{Attack: 0.0008, Decay: 0.3, Release: 0.05},
				Velocity: 0.7,
				Gain:     0.3,
			},
			Clips: []synth.Clip{
				{
					Name:  "a",
					Notes: []synth.Note{{Dur: 0.06, Key: 95, Vel: 0.8}, {At: 0.06, Dur: 0.12, Key: 100, Vel: 0.8}},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "jingle",
			Patch: synth.Patch{
				Oscs:     []synth.Osc{{FM: synth.FM{Ratio: 2.76, Index: 2.2, Decay: 0.05, Sustain: 0.25}}},
				Amp:      synth.Envelope{Attack: 0.0008, Decay: 0.3, Release: 0.05},
				Velocity: 0.7,
				Gain:     0.3,
			},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{At: 0.0912, Dur: 0.028, Key: 102, Vel: 0.34},
						{At: 0.1225, Dur: 0.028, Key: 98, Vel: 0.35},
						{At: 0.1492, Dur: 0.028, Key: 105, Vel: 0.34},
						{At: 0.1764, Dur: 0.028, Key: 100, Vel: 0.35},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -13}},
	Vary:     synth.Variation{Pitch: 0.3, Gain: 1},
	Voices:   4,
	Priority: 1,
}

// An item dropping from a monster and bouncing to rest, each bounce lower and sooner.
var dropSound = synth.Arrangement{
	ID:   "drop",
	Name: "drop",
	Tracks: []synth.Track{
		{
			Name: "plink",
			Patch: synth.Patch{
				Oscs:     []synth.Osc{{FM: synth.FM{Ratio: 3, Index: 1.2, Decay: 0.03}}},
				Amp:      synth.Envelope{Attack: 0.001, Decay: 0.09, Release: 0.02},
				Sweep:    synth.Sweep{Semis: 3, Time: 0.01},
				Velocity: 1,
				Gain:     0.35,
			},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{Dur: 0.03, Key: 79, Vel: 1},
						{At: 0.16, Dur: 0.03, Key: 79, Vel: 0.6},
						{At: 0.256, Dur: 0.03, Key: 79, Vel: 0.36},
						{At: 0.3136, Dur: 0.03, Key: 79, Vel: 0.216},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -10}},
	Priority: 1,
}

// Drinking a potion: three rising glugs, then a warm sparkle as it heals.
var potionSound = synth.Arrangement{
	ID:   "potion",
	Name: "potion",
	Tracks: []synth.Track{
		{
			Name: "glug",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{}, {Wave: synth.Triangle, Level: 0.3}},
				Amp:    synth.Envelope{Attack: 0.003, Decay: 0.07, Release: 0.02},
				Filter: synth.Filter{Cutoff: 1800},
				Sweep:  synth.Sweep{Semis: -8, Time: 0.025},
				Gain:   0.45,
			},
			Sends: []synth.Send{{Bus: "room", Level: -12}},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{Dur: 0.045, Key: 60, Vel: 0.8},
						{At: 0.09, Dur: 0.045, Key: 62, Vel: 0.8},
						{At: 0.18, Dur: 0.045, Key: 65, Vel: 0.8},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "heal",
			Patch: synth.Patch{
				Oscs: []synth.Osc{
					{FM: synth.FM{Ratio: 2, Index: 1.8, Decay: 0.08, Sustain: 0.2}},
					{Level: 0.25, Semi: 12, FM: synth.FM{Ratio: 3.5, Index: 1, Decay: 0.05}},
				},
				Amp:      synth.Envelope{Attack: 0.001, Decay: 0.4, Release: 0.1},
				Velocity: 0.6,
				Gain:     0.3,
			},
			Sends: []synth.Send{{Bus: "room", Level: -12}},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{At: 0.3, Dur: 0.05, Key: 84, Vel: 0.6},
						{At: 0.35, Dur: 0.05, Key: 88, Vel: 0.6},
						{At: 0.4, Dur: 0.05, Key: 91, Vel: 0.6},
						{At: 0.45, Dur: 0.15, Key: 96, Vel: 0.6},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Buses: []synth.Bus{
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.25, Damp: 0.6, PreDelay: 0.01}}},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -10}},
	Priority: 1,
}

// Casting a buff: a detuned stack rising and opening as it swirls, then a shimmer.
var buffSound = synth.Arrangement{
	ID:   "buff",
	Name: "buff",
	Tracks: []synth.Track{
		{
			Name: "rise",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Saw}, {Wave: synth.Saw, Level: 0.6, Semi: 7}},
				Unison: synth.Unison{Voices: 5, Detune: 30, Spread: 0.9},
				Amp:    synth.Envelope{Attack: 0.05, Decay: 0.6, Release: 0.15},
				Filter: synth.Filter{
					Cutoff:    300,
					Resonance: 0.35,
					Env:       synth.Envelope{Attack: 0.45, Decay: 0.3},
					EnvAmount: 4,
				},
				Gain: 0.12,
			},
			Inserts: []synth.Effect{synth.Chorus{Mix: 0.5}},
			Sends:   []synth.Send{{Bus: "room", Level: -9.1}},
			Clips:   []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.5, Key: 60, Slide: 12}}}},
			Play:    []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "shimmer",
			Patch: synth.Patch{
				Oscs: []synth.Osc{
					{FM: synth.FM{Ratio: 2, Index: 1.8, Decay: 0.08, Sustain: 0.2}},
					{Level: 0.25, Semi: 12, FM: synth.FM{Ratio: 3.5, Index: 1, Decay: 0.05}},
				},
				Amp:      synth.Envelope{Attack: 0.001, Decay: 0.35, Release: 0.1},
				Velocity: 0.6,
				Gain:     0.3,
			},
			Sends: []synth.Send{{Bus: "room", Level: -9.1}},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{At: 0.4, Dur: 0.03, Key: 91, Vel: 0.5},
						{At: 0.43, Dur: 0.03, Key: 96, Vel: 0.5},
						{At: 0.46, Dur: 0.03, Key: 100, Vel: 0.5},
						{At: 0.49, Dur: 0.03, Key: 103, Vel: 0.5},
						{At: 0.52, Dur: 0.09, Key: 108, Vel: 0.5},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Buses: []synth.Bus{
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.25, Damp: 0.6, PreDelay: 0.01}}},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -10}},
	Priority: 1,
}

// Stepping through a portal: a warbling rise through a phaser, swallowed by the room.
var portalSound = synth.Arrangement{
	ID:   "portal",
	Name: "portal",
	Tracks: []synth.Track{
		{
			Name: "warp",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Saw}, {Level: 0.6, Semi: -12}},
				Unison: synth.Unison{Voices: 3, Detune: 20, Spread: 0.8},
				Amp:    synth.Envelope{Attack: 0.08, Decay: 0.6, Release: 0.2},
				Filter: synth.Filter{
					Cutoff:    400,
					Resonance: 0.5,
					Env:       synth.Envelope{Attack: 0.3, Decay: 0.4},
					EnvAmount: 3,
				},
				LFOs: []synth.LFO{{Rate: 11, Depth: 50}},
				Gain: 0.3,
			},
			Inserts: []synth.Effect{synth.Phaser{Rate: 2, Depth: 2, Feedback: 0.6}},
			Sends:   []synth.Send{{Bus: "room", Level: -8}},
			Clips:   []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.5, Key: 48, Slide: 24}}}},
			Play:    []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "swirl",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Noise}},
				Amp:    synth.Envelope{Attack: 0.15, Decay: 0.5, Release: 0.01},
				Filter: synth.Filter{Mode: synth.BandPass, Cutoff: 1200, Resonance: 0.7},
				LFOs: []synth.LFO{
					{Target: synth.Wah, Rate: 6, Depth: 1},
					{Target: synth.AutoPan, Rate: 3, Depth: 0.6},
				},
				Gain: 0.55,
			},
			Sends: []synth.Send{{Bus: "room", Level: -8}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.4, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Buses: []synth.Bus{
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.25, Damp: 0.6, PreDelay: 0.01}}},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -10}},
	Priority: 1,
}

// A weapon cutting the air, carried across the field as it goes.
var swingSound = synth.Arrangement{
	ID:   "swing",
	Name: "swing",
	Tracks: []synth.Track{
		{
			Name: "air",
			Patch: synth.Patch{
				Oscs: []synth.Osc{{Wave: synth.Noise}},
				Amp:  synth.Envelope{Attack: 0.03, Decay: 0.14, Release: 0.01},
				Filter: synth.Filter{
					Mode:      synth.BandPass,
					Cutoff:    450,
					Resonance: 0.6,
					Env:       synth.Envelope{Attack: 0.05, Decay: 0.1},
					EnvAmount: 3,
				},
				LFOs: []synth.LFO{{Target: synth.AutoPan, Shape: synth.SawShape, Rate: 4, Depth: 0.6}},
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.1, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "edge",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Noise}},
				Amp:    synth.Envelope{Attack: 0.04, Decay: 0.08, Release: 0.01},
				Filter: synth.Filter{Mode: synth.HighPass, Cutoff: 5000, Resonance: 0.3},
				Gain:   0.15,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.05, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -2}},
	Vary:     synth.Variation{Pitch: 1, Gain: 1, Takes: 3},
	Voices:   3,
	Priority: 2,
}

// Landing a blow on a monster: crack, slap, and weight.
var hitSound = synth.Arrangement{
	ID:   "hit",
	Name: "hit",
	Tracks: []synth.Track{
		{
			Name: "crack",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Noise}},
				Amp:    synth.Envelope{Attack: 0.001, Decay: 0.045, Release: 0.01},
				Filter: synth.Filter{Mode: synth.HighPass, Cutoff: 1800, Resonance: 0.2},
				Gain:   0.5,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.02, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "slap",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{Wave: synth.Triangle, Fixed: 340}},
				Amp:   synth.Envelope{Attack: 0.0005, Decay: 0.05, Release: 0.01},
				Sweep: synth.Sweep{Semis: 14, Time: 0.012},
				Drive: 2,
				Gain:  0.3,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.02, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "body",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{Fixed: 110}},
				Amp:   synth.Envelope{Attack: 0.001, Decay: 0.12, Release: 0.02},
				Sweep: synth.Sweep{Semis: 24, Time: 0.02},
				Drive: 2,
				Gain:  0.38,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.05, Key: 36}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -2}},
	Vary:     synth.Variation{Pitch: 1.5, Gain: 1.5, Takes: 4},
	Voices:   4,
	Priority: 2,
}

// A critical: the same blow with a metallic ring and a sparkle on top of it.
var critSound = synth.Arrangement{
	ID:   "crit",
	Name: "crit",
	Tracks: []synth.Track{
		{
			Name: "crack",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Noise}},
				Amp:    synth.Envelope{Attack: 0.001, Decay: 0.045, Release: 0.01},
				Filter: synth.Filter{Mode: synth.HighPass, Cutoff: 1800, Resonance: 0.2},
				Gain:   0.5,
			},
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.02, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "slap",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{Wave: synth.Triangle, Fixed: 340}},
				Amp:   synth.Envelope{Attack: 0.0005, Decay: 0.05, Release: 0.01},
				Sweep: synth.Sweep{Semis: 14, Time: 0.012},
				Drive: 2,
				Gain:  0.3,
			},
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.02, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "body",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{Fixed: 110}},
				Amp:   synth.Envelope{Attack: 0.001, Decay: 0.12, Release: 0.02},
				Sweep: synth.Sweep{Semis: 24, Time: 0.02},
				Drive: 2,
				Gain:  0.38,
			},
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.05, Key: 36}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "ring",
			Patch: synth.Patch{
				Oscs:     []synth.Osc{{FM: synth.FM{Ratio: 3.5, Index: 3, Decay: 0.15, Sustain: 0.1}}},
				Amp:      synth.Envelope{Attack: 0.001, Decay: 0.5, Release: 0.1},
				Velocity: 0.5,
				Gain:     0.25,
			},
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.05, Key: 88}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "sparkle",
			Patch: synth.Patch{
				Oscs: []synth.Osc{
					{FM: synth.FM{Ratio: 2, Index: 1.8, Decay: 0.08, Sustain: 0.2}},
					{Level: 0.25, Semi: 12, FM: synth.FM{Ratio: 3.5, Index: 1, Decay: 0.05}},
				},
				Amp:      synth.Envelope{Attack: 0.001, Decay: 0.25, Release: 0.1},
				Velocity: 0.6,
				Gain:     0.3,
			},
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{At: 0.03, Dur: 0.021, Key: 95, Vel: 0.6},
						{At: 0.065, Dur: 0.021, Key: 100, Vel: 0.6},
						{At: 0.1, Dur: 0.021, Key: 107, Vel: 0.6},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Buses: []synth.Bus{
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.25, Damp: 0.6, PreDelay: 0.01}}},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -2}},
	Priority: 2,
}

// A bolt of magic: an FM zap falling from on high, with a crackle and an echo behind it.
var magicSound = synth.Arrangement{
	ID:   "magic",
	Name: "magic",
	Tracks: []synth.Track{
		{
			Name: "zap",
			Patch: synth.Patch{
				Oscs: []synth.Osc{
					{FM: synth.FM{Ratio: 1.41, Index: 6, Decay: 0.12, Sustain: 0.2}},
					{Wave: synth.Saw, Level: 0.25, Semi: 7},
				},
				Amp:    synth.Envelope{Attack: 0.002, Decay: 0.3, Release: 0.05},
				Filter: synth.Filter{Cutoff: 2500, Resonance: 0.4, Env: synth.Envelope{Decay: 0.15}, EnvAmount: 2},
				LFOs:   []synth.LFO{{Shape: synth.RandomShape, Rate: 40, Depth: 40}},
				Sweep:  synth.Sweep{Semis: 24, Time: 0.06},
				Gain:   0.6,
			},
			Inserts: []synth.Effect{synth.Delay{Time: 0.11, Feedback: 0.35, Damp: 3000, PingPong: true, Mix: 0.3}},
			Sends:   []synth.Send{{Bus: "room", Level: -14}},
			Clips:   []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.15, Key: 69}}}},
			Play:    []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "crackle",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Noise}},
				Amp:    synth.Envelope{Attack: 0.001, Decay: 0.2, Release: 0.01},
				Filter: synth.Filter{Mode: synth.HighPass, Cutoff: 3000, Resonance: 0.2},
				LFOs:   []synth.LFO{{Target: synth.Tremolo, Shape: synth.RandomShape, Rate: 60, Depth: 0.9}},
				Gain:   0.4,
			},
			Sends: []synth.Send{{Bus: "room", Level: -14}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{At: 0.01, Dur: 0.15, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Buses: []synth.Bus{
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.25, Damp: 0.6, PreDelay: 0.01}}},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -2}},
	Priority: 2,
}

// Taking damage: a crunchy, stuttering drop in pitch.
var hurtSound = synth.Arrangement{
	ID:   "hurt",
	Name: "hurt",
	Tracks: []synth.Track{
		{
			Name: "ouch",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Pulse, Width: 0.5}, {Wave: synth.Saw, Level: 0.5, Cents: 15}},
				Amp:    synth.Envelope{Attack: 0.002, Decay: 0.25, Release: 0.04},
				Filter: synth.Filter{Cutoff: 2500},
				LFOs:   []synth.LFO{{Target: synth.Tremolo, Shape: synth.SquareShape, Rate: 28, Depth: 0.6}},
				Gain:   0.25,
			},
			Inserts: []synth.Effect{synth.Crush{Bits: 6, Hold: 3}},
			Clips:   []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.2, Key: 74, Slide: -9}}}},
			Play:    []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "thud",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{Fixed: 90}},
				Amp:   synth.Envelope{Attack: 0.001, Decay: 0.1, Release: 0.02},
				Sweep: synth.Sweep{Semis: 12, Time: 0.02},
				Drive: 0.8,
				Gain:  0.4,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.05, Key: 36}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -2}},
	Vary:     synth.Variation{Pitch: 1, Gain: 1, Takes: 2},
	Voices:   2,
	Priority: 2,
}

// A monster's last: a falling squeal that dulls as it goes, and a puff of smoke.
var monsterdieSound = synth.Arrangement{
	ID:   "monsterdie",
	Name: "monsterdie",
	Tracks: []synth.Track{
		{
			Name: "squeal",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Saw}, {Wave: synth.Pulse, Level: 0.5, Semi: -5, Width: 0.3}},
				Amp:    synth.Envelope{Attack: 0.003, Decay: 0.4, Release: 0.05},
				Filter: synth.Filter{Cutoff: 500, Resonance: 0.3, Env: synth.Envelope{Decay: 0.35}, EnvAmount: 2.5},
				LFOs:   []synth.LFO{{Rate: 14, Depth: 60}},
				Gain:   0.38,
			},
			Sends: []synth.Send{{Bus: "room", Level: -12}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.3, Key: 67, Slide: -17}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "poof",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Noise}},
				Amp:    synth.Envelope{Attack: 0.02, Decay: 0.35, Release: 0.01},
				Filter: synth.Filter{Cutoff: 1500, Resonance: 0.15, Env: synth.Envelope{Decay: 0.3}, EnvAmount: 1},
				Gain:   0.6,
			},
			Sends: []synth.Send{{Bus: "room", Level: -12}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{At: 0.08, Dur: 0.25, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Buses: []synth.Bus{
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.25, Damp: 0.6, PreDelay: 0.01}}},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -2}},
	Vary:     synth.Variation{Pitch: 1, Gain: 1},
	Voices:   4,
	Priority: 2,
}

// A UI tick, quiet enough to hear on every button.
var clickSound = synth.Arrangement{
	ID:   "click",
	Name: "click",
	Tracks: []synth.Track{
		{
			Name: "tick",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{Fixed: 2200}, {Wave: synth.Noise, Level: 0.3}},
				Amp:   synth.Envelope{Attack: 0.0005, Decay: 0.018, Release: 0.005},
				Sweep: synth.Sweep{Semis: 4, Time: 0.004},
				Gain:  0.4,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.01, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -12}},
	Voices:   2,
	Priority: 2,
}

// A window opening, and closing: the same two tones, up and then down.
var openSound = synth.Arrangement{
	ID:   "open",
	Name: "open",
	Tracks: []synth.Track{
		{
			Name: "tone",
			Patch: synth.Patch{
				Oscs: []synth.Osc{{FM: synth.FM{Ratio: 1, Index: 0.8, Decay: 0.03}}},
				Amp:  synth.Envelope{Attack: 0.001, Decay: 0.08, Release: 0.02},
				Gain: 0.3,
			},
			Clips: []synth.Clip{
				{
					Name:  "a",
					Notes: []synth.Note{{Dur: 0.036, Key: 79, Vel: 0.8}, {At: 0.045, Dur: 0.036, Key: 86, Vel: 0.8}},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -12}},
	Priority: 2,
}

var closeSound = synth.Arrangement{
	ID:   "close",
	Name: "close",
	Tracks: []synth.Track{
		{
			Name: "tone",
			Patch: synth.Patch{
				Oscs: []synth.Osc{{FM: synth.FM{Ratio: 1, Index: 0.8, Decay: 0.03}}},
				Amp:  synth.Envelope{Attack: 0.001, Decay: 0.08, Release: 0.02},
				Gain: 0.3,
			},
			Clips: []synth.Clip{
				{
					Name:  "a",
					Notes: []synth.Note{{Dur: 0.036, Key: 86, Vel: 0.8}, {At: 0.045, Dur: 0.036, Key: 79, Vel: 0.8}},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -12}},
	Priority: 2,
}

// A chat bubble or an NPC's dialog appearing.
var popSound = synth.Arrangement{
	ID:   "pop",
	Name: "pop",
	Tracks: []synth.Track{
		{
			Name: "pop",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{}},
				Amp:   synth.Envelope{Attack: 0.001, Decay: 0.06, Release: 0.01},
				Sweep: synth.Sweep{Semis: -14, Time: 0.012},
				Gain:  0.45,
			},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.03, Key: 81}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -12}},
	Priority: 2,
}

// Not allowed: inventory full, not enough mana. Two low buzzes, the second a step down.
var errorSound = synth.Arrangement{
	ID:   "error",
	Name: "error",
	Tracks: []synth.Track{
		{
			Name: "buzz",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Pulse, Width: 0.4}, {Wave: synth.Saw, Level: 0.6, Cents: -12}},
				Amp:    synth.Envelope{Attack: 0.003, Decay: 0.1, Sustain: 0.7, Release: 0.03},
				Filter: synth.Filter{Cutoff: 1400, Resonance: 0.2},
				Gain:   0.2,
			},
			Clips: []synth.Clip{
				{
					Name:  "a",
					Notes: []synth.Note{{Dur: 0.075, Key: 52, Vel: 0.8}, {At: 0.1, Dur: 0.075, Key: 48, Vel: 0.8}},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -12}},
	Priority: 2,
}

// The player falls: a slow minor descent, and a tombstone landing hard after it.
var deathSound = synth.Arrangement{
	ID:   "death",
	Name: "death",
	Tracks: []synth.Track{
		{
			Name: "fall",
			Patch: synth.Patch{
				Name:   "lead",
				Oscs:   []synth.Osc{{Wave: synth.Pulse, Width: 0.25}, {Wave: synth.Triangle, Level: 0.5, Semi: -12}},
				Amp:    synth.Envelope{Attack: 0.005, Decay: 0.1, Sustain: 0.7, Release: 0.05},
				Filter: synth.Filter{Cutoff: 3500, KeyTrack: 0.3},
				LFOs:   []synth.LFO{{Rate: 5.5, Depth: 25, Delay: 0.15}},
				Gain:   0.2,
			},
			Sends: []synth.Send{{Bus: "room", Level: -10.5}},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{Dur: 0.144, Key: 67, Vel: 0.8},
						{At: 0.16, Dur: 0.144, Key: 63, Vel: 0.8},
						{At: 0.32, Dur: 0.144, Key: 60, Vel: 0.8},
						{At: 0.48, Dur: 0.576, Key: 59, Vel: 0.8},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "tomb",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{Fixed: 55}},
				Amp:   synth.Envelope{Attack: 0.001, Decay: 0.4, Release: 0.02},
				Sweep: synth.Sweep{Semis: 18, Time: 0.02},
				Drive: 1.5,
				Gain:  0.3,
			},
			Sends: []synth.Send{{Bus: "room", Level: -10.5}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{At: 1.05, Dur: 0.1, Key: 36}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "dust",
			Patch: synth.Patch{
				Oscs:   []synth.Osc{{Wave: synth.Noise}},
				Amp:    synth.Envelope{Attack: 0.001, Decay: 0.3, Release: 0.01},
				Filter: synth.Filter{Cutoff: 700, Resonance: 0.1},
				Gain:   1.5,
			},
			Sends: []synth.Send{{Bus: "room", Level: -10.5}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{At: 1.055, Dur: 0.2, Key: 60}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Buses: []synth.Bus{
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.25, Damp: 0.6, PreDelay: 0.01}}},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -3}},
	Priority: 3,
}

// Level up: a square-wave arpeggio onto a held chord, bells ringing it, a sparkle run
// climbing over the top.
var levelupSound = synth.Arrangement{
	ID:   "levelup",
	Name: "levelup",
	Tracks: []synth.Track{
		{
			Name: "lead",
			Patch: synth.Patch{
				Name:   "lead",
				Oscs:   []synth.Osc{{Wave: synth.Pulse, Width: 0.25}, {Wave: synth.Triangle, Level: 0.5, Semi: -12}},
				Amp:    synth.Envelope{Attack: 0.005, Decay: 0.1, Sustain: 0.7, Release: 0.05},
				Filter: synth.Filter{Cutoff: 3500, KeyTrack: 0.3},
				LFOs:   []synth.LFO{{Rate: 5.5, Depth: 25, Delay: 0.15}},
				Gain:   0.2,
			},
			Sends: []synth.Send{{Bus: "room", Level: -9.1}, {Bus: "echo", Level: -12}},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{Dur: 0.0765, Key: 67, Vel: 0.8},
						{At: 0.085, Dur: 0.0765, Key: 72, Vel: 0.8},
						{At: 0.17, Dur: 0.0765, Key: 76, Vel: 0.8},
						{At: 0.255, Dur: 0.0765, Key: 79, Vel: 0.8},
						{At: 0.425, Dur: 0.0765, Key: 76, Vel: 0.8},
						{At: 0.51, Dur: 0.612, Key: 84, Vel: 1},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "chord",
			Patch: synth.Patch{
				Name: "octave synth",
				Oscs: []synth.Osc{
					{Wave: synth.Saw, Level: 0.6},
					{Wave: synth.Saw, Level: 0.45, Cents: 7},
					{Wave: synth.Pulse, Level: 0.4, Semi: 12, Width: 0.35},
				},
				Unison: synth.Unison{Voices: 4, Detune: 16, Spread: 0.7},
				Amp:    synth.Envelope{Attack: 0.05, Decay: 0.6, Sustain: 0.8, Release: 0.6},
				Filter: synth.Filter{
					Cutoff:    1500,
					Resonance: 0.12,
					KeyTrack:  0.5,
					Env:       synth.Envelope{Attack: 0.6, Decay: 1.5, Sustain: 0.4, Release: 1},
					EnvAmount: 1.2,
				},
				LFOs:     []synth.LFO{{Target: synth.PWM, Rate: 0.35, Depth: 0.12}, {Rate: 5.2, Depth: 6, Delay: 0.6}},
				Velocity: 0.4,
				Gain:     0.09,
			},
			Level: 2.9,
			Sends: []synth.Send{{Bus: "room", Level: -9.1}},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{At: 0.51, Dur: 0.765, Key: 72, Vel: 0.7},
						{At: 0.51, Dur: 0.765, Key: 76, Vel: 0.7},
						{At: 0.51, Dur: 0.765, Key: 79, Vel: 0.7},
						{At: 0.51, Dur: 0.765, Key: 84, Vel: 0.7},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "bells",
			Patch: synth.Patch{
				Oscs: []synth.Osc{
					{FM: synth.FM{Ratio: 2, Index: 1.8, Decay: 0.08, Sustain: 0.2}},
					{Level: 0.25, Semi: 12, FM: synth.FM{Ratio: 3.5, Index: 1, Decay: 0.05}},
				},
				Amp:      synth.Envelope{Attack: 0.001, Decay: 1.2, Release: 0.1},
				Velocity: 0.6,
				Gain:     0.3,
			},
			Level: -4.4,
			Sends: []synth.Send{{Bus: "room", Level: -9.1}},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{At: 0.51, Dur: 0.51, Key: 84, Vel: 0.8},
						{At: 0.53, Dur: 0.49, Key: 88, Vel: 0.8},
						{At: 0.55, Dur: 0.47, Key: 91, Vel: 0.8},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "sparkle",
			Patch: synth.Patch{
				Oscs: []synth.Osc{
					{FM: synth.FM{Ratio: 2, Index: 1.8, Decay: 0.08, Sustain: 0.2}},
					{Level: 0.25, Semi: 12, FM: synth.FM{Ratio: 3.5, Index: 1, Decay: 0.05}},
				},
				Amp:      synth.Envelope{Attack: 0.001, Decay: 0.4, Release: 0.1},
				Velocity: 0.6,
				Gain:     0.3,
			},
			Sends: []synth.Send{{Bus: "room", Level: -9.1}},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{At: 0.595, Dur: 0.0425, Key: 96, Vel: 0.45},
						{At: 0.6375, Dur: 0.0425, Key: 100, Vel: 0.45},
						{At: 0.68, Dur: 0.0425, Key: 103, Vel: 0.45},
						{At: 0.7225, Dur: 0.0425, Key: 108, Vel: 0.45},
						{At: 0.765, Dur: 0.0425, Key: 103, Vel: 0.45},
						{At: 0.8075, Dur: 0.0425, Key: 108, Vel: 0.45},
						{At: 0.85, Dur: 0.17, Key: 112, Vel: 0.45},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "sub",
			Patch: synth.Patch{
				Oscs:  []synth.Osc{{Fixed: 65}},
				Amp:   synth.Envelope{Attack: 0.001, Decay: 0.6, Release: 0.02},
				Sweep: synth.Sweep{Semis: 12, Time: 0.02},
				Drive: 0.8,
				Gain:  0.2,
			},
			Sends: []synth.Send{{Bus: "room", Level: -9.1}},
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{At: 0.51, Dur: 0.3, Key: 36}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Buses: []synth.Bus{
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.25, Damp: 0.6, PreDelay: 0.01}}},
		{
			Name:    "echo",
			Effects: []synth.Effect{synth.Delay{Time: 0.255, Feedback: 0.3, Damp: 4000, PingPong: true}},
		},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -3}},
	Priority: 3,
}

// A quest done: a short bell figure resolving up, over a soft chord.
var questSound = synth.Arrangement{
	ID:   "quest",
	Name: "quest",
	Tracks: []synth.Track{
		{
			Name: "bells",
			Patch: synth.Patch{
				Oscs: []synth.Osc{
					{FM: synth.FM{Ratio: 2, Index: 1.8, Decay: 0.08, Sustain: 0.2}},
					{Level: 0.25, Semi: 12, FM: synth.FM{Ratio: 3.5, Index: 1, Decay: 0.05}},
				},
				Amp:      synth.Envelope{Attack: 0.001, Decay: 0.8, Release: 0.1},
				Velocity: 0.6,
				Gain:     0.3,
			},
			Sends: []synth.Send{{Bus: "room", Level: -10.5}},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{Dur: 0.075, Key: 86, Vel: 0.8},
						{At: 0.075, Dur: 0.075, Key: 91, Vel: 0.8},
						{At: 0.15, Dur: 0.075, Key: 95, Vel: 0.8},
						{At: 0.3, Dur: 0.075, Key: 93, Vel: 0.8},
						{At: 0.375, Dur: 0.375, Key: 98, Vel: 0.8},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
		{
			Name: "under",
			Patch: synth.Patch{
				Name: "octave synth",
				Oscs: []synth.Osc{
					{Wave: synth.Saw, Level: 0.6},
					{Wave: synth.Saw, Level: 0.45, Cents: 7},
					{Wave: synth.Pulse, Level: 0.4, Semi: 12, Width: 0.35},
				},
				Unison: synth.Unison{Voices: 4, Detune: 16, Spread: 0.7},
				Amp:    synth.Envelope{Attack: 0.03, Decay: 0.6, Sustain: 0.8, Release: 0.5},
				Filter: synth.Filter{
					Cutoff:    1500,
					Resonance: 0.12,
					KeyTrack:  0.5,
					Env:       synth.Envelope{Attack: 0.6, Decay: 1.5, Sustain: 0.4, Release: 1},
					EnvAmount: 1.2,
				},
				LFOs:     []synth.LFO{{Target: synth.PWM, Rate: 0.35, Depth: 0.12}, {Rate: 5.2, Depth: 6, Delay: 0.6}},
				Velocity: 0.4,
				Gain:     0.09,
			},
			Level: 5.1,
			Sends: []synth.Send{{Bus: "room", Level: -10.5}},
			Clips: []synth.Clip{
				{
					Name: "a",
					Notes: []synth.Note{
						{At: 0.3, Dur: 0.9, Key: 67, Vel: 0.8},
						{At: 0.3, Dur: 0.9, Key: 71, Vel: 0.8},
						{At: 0.3, Dur: 0.9, Key: 74, Vel: 0.8},
					},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
	Buses: []synth.Bus{
		{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.25, Damp: 0.6, PreDelay: 0.01}}},
	},
	Master:   []synth.Effect{synth.Normalize{Peak: -3}},
	Priority: 3,
}
