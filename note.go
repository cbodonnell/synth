package synth

import (
	"fmt"
	"math"
	"strconv"
)

// Note is one key press. Key is a MIDI number and may be fractional for microtuning.
type Note struct {
	At  float64 // seconds from the start of the Song
	Dur float64 // seconds the key is held; the patch's release rings past it
	Key float64
	Vel float64 // 0..1; 0 means 1

	// Semitones glided linearly over SlideTime: a portamento, or a guitarist's bend.
	Slide     float64
	SlideTime float64 // seconds; 0 spreads the slide over Dur
}

func (n Note) slideTime() float64 {
	if n.SlideTime > 0 {
		return n.SlideTime
	}
	return math.Max(n.Dur, 1e-6)
}

func (n Note) velocity() float64 {
	if n.Vel <= 0 {
		return 1
	}
	return math.Min(n.Vel, 1)
}

func Hz(key float64) float64 { return 440 * math.Pow(2, (key-69)/12) }

var letterSteps = map[byte]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}

// ParseKey reads scientific pitch: "C4" is 60, "Bb2" is 46, "F#-1" is 6.
func ParseKey(name string) (float64, error) {
	if len(name) < 2 {
		return 0, fmt.Errorf("synth: bad note %q", name)
	}
	step, ok := letterSteps[name[0]]
	if !ok {
		return 0, fmt.Errorf("synth: bad note %q", name)
	}
	i := 1
	for i < len(name) && (name[i] == '#' || name[i] == 'b') {
		if name[i] == '#' {
			step++
		} else {
			step--
		}
		i++
	}
	octave, err := strconv.Atoi(name[i:])
	if err != nil {
		return 0, fmt.Errorf("synth: bad note %q", name)
	}
	return float64((octave+1)*12 + step), nil
}

// MustKey panics on a typo so a score fails when it's built, not as a missing note.
func MustKey(name string) float64 {
	k, err := ParseKey(name)
	if err != nil {
		panic(err)
	}
	return k
}

var keyNames = [12]string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}

func KeyName(key int) string {
	return fmt.Sprintf("%s%d", keyNames[((key%12)+12)%12], int(math.Floor(float64(key)/12))-1)
}
