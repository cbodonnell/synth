package session

import (
	"slices"
	"strconv"
	"strings"

	"github.com/cbodonnell/synth"
)

// Scale is a song's Key: which pitch classes are in it, and how its notes are spelled.
type Scale struct {
	Tonic  int   // pitch class, 0 is C
	Steps  []int // semitones above the tonic of each degree
	letter int   // the tonic's letter, 0 is C
	sharps bool  // spells notes outside it with sharps
}

var modes = map[string][]int{
	"major":          {0, 2, 4, 5, 7, 9, 11},
	"ionian":         {0, 2, 4, 5, 7, 9, 11},
	"dorian":         {0, 2, 3, 5, 7, 9, 10},
	"phrygian":       {0, 1, 3, 5, 7, 8, 10},
	"lydian":         {0, 2, 4, 6, 7, 9, 11},
	"mixolydian":     {0, 2, 4, 5, 7, 9, 10},
	"minor":          {0, 2, 3, 5, 7, 8, 10},
	"aeolian":        {0, 2, 3, 5, 7, 8, 10},
	"locrian":        {0, 1, 3, 5, 6, 8, 10},
	"harmonic minor": {0, 2, 3, 5, 7, 8, 11},
	"melodic minor":  {0, 2, 3, 5, 7, 9, 11},
}

const letters = "CDEFGAB"

var naturals = [7]int{0, 2, 4, 5, 7, 9, 11}

// ParseScale reads a Key as songs write it: "D mixolydian", "Bb minor", "G".
func ParseScale(key string) (Scale, bool) {
	tonic, mode, _ := strings.Cut(strings.TrimSpace(key), " ")
	if tonic == "" {
		return Scale{}, false
	}
	mode = strings.ToLower(strings.Join(strings.Fields(mode), " "))
	if mode == "" {
		mode = "major"
	}
	steps, ok := modes[mode]
	l := strings.IndexByte(letters, tonic[0])
	if !ok || l < 0 {
		return Scale{}, false
	}
	pc := naturals[l]
	for _, r := range tonic[1:] {
		switch r {
		case '#':
			pc++
		case 'b':
			pc--
		default:
			return Scale{}, false
		}
	}
	s := Scale{Tonic: (pc + 12) % 12, Steps: steps, letter: l, sharps: true}
	for d := range steps {
		if n := s.degreeName(d); strings.Contains(n, "b") {
			s.sharps = false
		} else if strings.Contains(n, "#") {
			s.sharps = true
			break
		}
	}
	return s, true
}

// degreeName is the scale's degree d spelled without its octave: its own letter, sharp
// or flat as the scale needs.
func (s Scale) degreeName(d int) string {
	l := (s.letter + d) % 7
	acc := ((s.Tonic+s.Steps[d]-naturals[l])%12+18)%12 - 6
	switch {
	case acc > 0:
		return letters[l:l+1] + strings.Repeat("#", acc)
	case acc < 0:
		return letters[l:l+1] + strings.Repeat("b", -acc)
	}
	return letters[l : l+1]
}

// degree is the scale degree of key's pitch class, or -1.
func (s Scale) degree(key int) int {
	return slices.Index(s.Steps, ((key-s.Tonic)%12+12)%12)
}

// Has reports whether key is in the scale.
func (s Scale) Has(key int) bool { return s.degree(key) >= 0 }

var sharpNames = [12]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

// Spell names key as the notation writes it, spelled as the scale spells it.
func (s Scale) Spell(key int) string {
	if d := s.degree(key); d >= 0 {
		name := s.degreeName(d)
		l := strings.IndexByte(letters, name[0])
		natural := key - (strings.Count(name, "#") - strings.Count(name, "b"))
		return name + strconv.Itoa(floorDiv(natural-naturals[l], 12)-1)
	}
	if s.sharps {
		return sharpNames[((key%12)+12)%12] + strconv.Itoa(floorDiv(key, 12)-1)
	}
	return synth.KeyName(key)
}

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// Chord stacks n of the scale's thirds on the degree at or below key: 3 is its triad,
// 4 its seventh. A song without a Key stacks C major's.
func (s Scale) Chord(key, n int) []int {
	if len(s.Steps) == 0 {
		s = Scale{Steps: modes["major"]}
	}
	root := key
	for s.degree(root) < 0 {
		root--
	}
	d := s.degree(root)
	base := root - s.Steps[d]
	var keys []int
	for i := range n {
		k := d + 2*i
		keys = append(keys, base+12*(k/len(s.Steps))+s.Steps[k%len(s.Steps)])
	}
	return keys
}
