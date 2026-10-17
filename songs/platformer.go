package songs

import (
	"slices"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// Platformer is a side-scrolling MMO's effects: movement, combat, pickups, UI, and the
// little fanfares. Each is a Song of layered parts, the way a sound designer stacks a
// transient, a body, and a tail. Occasional sounds share one small room; anything a game
// fires many times a second stays dry, since even a faint tail keeps a voice alive and
// the player only holds so many. Noise layers play at C4, where a filter's key tracking
// leaves its cutoff alone.
//
// Priority follows the tiers, so a flood of footsteps or coins gives way before a blow,
// and a blow before a fanfare. Whatever a game fires over and over varies each play and
// is capped, so it can't crowd out the rest.
func Platformer() []synth.Song {
	return slices.Concat(
		pfTier(0,
			pfRepeat(pfJump(), synth.Variation{Pitch: 0.5, Gain: 1, Takes: 2}, 2),
			pfDoubleJump(),
			pfRepeat(pfLand(), synth.Variation{Pitch: 1, Gain: 1.5, Takes: 3}, 2),
			pfRepeat(pfClimb(), synth.Variation{Pitch: 1.5, Gain: 2, Takes: 3}, 2),
		),
		pfTier(1,
			pfRepeat(pfPickup(), synth.Variation{Pitch: 0.25, Gain: 1}, 3),
			pfRepeat(pfMeso(), synth.Variation{Pitch: 0.3, Gain: 1}, 4),
			pfDrop(), pfPotion(), pfBuff(), pfPortal(),
		),
		pfTier(2,
			pfRepeat(pfSwing(), synth.Variation{Pitch: 1, Gain: 1, Takes: 3}, 3),
			pfRepeat(pfHit(), synth.Variation{Pitch: 1.5, Gain: 1.5, Takes: 4}, 4),
			pfCrit(), pfMagic(),
			pfRepeat(pfHurt(), synth.Variation{Pitch: 1, Gain: 1, Takes: 2}, 2),
			pfRepeat(pfMonsterDie(), synth.Variation{Pitch: 1, Gain: 1}, 4),
			pfRepeat(pfClick(), synth.Variation{}, 2), pfOpen(), pfClose(), pfPop(), pfError(),
		),
		pfTier(3, pfDeath(), pfLevelUp(), pfQuest()),
	)
}

func pfTier(priority int, songs ...synth.Song) []synth.Song {
	for i := range songs {
		songs[i].Priority = priority
	}
	return songs
}

func pfRepeat(s synth.Song, v synth.Variation, voices int) synth.Song {
	s.Vary, s.Voices = v, voices
	return s
}

// pfSound mixes layers through a shared room and a limiter, so every effect lands at the
// same loudness however its layers stack.
func pfSound(id string, room float64, parts ...synth.Part) synth.Song {
	s := synth.Song{
		ID:     id,
		Name:   id,
		Parts:  parts,
		Master: []synth.Effect{synth.Limiter{Ceiling: -1}},
	}
	if room > 0 {
		s.Buses = []synth.Bus{{Name: "room", Effects: []synth.Effect{
			synth.Reverb{Size: 0.25, Damp: 0.6, PreDelay: 0.01},
		}}}
		for i := range s.Parts {
			s.Parts[i].Sends = append(s.Parts[i].Sends, synth.Send{Bus: "room", Level: room})
		}
	}
	return s
}

func pfLayer(name string, p synth.Patch, notes ...synth.Note) synth.Part {
	return synth.Part{Name: name, Patch: p, Notes: notes}
}

func pfNote(at float64, key string, dur float64) synth.Note {
	return synth.Note{At: at, Key: synth.MustKey(key), Dur: dur}
}

func pfNoise(mode synth.FilterMode, cutoff, res, decay float64) synth.Patch {
	return synth.Patch{
		Oscs:   []synth.Osc{{Wave: synth.Noise}},
		Amp:    synth.Envelope{Attack: 0.001, Decay: decay, Release: 0.01},
		Filter: synth.Filter{Mode: mode, Cutoff: cutoff, Resonance: res},
	}
}

// pfThump is a body: a sine dropping onto a fixed pitch.
func pfThump(hz, semis, decay float64) synth.Patch {
	return synth.Patch{
		Oscs:  []synth.Osc{{Wave: synth.Sine, Fixed: hz}},
		Sweep: synth.Sweep{Semis: semis, Time: 0.02},
		Amp:   synth.Envelope{Attack: 0.001, Decay: decay, Release: 0.02},
		Drive: 0.8,
	}
}

// pfChime is a soft FM bell for pickups and fanfares: bright on the strike, round after.
func pfChime(decay float64) synth.Patch {
	return synth.Patch{
		Oscs: []synth.Osc{
			{Wave: synth.Sine, FM: synth.FM{Ratio: 2, Index: 1.8, Decay: 0.08, Sustain: 0.2}},
			{Wave: synth.Sine, Semi: 12, Level: 0.25, FM: synth.FM{Ratio: 3.5, Index: 1, Decay: 0.05}},
		},
		Amp:      synth.Envelope{Attack: 0.001, Decay: decay, Release: 0.1},
		Velocity: 0.6,
		Gain:     0.3,
	}
}

// pfSquare is the chiptune voice for fanfares, a little softer than patches.Lead.
func pfSquare() synth.Patch {
	p := patches.Lead()
	p.Oscs = append(p.Oscs, synth.Osc{Wave: synth.Triangle, Semi: -12, Level: 0.5})
	p.Filter = synth.Filter{Cutoff: 3500, KeyTrack: 0.3}
	return p
}

// A springy blip that rises into its note, over a puff of air from the feet.
func pfJump() synth.Song {
	blip := synth.Patch{
		Oscs:   []synth.Osc{{Wave: synth.Pulse, Width: 0.3}, {Wave: synth.Triangle, Semi: -12, Level: 0.6}},
		Sweep:  synth.Sweep{Semis: -9, Time: 0.03},
		Amp:    synth.Envelope{Attack: 0.002, Decay: 0.13, Release: 0.03},
		Filter: synth.Filter{Cutoff: 3000},
		Gain:   0.3,
	}
	puff := pfNoise(synth.BandPass, 1400, 0.3, 0.05)
	puff.Gain = 0.25
	return pfSound("jump", 0,
		pfLayer("blip", blip, synth.Note{Key: synth.MustKey("A4"), Dur: 0.08, Slide: 4}),
		pfLayer("puff", puff, pfNote(0, "C4", 0.03)),
	)
}

// The second jump in the air: a whoosh that swells and passes, under a rising whistle.
func pfDoubleJump() synth.Song {
	whoosh := pfNoise(synth.BandPass, 500, 0.55, 0.28)
	whoosh.Amp.Attack = 0.04
	whoosh.Filter.Env = synth.Envelope{Attack: 0.08, Decay: 0.2}
	whoosh.Filter.EnvAmount = 2.5
	whoosh.LFOs = []synth.LFO{{Target: synth.AutoPan, Shape: synth.SawShape, Rate: 2.5, Depth: 0.5}}
	whoosh.Gain = 0.9
	whistle := synth.Patch{
		Oscs: []synth.Osc{{Wave: synth.Sine}, {Wave: synth.Triangle, Semi: 12, Level: 0.3}},
		Amp:  synth.Envelope{Attack: 0.01, Decay: 0.22, Release: 0.05},
		Gain: 0.2,
	}
	return pfSound("doublejump", 0.15,
		pfLayer("whoosh", whoosh, pfNote(0, "C4", 0.2)),
		pfLayer("whistle", whistle, synth.Note{Key: synth.MustKey("E5"), Dur: 0.15, Slide: 12, SlideTime: 0.15}),
	)
}

// Feet on the ground: a dull thud and a scuff.
func pfLand() synth.Song {
	thud := pfThump(80, 10, 0.09)
	thud.Gain = 0.3
	scuff := pfNoise(synth.LowPass, 900, 0.1, 0.05)
	scuff.Gain = 0.2
	return pfSound("land", 0,
		pfLayer("thud", thud, pfNote(0, "C2", 0.05)),
		pfLayer("scuff", scuff, pfNote(0.004, "C4", 0.03)),
	)
}

// One hand on a rope or rung: a wooden tock. A game plays it every few frames of climbing.
func pfClimb() synth.Song {
	tock := synth.Patch{
		Oscs:  []synth.Osc{{Wave: synth.Triangle, Fixed: 520}, {Wave: synth.Sine, Fixed: 1240, Level: 0.3}},
		Sweep: synth.Sweep{Semis: 5, Time: 0.005},
		Amp:   synth.Envelope{Attack: 0.0005, Decay: 0.035, Release: 0.01},
		Gain:  0.3,
	}
	grain := pfNoise(synth.BandPass, 2500, 0.4, 0.015)
	grain.Gain = 0.25
	return pfSound("climb", 0,
		pfLayer("tock", tock, pfNote(0, "C4", 0.02)),
		pfLayer("grain", grain, pfNote(0, "C4", 0.01)),
	)
}

// A weapon cutting the air, carried across the field as it goes.
func pfSwing() synth.Song {
	air := pfNoise(synth.BandPass, 450, 0.6, 0.14)
	air.Amp.Attack = 0.03
	air.Filter.Env = synth.Envelope{Attack: 0.05, Decay: 0.1}
	air.Filter.EnvAmount = 3
	air.LFOs = []synth.LFO{{Target: synth.AutoPan, Shape: synth.SawShape, Rate: 4, Depth: 0.6}}
	edge := pfNoise(synth.HighPass, 5000, 0.3, 0.08)
	edge.Amp.Attack = 0.04
	edge.Gain = 0.15
	return pfSound("swing", 0,
		pfLayer("air", air, pfNote(0, "C4", 0.1)),
		pfLayer("edge", edge, pfNote(0, "C4", 0.05)),
	)
}

func pfStrike() []synth.Part {
	crack := pfNoise(synth.HighPass, 1800, 0.2, 0.045)
	crack.Gain = 0.5
	slap := synth.Patch{
		Oscs:  []synth.Osc{{Wave: synth.Triangle, Fixed: 340}},
		Sweep: synth.Sweep{Semis: 14, Time: 0.012},
		Amp:   synth.Envelope{Attack: 0.0005, Decay: 0.05, Release: 0.01},
		Drive: 2,
		Gain:  0.3,
	}
	body := pfThump(110, 24, 0.12)
	body.Drive = 2
	body.Gain = 0.38
	return []synth.Part{
		pfLayer("crack", crack, pfNote(0, "C4", 0.02)),
		pfLayer("slap", slap, pfNote(0, "C4", 0.02)),
		pfLayer("body", body, pfNote(0, "C2", 0.05)),
	}
}

// Landing a blow on a monster: crack, slap, and weight.
func pfHit() synth.Song {
	return pfSound("hit", 0, pfStrike()...)
}

// A critical: the same blow with a metallic ring and a sparkle on top of it.
func pfCrit() synth.Song {
	ring := synth.Patch{
		Oscs:     []synth.Osc{{Wave: synth.Sine, FM: synth.FM{Ratio: 3.5, Index: 3, Decay: 0.15, Sustain: 0.1}}},
		Amp:      synth.Envelope{Attack: 0.001, Decay: 0.5, Release: 0.1},
		Velocity: 0.5,
		Gain:     0.25,
	}
	sparkle := synth.Phrase{At: 0.03, Step: 0.035, Gate: 0.6, Vel: 0.6}.Notes("B6 E7 B7")
	return pfSound("crit", 0.2, append(pfStrike(),
		pfLayer("ring", ring, pfNote(0, "E6", 0.05)),
		pfLayer("sparkle", pfChime(0.25), sparkle...),
	)...)
}

// A bolt of magic: an FM zap falling from on high, with a crackle and an echo behind it.
func pfMagic() synth.Song {
	zap := synth.Patch{
		Oscs: []synth.Osc{
			{Wave: synth.Sine, FM: synth.FM{Ratio: 1.41, Index: 6, Decay: 0.12, Sustain: 0.2}},
			{Wave: synth.Saw, Semi: 7, Level: 0.25},
		},
		Sweep:  synth.Sweep{Semis: 24, Time: 0.06},
		Amp:    synth.Envelope{Attack: 0.002, Decay: 0.3, Release: 0.05},
		Filter: synth.Filter{Cutoff: 2500, Resonance: 0.4, Env: synth.Envelope{Decay: 0.15}, EnvAmount: 2},
		LFOs:   []synth.LFO{{Target: synth.Vibrato, Shape: synth.RandomShape, Rate: 40, Depth: 40}},
		Gain:   0.6,
	}
	crackle := pfNoise(synth.HighPass, 3000, 0.2, 0.2)
	crackle.LFOs = []synth.LFO{{Target: synth.Tremolo, Shape: synth.RandomShape, Rate: 60, Depth: 0.9}}
	crackle.Gain = 0.4
	s := pfSound("magic", 0.2,
		pfLayer("zap", zap, pfNote(0, "A4", 0.15)),
		pfLayer("crackle", crackle, pfNote(0.01, "C4", 0.15)),
	)
	s.Parts[0].Inserts = []synth.Effect{synth.Delay{Time: 0.11, Feedback: 0.35, Damp: 3000, PingPong: true, Mix: 0.3}}
	return s
}

// Taking damage: a crunchy, stuttering drop in pitch.
func pfHurt() synth.Song {
	ouch := synth.Patch{
		Oscs:   []synth.Osc{{Wave: synth.Pulse, Width: 0.5}, {Wave: synth.Saw, Cents: 15, Level: 0.5}},
		Amp:    synth.Envelope{Attack: 0.002, Decay: 0.25, Release: 0.04},
		Filter: synth.Filter{Cutoff: 2500},
		LFOs:   []synth.LFO{{Target: synth.Tremolo, Shape: synth.SquareShape, Rate: 28, Depth: 0.6}},
		Gain:   0.25,
	}
	thud := pfThump(90, 12, 0.1)
	thud.Gain = 0.4
	s := pfSound("hurt", 0,
		pfLayer("ouch", ouch, synth.Note{Key: synth.MustKey("D5"), Dur: 0.2, Slide: -9}),
		pfLayer("thud", thud, pfNote(0, "C2", 0.05)),
	)
	s.Parts[0].Inserts = []synth.Effect{synth.Crush{Bits: 6, Hold: 3}}
	return s
}

// A monster's last: a falling squeal that dulls as it goes, and a puff of smoke.
func pfMonsterDie() synth.Song {
	squeal := synth.Patch{
		Oscs: []synth.Osc{{Wave: synth.Saw}, {Wave: synth.Pulse, Semi: -5, Level: 0.5, Width: 0.3}},
		Amp:  synth.Envelope{Attack: 0.003, Decay: 0.4, Release: 0.05},
		Filter: synth.Filter{
			Cutoff: 500, Resonance: 0.3,
			Env:       synth.Envelope{Decay: 0.35},
			EnvAmount: 2.5,
		},
		LFOs: []synth.LFO{{Target: synth.Vibrato, Rate: 14, Depth: 60}},
		Gain: 0.38,
	}
	poof := pfNoise(synth.LowPass, 1500, 0.15, 0.35)
	poof.Amp.Attack = 0.02
	poof.Filter.Env = synth.Envelope{Decay: 0.3}
	poof.Filter.EnvAmount = 1
	poof.Gain = 0.6
	return pfSound("monsterdie", 0.25,
		pfLayer("squeal", squeal, synth.Note{Key: synth.MustKey("G4"), Dur: 0.3, Slide: -17}),
		pfLayer("poof", poof, pfNote(0.08, "C4", 0.25)),
	)
}

// The player falls: a slow minor descent, and a tombstone landing hard after it.
func pfDeath() synth.Song {
	fall := synth.Phrase{Step: 0.16, Gate: 0.9, Vel: 0.8}.Notes("G4 Eb4 C4 B3 - - -")
	tomb := pfThump(55, 18, 0.4)
	tomb.Drive = 1.5
	tomb.Gain = 0.3
	dust := pfNoise(synth.LowPass, 700, 0.1, 0.3)
	dust.Gain = 1.5
	const landing = 1.05
	return pfSound("death", 0.3,
		pfLayer("fall", pfSquare(), fall...),
		pfLayer("tomb", tomb, pfNote(landing, "C2", 0.1)),
		pfLayer("dust", dust, pfNote(landing+0.005, "C4", 0.2)),
	)
}

// Picking an item up: a quick, soft bloop upward.
func pfPickup() synth.Song {
	bloop := synth.Patch{
		Oscs:  []synth.Osc{{Wave: synth.Sine, FM: synth.FM{Ratio: 2, Index: 1, Decay: 0.04}}, {Wave: synth.Triangle, Level: 0.4}},
		Sweep: synth.Sweep{Semis: -5, Time: 0.015},
		Amp:   synth.Envelope{Attack: 0.002, Decay: 0.12, Release: 0.03},
		Gain:  0.35,
	}
	return pfSound("pickup", 0,
		pfLayer("bloop", bloop, synth.Phrase{Step: 0.055, Gate: 0.8}.Notes("E5 B5")...),
	)
}

// Coins into the purse: two bright pings and the jingle of a few more behind them.
func pfMeso() synth.Song {
	ping := synth.Patch{
		Oscs:     []synth.Osc{{Wave: synth.Sine, FM: synth.FM{Ratio: 2.76, Index: 2.2, Decay: 0.05, Sustain: 0.25}}},
		Amp:      synth.Envelope{Attack: 0.0008, Decay: 0.3, Release: 0.05},
		Velocity: 0.7,
		Gain:     0.3,
	}
	jingle := synth.Phrase{At: 0.09, Step: 0.028, Vel: 0.35, Humanize: 0.01, Seed: 7}.Notes("F#7 D7 A7 E7")
	return pfSound("meso", 0,
		pfLayer("ping", ping, synth.Phrase{Step: 0.06}.Notes("B6 E7 -")...),
		pfLayer("jingle", ping, jingle...),
	)
}

// An item dropping from a monster and bouncing to rest, each bounce lower and sooner.
func pfDrop() synth.Song {
	plink := synth.Patch{
		Oscs:     []synth.Osc{{Wave: synth.Sine, FM: synth.FM{Ratio: 3, Index: 1.2, Decay: 0.03}}},
		Sweep:    synth.Sweep{Semis: 3, Time: 0.01},
		Amp:      synth.Envelope{Attack: 0.001, Decay: 0.09, Release: 0.02},
		Velocity: 1,
		Gain:     0.35,
	}
	var bounces []synth.Note
	at, gap, vel := 0.0, 0.16, 1.0
	for range 4 {
		bounces = append(bounces, synth.Note{At: at, Key: synth.MustKey("G5"), Dur: 0.03, Vel: vel})
		at += gap
		gap *= 0.6
		vel *= 0.6
	}
	return pfSound("drop", 0, pfLayer("plink", plink, bounces...))
}

// Drinking a potion: three rising glugs, then a warm sparkle as it heals.
func pfPotion() synth.Song {
	glug := synth.Patch{
		Oscs:   []synth.Osc{{Wave: synth.Sine}, {Wave: synth.Triangle, Level: 0.3}},
		Sweep:  synth.Sweep{Semis: -8, Time: 0.025},
		Amp:    synth.Envelope{Attack: 0.003, Decay: 0.07, Release: 0.02},
		Filter: synth.Filter{Cutoff: 1800},
		Gain:   0.45,
	}
	glugs := synth.Phrase{Step: 0.09, Gate: 0.5}.Notes("C4 D4 F4")
	heal := synth.Phrase{At: 0.3, Step: 0.05, Vel: 0.6}.Notes("C6 E6 G6 C7 - -")
	return pfSound("potion", 0.25,
		pfLayer("glug", glug, glugs...),
		pfLayer("heal", pfChime(0.4), heal...),
	)
}

// Casting a buff: a detuned stack rising and opening as it swirls, then a shimmer.
func pfBuff() synth.Song {
	rise := synth.Patch{
		Oscs:   []synth.Osc{{Wave: synth.Saw}, {Wave: synth.Saw, Semi: 7, Level: 0.6}},
		Unison: synth.Unison{Voices: 5, Detune: 30, Spread: 0.9},
		Amp:    synth.Envelope{Attack: 0.05, Decay: 0.6, Release: 0.15},
		Filter: synth.Filter{
			Cutoff: 300, Resonance: 0.35,
			Env:       synth.Envelope{Attack: 0.45, Decay: 0.3},
			EnvAmount: 4,
		},
		Gain: 0.12,
	}
	shimmer := synth.Phrase{At: 0.4, Step: 0.03, Vel: 0.5}.Notes("G6 C7 E7 G7 C8 - -")
	s := pfSound("buff", 0.35,
		pfLayer("rise", rise, synth.Note{Key: synth.MustKey("C4"), Dur: 0.5, Slide: 12}),
		pfLayer("shimmer", pfChime(0.35), shimmer...),
	)
	s.Parts[0].Inserts = []synth.Effect{synth.Chorus{Mix: 0.5}}
	return s
}

// Stepping through a portal: a warbling rise through a phaser, swallowed by the room.
func pfPortal() synth.Song {
	warp := synth.Patch{
		Oscs:   []synth.Osc{{Wave: synth.Saw}, {Wave: synth.Sine, Semi: -12, Level: 0.6}},
		Unison: synth.Unison{Voices: 3, Detune: 20, Spread: 0.8},
		Amp:    synth.Envelope{Attack: 0.08, Decay: 0.6, Release: 0.2},
		Filter: synth.Filter{
			Cutoff: 400, Resonance: 0.5,
			Env:       synth.Envelope{Attack: 0.3, Decay: 0.4},
			EnvAmount: 3,
		},
		LFOs: []synth.LFO{{Target: synth.Vibrato, Rate: 11, Depth: 50}},
		Gain: 0.3,
	}
	swirl := pfNoise(synth.BandPass, 1200, 0.7, 0.5)
	swirl.Amp.Attack = 0.15
	swirl.LFOs = []synth.LFO{
		{Target: synth.Wah, Rate: 6, Depth: 1},
		{Target: synth.AutoPan, Rate: 3, Depth: 0.6},
	}
	swirl.Gain = 0.55
	s := pfSound("portal", 0.4,
		pfLayer("warp", warp, synth.Note{Key: synth.MustKey("C3"), Dur: 0.5, Slide: 24}),
		pfLayer("swirl", swirl, pfNote(0, "C4", 0.4)),
	)
	s.Parts[0].Inserts = []synth.Effect{synth.Phaser{Rate: 2, Depth: 2, Feedback: 0.6}}
	return s
}

// Level up: a square-wave arpeggio onto a held chord, bells ringing it, a sparkle run
// climbing over the top.
func pfLevelUp() synth.Song {
	const step = 0.085
	lead := synth.Phrase{Step: step, Gate: 0.9}.Notes("G4 C5 E5 G5 . E5 C6! - - - - - - -")
	chord := synth.Phrase{At: 6 * step, Step: step, Vel: 0.7}.Notes("C5,E5,G5,C6 - - - - - - - -")
	bells := synth.Phrase{At: 6 * step, Step: step, Strum: 0.02, Vel: 0.8}.Notes("C6,E6,G6 - - - - -")
	sparkle := synth.Phrase{At: 7 * step, Step: step / 2, Vel: 0.45}.Notes("C7 E7 G7 C8 G7 C8 E8 - - -")
	sub := pfThump(65, 12, 0.6)
	sub.Gain = 0.2

	pad := patches.OctaveSynth()
	pad.Amp.Attack = 0.05
	pad.Amp.Release = 0.6
	s := pfSound("levelup", 0.35,
		pfLayer("lead", pfSquare(), lead...),
		pfLayer("chord", pad, chord...),
		synth.Part{Name: "bells", Patch: pfChime(1.2), Notes: bells, Gain: 0.6},
		pfLayer("sparkle", pfChime(0.4), sparkle...),
		pfLayer("sub", sub, pfNote(6*step, "C2", 0.3)),
	)
	s.Parts[1].Gain = 1.4
	s.Parts[0].Sends = append(s.Parts[0].Sends, synth.Send{Bus: "echo", Level: 0.25})
	s.Buses = append(s.Buses, synth.Bus{Name: "echo", Effects: []synth.Effect{
		synth.Delay{Time: 3 * step, Feedback: 0.3, Damp: 4000, PingPong: true},
	}})
	return s
}

// A quest done: a short bell figure resolving up, over a soft chord.
func pfQuest() synth.Song {
	bells := synth.Phrase{Step: 0.075, Vel: 0.8}.Notes("D6 G6 B6 . A6 D7 - - - -")
	under := synth.Phrase{At: 0.3, Step: 0.3}.Notes("G4,B4,D5 - -")
	pad := patches.OctaveSynth()
	pad.Amp.Attack = 0.03
	pad.Amp.Release = 0.5
	s := pfSound("quest", 0.3,
		pfLayer("bells", pfChime(0.8), bells...),
		pfLayer("under", pad, under...),
	)
	s.Parts[1].Gain = 1.8
	return s
}

// A UI tick, quiet enough to hear on every button.
func pfClick() synth.Song {
	tick := synth.Patch{
		Oscs:  []synth.Osc{{Wave: synth.Sine, Fixed: 2200}, {Wave: synth.Noise, Level: 0.3}},
		Amp:   synth.Envelope{Attack: 0.0005, Decay: 0.018, Release: 0.005},
		Sweep: synth.Sweep{Semis: 4, Time: 0.004},
		Gain:  0.4,
	}
	return pfSound("click", 0, pfLayer("tick", tick, pfNote(0, "C4", 0.01)))
}

func pfTwoTone(id, notes string) synth.Song {
	tone := synth.Patch{
		Oscs: []synth.Osc{{Wave: synth.Sine, FM: synth.FM{Ratio: 1, Index: 0.8, Decay: 0.03}}},
		Amp:  synth.Envelope{Attack: 0.001, Decay: 0.08, Release: 0.02},
		Gain: 0.3,
	}
	return pfSound(id, 0, pfLayer("tone", tone, synth.Phrase{Step: 0.045, Gate: 0.8}.Notes(notes)...))
}

// A window opening, and closing: the same two tones, up and then down.
func pfOpen() synth.Song  { return pfTwoTone("open", "G5 D6") }
func pfClose() synth.Song { return pfTwoTone("close", "D6 G5") }

// A chat bubble or an NPC's dialog appearing.
func pfPop() synth.Song {
	pop := synth.Patch{
		Oscs:  []synth.Osc{{Wave: synth.Sine}},
		Sweep: synth.Sweep{Semis: -14, Time: 0.012},
		Amp:   synth.Envelope{Attack: 0.001, Decay: 0.06, Release: 0.01},
		Gain:  0.45,
	}
	return pfSound("pop", 0, pfLayer("pop", pop, pfNote(0, "A5", 0.03)))
}

// Not allowed: inventory full, not enough mana. Two low buzzes, the second a step down.
func pfError() synth.Song {
	buzz := synth.Patch{
		Oscs:   []synth.Osc{{Wave: synth.Pulse, Width: 0.4}, {Wave: synth.Saw, Cents: -12, Level: 0.6}},
		Amp:    synth.Envelope{Attack: 0.003, Decay: 0.1, Sustain: 0.7, Release: 0.03},
		Filter: synth.Filter{Cutoff: 1400, Resonance: 0.2},
		Gain:   0.2,
	}
	return pfSound("error", 0, pfLayer("buzz", buzz, synth.Phrase{Step: 0.1, Gate: 0.75}.Notes("E3 C3")...))
}
