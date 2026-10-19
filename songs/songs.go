// Package songs holds example content: pieces of music and a palette of effects.
package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

func Arrangements() []synth.Arrangement {
	return []synth.Arrangement{goldenBrown, amberHours, playaSunrise, lavaLamp, brighterDays, smokeSignal, woodlands, interstellar}
}

func Songs() []synth.Song {
	var songs []synth.Song
	for _, a := range Arrangements() {
		songs = append(songs, a.Song())
	}
	return songs
}

func Sounds() []synth.Song {
	coin := synth.Phrase{Step: 0.07, Gate: 0.9}.Notes("B5 E6 -")
	return append([]synth.Song{
		synth.Hit("kick", patches.Kick(), synth.MustKey("C2"), 0.3),
		synth.Hit("snare", patches.Snare(), synth.MustKey("C3"), 0.15),
		synth.Hit("hihat", patches.HiHat(), synth.MustKey("C5"), 0.04),
		synth.Hit("laser", patches.Laser(), synth.MustKey("C5"), 0.2),
		synth.Hit("bell", patches.Bell(), synth.MustKey("E5"), 1),
		synth.Hit("epiano", patches.EPiano(), synth.MustKey("A3"), 1),
		synth.Hit("bass", patches.Bass(), synth.MustKey("E1"), 0.5),
		synth.Hit("harpsichord", patches.Harpsichord(), synth.MustKey("Bb3"), 1),
		{ID: "coin", Name: "coin", Parts: []synth.Part{{Name: "coin", Patch: patches.Coin(), Notes: coin}}},
	}, Platformer()...)
}

func Lookup(id string) (synth.Song, bool) {
	for _, s := range append(Songs(), Sounds()...) {
		if s.ID == id {
			return s, true
		}
	}
	return synth.Song{}, false
}
