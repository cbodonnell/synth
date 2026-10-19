// Package session is the studio without its window: what a song looks like laid out
// on a timeline, where the playhead is in bars, renders of what's heard, and the
// supervisor that rebuilds the studio when its source changes.
package session

import (
	"fmt"
	"math"

	"github.com/cbodonnell/synth"
)

// StepsPerBeat is how finely the transport counts a beat: bar.beat.step.
const StepsPerBeat = 4

// Timed reports whether a song keeps time in bars, rather than being a sound effect
// measured in seconds.
func Timed(s synth.Song) bool { return s.Meter.Beat > 0 }

// Format writes song time t as the transport shows it: bar.beat.step counted from 1, or
// seconds for a sound.
func Format(s synth.Song, t float64) string {
	if !Timed(s) {
		return fmt.Sprintf("%.3fs", t)
	}
	p := s.Meter.At(t)
	return fmt.Sprintf("%d.%d.%d", p.Bar+1, p.Beat+1, int(p.Phase*StepsPerBeat)+1)
}

// BarTime is where bar b, counted from 0, begins in seconds.
func BarTime(m synth.Meter, b int) float64 { return m.BarStart(b) * m.Beat }

// Bars is how many bars begin before the song's end.
func Bars(s synth.Song) int {
	if !Timed(s) {
		return 0
	}
	n := 0
	for BarTime(s.Meter, n) < s.Length-1e-9 {
		n++
	}
	return n
}

// BarAt is the bar, from 0, that holds song time t.
func BarAt(m synth.Meter, t float64) int {
	if m.Beat <= 0 {
		return 0
	}
	return m.At(t).Bar
}

// Seam is where to drop the playhead to hear the loop wrap: two bars before its end.
func Seam(s synth.Song) float64 {
	if !s.Loop {
		return 0
	}
	if !Timed(s) {
		return math.Max(s.LoopStart, s.Length-2)
	}
	end := Bars(s)
	return math.Max(s.LoopStart, BarTime(s.Meter, max(end-2, 0)))
}
