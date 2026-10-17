// Package songs holds example content: pieces of music and a palette of effects.
package songs

import (
	"slices"
	"strings"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

func Songs() []synth.Song {
	return []synth.Song{GoldenBrown(), AmberHours(), PlayaSunrise(), LavaLamp(), BrighterDays(), SmokeSignal(), Woodlands(), Interstellar()}
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

// form is a song's sections in order, a byte each, every pass the same number of steps.
type form struct {
	sections string
	steps    int
}

// score writes a line through the form, one pass at a time; play returns "" to rest.
func (f form) score(play func(pass int, section byte) string) string {
	var b strings.Builder
	for pass := range len(f.sections) {
		bar := play(pass, f.sections[pass])
		if bar == "" {
			bar = silence(f.steps)
		}
		b.WriteString(bar + " ")
	}
	return b.String()
}

// plays writes a section line on the passes in plays and rests through the others.
func (f form) plays(line map[byte]string, plays func(pass int) bool) string {
	return f.score(func(pass int, s byte) string {
		if !plays(pass) {
			return ""
		}
		return line[s]
	})
}

func from(first int) func(int) bool { return func(pass int) bool { return pass >= first } }

func only(passes ...int) func(int) bool {
	return func(pass int) bool { return slices.Contains(passes, pass) }
}
