package session

import (
	"encoding/json"
	"slices"

	"github.com/cbodonnell/synth"
)

// Library is a game's music and sound: its arrangements, any songs written as code
// instead, and its sound effects.
type Library struct {
	Arrangements []synth.Arrangement
	Songs        []synth.Song
	Sounds       []synth.Song
}

// Music is every song, compiled.
func (l Library) Music() []synth.Song {
	var songs []synth.Song
	for _, a := range l.Arrangements {
		songs = append(songs, a.Song())
	}
	return append(songs, l.Songs...)
}

// All is every song and sound, compiled.
func (l Library) All() []synth.Song { return append(l.Music(), l.Sounds...) }

// Arrangement is the arrangement with the ID, if the song is one.
func (l Library) Arrangement(id string) (synth.Arrangement, bool) {
	for _, a := range l.Arrangements {
		if a.ID == id {
			return a, true
		}
	}
	return synth.Arrangement{}, false
}

// Open lays out the song or sound with the ID.
func (l Library) Open(id string) (Layout, bool) {
	for _, a := range l.Arrangements {
		if a.ID == id {
			return Arranged(a), true
		}
	}
	for _, s := range slices.Concat(l.Songs, l.Sounds) {
		if s.ID == id {
			return Unarranged(s), true
		}
	}
	return Layout{}, false
}

// View is where the studio is looking, carried across a reload.
type View struct {
	ID      string
	At      float64 // the playhead, in seconds
	Playing bool
	Loop    bool
	Click   bool
	Solo    []string
	Track   string // the selected track
	Clip    string // the clip open in the editor, on the selected track

	Zoom    float64 // pixels a second; 0 fits the song
	Scroll  float64 // seconds at the timeline's left edge
	Top     float64 // pixels the lanes are scrolled down
	Browser bool
	Mixer   bool     // the mixer in place of the arrangement
	Folded  []string // group buses folded shut
}

func (v View) String() string {
	b, _ := json.Marshal(v)
	return string(b)
}

// ParseView reads a View as String writes it.
func ParseView(s string) (View, error) {
	var v View
	return v, json.Unmarshal([]byte(s), &v)
}
