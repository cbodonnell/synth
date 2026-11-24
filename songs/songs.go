// Package songs holds example content: pieces of music and a palette of effects.
package songs

import (
	"slices"

	"github.com/cbodonnell/synth"
)

// Arrangements are the example songs, then the sounds.
func Arrangements() []synth.Arrangement {
	return slices.Concat([]synth.Arrangement{
		goldenBrown, amberHours, playaSunrise, lavaLamp, brighterDays, smokeSignal, woodlands,
		interstellar, longLight, daybreak, goldenHour, elegy,
	}, Hits(), Platformer())
}

func Songs() []synth.Song { return compiled(func(a synth.Arrangement) bool { return a.Tempo > 0 }) }

func Sounds() []synth.Song { return compiled(func(a synth.Arrangement) bool { return a.Tempo == 0 }) }

func compiled(keep func(synth.Arrangement) bool) []synth.Song {
	var songs []synth.Song
	for _, a := range Arrangements() {
		if keep(a) {
			songs = append(songs, a.Song())
		}
	}
	return songs
}

func Lookup(id string) (synth.Song, bool) {
	for _, s := range append(Songs(), Sounds()...) {
		if s.ID == id {
			return s, true
		}
	}
	return synth.Song{}, false
}
